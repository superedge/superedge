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
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
)

type endpointSliceV1Beta1Handler struct {
	cache *storageCache
}

func (eh *endpointSliceV1Beta1Handler) add(endpointSlice *discoveryv1beta1.EndpointSlice) {
	sc := eh.cache

	sc.mu.Lock()

	endpointsKey := types.NamespacedName{Namespace: endpointSlice.Namespace, Name: endpointSlice.Name}
	klog.Infof("Adding endpointSlice %v", endpointsKey)
	newEps := pruneEndpointSliceV1Beta1(sc.hostName, sc.nodesMap, sc.servicesMap, endpointSlice, sc.localNodeInfo, sc.wrapperInCluster, sc.serviceAutonomyEnhancementEnabled)

	sc.endpointSliceV1Beta1Map[endpointsKey] = &endpointSliceV1Beta1Container{
		endpointSlice: endpointSlice,
		modified:      newEps,
	}

	sc.mu.Unlock()

	sc.endpointSliceV1Beta1Boardcaster.ActionOrDrop(
		watch.Added,
		newEps)

}

func (eh *endpointSliceV1Beta1Handler) update(endpointSlice *discoveryv1beta1.EndpointSlice) {
	sc := eh.cache

	sc.mu.Lock()
	endpointsKey := types.NamespacedName{Namespace: endpointSlice.Namespace, Name: endpointSlice.Name}
	klog.Infof("Updating endpointSlice %v", endpointsKey)

	endpointSliceContainer, found := sc.endpointSliceV1Beta1Map[endpointsKey]
	if !found {
		sc.mu.Unlock()
		klog.Errorf("Updating non-existed endpointSlice %v", endpointsKey)
		return
	}
	endpointSliceContainer.endpointSlice = endpointSlice
	newEps := pruneEndpointSliceV1Beta1(sc.hostName, sc.nodesMap, sc.servicesMap, endpointSlice, sc.localNodeInfo, sc.wrapperInCluster, sc.serviceAutonomyEnhancementEnabled)
	changed := !apiequality.Semantic.DeepEqual(endpointSliceContainer.modified, newEps)
	if changed {
		endpointSliceContainer.modified = newEps
	}
	sc.mu.Unlock()

	if changed {
		sc.endpointSliceV1Beta1Boardcaster.ActionOrDrop(
			watch.Modified,
			newEps)

	}
}

func (eh *endpointSliceV1Beta1Handler) delete(endpointSlice *discoveryv1beta1.EndpointSlice) {
	sc := eh.cache

	sc.mu.Lock()

	endpointsKey := types.NamespacedName{Namespace: endpointSlice.Namespace, Name: endpointSlice.Name}
	klog.Infof("Deleting endpointSlice %v", endpointsKey)
	endpointSliceContainer, found := sc.endpointSliceV1Beta1Map[endpointsKey]
	if !found {
		sc.mu.Unlock()
		klog.Errorf("Deleting non-existed endpointSlice %v", endpointsKey)
		return
	}
	delete(sc.endpointSliceV1Beta1Map, endpointsKey)

	sc.mu.Unlock()

	sc.endpointSliceV1Beta1Boardcaster.ActionOrDrop(
		watch.Deleted,
		endpointSliceContainer.modified)
}

func (eh *endpointSliceV1Beta1Handler) OnAdd(obj interface{}) {
	eps, ok := obj.(*discoveryv1beta1.EndpointSlice)
	if !ok {
		return
	}
	eh.add(eps)
}

func (eh *endpointSliceV1Beta1Handler) OnUpdate(_, newObj interface{}) {
	eps, ok := newObj.(*discoveryv1beta1.EndpointSlice)
	if !ok {
		return
	}
	eh.update(eps)
}

func (eh *endpointSliceV1Beta1Handler) OnDelete(obj interface{}) {
	eps, ok := obj.(*discoveryv1beta1.EndpointSlice)
	if !ok {
		return
	}
	eh.delete(eps)
}
