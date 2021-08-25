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
	discovery "k8s.io/api/discovery/v1beta1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
)

type endpointSliceHandler struct {
	cache *storageCache
}

func (eh *endpointSliceHandler) add(endpointSlice *discovery.EndpointSlice) {
	sc := eh.cache

	sc.mu.Lock()

	endpointsKey := types.NamespacedName{Namespace: endpointSlice.Namespace, Name: endpointSlice.Name}
	klog.Infof("Adding endpoints %v", endpointsKey)
	newEps := pruneEndpointSlice(sc.hostName, sc.nodesMap, sc.servicesMap, endpointSlice, sc.localNodeInfo, sc.wrapperInCluster, sc.serviceAutonomyEnhancementEnabled)
	//sc.endpointsMap[endpointsKey] = &endpointsContainer{
	//	endpoints: endpoints,
	//	modified:  newEps,
	//}
	sc.endpointSliceMap[endpointsKey] = &endpointSliceContainer{
		endpointSlice: endpointSlice,
		modified:      newEps,
	}

	sc.mu.Unlock()

	sc.endpointSliceChan <- watch.Event{
		Type:   watch.Added,
		Object: newEps,
	}
}

func (eh *endpointSliceHandler) update(endpointSlice *discovery.EndpointSlice) {
	sc := eh.cache

	sc.mu.Lock()
	endpointsKey := types.NamespacedName{Namespace: endpointSlice.Namespace, Name: endpointSlice.Name}
	klog.Infof("Updating endpoints %v", endpointsKey)

	endpointSliceContainer, found := sc.endpointSliceMap[endpointsKey]
	if !found {
		sc.mu.Unlock()
		klog.Errorf("Updating non-existed endpoints %v", endpointsKey)
		return
	}
	endpointSliceContainer.endpointSlice = endpointSlice
	newEps := pruneEndpointSlice(sc.hostName, sc.nodesMap, sc.servicesMap, endpointSlice, sc.localNodeInfo, sc.wrapperInCluster, sc.serviceAutonomyEnhancementEnabled)
	changed := !apiequality.Semantic.DeepEqual(endpointSliceContainer.modified, newEps)
	if changed {
		endpointSliceContainer.modified = newEps
	}
	sc.mu.Unlock()

	if changed {
		sc.endpointSliceChan <- watch.Event{
			Type:   watch.Modified,
			Object: newEps,
		}
	}
}

func (eh *endpointSliceHandler) delete(endpointSlice *discovery.EndpointSlice) {
	sc := eh.cache

	sc.mu.Lock()

	endpointsKey := types.NamespacedName{Namespace: endpointSlice.Namespace, Name: endpointSlice.Name}
	klog.Infof("Deleting endpoints %v", endpointsKey)
	endpointSliceContainer, found := sc.endpointSliceMap[endpointsKey]
	if !found {
		sc.mu.Unlock()
		klog.Errorf("Deleting non-existed endpoints %v", endpointsKey)
		return
	}
	delete(sc.endpointSliceMap, endpointsKey)

	sc.mu.Unlock()

	sc.endpointSliceChan <- watch.Event{
		Type:   watch.Deleted,
		Object: endpointSliceContainer.modified,
	}
}

func (eh *endpointSliceHandler) OnAdd(obj interface{}) {
	eps, ok := obj.(*discovery.EndpointSlice)
	if !ok {
		return
	}
	eh.add(eps)
}

func (eh *endpointSliceHandler) OnUpdate(_, newObj interface{}) {
	eps, ok := newObj.(*discovery.EndpointSlice)
	if !ok {
		return
	}
	eh.update(eps)
}

func (eh *endpointSliceHandler) OnDelete(obj interface{}) {
	eps, ok := obj.(*discovery.EndpointSlice)
	if !ok {
		return
	}
	eh.delete(eps)
}
