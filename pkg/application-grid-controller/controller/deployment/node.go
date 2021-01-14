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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
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
		dgc.deleteDeployment(node)
		return
	}
	dgs := dgc.getGridForNode(node)
	if len(dgs) == 0 {
		return
	}
	klog.V(4).Infof("Node %s added.", node.Name)
	for _, dg := range dgs {
		dgc.enqueueDeploymentGrid(dg)
	}
}

func (dgc *DeploymentGridController) updateNode(oldObj, newObj interface{}) {
	oldNode := oldObj.(*corev1.Node)
	curNode := newObj.(*corev1.Node)
	labelChanged := !reflect.DeepEqual(curNode.Labels, oldNode.Labels)
	// Only handles nodes whose label has changed.
	if labelChanged {
		dgs := dgc.getGridForNode(curNode)
		if len(dgs) == 0 {
			return
		}
		klog.V(4).Infof("Node %s updated.", curNode.Name)
		for _, dg := range dgs {
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
	klog.V(4).Infof("Node %s deleted.", node.Name)
	for _, dg := range dgs {
		dgc.enqueueDeploymentGrid(dg)
	}
}

// getGridForNode get deploymentGrids those gridUniqKey exists in node labels.
func (dgc *DeploymentGridController) getGridForNode(node *corev1.Node) []*crdv1.DeploymentGrid {
	if len(node.Labels) == 0 {
		return nil
	}
	dgList, err := dgc.dpGridLister.DeploymentGrids("").List(labels.Everything())
	if err != nil {
		return nil
	}
	var deploymentGrids []*crdv1.DeploymentGrid
	for _, dg := range dgList {
		if _, exist := node.Labels[dg.Spec.GridUniqKey]; exist {
			deploymentGrids = append(deploymentGrids, dg)
		}
	}
	if len(deploymentGrids) == 0 {
		return nil
	}
	return deploymentGrids
}
