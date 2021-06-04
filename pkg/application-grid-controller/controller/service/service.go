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
	"k8s.io/klog/v2"

	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/service/util"
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
		sg := sgc.resolveControllerRef(svc.Namespace, controllerRef)
		if sg == nil {
			return
		}
		klog.V(4).Infof("Service %s(its owner ServiceGrid %s) added.", svc.Name, sg.Name)
		sgc.enqueueServiceGrid(sg)
		return
	}

	// Otherwise, it's an orphan. Get a list of all matching ServiceGrids and sync
	// them to see if anyone wants to adopt it.
	sgs := sgc.getGridForService(svc)
	for _, sg := range sgs {
		klog.V(4).Infof("Orphan Service %s(its possible owner ServiceGrid %s) added.", svc.Name, sg.Name)
		sgc.enqueueServiceGrid(sg)
	}
}

func (sgc *ServiceGridController) updateService(oldObj, newObj interface{}) {
	oldSvc := oldObj.(*corev1.Service)
	curSvc := newObj.(*corev1.Service)
	if curSvc.ResourceVersion == oldSvc.ResourceVersion {
		// Periodic resync will send update events for all known Services.
		// Two different versions of the same Service will always have different RVs.
		return
	}

	curControllerRef := metav1.GetControllerOf(curSvc)
	oldControllerRef := metav1.GetControllerOf(oldSvc)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged && oldControllerRef != nil {
		// The ControllerRef was changed. Sync the old controller, if any.
		if sg := sgc.resolveControllerRef(oldSvc.Namespace, oldControllerRef); sg != nil {
			klog.V(4).Infof("Service %s(its old owner ServiceGrid %s) updated.", oldSvc.Name, sg.Name)
			sgc.enqueueServiceGrid(sg)
		}
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		sg := sgc.resolveControllerRef(curSvc.Namespace, curControllerRef)
		if sg == nil {
			return
		}
		klog.V(4).Infof("Service %s(its owner ServiceGrid %s) updated.", curSvc.Name, sg.Name)
		sgc.enqueueServiceGrid(sg)
		return
	}

	if !common.IsConcernedObject(curSvc.ObjectMeta) {
		return
	}

	// Otherwise, it's an orphan. If anything changed, sync matching controllers
	// to see if anyone wants to adopt it now.
	labelChanged := !reflect.DeepEqual(curSvc.Labels, oldSvc.Labels)
	if labelChanged || controllerRefChanged {
		sgs := sgc.getGridForService(curSvc)
		for _, sg := range sgs {
			klog.V(4).Infof("Orphan Service %s(its possible owner ServiceGrid %s) updated.", curSvc.Name, sg.Name)
			sgc.enqueueServiceGrid(sg)
		}
	}
}

func (sgc *ServiceGridController) deleteService(obj interface{}) {
	svc, ok := obj.(*corev1.Service)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		svc, ok = tombstone.Obj.(*corev1.Service)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object that is not a Service %#v", obj))
			return
		}
	}
	if !common.IsConcernedObject(svc.ObjectMeta) {
		return
	}
	controllerRef := metav1.GetControllerOf(svc)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}
	sg := sgc.resolveControllerRef(svc.Namespace, controllerRef)
	if sg == nil {
		return
	}
	klog.V(4).Infof("Service %s(its owner ServiceGrid %s) deleted.", svc.Name, sg.Name)
	sgc.enqueueServiceGrid(sg)
}

func (sgc *ServiceGridController) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *crdv1.ServiceGrid {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != util.ControllerKind.Kind {
		return nil
	}
	sg, err := sgc.svcGridLister.ServiceGrids(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}
	if sg.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return sg
}

func (sgc *ServiceGridController) getGridForService(svc *corev1.Service) []*crdv1.ServiceGrid {
	if len(svc.Labels) == 0 {
		return nil
	}

	sgList, err := sgc.svcGridLister.ServiceGrids(svc.Namespace).List(labels.Everything())
	if err != nil {
		return nil
	}

	var serviceGrids []*crdv1.ServiceGrid
	for _, sg := range sgList {
		selector, err := common.GetDefaultSelector(sg.Name)
		if err != nil {
			return nil
		}

		if !selector.Matches(labels.Set(svc.Labels)) {
			continue
		}
		serviceGrids = append(serviceGrids, sg)
	}

	if len(serviceGrids) > 1 {
		// ControllerRef will ensure we don't do anything crazy, but more than one
		// item in this list nevertheless constitutes user error.
		klog.V(4).Infof("user error! service %s/%s with labels: %#v selects more than one serviceGrid, returning %#v",
			svc.Namespace, svc.Name, svc.Labels, serviceGrids)
	}
	return serviceGrids
}
