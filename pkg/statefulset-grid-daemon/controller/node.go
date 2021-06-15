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

package controller

import (
	"fmt"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"reflect"
)

func (ssgdc *StatefulSetGridDaemonController) addNode(obj interface{}) {
	node := obj.(*corev1.Node)
	if node.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		ssgdc.deleteNode(node)
		return
	}

	sets := ssgdc.getStatefulSetForNode(node)
	for _, set := range sets {
		klog.V(4).Infof("Node %s(its relevant StatefulSet %s) added.", node.Name, set.Name)
		ssgdc.enqueueStatefulSet(set)
	}
}

func (ssgdc *StatefulSetGridDaemonController) updateNode(oldObj, newObj interface{}) {
	oldNode := oldObj.(*corev1.Node)
	curNode := newObj.(*corev1.Node)
	if curNode.ResourceVersion == oldNode.ResourceVersion {
		// Periodic resync will send update events for all known Nodes.
		// Two different versions of the same Node will always have different RVs.
		return
	}
	labelChanged := !reflect.DeepEqual(curNode.Labels, oldNode.Labels)
	// Only handles nodes whose label has changed.
	if labelChanged {
		sets := ssgdc.getStatefulSetForNode(curNode)
		for _, set := range sets {
			klog.V(4).Infof("Node %s(its relevant StatefulSet %s) updated.", curNode.Name, set.Name)
			ssgdc.enqueueStatefulSet(set)
		}
	}
}

func (ssgdc *StatefulSetGridDaemonController) deleteNode(obj interface{}) {
	node, ok := obj.(*corev1.Node)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		node, ok = tombstone.Obj.(*corev1.Node)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object is not a node %#v", obj))
			return
		}
	}
	sets := ssgdc.getStatefulSetForNode(node)
	for _, set := range sets {
		klog.V(4).Infof("Node %s(its relevant StatefulSet %s) deleted.", node.Name, set.Name)
		ssgdc.enqueueStatefulSet(set)
	}
}

func (ssgdc *StatefulSetGridDaemonController) getStatefulSetForNode(node *corev1.Node) []*appv1.StatefulSet {
	selector, err := common.GetNodesSelector(node)
	if err != nil {
		return nil
	}
	setList, err := ssgdc.setLister.List(selector)
	if err != nil {
		return nil
	}
	var sets []*appv1.StatefulSet
	for _, set := range setList {
		if rel, err := ssgdc.IsConcernedStatefulSet(set); err != nil || !rel {
			continue
		}
		sets = append(sets, set)
	}
	return sets
}
