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
	"fmt"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"syscall"
	"time"

	"k8s.io/klog"

	"superedge/pkg/lite-apiserver/cert"
	"superedge/pkg/lite-apiserver/storage"
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

	userAgent   string
	method      string
	commonName  string
	urlString   string
	backendUrl  string
	backendPort int
	timeout     int

	storage     storage.Storage
	transport   *EdgeTransport
	certManager *cert.CertManager
	cacher      *RequestCacheController

	watch      bool
	list       bool
	needCache  bool
	selfUpdate bool
}

func NewEdgeReverseProxy(r *http.Request, manager *cert.CertManager, backendUrl string, backendPort int, timeout int, s storage.Storage, cacher *RequestCacheController) *EdgeReverseProxy {
	isList, isWatch, needCache := getRequestProperties(r)
	p := &EdgeReverseProxy{
		certManager: manager,
		backendPort: backendPort,
		backendUrl:  backendUrl,
		timeout:     timeout,
		method:      r.Method,
		urlString:   r.URL.String(),
		storage:     s,
		cacher:      cacher,

		watch:     isWatch,
		list:      isList,
		needCache: needCache,
	}

	h, exist := r.Header[UserAgent]
	if !exist {
		p.userAgent = DefaultUserAgent
	} else {
		p.userAgent = strings.Split(h[0], " ")[0]
	}

	if r.TLS != nil {
		for _, cert := range r.TLS.PeerCertificates {
			if !cert.IsCA {
				p.commonName = cert.Subject.CommonName
				break
			}
		}
	}
	p.transport = p.newTransport()

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

	val := r.Header.Get(EdgeUpdateHeader)
	if len(val) != 0 {
		p.selfUpdate = true
		klog.V(4).Infof("Receive self update request, url->%s, time %s", r.URL.String(), val)
		r.Header.Del(EdgeUpdateHeader)
	}

	if (p.list || p.watch) && !p.selfUpdate {
		p.cacher.AddRequest(r, p.userAgent, p.list, p.watch)
	}

	klog.V(2).Infof("Create new reverse proxy, userAgent->%s, method->%s, url->%s", p.userAgent, p.method, p.urlString)
	return p
}

func (p *EdgeReverseProxy) makeDirector(req *http.Request) {
	klog.V(6).Infof("Director request is %+v", req)

	req.URL.Scheme = "https"
	req.URL.Host = fmt.Sprintf("%s:%d", p.backendUrl, p.backendPort)

	klog.V(6).Infof("Make new request %+v", req)
}

func (p *EdgeReverseProxy) newTransport() *EdgeTransport {
	klog.V(4).Infof("Receive request from %s", p.commonName)

	if len(p.commonName) == 0 {
		return &EdgeTransport{p.certManager.DefaultTransport()}
	}

	tr := p.certManager.Load(p.commonName)
	if tr == nil {
		klog.Warningf("Cannot load transport for %s. Use DefaultTransport", p.commonName)
		return &EdgeTransport{p.certManager.DefaultTransport()}
	}

	return &EdgeTransport{tr}
}

func (p *EdgeReverseProxy) modifyResponse(res *http.Response) error {
	if !p.needCache {
		return nil
	}

	// cache response data
	p.writeCache(NewEdgeResponseDataHolder(res))
	return nil
}

func (p *EdgeReverseProxy) handlerError(rw http.ResponseWriter, req *http.Request, err error) {
	klog.Warningf("Request url %s, %s error %v", req.URL.Host, req.URL, err)
	defer func() {
		if p.watch {
			p.cacher.DeleteRequest(req, p.userAgent)
		}
	}()

	// filter error, if not ECONNREFUSED or ETIMEDOUT, not read cache and ignore
	if p.filterErrorToIgnore(err) {
		klog.V(4).Infof("Receive not syscall error %v", err)
		rw.WriteHeader(http.StatusBadGateway)
		_, err := rw.Write([]byte(err.Error()))
		if err != nil {
			klog.Errorf("Write error response err: %v", err)
		}
		return
	}

	klog.V(4).Infof("Request error, need read data from cache")

	// read cache when request error
	data, cacheErr := p.readCache()
	if cacheErr != nil {
		klog.Errorf("Read cache error %v, write though error", cacheErr)
		rw.WriteHeader(http.StatusBadGateway)
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

func (p *EdgeReverseProxy) filterErrorToIgnore(err error) bool {
	// ignore watch error
	if p.watch {
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

func (p *EdgeReverseProxy) key() string {
	return strings.ReplaceAll(fmt.Sprintf("%s_%s", p.userAgent, p.urlString), "/", "_")
}

func (p *EdgeReverseProxy) readCache() (*EdgeResponseDataHolder, error) {
	data, err := p.storage.Load(p.key())
	if err != nil {
		return nil, err
	}

	res := &EdgeResponseDataHolder{}
	err = res.Input(data)
	if err != nil {
		klog.Errorf("Read cache unmarshal %s error: %v", p.key(), err)
		return nil, err
	}
	return res, nil
}

func (p *EdgeReverseProxy) writeCache(r *EdgeResponseDataHolder) {
	bodyBytes, err := r.Output()
	if err != nil {
		klog.Errorf("Write cache marshal %s error: %v", p.key(), err)
	}
	err = p.storage.Store(p.key(), bodyBytes)
	if err != nil {
		klog.Errorf("Write cache %s error: %v", p.key(), err)
	}
}

func getRequestProperties(r *http.Request) (isList bool, isWatch bool, needCache bool) {
	info, ok := apirequest.RequestInfoFrom(r.Context())
	if ok {
		isList = info.Verb == VerbList
		isWatch = info.Verb == VerbWatch

		klog.V(4).Infof("request resourceInfo=%+v", info)

		// only cache resource request
		if info.IsResourceRequest {
			needCache = true
			if isWatch || r.Method != http.MethodGet {
				needCache = false
			} else if info.Subresource == "log" {
				// do not cache logs
				needCache = false
			}
		}
	}

	return isList, isWatch, needCache
}