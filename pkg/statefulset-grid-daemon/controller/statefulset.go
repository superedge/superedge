package controller

import (
	"fmt"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

func (setc *StatefulSetController) addStatefulset(obj interface{}) {
	set := obj.(*appsv1.StatefulSet)


	if rel, err := setc.IsConcernedStatefulSet(set); err != nil || !rel {
		return
	}

	if set.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		setc.deleteStatefulset(set)
		return
	}

	setc.enqueueStatefulset(set)
}

func (setc *StatefulSetController) updateStatefulset(oldObj, newObj interface{}) {
	//oldS := oldObj.(*appsv1.StatefulSet)
	curS := newObj.(*appsv1.StatefulSet)

	if rel, err := setc.IsConcernedStatefulSet(curS); err != nil || !rel {
		return
	}

	setc.enqueueStatefulset(curS)

}

func (setc *StatefulSetController) deleteStatefulset(obj interface{}) {
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

	if rel, err := setc.IsConcernedStatefulSet(set); err != nil || !rel {
		return
	}

	controllerRef := metav1.GetControllerOf(set)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}

	klog.V(4).Infof("statefulSet %s deleted.", set.Name)
	setc.enqueueStatefulset(set)
}

func HasServiceName(set *appsv1.StatefulSet) bool {
	return !(set.Spec.ServiceName == "")
}

func (setc *StatefulSetController)IsConcernedStatefulSet(set *appsv1.StatefulSet) (bool, error) {
	if set.ObjectMeta.Labels == nil {
		return false, nil
	}

	//1. has GridSelectorName in label for ss
	_, found := set.ObjectMeta.Labels[common.GridSelectorName]
	if !found {
		return false, nil
	}

	//2. has serviceName in spec for ss
	if !HasServiceName(set) {
		return false, nil
	}

	//3. nodeselector related with hostNode
	node, err := setc.nodeLister.Get(setc.hostName)
	if err != nil {
		klog.Errorf("get host node err %v", err)
		return false, err
	}

	releated := false
	for k,v := range node.Labels{
		if val, ok := set.Spec.Template.Spec.NodeSelector[k]; ok && v == val{
			releated = true
			break
		}
	}
	if !releated {
		return false, nil
	}

	return true, nil
}


