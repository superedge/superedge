package controller

import (
	"fmt"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"reflect"
)

func (setc *StatefulSetController) addNode(obj interface{}) {
	node := obj.(*corev1.Node)
	if node.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		setc.deleteNode(node)
		return
	}

	sets := setc.getStatefulSetForNode(node)
	if len(sets) == 0 {
		return
	}
	klog.V(4).Infof("Node %s added.", node.Name)
	for _, set := range sets {
		setc.enqueueStatefulset(set)
	}
}

func (setc *StatefulSetController) updateNode(oldObj, newObj interface{}) {
	oldNode := oldObj.(*corev1.Node)
	curNode := newObj.(*corev1.Node)
	labelChanged := !reflect.DeepEqual(curNode.Labels, oldNode.Labels)
	// Only handles nodes whose label has changed.
	if labelChanged {
		sets := setc.getStatefulSetForNode(curNode)
		if len(sets) == 0 {
			return
		}
		klog.V(4).Infof("Node %s updated.", curNode.Name)
		for _, set := range sets {
			setc.enqueueStatefulset(set)
		}
	}
}

func (setc *StatefulSetController) deleteNode(obj interface{}) {
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
	sets := setc.getStatefulSetForNode(node)
	if len(sets) == 0 {
		return
	}
	klog.V(4).Infof("Node %s deleted.", node.Name)
	for _, set := range sets {
		setc.enqueueStatefulset(set)
	}
}

func (setc *StatefulSetController) getStatefulSetForNode(node *corev1.Node) []*appv1.StatefulSet {
	if len(node.Labels) == 0 {
		return nil
	}
	setList, err := setc.setLister.StatefulSets("").List(labels.Everything())
	if err != nil {
		return nil
	}
	var sets []*appv1.StatefulSet
	for _, set := range setList {
		for k, v := range node.Labels {
			if val, ok := set.Spec.Template.Spec.NodeSelector[k]; ok && v == val {
				sets = append(sets, set)
				break
			}
		}
	}
	if len(sets) == 0 {
		return nil
	}
	return sets
}
