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
	appsv1 "github.com/superedge/superedge/pkg/apps-manager/apis/apps/v1"
	"github.com/superedge/superedge/pkg/apps-manager/constant"
	"github.com/superedge/superedge/pkg/apps-manager/utils"

	"github.com/superedge/superedge/pkg/util"
	"k8s.io/klog/v2"
)

func (appsManager *SitesManagerController) addEDeploy(obj interface{}) {
	edeploy := obj.(*appsv1.EDeployment)
	klog.V(4).Infof("Get Add edge deploy: %s", util.ToJson(edeploy))

	if edeploy.DeletionTimestamp != nil {
		appsManager.deleteNodeUnit(edeploy) //todo
		return
	}

	selectedNodes, err := utils.SchedulableNode(appsManager.kubeClient, edeploy)
	if err != nil {
		klog.Errorf("Edeploy: %s selecter node error: %#v", edeploy.Name, err)
		return
	}
	if len(selectedNodes) == 0 {
		klog.V(1).Infof("Edeploy: %s selecter node nil", edeploy.Name)
		return
	}

	dstNode := appsManager.hostName
	for _, node := range selectedNodes {
		if node.Name == dstNode {
			if err := utils.WriteEdeployToStaticPod(appsManager.kubeClient, edeploy, constant.KubeManifestsDir); err != nil {
				klog.Errorf("Write edeploy: %s to static pod， error: %#v", edeploy.Name, err)
				return
			}
			return
		}
	}
	klog.V(4).Infof("Add EDeploy nodeUnit: %s success.", edeploy.Name)
}

func (appsManager *SitesManagerController) updateEDeploy(oldObj, newObj interface{}) {
	oldEDeployment := oldObj.(*appsv1.EDeployment)
	curEDeployment := newObj.(*appsv1.EDeployment)
	klog.V(4).Infof("Get Update edge deploy: %s, edge deploy: %s", util.ToJson(oldEDeployment), util.ToJson(curEDeployment))

	if oldEDeployment.ResourceVersion == curEDeployment.ResourceVersion {
		return
	}
	// 1. 是否符合自己的node // todo: 每个node只处理自己的
	selectedNodes, err := utils.SchedulableNode(appsManager.kubeClient, curEDeployment)
	if err != nil {
		klog.Errorf("Edeploy: %s selecter node error: %#v", curEDeployment.Name, err)
		return
	}
	if len(selectedNodes) == 0 {
		klog.V(1).Infof("Edeploy: %s selecter node nil", curEDeployment.Name)
		return
	}

	dstNode := appsManager.hostName
	for _, node := range selectedNodes {
		if node.Name == dstNode {
			if err := utils.UpdateEdeployToStaticPod(appsManager.kubeClient, oldEDeployment, curEDeployment, constant.KubeManifestsDir); err != nil {
				klog.Errorf("Update edeploy: %s to static pod， error: %#v", curEDeployment.Name, err)
				return
			}
			return
		}
	}
	klog.V(4).Infof("update EDeploy nodeUnit: %s success.", curEDeployment.Name)
}

func (appsManager *SitesManagerController) deleteEDeploy(obj interface{}) {
	edeploy := obj.(*appsv1.EDeployment)
	if edeploy.DeletionTimestamp != nil {
		appsManager.deleteNodeUnit(edeploy) //todo
		return
	}

	selectedNodes, err := utils.SchedulableNode(appsManager.kubeClient, edeploy)
	if err != nil {
		klog.Errorf("Edeploy: %s selecter node error: %#v", edeploy.Name, err)
		return
	}
	if len(selectedNodes) == 0 {
		klog.V(1).Infof("Edeploy: %s selecter node nil", edeploy.Name)
		return
	}

	dstNode := appsManager.hostName
	for _, node := range selectedNodes {
		if node.Name == dstNode {
			// 2. edploy to node yaml
			if err := utils.DeleteStaticPodFromEdeploy(appsManager.kubeClient, edeploy, constant.KubeManifestsDir); err != nil {
				klog.Errorf("Delete edeploy: %s to static pod， error: %#v", edeploy.Name, err)
				return
			}
			return
		}
	}
	klog.V(4).Infof("delete EDeploy nodeUnit: %s success.", edeploy.Name)
}

