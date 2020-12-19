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

package service

import (
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"

	crdv1 "superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	"superedge/pkg/application-grid-controller/controller/common"
)

func (sgc *ServiceGridController) addService(obj interface{}) {
	svc := obj.(*corev1.Service)
	if !common.IsConcernedObject(svc.ObjectMeta) {
		return
	}
	if svc.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		sgc.deleteService(svc)
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(svc); controllerRef != nil {
		g := sgc.resolveControllerRef(svc.Namespace, controllerRef)
		if g == nil {
			return
		}
		klog.V(4).Infof("ServiceGrid %s added.", g.Name)
		sgc.enqueueServiceGrid(g)
		return
	}

	// Otherwise, it's an orphan. Get a list of all matching ServiceGrids and sync
	// them to see if anyone wants to adopt it.
	gs := sgc.getGridForService(svc)
	if len(gs) == 0 {
		return
	}
	klog.V(4).Infof("Orphan Service %s added.", svc.Name)
	for _, g := range gs {
		sgc.enqueueServiceGrid(g)
	}
}

func (sgc *ServiceGridController) updateService(oldObj, newObj interface{}) {
	oldD := oldObj.(*corev1.Service)
	curD := newObj.(*corev1.Service)

	if !common.IsConcernedObject(curD.ObjectMeta) {
		return
	}

	curControllerRef := metav1.GetControllerOf(curD)
	oldControllerRef := metav1.GetControllerOf(oldD)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged && oldControllerRef != nil {
		// The ControllerRef was changed. Sync the old controller, if any.
		if g := sgc.resolveControllerRef(oldD.Namespace, oldControllerRef); g != nil {
			sgc.enqueueServiceGrid(g)
		}
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		g := sgc.resolveControllerRef(curD.Namespace, curControllerRef)
		if g == nil {
			return
		}
		klog.V(4).Infof("ServiceGrid %s updated.", curD.Name)
		sgc.enqueueServiceGrid(g)
		return
	}

	// Otherwise, it's an orphan. If anything changed, sync matching controllers
	// to see if anyone wants to adopt it now.
	labelChanged := !reflect.DeepEqual(curD.Labels, oldD.Labels)
	if labelChanged || controllerRefChanged {
		gs := sgc.getGridForService(curD)
		if len(gs) == 0 {
			return
		}
		klog.V(4).Infof("Orphan Service %s updated.", curD.Name)
		for _, g := range gs {
			sgc.enqueueServiceGrid(g)
		}
	}
}

func (sgc *ServiceGridController) deleteService(obj interface{}) {
	d, ok := obj.(*corev1.Service)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		d, ok = tombstone.Obj.(*corev1.Service)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object that is not a Service %#v", obj))
			return
		}
	}
	controllerRef := metav1.GetControllerOf(d)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}
	g := sgc.resolveControllerRef(d.Namespace, controllerRef)
	if g == nil {
		return
	}
	if !common.IsConcernedObject(d.ObjectMeta) {
		return
	}
	klog.V(4).Infof("Service %s deleted.", d.Name)
	sgc.enqueueServiceGrid(g)
}

func (sgc *ServiceGridController) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *crdv1.ServiceGrid {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}
	svc, err := sgc.svcGridLister.ServiceGrids(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}
	if svc.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return svc
}

func (sgc *ServiceGridController) getGridForService(svc *corev1.Service) []*crdv1.ServiceGrid {
	if len(svc.Labels) == 0 {
		return nil
	}

	dgList, err := sgc.svcGridLister.ServiceGrids(svc.Namespace).List(labels.Everything())
	if err != nil {
		return nil
	}

	var serviceGrids []*crdv1.ServiceGrid
	for _, g := range dgList {
		selector, err := common.GetDefaultSelector(g.Name)
		if err != nil {
			return nil
		}

		if !selector.Matches(labels.Set(svc.Labels)) {
			continue
		}
		serviceGrids = append(serviceGrids, g)
	}

	if len(serviceGrids) == 0 {
		return nil
	}

	if len(serviceGrids) > 1 {
		// ControllerRef will ensure we don't do anything crazy, but more than one
		// item in this list nevertheless constitutes user error.
		klog.V(4).Infof("user error! more than one deployment is selecting deployment %s/%s with labels: %#v, returning %s/%s",
			svc.Namespace, svc.Name, svc.Labels, serviceGrids[0].Namespace, serviceGrids[0].Name)
	}
	return serviceGrids
}
