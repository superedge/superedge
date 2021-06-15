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
	"net/http"
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/endpoints/filters"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/klog/v2"

	"github.com/superedge/superedge/pkg/lite-apiserver/cache"
	"github.com/superedge/superedge/pkg/lite-apiserver/config"
	"github.com/superedge/superedge/pkg/lite-apiserver/transport"
)

// EdgeServerHandler is the real handler for each request
type EdgeServerHandler struct {

	// apiserverInfo is used to proxy to.
	apiserverUrl  string
	apiserverPort int

	// transportManager is to transportManager all cert declared in config, and gen correct transport
	transportManager *transport.TransportManager

	transportChannel <-chan string

	// the default proxy
	defaultProxy *EdgeReverseProxy

	proxyMapLock sync.RWMutex
	// proxy with client tls cert
	reverseProxyMap map[string]*EdgeReverseProxy

	// cacheManager
	cacheManager *cache.CacheManager
}

func NewEdgeServerHandler(config *config.LiteServerConfig, transportManager *transport.TransportManager,
	cacheManager *cache.CacheManager, transportChannel <-chan string) (http.Handler, error) {
	h := &EdgeServerHandler{
		apiserverUrl:     config.KubeApiserverUrl,
		apiserverPort:    config.KubeApiserverPort,
		transportManager: transportManager,
		transportChannel: transportChannel,
		reverseProxyMap:  make(map[string]*EdgeReverseProxy),
		cacheManager:     cacheManager,
	}

	// init proxy
	h.initProxies()

	// start to handle new proxy
	h.start()

	return h.buildHandlerChain(config, h), nil
}

func (h *EdgeServerHandler) initProxies() {
	klog.Infof("init default proxy")
	h.defaultProxy = NewEdgeReverseProxy(h.transportManager.GetTransport(""), h.apiserverUrl, h.apiserverPort, h.cacheManager)

	h.proxyMapLock.Lock()
	defer h.proxyMapLock.Unlock()
	for commonName, t := range h.transportManager.GetTransportMap() {
		klog.Infof("init proxy for %s", commonName)
		proxy := NewEdgeReverseProxy(t, h.apiserverUrl, h.apiserverPort, h.cacheManager)
		h.reverseProxyMap[commonName] = proxy
	}

}

func (h *EdgeServerHandler) start() {
	go func() {
		for {
			select {
			case commonName := <-h.transportChannel:
				// receive new transport to create EdgeReverseProxy
				klog.Infof("receive new transport %s", commonName)
				t := h.transportManager.GetTransport(commonName)

				klog.Infof("add new proxy for %s", commonName)
				proxy := NewEdgeReverseProxy(t, h.apiserverUrl, h.apiserverPort, h.cacheManager)

				h.proxyMapLock.Lock()
				h.reverseProxyMap[commonName] = proxy
				h.proxyMapLock.Unlock()
			}
		}
	}()
}

func (h *EdgeServerHandler) buildHandlerChain(config *config.LiteServerConfig, handler http.Handler) http.Handler {
	if config.ModifyRequestAccept {
		handler = WithRequestAccept(handler)
	}

	cfg := &server.Config{
		LegacyAPIGroupPrefixes: sets.NewString(server.DefaultLegacyAPIPrefix),
	}
	resolver := server.NewRequestInfoResolver(cfg)
	handler = filters.WithRequestInfo(handler, resolver)
	return handler
}

func (h *EdgeServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	commonName := ""
	if r.TLS != nil {
		for _, cert := range r.TLS.PeerCertificates {
			if !cert.IsCA {
				commonName = cert.Subject.CommonName
				break
			}
		}
	}

	reverseProxy := h.getEdgeReverseProxy(commonName)
	reverseProxy.ServeHTTP(w, r)
}

func (h *EdgeServerHandler) getEdgeReverseProxy(commonName string) *EdgeReverseProxy {
	if len(commonName) == 0 {
		klog.V(6).Info("commonName is empty, use default proxy")
		return h.defaultProxy
	}

	h.proxyMapLock.RLock()
	defer h.proxyMapLock.RUnlock()
	proxy, ok := h.reverseProxyMap[commonName]
	if ok {
		klog.V(6).Infof("got proxy for commonName %s", commonName)
		return proxy
	}

	klog.V(2).Infof("couldn't get proxy for %s, use default proxy", commonName)
	return h.defaultProxy
}
