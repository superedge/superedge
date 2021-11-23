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
	"context"
	"encoding/json"
	"fmt"
	sitev1 "github.com/superedge/superedge/pkg/site-manager/apis/site/v1"
	"github.com/superedge/superedge/pkg/site-manager/constant"
	crdClientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	"github.com/superedge/superedge/pkg/util"
	utilkube "github.com/superedge/superedge/pkg/util/kubeclient"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func (siteManager *SitesManagerDaemonController) addNode(obj interface{}) {
	node := obj.(*corev1.Node)
	if node.DeletionTimestamp != nil {
		siteManager.deleteNode(node)
		return
	}

	// 1. get all nodeunit
	allNodeUnit, err := siteManager.crdClient.SiteV1().NodeUnits().List(context.TODO(), metav1.ListOptions{})
	if err != nil && !errors.IsConflict(err) {
		klog.Errorf("List nodeUnit error: %#v", err)
		return
	}

	// 2.node match nodeunit
	nodeLabels := node.Labels
	for _, nodeunit := range allNodeUnit.Items {
		var matchNum int = 0
		nodeunitSelector := nodeunit.Spec.Selector
		for key, value := range nodeunitSelector.MatchLabels {
			labelsValue, ok := nodeLabels[key]
			if !ok || labelsValue != value {
				break
			}
			if ok || labelsValue == value {
				matchNum++
			}
		}

		if len(nodeunitSelector.MatchLabels) == matchNum {
			unitStatus := &nodeunit.Status
			if utilkube.IsReadyNode(node) {
				unitStatus.ReadyNodes = append(unitStatus.ReadyNodes, node.Name)
			} else {
				unitStatus.NotReadyNodes = append(unitStatus.NotReadyNodes, node.Name)
			}
			unitStatus.ReadyRate = NodeUitReadyRateAdd(&nodeunit)

			_, err = siteManager.crdClient.SiteV1().NodeUnits().UpdateStatus(context.TODO(), &nodeunit, metav1.UpdateOptions{})
			if err != nil && !errors.IsConflict(err) {
				klog.Errorf("Update nodeUnit: %s error: %#v", nodeunit.Name, err)
				return
			}

			if err := SetNodeUnitAnnotations(siteManager.kubeClient, node, &nodeunit); err != nil {
				klog.Errorf("Set nodeunit: %s annotations error: %#v", nodeunit.Name, err)
				continue
			}
		}
	}

	klog.V(1).Infof("Add node: %s to all match node-unit success.", node.Name)
}

func SetNodeUnitAnnotations(kubeclient clientset.Interface, node *corev1.Node, nodeUnit *sitev1.NodeUnit) error {
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}
	value, ok := node.Annotations[constant.NodeUnitSuperedge]
	if !ok {
		return fmt.Errorf("Get annotations: %v nil", constant.NodeUnitSuperedge)
	}

	var nodeUnits []string
	if err := json.Unmarshal([]byte(value), &nodeUnits); err != nil {
		return err
	}

	nodeUnits = append(nodeUnits, nodeUnit.Name)
	node.Annotations[constant.NodeUnitSuperedge] = util.ToJson(value)
	if _, err := kubeclient.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{}); err != nil {
		klog.Errorf("Update Node: %s, error: %#v", node.Name, err)
		return err
	}

	return nil
}

func (siteManager *SitesManagerDaemonController) updateNode(oldObj, newObj interface{}) {
	oldNode, curNode := oldObj.(*corev1.Node), newObj.(*corev1.Node)
	if curNode.ResourceVersion == oldNode.ResourceVersion {
		return
	}
	if utilkube.IsReadyNode(oldNode) == utilkube.IsReadyNode(curNode) {
		return
	}

	nodeUnits, err := GetNodeUnitByNode(siteManager.crdClient, curNode)
	if err != nil {
		klog.Errorf("Get nodeUnit by node, error： %#v", err)
		return
	}
	for _, nodeUnit := range nodeUnits {
		unitStatus := &nodeUnit.Status
		if utilkube.IsReadyNode(oldNode) {
			unitStatus.NotReadyNodes = util.DeleteSliceElement(unitStatus.NotReadyNodes, curNode.Name)
			unitStatus.ReadyNodes = append(unitStatus.ReadyNodes, curNode.Name)
		}
		if !utilkube.IsReadyNode(oldNode) {
			unitStatus.ReadyNodes = util.DeleteSliceElement(unitStatus.ReadyNodes, curNode.Name)
			unitStatus.NotReadyNodes = append(unitStatus.NotReadyNodes, curNode.Name)
		}
		unitStatus.ReadyRate = GetNodeUitReadyRate(&nodeUnit)

		_, err = siteManager.crdClient.SiteV1().NodeUnits().UpdateStatus(context.TODO(), &nodeUnit, metav1.UpdateOptions{})
		if err != nil && !errors.IsConflict(err) {
			klog.Errorf("Update nodeUnit: %s error: %#v", nodeUnit.Name, err)
			return
		}
		klog.V(6).Infof("Updated nodeUnit: %s success", nodeUnit.Name)
	}
	klog.V(4).Infof("Node: %s status update with update nodeUnitsuccess", curNode.Name)
}

func GetNodeUnitByNode(crdClient *crdClientset.Clientset, node *corev1.Node) (nodeUnits []sitev1.NodeUnit, err error) {
	allNodeUnit, err := crdClient.SiteV1().NodeUnits().List(context.TODO(), metav1.ListOptions{})
	if err != nil && !errors.IsConflict(err) {
		klog.Errorf("List nodeUnit error: %#v", err)
		return nil, err
	}

	for _, nodeunit := range allNodeUnit.Items {
		for _, nodeName := range nodeunit.Status.ReadyNodes {
			if nodeName == node.Name {
				nodeUnits = append(nodeUnits, nodeunit)
			}
		}
		for _, nodeName := range nodeunit.Status.NotReadyNodes {
			if nodeName == node.Name {
				nodeUnits = append(nodeUnits, nodeunit)
			}
		}
	}
	return
}

func (siteManager *SitesManagerDaemonController) deleteNode(obj interface{}) {
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
	sets := siteManager.getStatefulSetForNode(node)
	for _, set := range sets {
		klog.V(4).Infof("Node %s(its relevant StatefulSet %s) deleted.", node.Name, set.Name)
		//siteManager.enqueueStatefulSet(set)
	}
}

func (siteManager *SitesManagerDaemonController) getStatefulSetForNode(node *corev1.Node) []*appv1.StatefulSet {
	//selector, err := common.GetNodesSelector(node)
	//if err != nil {
	//	return nil
	//}
	//setList, err := siteManager.setLister.List(selector)
	//if err != nil {
	//	return nil
	//}
	var sets []*appv1.StatefulSet
	//for _, set := range setList {
	//	if rel, err := siteManager.IsConcernedStatefulSet(set); err != nil || !rel {
	//		continue
	//	}
	//	sets = append(sets, set)
	//}
	return sets
}
