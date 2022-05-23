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
	// NodeCacheURL is cached url
	NodeCacheURL = "/api/v1/nodes"
)

func init() {
	RegisterMuxFactory(NodeCacheURL, &nodeCacheMuxFactory{})
}

type nodeCacheMuxFactory struct{}

var _ CacheMuxFactory = &nodeCacheMuxFactory{}

func (f *nodeCacheMuxFactory) Create(hostname string, informerFactory informers.SharedInformerFactory) (CacheMux, error) {
	return NewNodeMux(hostname, informerFactory)
}

type nodeMux struct {
	hostname     string
	broadcaster  *watch.Broadcaster
	nodeInformer cache.SharedIndexInformer
	nodeLister   listerv1.NodeLister
}

var _ CacheMux = &nodeMux{}

// NewNodeMux ...
func NewNodeMux(hostname string, informerFactory informers.SharedInformerFactory) (CacheMux, error) {
	ba := watch.NewLongQueueBroadcaster(maxQueuedEvents, watch.DropIfChannelFull)
	informer := informerFactory.Core().V1().Nodes()
	nodeMux := &nodeMux{
		hostname:     hostname,
		broadcaster:  ba,
		nodeInformer: informer.Informer(),
		nodeLister:   informer.Lister(),
	}
	// don't resync to downstream watchers
	informer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    nodeMux.OnAdd,
		UpdateFunc: nodeMux.OnUpdate,
		DeleteFunc: nodeMux.OnDelete,
	}, 0)

	return nodeMux, nil
}

func (nm *nodeMux) Name() string {
	return NodeCacheURL
}

func (nm *nodeMux) Match(method, URLPath string) bool {
	if method == http.MethodGet && strings.HasPrefix(URLPath, NodeCacheURL) {
		return true
	}
	return false
}

// watch will ignore resourceVersion and return all
func (nm *nodeMux) Watch(bookmark bool, ResourceVersion string) (watch.Interface, error) {
	// list all obj in all namespaces
	nodes, err := nm.nodeLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("List node error by node lister, will drop first, err %v", err)
		return nil, err
	}
	klog.V(4).Infof("WatchWithPrefix node list length %v", len(nodes))
	// build ADD event
	// TODO deal with resourceVersion param
	// sort obj by resourceVersion
	// index by resourceVersion
	evList := make([]watch.Event, 0, len(nodes))
	for _, n := range nodes {
		evList = append(evList, watch.Event{Type: watch.Added, Object: n})
	}

	// broadcast watch with prefix
	return nm.broadcaster.WatchWithPrefix(evList), nil
}

func (nm *nodeMux) ListObjects(selector labels.Selector, appendFn cache.AppendFunc) error {
	store := nm.nodeInformer.GetStore()
	return cache.ListAll(store, selector, appendFn)
}

func (nm *nodeMux) OnAdd(obj interface{}) {
	n, ok := obj.(*v1.Node)
	if !ok {
		return
	}
	nm.broadcaster.Action(watch.Added, n)
}

func (nm *nodeMux) OnUpdate(oldObj interface{}, newObj interface{}) {
	n, ok := newObj.(*v1.Node)
	if !ok {
		return
	}
	nm.broadcaster.Action(watch.Modified, n)
}

func (nm *nodeMux) OnDelete(obj interface{}) {
	n, ok := obj.(*v1.Node)
	if !ok {
		return
	}
	nm.broadcaster.Action(watch.Deleted, n)
}
