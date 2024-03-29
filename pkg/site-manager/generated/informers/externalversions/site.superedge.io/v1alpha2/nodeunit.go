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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha2

import (
	"context"
	time "time"

	sitesuperedgeiov1alpha2 "github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha2"
	versioned "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	internalinterfaces "github.com/superedge/superedge/pkg/site-manager/generated/informers/externalversions/internalinterfaces"
	v1alpha2 "github.com/superedge/superedge/pkg/site-manager/generated/listers/site.superedge.io/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// NodeUnitInformer provides access to a shared informer and lister for
// NodeUnits.
type NodeUnitInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha2.NodeUnitLister
}

type nodeUnitInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewNodeUnitInformer constructs a new informer for NodeUnit type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewNodeUnitInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredNodeUnitInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredNodeUnitInformer constructs a new informer for NodeUnit type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredNodeUnitInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SiteV1alpha2().NodeUnits().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SiteV1alpha2().NodeUnits().Watch(context.TODO(), options)
			},
		},
		&sitesuperedgeiov1alpha2.NodeUnit{},
		resyncPeriod,
		indexers,
	)
}

func (f *nodeUnitInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredNodeUnitInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *nodeUnitInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&sitesuperedgeiov1alpha2.NodeUnit{}, f.defaultInformer)
}

func (f *nodeUnitInformer) Lister() v1alpha2.NodeUnitLister {
	return v1alpha2.NewNodeUnitLister(f.Informer().GetIndexer())
}
