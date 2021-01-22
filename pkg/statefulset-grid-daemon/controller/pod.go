package controller

import (
	"fmt"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"reflect"
)

func (setc *StatefulSetController) addPod(obj interface{}) {
	pod := obj.(*corev1.Pod)
	if pod.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		setc.deletePod(pod)
		return
	}

	set := setc.getStatefulSetForPod(pod)
	if set == nil {
		return
	}

	if rel, err := setc.IsConcernedStatefulSet(set); err != nil || !rel {
		return
	}

	setc.enqueueStatefulset(set)
}

func (setc *StatefulSetController) updatePod(oldObj, newObj interface{}) {
	oldPod := oldObj.(*corev1.Pod)
	curPod := newObj.(*corev1.Pod)

	set := setc.getStatefulSetForPod(curPod)
	if set == nil {
		return
	}

	if rel, err := setc.IsConcernedStatefulSet(set); err != nil || !rel {
		return
	}

	curControllerRef := metav1.GetControllerOf(curPod)
	oldControllerRef := metav1.GetControllerOf(oldPod)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged && oldControllerRef != nil {
		// The ControllerRef was changed. Sync the old controller, if any.
		if ss := setc.getStatefulSetForPod(oldPod); ss != nil {
			setc.enqueueStatefulset(ss)
		}
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		ss := setc.getStatefulSetForPod(curPod)
		if ss == nil {
			return
		}
		klog.V(4).Infof("statefulset %s updated.", curPod.Name)
		setc.enqueueStatefulset(ss)
		return
	}
}

func (setc *StatefulSetController) deletePod(obj interface{}) {
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

	set := setc.getStatefulSetForPod(pod)
	if set == nil {
		return
	}

	if rel, err := setc.IsConcernedStatefulSet(set); err != nil || !rel {
		return
	}

	controllerRef := metav1.GetControllerOf(pod)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}

	klog.V(4).Infof("pod %s deleted.", pod.Name)
	setc.enqueueStatefulset(set)
}

func (setc *StatefulSetController) getStatefulSetForPod(pod *corev1.Pod) *appv1.StatefulSet {
	ownerRef := metav1.GetControllerOf(pod)
	if ownerRef != nil {
		if ownerRef.Kind != controllerKind.Kind {
			return nil
		}

		set, err := setc.setLister.StatefulSets(pod.Namespace).Get(ownerRef.Name)
		if err != nil {
			klog.Errorf("get %s StatefulSets err %v", ownerRef.Name, err)
			return nil
		}
		if set.UID != ownerRef.UID {
			// The controller we found with this Name is not the same one that the
			// ControllerRef points to.
			return nil
		}
		return set
	}
	return nil
}
