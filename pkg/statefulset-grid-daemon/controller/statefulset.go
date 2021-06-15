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
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/statefulset/util"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func (ssgdc *StatefulSetGridDaemonController) addStatefulSet(obj interface{}) {
	set := obj.(*appsv1.StatefulSet)

	if set.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		ssgdc.deleteStatefulSet(set)
		return
	}

	if rel, err := ssgdc.IsConcernedStatefulSet(set); err != nil || !rel {
		return
	}
	klog.V(4).Infof("StatefulSet %s added.", set.Name)
	ssgdc.enqueueStatefulSet(set)
}

func (ssgdc *StatefulSetGridDaemonController) updateStatefulSet(oldObj, newObj interface{}) {
	oldSet := oldObj.(*appsv1.StatefulSet)
	curSet := newObj.(*appsv1.StatefulSet)
	if curSet.ResourceVersion == oldSet.ResourceVersion {
		// Periodic resync will send update events for all known StatefulSet.
		// Two different versions of the same StatefulSet will always have different RVs.
		return
	}

	if rel, err := ssgdc.IsConcernedStatefulSet(curSet); err != nil || !rel {
		return
	}
	klog.V(4).Infof("StatefulSet %s updated.", curSet.Name)
	ssgdc.enqueueStatefulSet(curSet)
}

func (ssgdc *StatefulSetGridDaemonController) deleteStatefulSet(obj interface{}) {
	set, ok := obj.(*appsv1.StatefulSet)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		set, ok = tombstone.Obj.(*appsv1.StatefulSet)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object is not a StatefulSet %#v", obj))
			return
		}
	}

	if rel, err := ssgdc.IsConcernedStatefulSet(set); err != nil || !rel {
		return
	}

	klog.V(4).Infof("StatefulSet %s deleted.", set.Name)
	ssgdc.enqueueStatefulSet(set)
}

func (ssgdc *StatefulSetGridDaemonController) IsConcernedStatefulSet(set *appsv1.StatefulSet) (bool, error) {
	// Check statefulset controllerRef
	controllerRef := metav1.GetControllerOf(set)
	if controllerRef == nil || controllerRef.Kind != util.ControllerKind.Kind {
		// Never care about statefulset orphans
		return false, nil
	}
	// Check consistency of statefulset and never care about inconsistent ones
	// Check GridSelectorName labels consistency
	if set.ObjectMeta.Labels == nil {
		return false, nil
	}
	controllerName, found := set.ObjectMeta.Labels[common.GridSelectorName]
	if !found || controllerName != controllerRef.Name {
		return false, nil
	}
	// Check GridSelectorUniqKeyName labels consistency
	gridUniqKeyName, found := set.ObjectMeta.Labels[common.GridSelectorUniqKeyName]
	if !found {
		return false, nil
	}
	if ssg, err := ssgdc.setGridLister.StatefulSetGrids(set.Namespace).Get(controllerRef.Name); err == nil {
		if ssg.Spec.GridUniqKey != gridUniqKeyName {
			return false, nil
		}
		if controllerRef.UID != ssg.UID {
			// The controller we found with this Name is not the same one that the
			// ControllerRef points to.
			return false, nil
		}
	} else if errors.IsNotFound(err) {
		klog.V(4).Infof("StatefulSet %s relevant owner statefulset grid %s not found.", set.Name, controllerRef.Name)
	} else {
		klog.Errorf("Get statefulset grid %s err %v", controllerRef.Name, err)
		return false, err
	}

	// Never care about statefulset that does not has service name
	if set.Spec.ServiceName == "" {
		return false, nil
	}

	// Check NodeSelector consistency
	node, err := ssgdc.nodeLister.Get(ssgdc.hostName)
	if err != nil {
		klog.Errorf("Get host node %s err %v", ssgdc.hostName, err)
		return false, err
	}
	nodeGridValue, exist := node.Labels[gridUniqKeyName]
	if !exist {
		return false, nil
	}
	if setGridValue, exist := set.Spec.Template.Spec.NodeSelector[gridUniqKeyName]; !exist || !(setGridValue == nodeGridValue) {
		return false, nil
	}
	return true, nil
}
