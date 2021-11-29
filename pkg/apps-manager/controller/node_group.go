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

//import (
//	"context"
//	"fmt"
//	sitev1 "github.com/superedge/superedge/pkg/apps-manager/apis/site/v1"
//	"github.com/superedge/superedge/pkg/apps-manager/utils"
//	"github.com/superedge/superedge/pkg/util"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
//	"k8s.io/client-go/tools/cache"
//	"k8s.io/klog/v2"
//	"strings"
//)
//
//func (siteManager *SitesManagerDaemonController) addNodeGroup(obj interface{}) {
//	nodeGroup := obj.(*sitev1.NodeGroup)
//	klog.V(4).Infof("Get Add nodeGroup: %s", util.ToJson(nodeGroup))
//	if nodeGroup.DeletionTimestamp != nil {
//		siteManager.deleteNodeGroup(nodeGroup) //todo
//		return
//	}
//
//	units, err := utils.GetUnitsByNodeGroup(siteManager.crdClient, nodeGroup)
//	if err != nil {
//		if strings.Contains(err.Error(), "not found") {
//			units = []string{}
//			klog.Warningf("Get NodeGroup: %s unit nil", nodeGroup.Name)
//		} else {
//			klog.Errorf("Get NodeGroup unit error: %v", err)
//			return
//		}
//	}
//
//	// todo: set unit
//
//	nodeGroupStatus := &nodeGroup.Status
//	nodeGroupStatus.NodeUnits = units
//	nodeGroupStatus.UnitNumber = len(units)
//	_, err = siteManager.crdClient.SiteV1().NodeGroups().UpdateStatus(context.TODO(), nodeGroup, metav1.UpdateOptions{})
//	if err != nil {
//		klog.Errorf("Update nodeGroup: %s error: %#v", nodeGroup.Name, err)
//		return
//	}
//
//	//todo: Add nodegroup annotations to unit
//
//	klog.V(4).Infof("Add nodeGroup: %s success.", nodeGroup.Name)
//}
//
//func (siteManager *SitesManagerDaemonController) updateNodeGroup(oldObj, newObj interface{}) {
//	oldNodeGroup := oldObj.(*sitev1.NodeGroup)
//	curNodeGroup := newObj.(*sitev1.NodeGroup)
//	klog.V(4).Infof("Get oldNodeGroup: %s, curNodeGroup: %s", util.ToJson(oldNodeGroup), util.ToJson(curNodeGroup))
//
//	if oldNodeGroup.ResourceVersion == curNodeGroup.ResourceVersion {
//		return
//	}
//
//	/*
//		curNodeGroup
//	*/
//	units, err := utils.GetUnitsByNodeGroup(siteManager.crdClient, curNodeGroup)
//	if err != nil {
//		if strings.Contains(err.Error(), "not found") {
//			units = []string{}
//			klog.Warningf("Get nodeGroup: %s unit nil", curNodeGroup.Name)
//		} else {
//			klog.Errorf("Get NodeGroup unit error: %v", err)
//			return
//		}
//	}
//
//	// todo: set unit
//
//	nodeGroupStatus := &curNodeGroup.Status
//	nodeGroupStatus.NodeUnits = units
//	nodeGroupStatus.UnitNumber = len(units)
//	_, err = siteManager.crdClient.SiteV1().NodeGroups().UpdateStatus(context.TODO(), curNodeGroup, metav1.UpdateOptions{})
//	if err != nil {
//		klog.Errorf("Update nodeGroup: %s error: %#v", curNodeGroup.Name, err)
//		return
//	}
//
//	// todo: delete nodeGroup from oldNodeUnit And Add nodegroup annotations to curNodeUnit
//
//	klog.V(4).Infof("Updated nodeGroup: %s success", curNodeGroup.Name)
//}
//
//func (siteManager *SitesManagerDaemonController) deleteNodeGroup(obj interface{}) {
//	nodeGroup, ok := obj.(*sitev1.NodeGroup)
//	if !ok {
//		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
//		if !ok {
//			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v\n", obj))
//			return
//		}
//		nodeGroup, ok = tombstone.Obj.(*sitev1.NodeGroup)
//		if !ok {
//			utilruntime.HandleError(fmt.Errorf("Tombstone contained object is not a nodeGroup %#v\n", obj))
//			return
//		}
//	}
//
//	// todo: delete set unit
//
//	// todo: delete nodegroup annotations set in unit
//
//	klog.V(4).Infof("Delete NodeGroup: %s succes.", nodeGroup.Name)
//}
