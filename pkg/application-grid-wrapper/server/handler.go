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
	discoveryv1 "k8s.io/api/discovery/v1"
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/streaming"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	restclientwatch "k8s.io/client-go/rest/watch"
	"k8s.io/klog/v2"

	siteconstant "github.com/superedge/superedge/pkg/site-manager/constant"
)

const (
	SuperEdgeIngress = "superedge-ingress"
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
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/events") || strings.HasPrefix(r.URL.Path, "/superedge-ingress") {
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

			if s.k3sServiceInformer != nil {
				for _, svc := range s.k3sServiceInformer.GetStore().List() {
					k3sSvc := svc.(*v1.Service).DeepCopy()
					k3sSvc.Name = fmt.Sprintf("k3s-%s", k3sSvc.Name)
					svcItems = append(svcItems, *k3sSvc)
				}
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

		watchCH := s.serviceWatchBroadcaster.Watch()
		for {
			select {
			case <-r.Context().Done():
				return
			case <-timer.C:
				return
			case evt := <-watchCH.ResultChan():
				klog.V(4).Infof("Send service watch event: %+#v", evt)
				err := e.Encode(&evt)
				if err != nil {
					klog.Errorf("can't encode watch event, %v", err)
					return
				}

				if len(watchCH.ResultChan()) == 0 {
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
		endpointsWatch := s.endpointsBroadcaster.Watch()
		defer endpointsWatch.Stop()
		for {
			select {
			case <-r.Context().Done():
				return
			case <-timer.C:
				return
			case evt := <-endpointsWatch.ResultChan():
				klog.V(4).Infof("Send endpoint watch event: %+#v", evt)
				err := e.Encode(&evt)
				if err != nil {
					klog.Errorf("can't encode watch event, %v", err)
					return
				}

				if len(endpointsWatch.ResultChan()) == 0 {
					flusher.Flush()
				}
			}
		}
	})
}

func (s *interceptorServer) interceptEndpointSliceV1Request(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, "/apis/discovery.k8s.io/v1/endpointslices") {
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

		encoder := scheme.Codecs.EncoderForVersion(info.Serializer, discoveryv1.SchemeGroupVersion)
		// list request
		if queries.Get("watch") == "" {
			w.Header().Set("Content-Type", info.MediaType)
			allEndpointSlices := s.cache.GetEndpointSliceV1()
			epsItems := make([]discoveryv1.EndpointSlice, 0, len(allEndpointSlices))
			for _, eps := range allEndpointSlices {
				epsItems = append(epsItems, *eps)
			}
			//添加k3s endpointslice
			if s.k3sEndpointSliceV1Informer != nil {
				for _, epsV1 := range s.k3sEndpointSliceV1Informer.GetStore().List() {
					k3sEpsV1 := epsV1.(*discoveryv1.EndpointSlice)
					superedgeEpsV1 := k3sEpsV1.DeepCopy()
					superedgeEpsV1.Name = fmt.Sprintf("k3s-%s", superedgeEpsV1.Name)
					superedgeEpsV1.Labels["kubernetes.io/service-name"] = fmt.Sprintf("k3s-%s", superedgeEpsV1.Labels["kubernetes.io/service-name"])
					epsItems = append(epsItems, *superedgeEpsV1)
				}
			}
			epsList := &discoveryv1.EndpointSliceList{
				Items: epsItems,
			}

			err := encoder.Encode(epsList, w)
			if err != nil {
				klog.Errorf("can't marshal endpointSlice list, %v", err)
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
		watch := s.endpointSliceV1WatchBroadcaster.Watch()
		for {
			select {
			case <-r.Context().Done():
				return
			case <-timer.C:
				return
			case evt := <-watch.ResultChan():
				klog.V(4).Infof("Send endpointSlice watch event: %+#v", evt)
				err := e.Encode(&evt)
				if err != nil {
					klog.Errorf("can't encode watch event, %v", err)
					return
				}
				if len(watch.ResultChan()) == 0 {
					flusher.Flush()
				}
			}
		}
	})
}

func (s *interceptorServer) interceptEndpointSliceV1Beta1Request(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, "/apis/discovery.k8s.io/v1beta1/endpointslices") {
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

		encoder := scheme.Codecs.EncoderForVersion(info.Serializer, discoveryv1beta1.SchemeGroupVersion)
		// list request
		if queries.Get("watch") == "" {
			w.Header().Set("Content-Type", info.MediaType)
			allEndpointSlices := s.cache.GetEndpointSliceV1Beta1()
			epsItems := make([]discoveryv1beta1.EndpointSlice, 0, len(allEndpointSlices))
			for _, eps := range allEndpointSlices {
				epsItems = append(epsItems, *eps)
			}
			//添加endpointsliceV1Beta
			if s.k3sEndpointSliceV1Beta1Informer != nil {
				for _, epsV1Beta := range s.k3sEndpointSliceV1Beta1Informer.GetStore().List() {
					k3sEpsV1beta1 := epsV1Beta.(*discoveryv1beta1.EndpointSlice)
					superedgeEpsV1beta1 := k3sEpsV1beta1.DeepCopy()
					superedgeEpsV1beta1.Name = fmt.Sprintf("k3s-%s", superedgeEpsV1beta1.Name)
					superedgeEpsV1beta1.Labels["kubernetes.io/service-name"] = fmt.Sprintf("k3s-%s", superedgeEpsV1beta1.Labels["kubernetes.io/service-name"])
					epsItems = append(epsItems, *superedgeEpsV1beta1)
				}
			}
			epsList := &discoveryv1beta1.EndpointSliceList{
				Items: epsItems,
			}

			err := encoder.Encode(epsList, w)
			if err != nil {
				klog.Errorf("can't marshal endpointSlice list, %v", err)
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
		watch := s.endpointSliceV1Beta1WatchBroadcaster.Watch()
		for {
			select {
			case <-r.Context().Done():
				return
			case <-timer.C:
				return
			case evt := <-watch.ResultChan():
				klog.V(4).Infof("Send endpointSlice watch event: %+#v", evt)
				err := e.Encode(&evt)
				if err != nil {
					klog.Errorf("can't encode watch event, %v", err)
					return
				}
				if len(watch.ResultChan()) == 0 {
					flusher.Flush()
				}
			}
		}
	})
}

func (s *interceptorServer) interceptIngressRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/superedge-ingress") {
			ingress, path, err := getIngressPath(r.URL.Path)
			if err == nil {
				r.URL.Path = path
				r.Header.Add("superedge-ingress", ingress)
			}
		}
		handler.ServeHTTP(w, r)
		return
	})
}
func (s *interceptorServer) interceptIngressEndpointsRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// header SuperEdgeIngress 不为空且 path 是 /api/v1/endpoints 的请求需要处理，否则直接返回即可
		if r.Header.Get(SuperEdgeIngress) == "" || !strings.Contains(r.URL.Path, "/api/v1/endpoints") {
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
		cacheEvts := []watch.Event{}
		for _, v := range s.cache.GetEndpoints() {
			cacheEvts = append(cacheEvts, watch.Event{
				Type:   watch.Added,
				Object: v,
			})
		}
		endpointsWatch := s.endpointsBroadcaster.WatchWithPrefix(cacheEvts)
		defer endpointsWatch.Stop()
		nodeWatch := s.nodeBoradcaster.Watch()
		for {
			select {
			case <-r.Context().Done():
				return
			case <-timer.C:
				return
			case evt := <-endpointsWatch.ResultChan():
				klog.V(4).Infof("Send endpoint watch event: %+#v", evt)

				// Filter endpoints based on nodeunit nodes
				if ep, ok := evt.Object.(*v1.Endpoints); ok {
					deepEp := ep.DeepCopy()
					s.filerIngressEndpoints(deepEp, r.Header.Get(SuperEdgeIngress), siteconstant.NodeUnitSuperedge)
					evt.Object = deepEp
				}

				err := e.Encode(&evt)
				if err != nil {
					klog.Errorf("can't encode watch event, %v", err)
					return
				}

				if len(endpointsWatch.ResultChan()) == 0 {
					flusher.Flush()
				}
			case modifyNode := <-nodeWatch.ResultChan():

				for _, endpoints := range s.cache.GetEndpoints() {
					deepEndpoints := endpoints.DeepCopy()
					resyncflag := false
					for _, subset := range deepEndpoints.Subsets {
						for _, addr := range subset.Addresses {
							if addr.NodeName != nil {
								nodeName := addr.NodeName
								if *nodeName == modifyNode.Object.(*v1.Node).Name {
									resyncflag = true
								}
							}
						}
					}
					if resyncflag {
						s.filerIngressEndpoints(deepEndpoints, r.Header.Get(SuperEdgeIngress), siteconstant.NodeUnitSuperedge)
						err := e.Encode(&watch.Event{
							Type:   watch.Modified,
							Object: deepEndpoints,
						})
						if err != nil {
							klog.Errorf("can't encode watch event, %v", err)
							return
						}
					}
				}
			}
		}
	})
}

func (s *interceptorServer) filerIngressEndpoints(ep *v1.Endpoints, key, value string) {
	// Get the node of the nodeunit where nginx-ingress-controller is located
	unitnodes, err := s.nodeIndexer.ByIndex(NODELABELS_INDEXER, fmt.Sprintf("%s=%s", key, value))
	if err != nil {
		klog.Errorf("Failed to get unit %s nodes, error: %v", fmt.Sprintf("%s=%s", key, value), err)
	} else if len(unitnodes) != 0 {
		filterSubsets := []v1.EndpointSubset{}
		for _, subset := range ep.Subsets {
			filterAddress := []v1.EndpointAddress{}
			for _, addr := range subset.Addresses {
				if addr.NodeName != nil {
					// Filter by node name
					nodeName := addr.NodeName
					addflag := false
					for _, node := range unitnodes {
						if node.(*v1.Node).Name == *nodeName {
							addflag = true
						}
					}
					if addflag {
						filterAddress = append(filterAddress, addr)
					}
				}
			}
			if len(filterAddress) != len(subset.Addresses) {
				subset.Addresses = filterAddress
			}
			filterSubsets = append(filterSubsets, subset)
		}
		ep.Subsets = filterSubsets
	}
}
