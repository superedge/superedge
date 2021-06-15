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
	"fmt"
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
)

type baseControllerRefManager struct {
	controller metav1.Object
	selector   labels.Selector

	canAdoptErr  error
	canAdoptOnce sync.Once
	canAdoptFunc func() error
}

// DeploymentControllerRefManager is responsible for claiming deployments
type DeploymentControllerRefManager struct {
	baseControllerRefManager
	controllerKind schema.GroupVersionKind
	dpClient       DeployClientInterface
}

func NewDeploymentControllerRefManager(
	dpClient DeployClientInterface,
	controller metav1.Object,
	selector labels.Selector,
	controllerKind schema.GroupVersionKind,
	canAdopt func() error,
) *DeploymentControllerRefManager {
	return &DeploymentControllerRefManager{
		baseControllerRefManager: baseControllerRefManager{
			controller:   controller,
			selector:     selector,
			canAdoptFunc: canAdopt,
		},
		controllerKind: controllerKind,
		dpClient:       dpClient,
	}
}

func (dcrm *DeploymentControllerRefManager) ClaimDeployment(dpList []*appsv1.Deployment) ([]*appsv1.Deployment, error) {
	var claimed []*appsv1.Deployment
	var errList []error

	match := func(obj metav1.Object) bool {
		return dcrm.selector.Matches(labels.Set(obj.GetLabels()))
	}
	adopt := func(obj metav1.Object) error {
		return dcrm.adoptDeployment(obj.(*appsv1.Deployment))
	}
	release := func(obj metav1.Object) error {
		return dcrm.releaseDeployment(obj.(*appsv1.Deployment))
	}

	for _, dp := range dpList {
		ok, err := dcrm.claimObject(dp, match, adopt, release)
		if err != nil {
			errList = append(errList, err)
			continue
		}
		if ok {
			claimed = append(claimed, dp)
		}
	}
	return claimed, utilerrors.NewAggregate(errList)
}

func (dcrm *DeploymentControllerRefManager) adoptDeployment(dp *appsv1.Deployment) error {
	if err := dcrm.canAdopt(); err != nil {
		return fmt.Errorf("can't adopt Deployment %s/%s (%v): %v", dp.Namespace, dp.Name, dp.UID, err)
	}
	klog.V(2).Infof("Adopting deployment %s/%s to controllerRef %s/%s:%s",
		dp.Namespace, dp.Name, dcrm.controllerKind.GroupVersion(), dcrm.controllerKind.Kind, dcrm.controller.GetName())
	addControllerPatch := fmt.Sprintf(
		`{"metadata":{"ownerReferences":[{"apiVersion":"%s","kind":"%s","name":"%s","uid":"%s","controller":true,"blockOwnerDeletion":true}],"uid":"%s"}}`,
		dcrm.controllerKind.GroupVersion(), dcrm.controllerKind.Kind,
		dcrm.controller.GetName(), dcrm.controller.GetUID(), dp.UID)
	return dcrm.dpClient.PatchDeployment(dp.Namespace, dp.Name, []byte(addControllerPatch))
}

func (dcrm *DeploymentControllerRefManager) releaseDeployment(dp *appsv1.Deployment) error {
	klog.V(2).Infof("Deleting deployment %s/%s as it owned by controllerRef %s/%s:%s but selector %#v doesn't match",
		dp.Namespace, dp.Name, dcrm.controllerKind.GroupVersion(), dcrm.controllerKind.Kind, dcrm.controller.GetName(), dp.Labels)
	err := dcrm.dpClient.DeleteDeployment(dp.Namespace, dp.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		if errors.IsInvalid(err) {
			return nil
		}
	}
	return err
}

// StatefulSetControllerRefManager is responsible for claiming statefulsets
type StatefulSetControllerRefManager struct {
	baseControllerRefManager
	controllerKind schema.GroupVersionKind
	setClient      SetClientInterface
}

func NewStatefulSetControllerRefManager(
	setClient SetClientInterface,
	controller metav1.Object,
	selector labels.Selector,
	controllerKind schema.GroupVersionKind,
	canAdopt func() error,
) *StatefulSetControllerRefManager {
	return &StatefulSetControllerRefManager{
		baseControllerRefManager: baseControllerRefManager{
			controller:   controller,
			selector:     selector,
			canAdoptFunc: canAdopt,
		},
		controllerKind: controllerKind,
		setClient:      setClient,
	}
}

