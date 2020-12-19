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
	v1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog"
)

type endpointsHandler struct {
	cache *storageCache
}

func (eh *endpointsHandler) add(endpoints *v1.Endpoints) {
	sc := eh.cache

	sc.mu.Lock()

	endpointsKey := types.NamespacedName{Namespace: endpoints.Namespace, Name: endpoints.Name}
	klog.Infof("Adding endpoints %v", endpointsKey)
	newEps := pruneEndpoints(sc.hostName, sc.nodesMap, sc.servicesMap, endpoints, sc.wrapperInCluster)
	sc.endpointsMap[endpointsKey] = &endpointsContainer{
		c:        endpoints,
		modified: newEps,
	}

	sc.mu.Unlock()

	sc.endpointsChan <- watch.Event{
		Type:   watch.Added,
		Object: newEps,
	}
}

func (eh *endpointsHandler) update(endpoints *v1.Endpoints) {
	sc := eh.cache

	sc.mu.Lock()
	endpointsKey := types.NamespacedName{Namespace: endpoints.Namespace, Name: endpoints.Name}
	klog.Infof("Updating endpoints %v", endpointsKey)

	oldEndpoints, found := sc.endpointsMap[endpointsKey]
	if !found {
		sc.mu.Unlock()
		klog.Errorf("Updating non-existed endpoints %v", endpointsKey)
		return
	}
	oldEndpoints.c = endpoints
	newEps := pruneEndpoints(sc.hostName, sc.nodesMap, sc.servicesMap, endpoints, sc.wrapperInCluster)
	changed := !apiequality.Semantic.DeepEqual(oldEndpoints.modified, newEps)
	if changed {
		oldEndpoints.modified = newEps
	}
	sc.mu.Unlock()

	if changed {
		sc.endpointsChan <- watch.Event{
			Type:   watch.Modified,
			Object: newEps,
		}
	}
}

func (eh *endpointsHandler) delete(endpoints *v1.Endpoints) {
	sc := eh.cache

	sc.mu.Lock()

	endpointsKey := types.NamespacedName{Namespace: endpoints.Namespace, Name: endpoints.Name}
	klog.Infof("Deleting endpoints %v", endpointsKey)
	oldEndpoints, found := sc.endpointsMap[endpointsKey]
	if !found {
		sc.mu.Unlock()
		klog.Errorf("Updating non-existed endpoints %v", endpointsKey)
		return
	}
	delete(sc.endpointsMap, endpointsKey)

	sc.mu.Unlock()

	sc.endpointsChan <- watch.Event{
		Type:   watch.Deleted,
		Object: oldEndpoints.modified,
	}
}

func (eh *endpointsHandler) OnAdd(obj interface{}) {
	eps, ok := obj.(*v1.Endpoints)
	if !ok {
		return
	}
	eh.add(eps)
}

func (eh *endpointsHandler) OnUpdate(_, newObj interface{}) {
	eps, ok := newObj.(*v1.Endpoints)
	if !ok {
		return
	}
	eh.update(eps)
}

func (eh *endpointsHandler) OnDelete(obj interface{}) {
	eps, ok := obj.(*v1.Endpoints)
	if !ok {
		return
	}
	eh.delete(eps)
}
