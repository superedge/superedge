/*
Copyright 2021 The SuperEdge Authors.

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
	"fmt"
	"github.com/superedge/superedge/pkg/site-manager/constant"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"strings"

	sitev1 "github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha1"
	"github.com/superedge/superedge/pkg/site-manager/utils"
	"github.com/superedge/superedge/pkg/util"
)

func (siteManager *SitesManagerDaemonController) addNodeUnit(obj interface{}) {
	nodeUnit := obj.(*sitev1.NodeUnit)
	klog.V(4).Infof("Get Add nodeUnit: %s", util.ToJson(nodeUnit))
	if nodeUnit.DeletionTimestamp != nil {
		siteManager.deleteNodeUnit(nodeUnit)
		return
	}

	// set NodeUnit name to node
	if nodeUnit.Spec.SetNode.Labels == nil {
		nodeUnit.Spec.SetNode.Labels = make(map[string]string)
	}
	nodeUnit.Spec.SetNode.Labels[nodeUnit.Name] = constant.NodeUnitSuperedge
	readyNodes, notReadyNodes, err := utils.GetNodesByUnit(siteManager.kubeClient, nodeUnit)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			readyNodes, notReadyNodes = []string{}, []string{}
			klog.Warningf("Get unit: %s node nil", nodeUnit.Name)
		} else {
			klog.Errorf("Get NodeUnit Nodes error: %v", err)
			return
		}
	}

	nodeUnitStatus := &nodeUnit.Status
	nodeUnitStatus.ReadyNodes = readyNodes
	nodeUnitStatus.ReadyRate = fmt.Sprintf("%d/%d", len(readyNodes), len(readyNodes)+len(notReadyNodes))
	nodeUnitStatus.NotReadyNodes = notReadyNodes
	nodeUnit, err = siteManager.crdClient.SiteV1alpha1().NodeUnits().Update(context.TODO(), nodeUnit, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("Update nodeUnit: %s error: %#v", nodeUnit.Name, err)
		return
	}
	/*
	 nodeGroup action
	*/
	nodeGroups, err := utils.UnitMatchNodeGroups(siteManager.kubeClient, siteManager.crdClient, nodeUnit.Name)
	if err != nil {
		klog.Errorf("Get NodeGroups error: %v", err)
		return
	}

	// Update nodegroups
	for _, nodeGroup := range nodeGroups {
		nodeGroupStatus := &nodeGroup.Status
		nodeGroupStatus.NodeUnits = append(nodeGroupStatus.NodeUnits, nodeUnit.Name)
		nodeGroupStatus.NodeUnits = util.RemoveDuplicateElement(nodeGroupStatus.NodeUnits)
		nodeGroupStatus.UnitNumber = len(nodeGroupStatus.NodeUnits)
		_, err = siteManager.crdClient.SiteV1alpha1().NodeGroups().UpdateStatus(context.TODO(), nodeGroup, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("Update nodeGroup: %s error: %#v", nodeGroup.Name, err)
			continue
		}
		nodeUnit.Spec.SetNode.Labels[nodeGroup.Name] = nodeUnit.Name
	}

	nodeNames := append(readyNodes, notReadyNodes...)
	utils.SetNodeToNodes(siteManager.kubeClient, nodeUnit.Spec.SetNode, nodeNames)

	klog.V(4).Infof("Add nodeUnit success: %s", util.ToJson(nodeUnit))
}

