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
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"reflect"
)

func (ssgdc *StatefulSetGridDaemonController) addPod(obj interface{}) {
	pod := obj.(*corev1.Pod)
	if pod.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		ssgdc.deletePod(pod)
		return
	}

	set := ssgdc.getStatefulSetForPod(pod)
	if set == nil {
		return
	}

	if rel, err := ssgdc.IsConcernedStatefulSet(set); err != nil || !rel {
		return
	}
	klog.V(4).Infof("Pod %s(its owner statefulset %s) added.", pod.Name, set.Name)
	ssgdc.enqueueStatefulSet(set)
}

func (ssgdc *StatefulSetGridDaemonController) updatePod(oldObj, newObj interface{}) {
	oldPod := oldObj.(*corev1.Pod)
	curPod := newObj.(*corev1.Pod)
	if curPod.ResourceVersion == oldPod.ResourceVersion {
		// Periodic resync will send update events for all known Pods.
		// Two different versions of the same Pod will always have different RVs.
		return
	}

	curControllerRef := metav1.GetControllerOf(curPod)
	oldControllerRef := metav1.GetControllerOf(oldPod)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged && oldControllerRef != nil {
		// The ControllerRef was changed. Sync the old controller, if any.
		if set := ssgdc.getStatefulSetForPod(oldPod); set != nil {
			if rel, err := ssgdc.IsConcernedStatefulSet(set); err == nil && rel {
				klog.V(4).Infof("Pod %s(its old owner statefulset %s) updated.", oldPod.Name, set.Name)
				ssgdc.enqueueStatefulSet(set)
			}
		}
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		if set := ssgdc.getStatefulSetForPod(curPod); set != nil {
			if rel, err := ssgdc.IsConcernedStatefulSet(set); err == nil && rel {
				klog.V(4).Infof("Pod %s(its owner statefulset %s) updated.", curPod.Name, set.Name)
				ssgdc.enqueueStatefulSet(set)
			}
		}
	}
	return
}

func (ssgdc *StatefulSetGridDaemonController) deletePod(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		pod, ok = tombstone.Obj.(*corev1.Pod)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object is not a pod %#v", obj))
			return
		}
	}

	set := ssgdc.getStatefulSetForPod(pod)
	if set == nil {
		return
	}

	if rel, err := ssgdc.IsConcernedStatefulSet(set); err != nil || !rel {
		return
	}

	klog.V(4).Infof("Pod %s(its owner statefulset %s) deleted.", pod.Name, set.Name)
	ssgdc.enqueueStatefulSet(set)
}

func (ssgdc *StatefulSetGridDaemonController) getStatefulSetForPod(pod *corev1.Pod) *appv1.StatefulSet {
	controllerRef := metav1.GetControllerOf(pod)
	if controllerRef != nil {
		if controllerRef.Kind != controllerKind.Kind {
			return nil
		}

		set, err := ssgdc.setLister.StatefulSets(pod.Namespace).Get(controllerRef.Name)
		if err != nil {
			klog.Errorf("get %s StatefulSets err %v", controllerRef.Name, err)
			return nil
		}
		if set.UID != controllerRef.UID {
			// The controller we found with this Name is not the same one that the
			// ControllerRef points to.
			return nil
		}
		return set
	}
	return nil
}
