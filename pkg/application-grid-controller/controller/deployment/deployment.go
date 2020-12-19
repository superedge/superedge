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
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"

	crdv1 "superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	"superedge/pkg/application-grid-controller/controller/common"
)

func (dgc *DeploymentGridController) addDeployment(obj interface{}) {
	d := obj.(*appsv1.Deployment)

	if !common.IsConcernedObject(d.ObjectMeta) {
		return
	}

	if d.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		dgc.deleteDeployment(d)
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(d); controllerRef != nil {
		g := dgc.resolveControllerRef(d.Namespace, controllerRef)
		if g == nil {
			return
		}
		klog.V(4).Infof("DeploymentGrid %s added.", g.Name)
		dgc.enqueueDeploymentGrid(g)
		return
	}

	// Otherwise, it's an orphan. Get a list of all matching DeploymentGrids and sync
	// them to see if anyone wants to adopt it.
	gs := dgc.getGridForDeployment(d)
	if len(gs) == 0 {
		return
	}
	klog.V(4).Infof("Orphan Deployment %s added.", d.Name)
	for _, g := range gs {
		dgc.enqueueDeploymentGrid(g)
	}
}

func (dgc *DeploymentGridController) updateDeployment(oldObj, newObj interface{}) {
	oldD := oldObj.(*appsv1.Deployment)
	curD := newObj.(*appsv1.Deployment)

	if !common.IsConcernedObject(curD.ObjectMeta) {
		return
	}

	curControllerRef := metav1.GetControllerOf(curD)
	oldControllerRef := metav1.GetControllerOf(oldD)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged && oldControllerRef != nil {
		// The ControllerRef was changed. Sync the old controller, if any.
		if g := dgc.resolveControllerRef(oldD.Namespace, oldControllerRef); g != nil {
			dgc.enqueueDeploymentGrid(g)
		}
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		g := dgc.resolveControllerRef(curD.Namespace, curControllerRef)
		if g == nil {
			return
		}
		klog.V(4).Infof("DeploymentGrid %s updated.", curD.Name)
		dgc.enqueueDeploymentGrid(g)
		return
	}

	// Otherwise, it's an orphan. If anything changed, sync matching controllers
	// to see if anyone wants to adopt it now.
	labelChanged := !reflect.DeepEqual(curD.Labels, oldD.Labels)
	if labelChanged || controllerRefChanged {
		gs := dgc.getGridForDeployment(curD)
		if len(gs) == 0 {
			return
		}
		klog.V(4).Infof("Orphan Deployment %s updated.", curD.Name)
		for _, g := range gs {
			dgc.enqueueDeploymentGrid(g)
		}
	}
}

func (dgc *DeploymentGridController) deleteDeployment(obj interface{}) {
	d, ok := obj.(*appsv1.Deployment)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		d, ok = tombstone.Obj.(*appsv1.Deployment)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object that is not a Deployment %#v", obj))
			return
		}
	}
	controllerRef := metav1.GetControllerOf(d)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}
	g := dgc.resolveControllerRef(d.Namespace, controllerRef)
	if g == nil {
		return
	}
	if !common.IsConcernedObject(d.ObjectMeta) {
		return
	}
	klog.V(4).Infof("Deployment %s deleted.", d.Name)
	dgc.enqueueDeploymentGrid(g)
}

func (dgc *DeploymentGridController) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *crdv1.DeploymentGrid {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}
	d, err := dgc.dpGridLister.DeploymentGrids(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}
	if d.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return d
}

func (dgc *DeploymentGridController) getGridForDeployment(d *appsv1.Deployment) []*crdv1.DeploymentGrid {
	if len(d.Labels) == 0 {
		return nil
	}

	dgList, err := dgc.dpGridLister.DeploymentGrids(d.Namespace).List(labels.Everything())
	if err != nil {
		return nil
	}

	var deploymentGrids []*crdv1.DeploymentGrid
	for _, g := range dgList {
		selector, err := common.GetDefaultSelector(g.Name)
		if err != nil {
			return nil
		}

		if !selector.Matches(labels.Set(d.Labels)) {
			continue
		}
		deploymentGrids = append(deploymentGrids, g)
	}

	if len(deploymentGrids) == 0 {
		return nil
	}

	if len(deploymentGrids) > 1 {
		// ControllerRef will ensure we don't do anything crazy, but more than one
		// item in this list nevertheless constitutes user error.
		klog.V(4).Infof("user error! more than one deployment is selecting deployment %s/%s with labels: %#v, returning %s/%s",
			d.Namespace, d.Name, d.Labels, deploymentGrids[0].Namespace, deploymentGrids[0].Name)
	}
	return deploymentGrids
}
