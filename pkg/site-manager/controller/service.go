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

//
//import (
//	"fmt"
//	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
//	"github.com/superedge/superedge/pkg/application-grid-controller/controller/service/util"
//	appv1 "k8s.io/api/apps/v1"
//	corev1 "k8s.io/api/core/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/labels"
//	"k8s.io/apimachinery/pkg/selection"
//	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
//	"k8s.io/client-go/tools/cache"
//	"k8s.io/klog/v2"
//	"reflect"
//)
//
//func (siteManager *SitesManagerDaemonController) addService(obj interface{}) {
//	svc := obj.(*corev1.Service)
//	if svc.DeletionTimestamp != nil {
//		// On a restart of the controller manager, it's possible for an object to
//		// show up in a state that is already pending deletion.
//		siteManager.deleteService(svc)
//		return
//	}
//	setList := siteManager.getStatefulSetForService(svc)
//	for _, set := range setList {
//		if rel, err := siteManager.IsConcernedStatefulSet(set); err != nil || !rel {
//			continue
//		}
//		klog.V(4).Infof("Service %s(its relevant statefulset %s) added.", svc.Name, set.Name)
//		siteManager.enqueueStatefulSet(set)
//	}
//}
//
//func (siteManager *SitesManagerDaemonController) updateService(oldObj, newObj interface{}) {
//	oldSvc := oldObj.(*corev1.Service)
//	curSvc := newObj.(*corev1.Service)
//	if curSvc.ResourceVersion == oldSvc.ResourceVersion {
//		// Periodic resync will send update events for all known Services.
//		// Two different versions of the same Service will always have different RVs.
//		return
//	}
//
//	labelChanged := !reflect.DeepEqual(curSvc.Labels, oldSvc.Labels)
//	if labelChanged {
//		if _, exist := oldSvc.Labels[common.GridSelectorUniqKeyName]; exist {
//			setList := siteManager.getStatefulSetForService(oldSvc)
//			for _, set := range setList {
//				if rel, err := siteManager.IsConcernedStatefulSet(set); err != nil || !rel {
//					continue
//				}
//				klog.V(4).Infof("Service %s(its old relevant statefulset %s) updated.", oldSvc.Name, set.Name)
//				siteManager.enqueueStatefulSet(set)
//			}
//		}
//	}
//
//	// If it has a ControllerRef, that's all that matters.
//	if _, exist := curSvc.Labels[common.GridSelectorUniqKeyName]; exist {
//		setList := siteManager.getStatefulSetForService(curSvc)
//		for _, set := range setList {
//			if rel, err := siteManager.IsConcernedStatefulSet(set); err != nil || !rel {
//				continue
//			}
//			klog.V(4).Infof("Service %s(its new relevant statefulset %s) updated.", curSvc.Name, set.Name)
//			siteManager.enqueueStatefulSet(set)
//		}
//	}
//	return
//}
//
//func (siteManager *SitesManagerDaemonController) deleteService(obj interface{}) {
//	svc, ok := obj.(*corev1.Service)
//	if !ok {
//		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
//		if !ok {
//			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
//			return
//		}
//		svc, ok = tombstone.Obj.(*corev1.Service)
//		if !ok {
//			utilruntime.HandleError(fmt.Errorf("Tombstone contained object is not a service %#v", obj))
//			return
//		}
//	}
//	setList := siteManager.getStatefulSetForService(svc)
//	for _, set := range setList {
//		if rel, err := siteManager.IsConcernedStatefulSet(set); err != nil || !rel {
//			continue
//		}
//		klog.V(4).Infof("Service %s(its relevant statefulset %s) deleted.", svc.Name, set.Name)
//		siteManager.enqueueStatefulSet(set)
//	}
//}
//
//func (siteManager *SitesManagerDaemonController) getStatefulSetForService(svc *corev1.Service) []*appv1.StatefulSet {
//	if svc.Labels == nil {
//		klog.V(4).Infof("Service %s no labels.", svc.Name)
//		return nil
//	}
//	controllerName, found := svc.Labels[common.GridSelectorName]
//	if !found {
//		klog.V(4).Infof("Service %s no GridSelectorName label.", svc.Name)
//		return nil
//	}
//	controllerRef := metav1.GetControllerOf(svc)
//	if controllerRef == nil || controllerRef.Name != controllerName || controllerRef.Kind != util.ControllerKind.Kind {
//		klog.V(4).Infof("Service %s no correct controller ownerReferences %v.", svc.Name, controllerRef)
//		return nil
//	}
//	gridSelectorUniqKey, found := svc.Labels[common.GridSelectorUniqKeyName]
//	if !found {
//		klog.V(4).Infof("Service %s no GridSelectorUniqKeyName label.", svc.Name)
//		return nil
//	}
//	requirement, err := labels.NewRequirement(common.GridSelectorUniqKeyName, selection.Equals, []string{gridSelectorUniqKey})
//	if err != nil {
//		klog.V(4).Infof("Service %s new labels requirement err %v.", svc.Name, err)
//		return nil
//	}
//	labelSelector := labels.NewSelector()
//	labelSelector = labelSelector.Add(*requirement)
//
//	setList, err := siteManager.setLister.StatefulSets(svc.Namespace).List(labelSelector)
//	if err != nil {
//		klog.V(4).Infof("List statefulset err %v", err)
//		return nil
//	}
//	var targetSetList []*appv1.StatefulSet
//	for _, set := range setList {
//		if set.Spec.ServiceName == svc.Name {
//			targetSetList = append(targetSetList, set)
//		}
//	}
//	return targetSetList
//}
