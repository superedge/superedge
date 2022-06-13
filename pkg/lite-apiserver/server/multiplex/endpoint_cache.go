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

package multiplex

import (
	"net/http"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	listerv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	"k8s.io/klog/v2"
)

const (
	// EndpointCacheURL is cached url
	EndpointCacheURL = "/api/v1/endpoints"
)

func init() {
	RegisterMuxFactory(EndpointCacheURL, &endpointCacheMuxFactory{})
}

type endpointCacheMuxFactory struct{}

var _ CacheMuxFactory = &endpointCacheMuxFactory{}

func (f *endpointCacheMuxFactory) Create(hostname string, informerFactory informers.SharedInformerFactory) (CacheMux, error) {
	return NewEndpointMux(hostname, informerFactory)
}

type endpointMux struct {
	hostname    string
	broadcaster *watch.Broadcaster
	epInformer  cache.SharedIndexInformer
	epLister    listerv1.EndpointsLister
}

var _ CacheMux = &endpointMux{}

// NewEndpointMux ...
func NewEndpointMux(hostname string, informerFactory informers.SharedInformerFactory) (CacheMux, error) {
	ba := watch.NewLongQueueBroadcaster(maxQueuedEvents, watch.DropIfChannelFull)
	informer := informerFactory.Core().V1().Endpoints()
	epMux := &endpointMux{
		hostname:    hostname,
		broadcaster: ba,
		epInformer:  informer.Informer(),
		epLister:    informer.Lister(),
	}
	// don't resync to downstream watchers
	informer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    epMux.OnAdd,
		UpdateFunc: epMux.OnUpdate,
		DeleteFunc: epMux.OnDelete,
	}, 0)

	// run with informerFactory

	// stop := make(chan struct{})
	// go informer.Informer().Run(stop)
	// if !cache.WaitForNamedCacheSync("endpointMux", stop, informer.Informer().HasSynced) {
	// 	return nil, fmt.Errorf("can't sync informers")
	// }

	return epMux, nil
}

func (em *endpointMux) Name() string {
	return EndpointCacheURL
}

func (em *endpointMux) Match(method, URLPath string) bool {
	if method == http.MethodGet && strings.HasPrefix(URLPath, EndpointCacheURL) {
		return true
	}
	return false
}

// watch will ignore resourceVersion and return all
func (em *endpointMux) Watch(bookmark bool, ResourceVersion string) (watch.Interface, error) {
	// list all obj in all namespaces
	eps, err := em.epLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("List endpoints error by endpoint lister, will drop first, err %v", err)
		return nil, err
	}
	// build ADD event
	// TODO deal with resourceVersion param
	evList := make([]watch.Event, 0, len(eps))
	for _, ep := range eps {
		evList = append(evList, watch.Event{Type: watch.Added, Object: ep})
	}

	// broadcast watch with prefix
	return em.broadcaster.WatchWithPrefix(evList), nil
}

func (em *endpointMux) ListObjects(selector labels.Selector, appendFn cache.AppendFunc) error {
	store := em.epInformer.GetStore()
	return cache.ListAll(store, selector, appendFn)
}

func (em *endpointMux) OnAdd(obj interface{}) {
	eps, ok := obj.(*v1.Endpoints)
	if !ok {
		return
	}
	em.broadcaster.Action(watch.Added, eps)
}

func (em *endpointMux) OnUpdate(oldObj interface{}, newObj interface{}) {
	eps, ok := newObj.(*v1.Endpoints)
	if !ok {
		return
	}
	em.broadcaster.Action(watch.Modified, eps)
}

func (em *endpointMux) OnDelete(obj interface{}) {
	eps, ok := obj.(*v1.Endpoints)
	if !ok {
		return
	}
	em.broadcaster.Action(watch.Deleted, eps)
}
