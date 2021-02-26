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
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/endpoints/filters"
	"k8s.io/apiserver/pkg/server"
	"net/http"

	"github.com/superedge/superedge/pkg/lite-apiserver/cert"
	"github.com/superedge/superedge/pkg/lite-apiserver/config"
	"github.com/superedge/superedge/pkg/lite-apiserver/storage"
)

// EdgeServerHandler is the real handler for each request
type EdgeServerHandler struct {
	timeout int

	// certManager is to certManager all cert declared in config, and gen correct transport
	certManager *cert.CertManager

	// the default proxy
	defaultProxy *EdgeReverseProxy
	// proxy with client tls cert
	reverseProxyMap map[string]*EdgeReverseProxy

	// apiserverInfo is used to proxy to.
	apiserverUrl  string
	apiserverPort int

	// storage is to store/load history data
	storage storage.Storage

	// cacher hold request pair (list and watch for same path)
	// and update list data periodically
	cacher *RequestCacheController
}

func NewEdgeServerHandler(config *config.LiteServerConfig, certManager *cert.CertManager, cacher *RequestCacheController) (http.Handler, error) {
	h := &EdgeServerHandler{
		apiserverUrl:    config.KubeApiserverUrl,
		apiserverPort:   config.KubeApiserverPort,
		certManager:     certManager,
		reverseProxyMap: make(map[string]*EdgeReverseProxy),
		storage:         storage.NewFileStorage(config),
		cacher:          cacher,
		timeout:         config.BackendTimeout,
	}

	h.initProxies()

	return h.buildHandlerChain(h), nil
}

func (h *EdgeServerHandler) initProxies() {
	h.defaultProxy = NewEdgeReverseProxy(h.certManager.DefaultTransport(), h.apiserverUrl, h.apiserverPort, h.timeout, h.storage, h.cacher)

	for commonName, transport := range h.certManager.GetTransportMap() {
		proxy := NewEdgeReverseProxy(transport, h.apiserverUrl, h.apiserverPort, h.timeout, h.storage, h.cacher)
		h.reverseProxyMap[commonName] = proxy
	}

}

func (h *EdgeServerHandler) buildHandlerChain(handler http.Handler) http.Handler {
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
		return h.defaultProxy
	}

	proxy, e := h.reverseProxyMap[commonName]
	if !e {
		return h.defaultProxy
	}
	return proxy
}
