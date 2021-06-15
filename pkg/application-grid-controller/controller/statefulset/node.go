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

package statefulset

import (
	"fmt"
	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"reflect"
)

func (ssgc *StatefulSetGridController) addNode(obj interface{}) {
	node := obj.(*corev1.Node)
	if node.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		ssgc.deleteNode(node)
		return
	}
	ssgs := ssgc.getGridForNode(node)
	for _, ssg := range ssgs {
		klog.V(4).Infof("Node %s(its relevant StatefulSetGrid %s) added.", node.Name, ssg.Name)
		ssgc.enqueueStatefulSetGrid(ssg)
	}
}

func (ssgc *StatefulSetGridController) updateNode(oldObj, newObj interface{}) {
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
		ssgs := ssgc.getGridForNode(oldNode, curNode)
		for _, ssg := range ssgs {
			klog.V(4).Infof("Node %s(its relevant StatefulSetGrid %s) updated.", curNode.Name, ssg.Name)
			ssgc.enqueueStatefulSetGrid(ssg)
		}
	}
}

func (ssgc *StatefulSetGridController) deleteNode(obj interface{}) {
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
	ssgs := ssgc.getGridForNode(node)
	for _, ssg := range ssgs {
		klog.V(4).Infof("Node %s(its relevant StatefulSetGrid %s) deleted.", node.Name, ssg.Name)
		ssgc.enqueueStatefulSetGrid(ssg)
	}
}

// getGridForNode filters statefulsetGrids those gridUniqKey exists in node labels.
func (ssgc *StatefulSetGridController) getGridForNode(nodes ...*corev1.Node) []*crdv1.StatefulSetGrid {
	// Return directly when there is no labels at all
	needCheck := false
	for _, node := range nodes {
		if len(node.Labels) == 0 {
			continue
		} else {
			needCheck = true
			break
		}
	}
	if !needCheck {
		return nil
	}
	// Filter relevant grids of nodes by labels
	ssgs, err := ssgc.setGridLister.List(labels.Everything())
	if err != nil {
		return nil
	}
	var targetSsgs []*crdv1.StatefulSetGrid
	for _, ssg := range ssgs {
		for _, node := range nodes {
			if _, exist := node.Labels[ssg.Spec.GridUniqKey]; exist {
				targetSsgs = append(targetSsgs, ssg)
			}
		}
	}
	return targetSsgs
}
