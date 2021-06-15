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
	"k8s.io/klog/v2"
)

type nodeHandler struct {
	cache *storageCache
}

func (nh *nodeHandler) add(node *v1.Node) {
	sc := nh.cache

	sc.mu.Lock()

	nodeKey := types.NamespacedName{Namespace: node.Namespace, Name: node.Name}
	klog.Infof("Adding node %v", nodeKey)
	sc.nodesMap[nodeKey] = &nodeContainer{
		node:   node,
		labels: node.Labels,
	}
	// update endpoints
	changedEps := sc.rebuildEndpointsMap()

	sc.mu.Unlock()

	for _, eps := range changedEps {
		sc.endpointsChan <- eps
	}
}

func (nh *nodeHandler) update(node *v1.Node) {
	sc := nh.cache

	sc.mu.Lock()

	nodeKey := types.NamespacedName{Namespace: node.Namespace, Name: node.Name}
	klog.Infof("Updating node %v", nodeKey)
	nodeContainer, found := sc.nodesMap[nodeKey]
	if !found {
		sc.mu.Unlock()
		klog.Errorf("Updating non-existed node %v", nodeKey)
		return
	}

	nodeContainer.node = node
	// return directly when labels of node stay unchanged
	if reflect.DeepEqual(node.Labels, nodeContainer.labels) {
		sc.mu.Unlock()
		return
	}
	nodeContainer.labels = node.Labels

	// update endpoints
	changedEps := sc.rebuildEndpointsMap()

	sc.mu.Unlock()

	for _, eps := range changedEps {
		sc.endpointsChan <- eps
	}
}

func (nh *nodeHandler) delete(node *v1.Node) {
	sc := nh.cache

	sc.mu.Lock()

	nodeKey := types.NamespacedName{Namespace: node.Namespace, Name: node.Name}
	klog.Infof("Deleting node %v", nodeKey)
	delete(sc.nodesMap, nodeKey)

	// update endpoints
	changedEps := sc.rebuildEndpointsMap()

	sc.mu.Unlock()

	for _, eps := range changedEps {
		sc.endpointsChan <- eps
	}
}

func (nh *nodeHandler) OnAdd(obj interface{}) {
	node, ok := obj.(*v1.Node)
	if !ok {
		return
	}
	nh.add(node)
}

func (nh *nodeHandler) OnUpdate(_, newObj interface{}) {
	node, ok := newObj.(*v1.Node)
	if !ok {
		return
	}
	nh.update(node)
}

func (nh *nodeHandler) OnDelete(obj interface{}) {
	node, ok := obj.(*v1.Node)
	if !ok {
		return
	}
	nh.delete(node)
}
