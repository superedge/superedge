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

	"superedge/pkg/lite-apiserver/cert"
	"superedge/pkg/lite-apiserver/config"
	"superedge/pkg/lite-apiserver/storage"
)

// EdgeServerHandler is the real handler for each request
type EdgeServerHandler struct {
	timeout int
	// manager is to manager all cert declared in config, and gen correct transport
	manager *cert.CertManager

	// apiserverInfo is used to proxy to.
	apiserverUrl  string
	apiserverPort int

	// storage is to store/load history data
	storage storage.Storage

	// cacher hold request pair (list and watch for same path)
	// and update list data periodically
	cacher *RequestCacheController
}

func NewEdgeServerHandler(config *config.LiteServerConfig, manager *cert.CertManager, cacher *RequestCacheController) (http.Handler, error) {
	e := &EdgeServerHandler{
		apiserverUrl:  config.KubeApiserverUrl,
		apiserverPort: config.KubeApiserverPort,
		manager:       manager,
		storage:       storage.NewFileStorage(config),
		cacher:        cacher,
		timeout:       config.BackendTimeout,
	}
	return e.buildHandlerChain(e), nil
}

func (h *EdgeServerHandler) buildHandlerChain(edgeHandler http.Handler) http.Handler {
	cfg := &server.Config{
		LegacyAPIGroupPrefixes: sets.NewString(server.DefaultLegacyAPIPrefix),
	}
	resolver := server.NewRequestInfoResolver(cfg)
	handler := filters.WithRequestInfo(edgeHandler, resolver)
	return handler
}

func (h *EdgeServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reverseProxy := NewEdgeReverseProxy(r, h.manager, h.apiserverUrl, h.apiserverPort, h.timeout, h.storage, h.cacher)
	reverseProxy.backendProxy.ServeHTTP(w, r)
}
