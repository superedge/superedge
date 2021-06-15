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
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/statefulset/util"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
)

func (ssgc *StatefulSetGridController) addStatefulSet(obj interface{}) {
	set := obj.(*appsv1.StatefulSet)

	if !common.IsConcernedObject(set.ObjectMeta) {
		return
	}

	if set.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		ssgc.deleteStatefulSet(set)
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(set); controllerRef != nil {
		ssg := ssgc.resolveControllerRef(set.Namespace, controllerRef)
		if ssg == nil {
			return
		}
		klog.V(4).Infof("StatefulSet %s(its owner StatefulSetGrid %s) added.", set.Name, ssg.Name)
		ssgc.enqueueStatefulSetGrid(ssg)
		return
	}

	// Otherwise, it's an orphan. Get a list of all matching StatefulSetGrids and sync
	// them to see if anyone wants to adopt it.
	ssgs := ssgc.getGridForStatefulSet(set)
	for _, ssg := range ssgs {
		klog.V(4).Infof("Orphan StatefulSet %s(its possible owner StatefulSetGrid %s) added.", set.Name, ssg.Name)
		ssgc.enqueueStatefulSetGrid(ssg)
	}
}

func (ssgc *StatefulSetGridController) updateStatefulSet(oldObj, newObj interface{}) {
	oldSet := oldObj.(*appsv1.StatefulSet)
	curSet := newObj.(*appsv1.StatefulSet)
	if curSet.ResourceVersion == oldSet.ResourceVersion {
		// Periodic resync will send update events for all known StatefulSets.
		// Two different versions of the same StatefulSet will always have different RVs.
		return
	}

	oldControllerRef := metav1.GetControllerOf(oldSet)
	curControllerRef := metav1.GetControllerOf(curSet)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged && oldControllerRef != nil {
		// The ControllerRef was changed. Sync the old controller, if any.
		if ssg := ssgc.resolveControllerRef(oldSet.Namespace, oldControllerRef); ssg != nil {
			klog.V(4).Infof("StatefulSet %s(its old owner StatefulSetGrid %s) updated.", oldSet.Name, ssg.Name)
			ssgc.enqueueStatefulSetGrid(ssg)
		}
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		ssg := ssgc.resolveControllerRef(curSet.Namespace, curControllerRef)
		if ssg == nil {
			return
		}
		klog.V(4).Infof("StatefulSet %s(its owner StatefulSetGrid %s) updated.", curSet.Name, ssg.Name)
		ssgc.enqueueStatefulSetGrid(ssg)
		return
	}

	if !common.IsConcernedObject(curSet.ObjectMeta) {
		return
	}

	// Otherwise, it's an orphan. If anything changed, sync matching controllers
	// to see if anyone wants to adopt it now.
	labelChanged := !reflect.DeepEqual(curSet.Labels, oldSet.Labels)
	if labelChanged || controllerRefChanged {
		ssgs := ssgc.getGridForStatefulSet(curSet)
		for _, ssg := range ssgs {
			klog.V(4).Infof("Orphan StatefulSet %s(its possible owner StatefulSetGrid %s) updated.", curSet.Name, ssg.Name)
			ssgc.enqueueStatefulSetGrid(ssg)
		}
	}
}

func (ssgc *StatefulSetGridController) deleteStatefulSet(obj interface{}) {
	set, ok := obj.(*appsv1.StatefulSet)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		set, ok = tombstone.Obj.(*appsv1.StatefulSet)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object is not a statefulset %#v", obj))
			return
		}
	}
	if !common.IsConcernedObject(set.ObjectMeta) {
		return
	}
	controllerRef := metav1.GetControllerOf(set)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}
	ssg := ssgc.resolveControllerRef(set.Namespace, controllerRef)
	if ssg == nil {
		return
	}
	klog.V(4).Infof("StatefulSet %s(its owner StatefulSetGrid %s) deleted.", set.Name, ssg.Name)
	ssgc.enqueueStatefulSetGrid(ssg)
}

func (ssgc *StatefulSetGridController) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *crdv1.StatefulSetGrid {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != util.ControllerKind.Kind {
		return nil
	}
	ssg, err := ssgc.setGridLister.StatefulSetGrids(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}
	if ssg.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return ssg
}

func (ssgc *StatefulSetGridController) getGridForStatefulSet(set *appsv1.StatefulSet) []*crdv1.StatefulSetGrid {
	if len(set.Labels) == 0 {
		return nil
	}

	ssgList, err := ssgc.setGridLister.StatefulSetGrids(set.Namespace).List(labels.Everything())
	if err != nil {
		return nil
	}

	var statefulSetGrids []*crdv1.StatefulSetGrid
	for _, ssg := range ssgList {
		selector, err := common.GetDefaultSelector(ssg.Name)
		if err != nil {
			return nil
		}

		if !selector.Matches(labels.Set(set.Labels)) {
			continue
		}
		statefulSetGrids = append(statefulSetGrids, ssg)
	}

	if len(statefulSetGrids) > 1 {
		// ControllerRef will ensure we don't do anything crazy, but more than one
		// item in this list nevertheless constitutes user error.
		klog.V(4).Infof("user error! statefulset %s/%s with labels: %#v selects more than one statefulSetGrid, returning %#v",
			set.Namespace, set.Name, set.Labels, statefulSetGrids)
	}
	return statefulSetGrids
}