func (sscrm *StatefulSetControllerRefManager) ClaimStatefulSet(setList []*appsv1.StatefulSet) ([]*appsv1.StatefulSet, error) {
	var claimedSets []*appsv1.StatefulSet
	var errList []error

	match := func(obj metav1.Object) bool {
		return sscrm.selector.Matches(labels.Set(obj.GetLabels()))
	}
	adopt := func(obj metav1.Object) error {
		return sscrm.adoptStatefulSet(obj.(*appsv1.StatefulSet))
	}
	release := func(obj metav1.Object) error {
		return sscrm.releaseStatefulSet(obj.(*appsv1.StatefulSet))
	}

	for _, set := range setList {
		ok, err := sscrm.claimObject(set, match, adopt, release)
		if err != nil {
			errList = append(errList, err)
			continue
		}
		if ok {
			claimedSets = append(claimedSets, set)
		}
	}
	return claimedSets, utilerrors.NewAggregate(errList)
}

func (sscrm *StatefulSetControllerRefManager) adoptStatefulSet(set *appsv1.StatefulSet) error {
	if err := sscrm.canAdopt(); err != nil {
		return fmt.Errorf("Can't adopt statefulset %s/%s (%v): %v", set.Namespace, set.Name, set.UID, err)
	}
	klog.V(2).Infof("Adopting statefulset %s/%s to controllerRef %s/%s:%s",
		set.Namespace, set.Name, sscrm.controllerKind.GroupVersion(), sscrm.controllerKind.Kind, sscrm.controller.GetName())
	addControllerOwnerRefsPatch := fmt.Sprintf(
		`{"metadata":{"ownerReferences":[{"apiVersion":"%s","kind":"%s","name":"%s","uid":"%s","controller":true,"blockOwnerDeletion":true}],"uid":"%s"}}`,
		sscrm.controllerKind.GroupVersion(), sscrm.controllerKind.Kind,
		sscrm.controller.GetName(), sscrm.controller.GetUID(), set.UID)
	return sscrm.setClient.PatchStatefulSet(set.Namespace, set.Name, []byte(addControllerOwnerRefsPatch))
}

func (sscrm *StatefulSetControllerRefManager) releaseStatefulSet(set *appsv1.StatefulSet) error {
	klog.V(2).Infof("Deleting statefulset %s/%s as it owned by controllerRef %s/%s:%s but selector %#v doesn't match",
		set.Namespace, set.Name, sscrm.controllerKind.GroupVersion(), sscrm.controllerKind.Kind, sscrm.controller.GetName(), set.Labels)
	err := sscrm.setClient.DeleteStatefulSet(set.Namespace, set.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		if errors.IsInvalid(err) {
			return nil
		}
	}
	return err
}

// ServiceControllerRefManager is responsible for claiming services
type ServiceControllerRefManager struct {
	baseControllerRefManager
	controllerKind schema.GroupVersionKind
	svcClient      SvcClientInterface
}

func NewServiceControllerRefManager(
	svcClient SvcClientInterface,
	controller metav1.Object,
	selector labels.Selector,
	controllerKind schema.GroupVersionKind,
	canAdopt func() error,
) *ServiceControllerRefManager {
	return &ServiceControllerRefManager{
		baseControllerRefManager: baseControllerRefManager{
			controller:   controller,
			selector:     selector,
			canAdoptFunc: canAdopt,
		},
		controllerKind: controllerKind,
		svcClient:      svcClient,
	}
}

func (scrm *ServiceControllerRefManager) ClaimService(svcList []*corev1.Service) ([]*corev1.Service, error) {
	var claimed []*corev1.Service
	var errList []error

	match := func(obj metav1.Object) bool {
		return scrm.selector.Matches(labels.Set(obj.GetLabels()))
	}
	adopt := func(obj metav1.Object) error {
		return scrm.adoptService(obj.(*corev1.Service))
	}
	release := func(obj metav1.Object) error {
		return scrm.releaseService(obj.(*corev1.Service))
	}

	for _, svc := range svcList {
		ok, err := scrm.claimObject(svc, match, adopt, release)
		if err != nil {
			errList = append(errList, err)
			continue
		}
		if ok {
			claimed = append(claimed, svc)
		}
	}
	return claimed, utilerrors.NewAggregate(errList)
}

