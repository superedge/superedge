/*
Copyright 2020 The SuperEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package proxy

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"syscall"

	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/klog/v2"

	"github.com/superedge/superedge/pkg/lite-apiserver/cache"
	"github.com/superedge/superedge/pkg/lite-apiserver/constant"
	"github.com/superedge/superedge/pkg/lite-apiserver/transport"
)

// EdgeReverseProxy represents a real pair of http request and response
type EdgeReverseProxy struct {
	backendProxy *httputil.ReverseProxy

	backendUrl  string
	backendPort int

	transport    *transport.EdgeTransport
	cacheManager *cache.CacheManager
}

func NewEdgeReverseProxy(transport *transport.EdgeTransport, backendUrl string, backendPort int, cacheManager *cache.CacheManager) *EdgeReverseProxy {
	p := &EdgeReverseProxy{
		backendPort:  backendPort,
		backendUrl:   backendUrl,
		transport:    transport,
		cacheManager: cacheManager,
	}

	reverseProxy := &httputil.ReverseProxy{
		Director:       p.makeDirector,
		Transport:      p.transport,
		ModifyResponse: p.modifyResponse,
		ErrorHandler:   p.handlerError,
	}

	reverseProxy.FlushInterval = -1

	p.backendProxy = reverseProxy
	return p
}

func (p *EdgeReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	klog.V(2).Infof("New request: method->%s, url->%s", r.Method, r.URL.String())

	// handle http
	p.backendProxy.ServeHTTP(w, r)
}

func (p *EdgeReverseProxy) makeDirector(req *http.Request) {
	req.URL.Scheme = "https"
	req.URL.Host = fmt.Sprintf("%s:%d", p.backendUrl, p.backendPort)
}

func (p *EdgeReverseProxy) modifyResponse(resp *http.Response) error {
	if resp == nil || resp.Request == nil {
		klog.Infof("no response or request, skip cache response")
		return nil
	}

	isNeedCache := needCache(resp.Request)
	if !isNeedCache {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		klog.V(4).Infof("resp status is %d, skip cache response", resp.StatusCode)
		return nil
	}

	// validate watch Content-Type
	info, ok := apirequest.RequestInfoFrom(resp.Request.Context())
	if !ok {
		return nil
	}
	if info.Verb == constant.VerbWatch {
		contentType := resp.Header.Get(constant.ContentType)
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			klog.Warningf("Content-Type %s is not recognized: %v", contentType, err)
			return nil
		}
		klog.V(8).Infof("mediaType is %s", mediaType)
		if mediaType != constant.Json && mediaType != constant.Yaml {
			return nil
		}
	}

	// cache response data
	multiRead := MultiWrite(resp.Body, 2)
	if multiRead == nil {
		return fmt.Errorf("The number of Reads specified by MultiWrite is less than 2")
	}
	go func(req *http.Request, header http.Header, statusCode int, pipeReader io.ReadCloser) {
		err := p.writeCache(req, header, statusCode, pipeReader)
		if (err != nil) && (err != io.EOF) && (err != context.Canceled) {
			klog.Errorf("Write cache error: %v", err)
		}
	}(resp.Request, resp.Header.Clone(), resp.StatusCode, multiRead[1])

	resp.Body = multiRead[0]

	return nil
}

func (p *EdgeReverseProxy) handlerError(rw http.ResponseWriter, req *http.Request, err error) {
	klog.V(2).Infof("Request url=%s, error=%v", req.URL, err)

	// filter error. if true, not read cache and ignore
	if p.ignoreCache(req, err) {
		klog.V(6).Infof("Ignore request %s", req.URL)
		rw.WriteHeader(http.StatusServiceUnavailable)
		_, err := rw.Write([]byte(err.Error()))
		if err != nil {
			klog.Errorf("Write error response err: %v", err)
		}
		return
	}

	klog.V(4).Infof("Request error, need read data from cache")

	// read cache when request error
	data, cacheErr := p.readCache(req)
	if cacheErr != nil {
		klog.Errorf("Read cache error %v, write though error", cacheErr)
		rw.WriteHeader(http.StatusNotFound)
		_, err := rw.Write([]byte(err.Error()))
		if err != nil {
			klog.Errorf("Write read cache error: %v", err)
		}
		return
	}

	CopyHeader(rw.Header(), data.Header)
	rw.WriteHeader(data.StatusCode)
	_, err = rw.Write(data.Body)
	if err != nil {
		klog.Errorf("Write cache response for %s err: %v", req.URL, err)
	}
}

func (p *EdgeReverseProxy) ignoreCache(r *http.Request, err error) bool {
	// ignore those requests that do not need cache
	if !needCache(r) {
		return true
	}

	if (err == context.Canceled) || (err == context.DeadlineExceeded) {
		return false
	}

	netErr, ok := err.(net.Error)
	if !ok {
		klog.V(4).Infof("Request error is not net err: %+v", err)
		return true
	}
	if netErr.Timeout() {
		return false
	}

	opError, ok := netErr.(*net.OpError)
	if !ok {
		klog.V(4).Infof("Request error is not netop err: %+v", err)
		return true
	}

	switch t := opError.Err.(type) {
	case *os.SyscallError:
		if errno, ok := t.Err.(syscall.Errno); ok {
			klog.V(4).Infof("Request errorno is %+v", errno)
			switch errno {
			case syscall.ECONNREFUSED, syscall.ETIMEDOUT:
				return false
			default:
				return true
			}
		}
	}

	return false
}

func (p *EdgeReverseProxy) readCache(r *http.Request) (*cache.EdgeCache, error) {
	return p.cacheManager.Query(r)
}

func (p *EdgeReverseProxy) writeCache(req *http.Request, header http.Header, statusCode int, body io.ReadCloser) error {
	return p.cacheManager.Cache(req, statusCode, header, body)
}

func needCache(r *http.Request) (needCache bool) {
	info, ok := apirequest.RequestInfoFrom(r.Context())
	if ok {
		klog.V(4).Infof("request resourceInfo=%+v", info)

		// only cache resource request
		if info.IsResourceRequest {
			needCache = true
			if r.Method != http.MethodGet {
				needCache = false
			}

			if info.Subresource == "log" {
				// do not cache logs
				needCache = false
			}
		}
	} else {
		klog.Errorf("no RequestInfo found in the context")
	}
	return needCache
}
