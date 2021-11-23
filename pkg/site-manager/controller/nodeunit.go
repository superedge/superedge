///*
//Copyright 2020 The SuperEdge Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//*/
//
package controller

import (
	"context"
	"fmt"
	sitev1 "github.com/superedge/superedge/pkg/site-manager/apis/site/v1"
	"github.com/superedge/superedge/pkg/util"
	utilkube "github.com/superedge/superedge/pkg/util/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func (siteManager *SitesManagerDaemonController) addNodeUnit(obj interface{}) {
	nodeUnit := obj.(*sitev1.NodeUnit)
	klog.V(4).Infof("Get Add nodeUnit: %s", util.ToJson(nodeUnit))
	if nodeUnit.DeletionTimestamp != nil {
		siteManager.deleteNodeUnit(nodeUnit) //todo
		return
	}

	readyNodes, notReadyNodes, err := GetNodeUnitNodes(siteManager.kubeClient, nodeUnit)
	if err != nil {
		klog.Errorf("Get NodeUnit Nodes error: %v", err)
		return
	}

	// todo: set node

	nodeUnitStatus := &nodeUnit.Status
	nodeUnitStatus.ReadyNodes = readyNodes
	nodeUnitStatus.ReadyRate = fmt.Sprintf("%d/%d", len(readyNodes), len(readyNodes)+len(notReadyNodes))
	nodeUnitStatus.NotReadyNodes = notReadyNodes

	_, err = siteManager.crdClient.SiteV1().NodeUnits().UpdateStatus(context.TODO(), nodeUnit, metav1.UpdateOptions{})
	if err != nil && !errors.IsConflict(err) {
		klog.Errorf("Update nodeUnit: %s error: %#v", nodeUnit.Name, err)
		return
	}

	klog.V(4).Infof("Add nodeUnit: %s success.", nodeUnit.Name)

	siteManager.enqueueNodeUnit(nodeUnit) //todo dele?
}

func (siteManager *SitesManagerDaemonController) updateNodeUnit(oldObj, newObj interface{}) {
	oldNodeUnit := oldObj.(*sitev1.NodeUnit)
	curNodeUnit := newObj.(*sitev1.NodeUnit)
	klog.V(4).Infof("Get oldNodeUnit: %s, curNodeUnit: %s", util.ToJson(oldNodeUnit), util.ToJson(curNodeUnit))

	if oldNodeUnit.ResourceVersion == curNodeUnit.ResourceVersion {
		return
	}

	readyNodes, notReadyNodes, err := GetNodeUnitNodes(siteManager.kubeClient, curNodeUnit)
	if err != nil {
		klog.Errorf("Get NodeUnit Nodes error: %v", err)
		return
	}

	// todo: set node

	nodeUnitStatus := &curNodeUnit.Status
	nodeUnitStatus.ReadyNodes = readyNodes
	nodeUnitStatus.ReadyRate = fmt.Sprintf("%d/%d", len(readyNodes), len(readyNodes)+len(notReadyNodes))
	nodeUnitStatus.NotReadyNodes = notReadyNodes

	curNodeUnit, err = siteManager.crdClient.SiteV1().NodeUnits().UpdateStatus(context.TODO(), curNodeUnit, metav1.UpdateOptions{})
	if err != nil && !errors.IsConflict(err) {
		klog.Errorf("Update nodeUnit: %s error: %#v", curNodeUnit.Name, err)
		return
	}
	klog.V(4).Infof("Updated nodeUnit: %s success", curNodeUnit.Name)

	siteManager.enqueueNodeUnit(curNodeUnit) //todo dele?
}

func (siteManager *SitesManagerDaemonController) deleteNodeUnit(obj interface{}) {
	nodeUnit, ok := obj.(*sitev1.NodeUnit)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v\n", obj))
			return
		}
		nodeUnit, ok = tombstone.Obj.(*sitev1.NodeUnit)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object is not a nodeUnit %#v\n", obj))
			return
		}
	}

	// todo: delete set node

	klog.V(4).Infof("Delete NodeUnit: %s succes.", nodeUnit.Name)
	siteManager.enqueueNodeUnit(nodeUnit) //todo dele?
}

func GetNodeUnitNodes(kubeclient clientset.Interface, nodeUnit *sitev1.NodeUnit) (readyNodes, notReadyNodes []string, err error) {
	selector := nodeUnit.Spec.Selector
	var nodes []corev1.Node

	// Get Nodes by selector
	if selector != nil {
		if len(selector.MatchLabels) > 0 || len(selector.MatchExpressions) > 0 {
			labelSelector := &metav1.LabelSelector{
				MatchLabels:      selector.MatchLabels,
				MatchExpressions: selector.MatchExpressions,
			}
			selector, err := metav1.LabelSelectorAsSelector(labelSelector)
			if err != nil {
				return readyNodes, notReadyNodes, err
			}
			listOptions := metav1.ListOptions{LabelSelector: selector.String()}
			nodeList, err := kubeclient.CoreV1().Nodes().List(context.TODO(), listOptions)
			if err != nil {
				klog.Errorf("Get nodes by selector, error: %v", err)
				return readyNodes, notReadyNodes, err
			}
			nodes = append(nodes, nodeList.Items...)
		}

		if len(selector.Annotations) > 0 { //todo: add Annotations selector

		}
	}

	// Get Nodes by nodeName
	nodeNames := nodeUnit.Spec.Nodes
	for _, nodeName := range nodeNames {
		node, err := kubeclient.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Get nodes by node name, error: %v", err)
			return readyNodes, notReadyNodes, err
		}
		nodes = append(nodes, *node)
	}

	readyNodes, notReadyNodes = utilkube.GetNodeStatus(nodes) // get all readynode and notReadyNodes
	return util.RemoveDuplicateElement(readyNodes), util.RemoveDuplicateElement(notReadyNodes), nil
}
