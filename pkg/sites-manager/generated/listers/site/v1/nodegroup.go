/*
Copyright 2021 The SuperEdge authors.

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
// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/superedge/superedge/pkg/sites-manager/apis/site/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// NodeGroupLister helps list NodeGroups.
// All objects returned here must be treated as read-only.
type NodeGroupLister interface {
	// List lists all NodeGroups in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.NodeGroup, err error)
	// NodeGroups returns an object that can list and get NodeGroups.
	NodeGroups(namespace string) NodeGroupNamespaceLister
	NodeGroupListerExpansion
}

// nodeGroupLister implements the NodeGroupLister interface.
type nodeGroupLister struct {
	indexer cache.Indexer
}

// NewNodeGroupLister returns a new NodeGroupLister.
func NewNodeGroupLister(indexer cache.Indexer) NodeGroupLister {
	return &nodeGroupLister{indexer: indexer}
}

// List lists all NodeGroups in the indexer.
func (s *nodeGroupLister) List(selector labels.Selector) (ret []*v1.NodeGroup, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.NodeGroup))
	})
	return ret, err
}

// NodeGroups returns an object that can list and get NodeGroups.
func (s *nodeGroupLister) NodeGroups(namespace string) NodeGroupNamespaceLister {
	return nodeGroupNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// NodeGroupNamespaceLister helps list and get NodeGroups.
// All objects returned here must be treated as read-only.
type NodeGroupNamespaceLister interface {
	// List lists all NodeGroups in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.NodeGroup, err error)
	// Get retrieves the NodeGroup from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.NodeGroup, error)
	NodeGroupNamespaceListerExpansion
}

// nodeGroupNamespaceLister implements the NodeGroupNamespaceLister
// interface.
type nodeGroupNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all NodeGroups in the indexer for a given namespace.
func (s nodeGroupNamespaceLister) List(selector labels.Selector) (ret []*v1.NodeGroup, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.NodeGroup))
	})
	return ret, err
}

// Get retrieves the NodeGroup from the indexer for a given namespace and name.
func (s nodeGroupNamespaceLister) Get(name string) (*v1.NodeGroup, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("nodegroup"), name)
	}
	return obj.(*v1.NodeGroup), nil
}
