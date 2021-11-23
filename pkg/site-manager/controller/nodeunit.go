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
	"github.com/superedge/superedge/pkg/site-manager/utils"
	"github.com/superedge/superedge/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
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

	readyNodes, notReadyNodes, err := utils.GetNodeUnitNodes(siteManager.kubeClient, nodeUnit)
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

	readyNodes, notReadyNodes, err := utils.GetNodeUnitNodes(siteManager.kubeClient, curNodeUnit)
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

	readyNodes, notReadyNodes, err := utils.GetNodeUnitNodes(siteManager.kubeClient, nodeUnit)
	if err != nil {
		klog.Errorf("Get NodeUnit Nodes error: %v", err)
		return
	}
	nodeNames := append(readyNodes, notReadyNodes...)
	for _, nodeName := range nodeNames {
		node, err := siteManager.kubeClient.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Get Node: %s, error: %#v", nodeName, err)
			continue
		}

		// todo: 删除有错 不能更新
		if err := utils.RemoveNodeUnitAnnotations(siteManager.kubeClient, node, []string{nodeUnit.Name}); err != nil {
			klog.Errorf("Remove node: %s annotations nodeunit: %s flags error: %#v", nodeName, nodeUnit.Name, err)
			continue
		}
	}

	klog.V(4).Infof("Delete NodeUnit: %s succes.", nodeUnit.Name)
}
