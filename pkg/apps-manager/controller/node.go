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

//import (
//	"context"
//	"fmt"
//	"github.com/superedge/superedge/pkg/apps-manager/utils"
//	"github.com/superedge/superedge/pkg/util"
//	utilkube "github.com/superedge/superedge/pkg/util/kubeclient"
//	corev1 "k8s.io/api/core/v1"
//	"k8s.io/apimachinery/pkg/api/errors"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
//	"k8s.io/client-go/tools/cache"
//	"k8s.io/klog/v2"
//)
//
//func (siteManager *SitesManagerDaemonController) addNode(obj interface{}) {
//	node := obj.(*corev1.Node)
//	if node.DeletionTimestamp != nil {
//		siteManager.deleteNode(node)
//		return
//	}
//
//	// set node role
//	if err := utils.SetNodeRole(siteManager.kubeClient, node); err != nil {
//		klog.Errorf("Set node: %s role error: %#v", err)
//	}
//
//	// 1. get all nodeunit
//	allNodeUnit, err := siteManager.crdClient.SiteV1().NodeUnits().List(context.TODO(), metav1.ListOptions{})
//	if err != nil && !errors.IsConflict(err) {
//		klog.Errorf("List nodeUnit error: %#v", err)
//		return
//	}
//
//	// 2.node match nodeunit
//	nodeLabels := node.Labels
//	var needUpdateNodeUnit []string
//	for _, nodeunit := range allNodeUnit.Items {
//		var matchNum int = 0
//		nodeunitSelector := nodeunit.Spec.Selector
//		for key, value := range nodeunitSelector.MatchLabels { //todo: MatchExpressions && Annotations
//			labelsValue, ok := nodeLabels[key]
//			if !ok || labelsValue != value {
//				break
//			}
//			if ok || labelsValue == value {
//				matchNum++
//			}
//		}
//
//		if len(nodeunitSelector.MatchLabels) == matchNum {
//			unitStatus := &nodeunit.Status
//			if utilkube.IsReadyNode(node) {
//				unitStatus.ReadyNodes = append(unitStatus.ReadyNodes, node.Name)
//			} else {
//				unitStatus.NotReadyNodes = append(unitStatus.NotReadyNodes, node.Name)
//			}
//			unitStatus.ReadyRate = utils.AddNodeUitReadyRate(&nodeunit)
//
//			_, err = siteManager.crdClient.SiteV1().NodeUnits().UpdateStatus(context.TODO(), &nodeunit, metav1.UpdateOptions{})
//			if err != nil && !errors.IsConflict(err) {
//				klog.Errorf("Update nodeUnit: %s error: %#v", nodeunit.Name, err)
//				return
//			}
//			needUpdateNodeUnit = append(needUpdateNodeUnit, nodeunit.Name)
//		}
//	}
//
//	if err := utils.AddNodesAnnotations(siteManager.kubeClient, []string{node.Name}, needUpdateNodeUnit); err != nil {
//		klog.Errorf("Set node: %s annotations error: %#v", node.Name, err)
//		return
//	}
//
//	klog.V(1).Infof("Add node: %s to all match node-unit success.", node.Name)
//}
//
//func (siteManager *SitesManagerDaemonController) updateNode(oldObj, newObj interface{}) {
//	oldNode, curNode := oldObj.(*corev1.Node), newObj.(*corev1.Node)
//	if curNode.ResourceVersion == oldNode.ResourceVersion {
//		return
//	}
//	if utilkube.IsReadyNode(oldNode) == utilkube.IsReadyNode(curNode) {
//		return
//	}
//
//	// set node role
//	if err := utils.SetNodeRole(siteManager.kubeClient, curNode); err != nil {
//		klog.Errorf("Set node: %s role error: %#v", err)
//	}
//
//	nodeUnits, err := utils.GetUnitsByNode(siteManager.crdClient, curNode)
//	if err != nil {
//		klog.Errorf("Get nodeUnit by node, error： %#v", err)
//		return
//	}
//
//	/*
//	 only node status
//	*/
//	for _, nodeUnit := range nodeUnits {
//		unitStatus := &nodeUnit.Status
//		if utilkube.IsReadyNode(oldNode) {
//			unitStatus.NotReadyNodes = util.DeleteSliceElement(unitStatus.NotReadyNodes, curNode.Name)
//			unitStatus.ReadyNodes = append(unitStatus.ReadyNodes, curNode.Name)
//		}
//		if !utilkube.IsReadyNode(oldNode) {
//			unitStatus.ReadyNodes = util.DeleteSliceElement(unitStatus.ReadyNodes, curNode.Name)
//			unitStatus.NotReadyNodes = append(unitStatus.NotReadyNodes, curNode.Name)
//		}
//		unitStatus.ReadyRate = utils.GetNodeUitReadyRate(&nodeUnit)
//
//		_, err = siteManager.crdClient.SiteV1().NodeUnits().UpdateStatus(context.TODO(), &nodeUnit, metav1.UpdateOptions{})
//		if err != nil && !errors.IsConflict(err) {
//			klog.Errorf("Update nodeUnit: %s error: %#v", nodeUnit.Name, err)
//			return
//		}
//		klog.V(6).Infof("Updated nodeUnit: %s success", nodeUnit.Name)
//	}
//	/*
//		todo: if update node annotations, such as nodeunit annotations deleted
//	*/
//
//	klog.V(4).Infof("Node: %s status update with update nodeUnit success", curNode.Name)
//}
//
//func (siteManager *SitesManagerDaemonController) deleteNode(obj interface{}) {
//	node, ok := obj.(*corev1.Node)
//	if !ok {
//		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
//		if !ok {
//			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v\n", obj))
//			return
//		}
//		node, ok = tombstone.Obj.(*corev1.Node)
//		if !ok {
//			utilruntime.HandleError(fmt.Errorf("Tombstone contained object is not a node %#v\n", obj))
//			return
//		}
//	}
//
//	nodeUnits, err := utils.GetUnitsByNode(siteManager.crdClient, node)
//	if err != nil {
//		klog.Errorf("Get nodeUnit by node, error： %#v", err)
//		return
//	}
//
//	for _, nodeUnit := range nodeUnits {
//		unitStatus := &nodeUnit.Status
//		unitStatus.ReadyNodes = util.DeleteSliceElement(unitStatus.ReadyNodes, node.Name)
//		unitStatus.NotReadyNodes = util.DeleteSliceElement(unitStatus.NotReadyNodes, node.Name)
//		unitStatus.ReadyRate = utils.GetNodeUitReadyRate(&nodeUnit)
//
//		_, err = siteManager.crdClient.SiteV1().NodeUnits().UpdateStatus(context.TODO(), &nodeUnit, metav1.UpdateOptions{})
//		if err != nil && !errors.IsConflict(err) {
//			klog.Errorf("Update nodeUnit: %s error: %#v", nodeUnit.Name, err)
//			return
//		}
//		klog.V(6).Infof("Updated nodeUnit: %s success", nodeUnit.Name)
//	}
//	klog.V(1).Infof("Delete Node: %s Update nodeUnit success", node.Name)
//}
