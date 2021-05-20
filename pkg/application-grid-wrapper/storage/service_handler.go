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
	"k8s.io/klog"
)

type serviceHandler struct {
	cache *storageCache
}

func (sh *serviceHandler) add(service *v1.Service) {
	sc := sh.cache

	sc.mu.Lock()

	serviceKey := types.NamespacedName{Namespace: service.Namespace, Name: service.Name}
	klog.Infof("Adding service %v", serviceKey)
	sc.serviceChan <- watch.Event{
		Type:   watch.Added,
		Object: service,
	}
	sc.servicesMap[serviceKey] = &serviceContainer{
		svc:  service,
		keys: getTopologyKeys(&service.ObjectMeta),
	}

	// update endpoints
	changedEps := sc.rebuildEndpointsMap()

	sc.mu.Unlock()

	for _, eps := range changedEps {
		sc.endpointsChan <- eps
	}
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

	sc.serviceChan <- watch.Event{
		Type:   watch.Modified,
		Object: service,
	}

	serviceContainer.svc = service
	// return directly when topologyKeys of service stay unchanged
	if reflect.DeepEqual(serviceContainer.keys, newTopologyKeys) {
		sc.mu.Unlock()
		return
	}

	serviceContainer.keys = newTopologyKeys

	// update endpoints
	changedEps := sc.rebuildEndpointsMap()
	sc.mu.Unlock()

	for _, eps := range changedEps {
		sc.endpointsChan <- eps
	}
}

func (sh *serviceHandler) delete(service *v1.Service) {
	sc := sh.cache

	sc.mu.Lock()

	serviceKey := types.NamespacedName{Namespace: service.Namespace, Name: service.Name}
	klog.Infof("Deleting service %v", serviceKey)
	sc.serviceChan <- watch.Event{
		Type:   watch.Deleted,
		Object: service,
	}
	delete(sc.servicesMap, serviceKey)

	// update endpoints
	changedEps := sc.rebuildEndpointsMap()

	sc.mu.Unlock()

	for _, eps := range changedEps {
		sc.endpointsChan <- eps
	}
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
