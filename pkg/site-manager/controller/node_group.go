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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	sitev1 "github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha1"
	"github.com/superedge/superedge/pkg/site-manager/utils"
	"github.com/superedge/superedge/pkg/util"
)

func (siteManager *SitesManagerDaemonController) addNodeGroup(obj interface{}) {
	nodeGroup := obj.(*sitev1.NodeGroup)
	klog.V(4).Infof("Get Add nodeGroup: %s", util.ToJson(nodeGroup))
	if nodeGroup.DeletionTimestamp != nil {
		siteManager.deleteNodeGroup(nodeGroup) //todo
		return
	}

	if len(nodeGroup.Finalizers) == 0 {
		nodeGroup.Finalizers = append(nodeGroup.Finalizers, finalizerID)
	}

	if len(nodeGroup.Spec.AutoFindNodeKeys) > 0 {
		utils.AutoFindNodeKeysbyNodeGroup(siteManager.kubeClient, siteManager.crdClient, nodeGroup)
	}

	units, err := utils.GetUnitsByNodeGroup(siteManager.kubeClient, siteManager.crdClient, nodeGroup)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			units = []string{}
			klog.Warningf("Get NodeGroup: %s unit nil", nodeGroup.Name)
		} else {
			klog.Errorf("Get NodeGroup unit error: %v", err)
			return
		}
	}

	nodeGroup.Status.NodeUnits = units
	nodeGroup.Status.UnitNumber = len(units)
	_, err = siteManager.crdClient.SiteV1alpha1().NodeGroups().UpdateStatus(context.TODO(), nodeGroup, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("Update nodeGroup: %s error: %#v", nodeGroup.Name, err)
		return
	}

	klog.V(4).Infof("Add nodeGroup: %s success.", nodeGroup.Name)
}

func (siteManager *SitesManagerDaemonController) updateNodeGroup(oldObj, newObj interface{}) {
	oldNodeGroup := oldObj.(*sitev1.NodeGroup)
	curNodeGroup := newObj.(*sitev1.NodeGroup)
	klog.V(4).Infof("Get oldNodeGroup: %s, curNodeGroup: %s", util.ToJson(oldNodeGroup), util.ToJson(curNodeGroup))

	if len(curNodeGroup.Finalizers) == 0 {
		curNodeGroup.Finalizers = append(curNodeGroup.Finalizers, finalizerID)
	}

	if curNodeGroup.DeletionTimestamp != nil {
		siteManager.deleteNodeGroup(curNodeGroup) //todo
		return
	}

	if oldNodeGroup.ResourceVersion == curNodeGroup.ResourceVersion {
		return
	}

	if len(curNodeGroup.Spec.AutoFindNodeKeys) > 0 {
		utils.AutoFindNodeKeysbyNodeGroup(siteManager.kubeClient, siteManager.crdClient, curNodeGroup)
	}
	/*
		curNodeGroup
	*/

	units, err := utils.GetUnitsByNodeGroup(siteManager.kubeClient, siteManager.crdClient, curNodeGroup)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			units = []string{}
			klog.Warningf("Get nodeGroup: %s unit nil", curNodeGroup.Name)
		} else {
			klog.Errorf("Get NodeGroup unit error: %v", err)
			return
		}
	}

	curNodeGroup.Status.NodeUnits = units
	curNodeGroup.Status.UnitNumber = len(units)
	_, err = siteManager.crdClient.SiteV1alpha1().NodeGroups().UpdateStatus(context.TODO(), curNodeGroup, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("Update nodeGroup: %s error: %#v", curNodeGroup.Name, err)
		return
	}

	// todo: delete nodeGroup from oldNodeUnit And Add nodegroup annotations to curNodeUnit

	klog.V(4).Infof("Updated nodeGroup: %s success", curNodeGroup.Name)
}

func (siteManager *SitesManagerDaemonController) deleteNodeGroup(obj interface{}) {
	nodeGroup, ok := obj.(*sitev1.NodeGroup)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v\n", obj))
			return
		}
		nodeGroup, ok = tombstone.Obj.(*sitev1.NodeGroup)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object is not a nodeGroup %#v\n", obj))
			return
		}
	}

	// check all nodes, if which have the label with nodegroup name then remove
	for _, nu := range nodeGroup.Status.NodeUnits {

		obj, err := siteManager.crdClient.SiteV1alpha1().NodeUnits().Get(context.TODO(), nu, metav1.GetOptions{})
		if err != nil {
			klog.Error("List nodeunit fail ", err)
		}
		if obj.Spec.SetNode.Labels != nil {
			delete(obj.Spec.SetNode.Labels, nodeGroup.Name)
		}
		_, err = siteManager.crdClient.SiteV1alpha1().NodeUnits().Update(context.TODO(), obj, metav1.UpdateOptions{})
		if err != nil {
			klog.Error("Update nodeunit fail ", err)
		}
	}

	klog.V(4).Infof("Delete NodeGroup: %s succes.", nodeGroup.Name)
	return
}
