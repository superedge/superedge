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
	v1 "github.com/superedge/superedge/pkg/apps-manager/apis/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// EDeploymentLister helps list EDeployments.
// All objects returned here must be treated as read-only.
type EDeploymentLister interface {
	// List lists all EDeployments in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.EDeployment, err error)
	// EDeployments returns an object that can list and get EDeployments.
	EDeployments(namespace string) EDeploymentNamespaceLister
	EDeploymentListerExpansion
}

// eDeploymentLister implements the EDeploymentLister interface.
type eDeploymentLister struct {
	indexer cache.Indexer
}

// NewEDeploymentLister returns a new EDeploymentLister.
func NewEDeploymentLister(indexer cache.Indexer) EDeploymentLister {
	return &eDeploymentLister{indexer: indexer}
}

// List lists all EDeployments in the indexer.
func (s *eDeploymentLister) List(selector labels.Selector) (ret []*v1.EDeployment, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.EDeployment))
	})
	return ret, err
}

// EDeployments returns an object that can list and get EDeployments.
func (s *eDeploymentLister) EDeployments(namespace string) EDeploymentNamespaceLister {
	return eDeploymentNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// EDeploymentNamespaceLister helps list and get EDeployments.
// All objects returned here must be treated as read-only.
type EDeploymentNamespaceLister interface {
	// List lists all EDeployments in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.EDeployment, err error)
	// Get retrieves the EDeployment from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.EDeployment, error)
	EDeploymentNamespaceListerExpansion
}

// eDeploymentNamespaceLister implements the EDeploymentNamespaceLister
// interface.
type eDeploymentNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all EDeployments in the indexer for a given namespace.
func (s eDeploymentNamespaceLister) List(selector labels.Selector) (ret []*v1.EDeployment, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.EDeployment))
	})
	return ret, err
}

// Get retrieves the EDeployment from the indexer for a given namespace and name.
func (s eDeploymentNamespaceLister) Get(name string) (*v1.EDeployment, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("edeployment"), name)
	}
	return obj.(*v1.EDeployment), nil
}
