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

package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/http/pprof"
	"net/url"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/streaming"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	restclientwatch "k8s.io/client-go/rest/watch"
	"k8s.io/klog/v2"
)

func (s *interceptorServer) logger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		klog.Infof("method %s, request: %s", r.Method, r.URL.String())
		handler.ServeHTTP(w, r)
	})
}

func (s *interceptorServer) debugger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/debug/pprof") {
			handler.ServeHTTP(w, r)
			return
		}
		pprof.Index(w, r)
	})
}

func (s *interceptorServer) interceptEventRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/events") {
			handler.ServeHTTP(w, r)
			return
		}

		targetURL, _ := url.Parse(s.restConfig.Host)
		reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
		reverseProxy.Transport, _ = rest.TransportFor(s.restConfig)
		reverseProxy.ServeHTTP(w, r)
	})
}

func (s *interceptorServer) interceptNodeRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, "/api/v1/nodes") {
			handler.ServeHTTP(w, r)
			return
		}

		pathParts := strings.Split(r.URL.Path, "/")
		hostName := pathParts[len(pathParts)-1]

		node := s.cache.GetNode(hostName)
		if node == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		acceptType := r.Header.Get("Accept")
		info, found := s.parseAccept(acceptType, s.mediaSerializer)
		if !found {
			klog.Errorf("can't find %s serializer", acceptType)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		encoder := scheme.Codecs.EncoderForVersion(info.Serializer, v1.SchemeGroupVersion)
		w.Header().Set("Content-Type", info.MediaType)
		err := encoder.Encode(node, w)
		if err != nil {
			klog.Errorf("can't marshal node, %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

func (s *interceptorServer) interceptServiceRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, "/api/v1/services") {
			handler.ServeHTTP(w, r)
			return
		}

		queries := r.URL.Query()
		acceptType := r.Header.Get("Accept")
		info, found := s.parseAccept(acceptType, s.mediaSerializer)
		if !found {
			klog.Errorf("can't find %s serializer", acceptType)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		encoder := scheme.Codecs.EncoderForVersion(info.Serializer, v1.SchemeGroupVersion)
		// list request
		if queries.Get("watch") == "" {
			w.Header().Set("Content-Type", info.MediaType)

			allServices := s.cache.GetServices()
			svcItems := make([]v1.Service, 0, len(allServices))
			for _, svc := range allServices {
				svcItems = append(svcItems, *svc)
			}

			svcList := &v1.ServiceList{
				Items: svcItems,
			}

			err := encoder.Encode(svcList, w)
			if err != nil {
				klog.Errorf("can't marshal endpoints list, %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			return
		}

		// watch request
		timeoutSecondsStr := r.URL.Query().Get("timeoutSeconds")
		timeout := time.Minute
		if timeoutSecondsStr != "" {
			timeout, _ = time.ParseDuration(fmt.Sprintf("%ss", timeoutSecondsStr))
		}

		timer := time.NewTimer(timeout)
		defer timer.Stop()

		flusher, ok := w.(http.Flusher)
		if !ok {
			klog.Errorf("unable to start watch - can't get http.Flusher: %#v", w)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		e := restclientwatch.NewEncoder(
			streaming.NewEncoder(info.StreamSerializer.Framer.NewFrameWriter(w),
				scheme.Codecs.EncoderForVersion(info.StreamSerializer, v1.SchemeGroupVersion)),
			encoder)
		if info.MediaType == runtime.ContentTypeProtobuf {
			w.Header().Set("Content-Type", runtime.ContentTypeProtobuf+";stream=watch")
		} else {
			w.Header().Set("Content-Type", runtime.ContentTypeJSON)
		}
		w.Header().Set("Transfer-Encoding", "chunked")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-timer.C:
				return
			case evt := <-s.serviceWatchCh:
				klog.V(4).Infof("Send service watch event: %+#v", evt)
				err := e.Encode(&evt)
				if err != nil {
					klog.Errorf("can't encode watch event, %v", err)
					return
				}

				if len(s.serviceWatchCh) == 0 {
					flusher.Flush()
				}
			}
		}
	})
}

func (s *interceptorServer) interceptEndpointsRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, "/api/v1/endpoints") {
			handler.ServeHTTP(w, r)
			return
		}

		queries := r.URL.Query()
		acceptType := r.Header.Get("Accept")
		info, found := s.parseAccept(acceptType, s.mediaSerializer)
		if !found {
			klog.Errorf("can't find %s serializer", acceptType)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		encoder := scheme.Codecs.EncoderForVersion(info.Serializer, v1.SchemeGroupVersion)
		// list request
		if queries.Get("watch") == "" {
			w.Header().Set("Content-Type", info.MediaType)
			allEndpoints := s.cache.GetEndpoints()
			epsItems := make([]v1.Endpoints, 0, len(allEndpoints))
			for _, eps := range allEndpoints {
				epsItems = append(epsItems, *eps)
			}

			epsList := &v1.EndpointsList{
				Items: epsItems,
			}

			err := encoder.Encode(epsList, w)
			if err != nil {
				klog.Errorf("can't marshal endpoints list, %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			return
		}

		// watch request
		timeoutSecondsStr := r.URL.Query().Get("timeoutSeconds")
		timeout := time.Minute
		if timeoutSecondsStr != "" {
			timeout, _ = time.ParseDuration(fmt.Sprintf("%ss", timeoutSecondsStr))
		}

		timer := time.NewTimer(timeout)
		defer timer.Stop()

		flusher, ok := w.(http.Flusher)
		if !ok {
			klog.Errorf("unable to start watch - can't get http.Flusher: %#v", w)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		e := restclientwatch.NewEncoder(
			streaming.NewEncoder(info.StreamSerializer.Framer.NewFrameWriter(w),
				scheme.Codecs.EncoderForVersion(info.StreamSerializer, v1.SchemeGroupVersion)),
			encoder)
		if info.MediaType == runtime.ContentTypeProtobuf {
			w.Header().Set("Content-Type", runtime.ContentTypeProtobuf+";stream=watch")
		} else {
			w.Header().Set("Content-Type", runtime.ContentTypeJSON)
		}
		w.Header().Set("Transfer-Encoding", "chunked")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()
		for {
			select {
			case <-r.Context().Done():
				return
			case <-timer.C:
				return
			case evt := <-s.endpointsWatchCh:
				klog.V(4).Infof("Send endpoint watch event: %+#v", evt)
				err := e.Encode(&evt)
				if err != nil {
					klog.Errorf("can't encode watch event, %v", err)
					return
				}

				if len(s.endpointsWatchCh) == 0 {
					flusher.Flush()
				}
			}
		}
	})
}
