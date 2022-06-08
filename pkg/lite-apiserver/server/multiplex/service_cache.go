/*
Copyright 2020 The SuperEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file excsvct in compliance with the License.
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
	// serviceCacheURL is cached url
	ServiceCacheURL = "/api/v1/services"
)

func init() {
	RegisterMuxFactory(ServiceCacheURL, &serviceCacheMuxFactory{})
}

type serviceCacheMuxFactory struct{}

var _ CacheMuxFactory = &serviceCacheMuxFactory{}

func (f *serviceCacheMuxFactory) Create(hostname string, informerFactory informers.SharedInformerFactory) (CacheMux, error) {
	return NewServiceMux(hostname, informerFactory)
}

type serviceMux struct {
	hostname    string
	broadcaster *watch.Broadcaster
	svcInformer cache.SharedIndexInformer
	svcLister   listerv1.ServiceLister
}

var _ CacheMux = &serviceMux{}

// NewserviceMux ...
func NewServiceMux(hostname string, informerFactory informers.SharedInformerFactory) (CacheMux, error) {
	ba := watch.NewLongQueueBroadcaster(maxQueuedEvents, watch.DropIfChannelFull)
	informer := informerFactory.Core().V1().Services()
	svcMux := &serviceMux{
		hostname:    hostname,
		broadcaster: ba,
		svcInformer: informer.Informer(),
		svcLister:   informer.Lister(),
	}
	// don't resync to downstream watchers
	informer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    svcMux.OnAdd,
		UpdateFunc: svcMux.OnUpdate,
		DeleteFunc: svcMux.OnDelete,
	}, 0)

	// run with informerFactory

	// stop := make(chan struct{})
	// go informer.Informer().Run(stop)
	// if !cache.WaitForNamedCacheSync("serviceMux", stop, informer.Informer().HasSynced) {
	// 	return nil, fmt.Errorf("can't sync informers")
	// }

	return svcMux, nil
}

func (em *serviceMux) Name() string {
	return ServiceCacheURL
}

func (sm *serviceMux) Match(method, URLPath string) bool {
	if method == http.MethodGet && strings.HasPrefix(URLPath, ServiceCacheURL) {
		return true
	}
	return false
}

// watch will ignore resourceVersion and return all
func (sm *serviceMux) Watch(bookmark bool, ResourceVersion string) (watch.Interface, error) {
	// list all obj in all namespaces
	svcs, err := sm.svcLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("List services error by service lister, will drop first, err %v", err)
		return nil, err
	}
	// build ADD event
	// TODO deal with resourceVersion param
	evList := make([]watch.Event, 0, len(svcs))
	for _, svc := range svcs {
		evList = append(evList, watch.Event{Type: watch.Added, Object: svc})
	}

	// broadcast watch with prefix
	return sm.broadcaster.WatchWithPrefix(evList), nil
}

func (sm *serviceMux) ListObjects(selector labels.Selector, appendFn cache.AppendFunc) error {
	store := sm.svcInformer.GetStore()
	return cache.ListAll(store, selector, appendFn)
}

func (sm *serviceMux) OnAdd(obj interface{}) {
	svcs, ok := obj.(*v1.Service)
	if !ok {
		return
	}
	sm.broadcaster.Action(watch.Added, svcs)
}

func (sm *serviceMux) OnUpdate(oldObj interface{}, newObj interface{}) {
	svcs, ok := newObj.(*v1.Service)
	if !ok {
		return
	}
	sm.broadcaster.Action(watch.Modified, svcs)
}

func (sm *serviceMux) OnDelete(obj interface{}) {
	svcs, ok := obj.(*v1.Service)
	if !ok {
		return
	}
	sm.broadcaster.Action(watch.Deleted, svcs)
}