func (scrm *ServiceControllerRefManager) adoptService(svc *corev1.Service) error {
	if err := scrm.canAdopt(); err != nil {
		return fmt.Errorf("can't adopt Service %s/%s (%v): %v", svc.Namespace, svc.Name, svc.UID, err)
	}
	klog.V(2).Infof("Adopting service %s/%s to controllerRef %s/%s:%s",
		svc.Namespace, svc.Name, scrm.controllerKind.GroupVersion(), scrm.controllerKind.Kind, scrm.controller.GetName())
	addControllerOwnerRefsPatch := fmt.Sprintf(
		`{"metadata":{"ownerReferences":[{"apiVersion":"%s","kind":"%s","name":"%s","uid":"%s","controller":true,"blockOwnerDeletion":true}],"uid":"%s"}}`,
		scrm.controllerKind.GroupVersion(), scrm.controllerKind.Kind,
		scrm.controller.GetName(), scrm.controller.GetUID(), svc.UID)
	return scrm.svcClient.PatchService(svc.Namespace, svc.Name, []byte(addControllerOwnerRefsPatch))
}

func (scrm *ServiceControllerRefManager) releaseService(svc *corev1.Service) error {
	klog.V(2).Infof("Deleting service %s/%s as it owned by controllerRef %s/%s:%s but selector %#v doesn't match",
		svc.Namespace, svc.Name, scrm.controllerKind.GroupVersion(), scrm.controllerKind.Kind, scrm.controller.GetName(), svc.Labels)
	err := scrm.svcClient.DeleteService(svc.Namespace, svc.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		if errors.IsInvalid(err) {
			return nil
		}
	}
	return err
}

// claimObject claims receiving object according to its labels and controllerRef
func (bcrm *baseControllerRefManager) claimObject(obj metav1.Object, match func(metav1.Object) bool, adopt, release func(metav1.Object) error) (bool, error) {
	controllerRef := metav1.GetControllerOf(obj)
	if controllerRef != nil {
		if controllerRef.UID != bcrm.controller.GetUID() {
			// Owned by someone else. Ignore.
			return false, nil
		}
		if match(obj) {
			// We already own it and the selector matches.
			// Return true (successfully claimed) before checking deletion timestamp.
			// We're still allowed to claim things we already own while being deleted
			// because doing so requires taking no actions.
			return true, nil
		}
		// Owned by us but selector doesn't match.
		// Try to release, unless we're being deleted.
		if bcrm.controller.GetDeletionTimestamp() != nil {
			return false, nil
		}
		if obj.GetDeletionTimestamp() != nil {
			// Ignore if the object is being deleted
			return false, nil
		}
		if err := release(obj); err != nil {
			// If the object no longer exists, ignore the error.
			if errors.IsNotFound(err) {
				return false, nil
			}
			// Either someone else released it first, or there was a transient error.
			// The controller should requeue and try again if it's still stale.
			return false, err
		}
		// Successfully released.
		return false, nil
	}

	// It's an orphan.
	if bcrm.controller.GetDeletionTimestamp() != nil || !match(obj) {
		// Ignore if we're being deleted or selector doesn't match.
		return false, nil
	}
	if obj.GetDeletionTimestamp() != nil {
		// Ignore if the object is being deleted
		return false, nil
	}
	// Selector matches. Try to adopt.
	if err := adopt(obj); err != nil {
		// If the object no longer exists, ignore the error.
		if errors.IsNotFound(err) {
			return false, nil
		}
		// Either someone else adopted it first, or there was a transient error.
		// The controller should requeue and try again if it's still orphaned.
		return false, err
	}
	// Successfully adopted.
	return true, nil
}

func (bcrm *baseControllerRefManager) canAdopt() error {
	bcrm.canAdoptOnce.Do(func() {
		if bcrm.canAdoptFunc != nil {
			bcrm.canAdoptErr = bcrm.canAdoptFunc()
		}
	})
	return bcrm.canAdoptErr
}
