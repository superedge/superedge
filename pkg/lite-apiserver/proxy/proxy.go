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
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"syscall"
	"time"

	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/klog"

	"github.com/superedge/superedge/pkg/lite-apiserver/cert"
	"github.com/superedge/superedge/pkg/lite-apiserver/storage"
)

const EdgeUpdateHeader = "Edge-Update-Request"

const (
	UserAgent        = "User-Agent"
	DefaultUserAgent = "default"
)

const (
	VerbList  = "list"
	VerbWatch = "watch"
)

// EdgeReverseProxy represents a real pair of http request and response
type EdgeReverseProxy struct {
	backendProxy *httputil.ReverseProxy

	backendUrl  string
	backendPort int
	timeout     int

	storage     storage.Storage
	transport   *EdgeTransport
	certManager *cert.CertManager
	cacher      *RequestCacheController
}

func NewEdgeReverseProxy(transport *http.Transport, backendUrl string, backendPort int, timeout int, s storage.Storage, cacher *RequestCacheController) *EdgeReverseProxy {
	p := &EdgeReverseProxy{
		backendPort: backendPort,
		backendUrl:  backendUrl,
		timeout:     timeout,
		storage:     s,
		cacher:      cacher,
	}

	p.transport = p.newTransport(transport)

	// set timeout for request, if overtime, we think request failed, and read cache
	if p.timeout > 0 {
		p.transport.tr.DialContext = (&net.Dialer{
			Timeout: time.Duration(p.timeout) * time.Second,
		}).DialContext
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
	isList, isWatch := getRequestVerb(r)

	userAgent := getUserAgent(r)

	selfUpdate := false
	val := r.Header.Get(EdgeUpdateHeader)
	if len(val) != 0 {
		selfUpdate = true
		klog.V(4).Infof("Receive self update request, url->%s, time %s", r.URL.String(), val)
		r.Header.Del(EdgeUpdateHeader)
	}

	if (isList || isWatch) && !selfUpdate {
		p.cacher.AddRequest(r, userAgent, isList, isWatch)
	}

	klog.V(2).Infof("New request: userAgent->%s, method->%s, url->%s", userAgent, r.Method, r.URL.String())

	// handle http
	p.backendProxy.ServeHTTP(w, r)
}

func (p *EdgeReverseProxy) makeDirector(req *http.Request) {
	req.URL.Scheme = "https"
	req.URL.Host = fmt.Sprintf("%s:%d", p.backendUrl, p.backendPort)
}

func (p *EdgeReverseProxy) newTransport(tr *http.Transport) *EdgeTransport {
	return &EdgeTransport{tr}
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

	// cache response data
	dupReader, pipeReader := NewDupReadCloser(resp.Body)

	go func(req *http.Request, header http.Header, statusCode int, pipeReader io.ReadCloser) {
		err := p.writeCache(req, header, statusCode, pipeReader)
		if err != nil && err != io.EOF {
			klog.Errorf("Write cache error: %v", err)
		}
	}(resp.Request, resp.Header, resp.StatusCode, pipeReader)

	resp.Body = dupReader

	return nil
}

func (p *EdgeReverseProxy) handlerError(rw http.ResponseWriter, req *http.Request, err error) {
	klog.Warningf("Request url %s, %s error %v", req.URL.Host, req.URL, err)

	_, isWatch := getRequestVerb(req)
	cache := needCache(req)
	userAgent := getUserAgent(req)

	defer func() {
		if isWatch {
			p.cacher.DeleteRequest(req, userAgent)
		}
	}()

	// filter error, if not ECONNREFUSED or ETIMEDOUT, not read cache and ignore
	if p.filterErrorToIgnore(cache, err) {
		klog.V(4).Infof("Receive not syscall error %v", err)
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

	for k, v := range data.Header {
		for i := range v {
			rw.Header().Set(k, v[i])
		}
	}

	CopyHeader(rw.Header(), data.Header)
	rw.WriteHeader(data.Code)
	_, err = rw.Write(data.Body)
	if err != nil {
		klog.Errorf("Write cache response err: %v", err)
	}
}

func (p *EdgeReverseProxy) filterErrorToIgnore(needCache bool, err error) bool {
	// ignore those requests that do not need cache
	if !needCache {
		return true
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
			switch errno {
			case syscall.ECONNREFUSED, syscall.ETIMEDOUT:
				return false
			default:
				return true
			}
		}
	}

	return true
}

func (p *EdgeReverseProxy) key(r *http.Request) string {
	userAgent := getUserAgent(r)
	uri := r.URL.RequestURI()
	return strings.ReplaceAll(fmt.Sprintf("%s_%s", userAgent, uri), "/", "_")
}

func (p *EdgeReverseProxy) readCache(r *http.Request) (*EdgeResponseDataHolder, error) {
	key := p.key(r)
	data, err := p.storage.Load(key)
	if err != nil {
		return nil, err
	}

	res := &EdgeResponseDataHolder{}
	err = res.Input(data)
	if err != nil {
		klog.Errorf("Read cache unmarshal %s error: %v", key, err)
		return nil, err
	}
	return res, nil
}

func (p *EdgeReverseProxy) writeCache(req *http.Request, header http.Header, statusCode int, pipeReader io.ReadCloser) error {
	var buf bytes.Buffer
	n, err := buf.ReadFrom(pipeReader)
	if err != nil {
		klog.Errorf("Failed to get cache response, %v", err)
		return err
	}

	key := p.key(req)
	klog.V(4).Infof("Cache %d bytes from response for %s", n, key)

	if n == 0 {
		return nil
	}

	holder := &EdgeResponseDataHolder{
		Code:   statusCode,
		Body:   buf.Bytes(),
		Header: header,
	}
	bodyBytes, err := holder.Output()
	if err != nil {
		klog.Errorf("Write cache marshal %s error: %v", key, err)
		return err
	}
	err = p.storage.Store(key, bodyBytes)
	if err != nil {
		klog.Errorf("Write cache %s error: %v", key, err)
		return err
	}

	return nil
}

func getRequestVerb(r *http.Request) (isList bool, isWatch bool) {
	info, ok := apirequest.RequestInfoFrom(r.Context())
	if ok {
		klog.V(4).Infof("request resourceInfo=%+v", info)

		isList = info.Verb == VerbList
		isWatch = info.Verb == VerbWatch
	} else {
		klog.Errorf("parse requestInfo error")
	}

	return isList, isWatch
}

func needCache(r *http.Request) (needCache bool) {
	info, ok := apirequest.RequestInfoFrom(r.Context())
	if ok {
		klog.V(4).Infof("request resourceInfo=%+v", info)

		// only cache resource request
		if info.IsResourceRequest {
			needCache = true
			if info.Verb == VerbWatch || r.Method != http.MethodGet {
				needCache = false
			} else if info.Subresource == "log" {
				// do not cache logs
				needCache = false
			}
		}
	} else {
		klog.Errorf("parse requestInfo error")
	}
	return needCache
}

func getUserAgent(r *http.Request) string {
	userAgent := DefaultUserAgent
	h, exist := r.Header[UserAgent]
	if exist {
		userAgent = strings.Split(h[0], " ")[0]
	}
	return userAgent
}