func (appsManager *SitesManagerController) updateNodeUnit(oldObj, newObj interface{}) {
	//oldNodeUnit := oldObj.(*sitev1.NodeUnit)
	//curNodeUnit := newObj.(*sitev1.NodeUnit)
	//klog.V(4).Infof("Get oldNodeUnit: %s, curNodeUnit: %s", util.ToJson(oldNodeUnit), util.ToJson(curNodeUnit))
	//
	//if oldNodeUnit.ResourceVersion == curNodeUnit.ResourceVersion {
	//	return
	//}
	//
	///*
	//	oldNodeUnit
	//*/
	//nodeNames := append(oldNodeUnit.Status.ReadyNodes, oldNodeUnit.Status.NotReadyNodes...)
	//utils.RemoveNodesAnnotations(siteManager.kubeClient, nodeNames, []string{oldNodeUnit.Name})
	//
	///*
	//	curNodeUnit
	//*/
	//readyNodes, notReadyNodes, err := utils.GetNodesByUnit(siteManager.kubeClient, curNodeUnit)
	//if err != nil {
	//	if strings.Contains(err.Error(), "not found") {
	//		readyNodes, notReadyNodes = []string{}, []string{}
	//		klog.Warningf("Get unit: %s node nil", curNodeUnit.Name)
	//	} else {
	//		klog.Errorf("Get NodeUnit Nodes error: %v", err)
	//		return
	//	}
	//}
	//
	//// todo: set node
	//
	//nodeUnitStatus := &curNodeUnit.Status
	//nodeUnitStatus.ReadyNodes = readyNodes
	//nodeUnitStatus.ReadyRate = fmt.Sprintf("%d/%d", len(readyNodes), len(readyNodes)+len(notReadyNodes))
	//nodeUnitStatus.NotReadyNodes = notReadyNodes
	//curNodeUnit, err = siteManager.crdClient.SiteV1().NodeUnits().UpdateStatus(context.TODO(), curNodeUnit, metav1.UpdateOptions{})
	//if err != nil && !errors.IsConflict(err) {
	//	klog.Errorf("Update nodeUnit: %s error: %#v", curNodeUnit.Name, err)
	//	return
	//}
	//
	//nodeNames = append(readyNodes, notReadyNodes...)
	//if err := utils.AddNodesAnnotations(siteManager.kubeClient, nodeNames, []string{curNodeUnit.Name}); err != nil {
	//	klog.Errorf("Add nodes annotations: %s, error: %#v", curNodeUnit.Name, err)
	//	return
	//}
	//
	///*
	// nodeGroup action
	//*/
	//nodeGroups, err := utils.UnitMatchNodeGroups(siteManager.crdClient, curNodeUnit.Name)
	//if err != nil {
	//	klog.Errorf("Get NodeGroups error: %v", err)
	//	return
	//}
	//
	//// Update nodegroups
	//for _, nodeGroup := range nodeGroups {
	//	nodeGroupStatus := &nodeGroup.Status
	//	nodeGroupStatus.NodeUnits = append(nodeGroupStatus.NodeUnits, curNodeUnit.Name)
	//	nodeGroupStatus.NodeUnits = util.RemoveDuplicateElement(nodeGroupStatus.NodeUnits)
	//	nodeGroupStatus.UnitNumber = len(nodeGroupStatus.NodeUnits)
	//	_, err = siteManager.crdClient.SiteV1().NodeGroups().UpdateStatus(context.TODO(), nodeGroup, metav1.UpdateOptions{})
	//	if err != nil {
	//		klog.Errorf("Update nodeGroup: %s error: %#v", nodeGroup.Name, err)
	//		continue
	//	}
	//}

	// klog.V(4).Infof("Updated nodeUnit: %s success", curNodeUnit.Name)
}

func (appsManager *SitesManagerController) deleteNodeUnit(obj interface{}) {
	//nodeUnit, ok := obj.(*sitev1.NodeUnit)
	//if !ok {
	//	tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
	//	if !ok {
	//		utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v\n", obj))
	//		return
	//	}
	//	nodeUnit, ok = tombstone.Obj.(*sitev1.NodeUnit)
	//	if !ok {
	//		utilruntime.HandleError(fmt.Errorf("Tombstone contained object is not a nodeUnit %#v\n", obj))
	//		return
	//	}
	//}
	//
	///*
	// node action
	//*/
	//// todo: delete set node
	//
	//readyNodes, notReadyNodes, err := utils.GetNodesByUnit(siteManager.kubeClient, nodeUnit)
	//if err != nil {
	//	if strings.Contains(err.Error(), "not found") {
	//		klog.Warningf("Get unit: %s node nil", nodeUnit.Name)
	//		return
	//	} else {
	//		klog.Errorf("Get NodeUnit Nodes error: %v", err)
	//		return
	//	}
	//}
	//nodeNames := append(readyNodes, notReadyNodes...)
	//utils.RemoveNodesAnnotations(siteManager.kubeClient, nodeNames, []string{nodeUnit.Name})
	//
	///*
	// nodeGroup action
	//*/
	//nodeGroups, err := utils.GetNodeGroupsByUnit(siteManager.crdClient, nodeUnit.Name)
	//if err != nil {
	//	klog.Errorf("Get NodeGroups error: %v", err)
	//	return
	//}
	//
	//// Update nodegroups
	//for _, nodeGroup := range nodeGroups {
	//	nodeGroupStatus := &nodeGroup.Status
	//	nodeGroupStatus.NodeUnits = util.DeleteSliceElement(nodeGroupStatus.NodeUnits, nodeUnit.Name)
	//	nodeGroupStatus.UnitNumber = len(nodeGroupStatus.NodeUnits)
	//	_, err = siteManager.crdClient.SiteV1().NodeGroups().UpdateStatus(context.TODO(), nodeGroup, metav1.UpdateOptions{})
	//	if err != nil {
	//		klog.Errorf("Update nodeGroup: %s error: %#v", nodeGroup.Name, err)
	//		continue
	//	}
	//}

	// klog.V(4).Infof("Delete NodeUnit: %s succes.", nodeUnit.Name)
}