func (siteManager *SitesManagerDaemonController) updateNodeUnit(oldObj, newObj interface{}) {
	oldNodeUnit := oldObj.(*sitev1.NodeUnit)
	curNodeUnit := newObj.(*sitev1.NodeUnit)

	klog.V(4).Infof("Get oldNodeUnit: %s, curNodeUnit: %s", util.ToJson(oldNodeUnit), util.ToJson(curNodeUnit))
	if oldNodeUnit.ResourceVersion == curNodeUnit.ResourceVersion {
		return
	}
	/*
		curNodeUnit
	*/
	readyNodes, notReadyNodes, err := utils.GetNodesByUnit(siteManager.kubeClient, curNodeUnit)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			readyNodes, notReadyNodes = []string{}, []string{}
			klog.Warningf("Get unit: %s node nil", curNodeUnit.Name)
		} else {
			klog.Errorf("Get NodeUnit Nodes error: %v", err)
			return
		}
	}

	nodeUnitStatus := &curNodeUnit.Status
	nodeUnitStatus.ReadyNodes = readyNodes
	nodeUnitStatus.ReadyRate = fmt.Sprintf("%d/%d", len(readyNodes), len(readyNodes)+len(notReadyNodes))
	nodeUnitStatus.NotReadyNodes = notReadyNodes
	curNodeUnit, err = siteManager.crdClient.SiteV1alpha1().NodeUnits().Update(context.TODO(), curNodeUnit, metav1.UpdateOptions{})
	if err != nil && !errors.IsConflict(err) {
		klog.Errorf("Update nodeUnit: %s error: %#v", curNodeUnit.Name, err)
		return
	}

	// new node add or old node remove
	nodeNamesCur := append(readyNodes, notReadyNodes...)
	nodeNamesOld := append(oldNodeUnit.Status.ReadyNodes, oldNodeUnit.Status.NotReadyNodes...)
	removeNodes, updateNodes := utils.NeedUpdateNode(nodeNamesOld, nodeNamesCur)
	klog.V(4).Infof("NodeUnit: %s need remove nodes: %s setNode: %s", oldNodeUnit.Name, util.ToJson(removeNodes), util.ToJson(oldNodeUnit.Spec.SetNode))
	if err := utils.RemoveSetNode(siteManager.kubeClient, oldNodeUnit, removeNodes); err != nil {
		klog.Errorf("Remove node NodeUnit setNode error: %v", err)
		return
	}
	/*
	   nodeGroup action
	*/
	nodeGroups, err := utils.UnitMatchNodeGroups(siteManager.kubeClient, siteManager.crdClient, curNodeUnit.Name)
	if err != nil {
		klog.Errorf("Get NodeGroups error: %v", err)
		return
	}

	// Update nodegroups
	setNode := &curNodeUnit.Spec.SetNode
	if setNode.Labels == nil {
		setNode.Labels = make(map[string]string)
	}
	for _, nodeGroup := range nodeGroups {
		nodeGroupStatus := &nodeGroup.Status
		nodeGroupStatus.NodeUnits = append(nodeGroupStatus.NodeUnits, curNodeUnit.Name)
		nodeGroupStatus.NodeUnits = util.RemoveDuplicateElement(nodeGroupStatus.NodeUnits)
		nodeGroupStatus.UnitNumber = len(nodeGroupStatus.NodeUnits)
		_, err = siteManager.crdClient.SiteV1alpha1().NodeGroups().UpdateStatus(context.TODO(), nodeGroup, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("Update nodeGroup: %s error: %#v", nodeGroup.Name, err)
			continue
		}
		setNode.Labels[nodeGroup.Name] = curNodeUnit.Name
	}

	//todo setNode.Labels need update
	curNodeUnit, err = siteManager.crdClient.SiteV1alpha1().NodeUnits().Update(context.TODO(), curNodeUnit, metav1.UpdateOptions{})
	if err != nil && !errors.IsConflict(err) {
		klog.Errorf("Update nodeUnit: %s error: %#v", curNodeUnit.Name, err)
		return
	}
	utils.UpdtateNodeFromSetNode(siteManager.kubeClient, oldNodeUnit.Spec.SetNode, curNodeUnit.Spec.SetNode, updateNodes)

	klog.V(4).Infof("Add nodeUnit success: %s", util.ToJson(curNodeUnit))
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

	/*
	 nodeGroup action
	*/
	nodeGroups, err := utils.GetNodeGroupsByUnit(siteManager.crdClient, nodeUnit.Name)
	if err != nil {
		klog.Errorf("Get NodeGroups error: %v", err)
		return
	}

	// Update nodegroups
	for _, nodeGroup := range nodeGroups {
		nodeGroupStatus := &nodeGroup.Status
		nodeGroupStatus.NodeUnits = util.DeleteSliceElement(nodeGroupStatus.NodeUnits, nodeUnit.Name)
		nodeGroupStatus.UnitNumber = len(nodeGroupStatus.NodeUnits)
		_, err = siteManager.crdClient.SiteV1alpha1().NodeGroups().UpdateStatus(context.TODO(), nodeGroup, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("Update nodeGroup: %s error: %#v", nodeGroup.Name, err)
			continue
		}
	}

	readyNodes, notReadyNodes, err := utils.GetNodesByUnit(siteManager.kubeClient, nodeUnit)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			klog.Warningf("Get unit: %s node nil", nodeUnit.Name)
			return
		} else {
			klog.Errorf("Get NodeUnit Nodes error: %v", err)
			return
		}
	}
	nodeNames := append(readyNodes, notReadyNodes...)
	utils.DeleteNodesFromSetNode(siteManager.kubeClient, nodeUnit.Spec.SetNode, nodeNames)

	klog.V(4).Infof("Delete NodeUnit: %s succes.", nodeUnit.Name)
}
