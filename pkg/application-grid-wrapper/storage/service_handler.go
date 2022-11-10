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

package storage

import (
	"reflect"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
)

type serviceHandler struct {
	cache *storageCache
}

func (sh *serviceHandler) add(service *v1.Service) {
	sc := sh.cache

	sc.mu.Lock()

	serviceKey := types.NamespacedName{Namespace: service.Namespace, Name: service.Name}
	klog.Infof("Adding service %v", serviceKey)
	sc.serviceBroardcaster.ActionOrDrop(
		watch.Added, service)
	sc.servicesMap[serviceKey] = &serviceContainer{
		svc:  service,
		keys: getTopologyKeys(&service.ObjectMeta),
	}
	sc.mu.Unlock()
	sc.resync()
}

func (sh *serviceHandler) update(service *v1.Service) {
	sc := sh.cache

	sc.mu.Lock()
	serviceKey := types.NamespacedName{Namespace: service.Namespace, Name: service.Name}
	klog.Infof("Updating service %v", serviceKey)
	newTopologyKeys := getTopologyKeys(&service.ObjectMeta)
	serviceContainer, found := sc.servicesMap[serviceKey]
	if !found {
		sc.mu.Unlock()
		klog.Errorf("update non-existed service, %v", serviceKey)
		return
	}

	sc.serviceBroardcaster.ActionOrDrop(
		watch.Modified, service)

	serviceContainer.svc = service
	// return directly when topologyKeys of service stay unchanged
	if reflect.DeepEqual(serviceContainer.keys, newTopologyKeys) {
		sc.mu.Unlock()
		return
	}

	serviceContainer.keys = newTopologyKeys
	sc.mu.Unlock()

	sc.resync()
}

func (sh *serviceHandler) delete(service *v1.Service) {
	sc := sh.cache

	sc.mu.Lock()

	serviceKey := types.NamespacedName{Namespace: service.Namespace, Name: service.Name}
	klog.Infof("Deleting service %v", serviceKey)
	sc.serviceBroardcaster.ActionOrDrop(watch.Deleted, service)
	delete(sc.servicesMap, serviceKey)

	sc.mu.Unlock()
	sc.resync()
}

func (sh *serviceHandler) OnAdd(obj interface{}) {
	svc, ok := obj.(*v1.Service)
	if !ok {
		return
	}
	sh.add(svc)
}

func (sh *serviceHandler) OnUpdate(_, newObj interface{}) {
	svc, ok := newObj.(*v1.Service)
	if !ok {
		return
	}
	sh.update(svc)
}

func (sh *serviceHandler) OnDelete(obj interface{}) {
	svc, ok := obj.(*v1.Service)
	if !ok {
		return
	}
	sh.delete(svc)
}
