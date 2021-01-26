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

package deployment

import (
	"fmt"
	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"reflect"
)

func (dgc *DeploymentGridController) addNode(obj interface{}) {
	node := obj.(*corev1.Node)
	if node.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		dgc.deleteNode(node)
		return
	}
	dgs := dgc.getGridForNode(node)
	if len(dgs) == 0 {
		return
	}
	for _, dg := range dgs {
		klog.V(4).Infof("Node %s(its relevant DeploymentGrid %s) added.", node.Name, dg.Name)
		dgc.enqueueDeploymentGrid(dg)
	}
}

func (dgc *DeploymentGridController) updateNode(oldObj, newObj interface{}) {
	oldNode := oldObj.(*corev1.Node)
	curNode := newObj.(*corev1.Node)
	if curNode.ResourceVersion == oldNode.ResourceVersion {
		// Periodic resync will send update events for all known Nodes.
		// Two different versions of the same Nodes will always have different RVs.
		return
	}
	labelChanged := !reflect.DeepEqual(curNode.Labels, oldNode.Labels)
	// Only handles nodes whose label has changed.
	if labelChanged {
		dgs := dgc.getGridForNode(oldNode, curNode)
		if len(dgs) == 0 {
			return
		}
		for _, dg := range dgs {
			klog.V(4).Infof("Node %s(its relevant StatefulSetGrid %s) updated.", curNode.Name, dg.Name)
			dgc.enqueueDeploymentGrid(dg)
		}
	}
}

func (dgc *DeploymentGridController) deleteNode(obj interface{}) {
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
	dgs := dgc.getGridForNode(node)
	if len(dgs) == 0 {
		return
	}
	for _, dg := range dgs {
		klog.V(4).Infof("Node %s(its relevant StatefulSetGrid %s) deleted.", node.Name, dg.Name)
		dgc.enqueueDeploymentGrid(dg)
	}
}

// getGridForNode get deploymentGrids those gridUniqKey exists in node labels.
func (dgc *DeploymentGridController) getGridForNode(nodes ...*corev1.Node) []*crdv1.DeploymentGrid {
	selector, err := common.GetNodesSelector(nodes...)
	if err != nil {
		return nil
	}
	dgs, err := dgc.dpGridLister.DeploymentGrids("").List(selector)
	if err != nil {
		return nil
	}
	return dgs
}
