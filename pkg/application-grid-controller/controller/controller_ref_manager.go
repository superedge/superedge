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
	"k8s.io/klog"
)

type BaseControllerRefManager struct {
	Controller metav1.Object
	Selector   labels.Selector

	canAdoptErr  error
	canAdoptOnce sync.Once
	CanAdoptFunc func() error
}

type DeploymentControllerRefManager struct {
	BaseControllerRefManager
	controllerKind schema.GroupVersionKind
	dpControl      DPControlInterface
}

func NewDeploymentControllerRefManager(
	dpControl DPControlInterface,
	controller metav1.Object,
	selector labels.Selector,
	controllerKind schema.GroupVersionKind,
	canAdopt func() error,
) *DeploymentControllerRefManager {
	return &DeploymentControllerRefManager{
		BaseControllerRefManager: BaseControllerRefManager{
			Controller:   controller,
			Selector:     selector,
			CanAdoptFunc: canAdopt,
		},
		controllerKind: controllerKind,
		dpControl:      dpControl,
	}
}

func (m *DeploymentControllerRefManager) ClaimDeployment(sets []*appsv1.Deployment) ([]*appsv1.Deployment, error) {
	var claimed []*appsv1.Deployment
	var errlist []error

	match := func(obj metav1.Object) bool {
		return m.Selector.Matches(labels.Set(obj.GetLabels()))
	}
	adopt := func(obj metav1.Object) error {
		return m.AdoptDeployment(obj.(*appsv1.Deployment))
	}
	release := func(obj metav1.Object) error {
		return m.ReleaseDeployment(obj.(*appsv1.Deployment))
	}

	for _, dp := range sets {
		ok, err := m.ClaimObject(dp, match, adopt, release)
		if err != nil {
			errlist = append(errlist, err)
			continue
		}
		if ok {
			claimed = append(claimed, dp)
		}
	}
	return claimed, utilerrors.NewAggregate(errlist)
}

func (m *DeploymentControllerRefManager) AdoptDeployment(dp *appsv1.Deployment) error {
	if err := m.CanAdopt(); err != nil {
		return fmt.Errorf("can't adopt Deployment %v/%v (%v): %v", dp.Namespace, dp.Name, dp.UID, err)
	}
	addControllerPatch := fmt.Sprintf(
		`{"metadata":{"ownerReferences":[{"apiVersion":"%s","kind":"%s","name":"%s","uid":"%s","controller":true,"blockOwnerDeletion":true}],"uid":"%s"}}`,
		m.controllerKind.GroupVersion(), m.controllerKind.Kind,
		m.Controller.GetName(), m.Controller.GetUID(), dp.UID)
	return m.dpControl.PatchDeployment(dp.Namespace, dp.Name, []byte(addControllerPatch))
}

func (m *DeploymentControllerRefManager) CanAdopt() error {
	m.canAdoptOnce.Do(func() {
		if m.CanAdoptFunc != nil {
			m.canAdoptErr = m.CanAdoptFunc()
		}
	})
	return m.canAdoptErr
}

func (m *DeploymentControllerRefManager) ReleaseDeployment(dp *appsv1.Deployment) error {
	klog.V(2).Infof("Patching Deployment %s_%s to remove its controllerRef to %s/%s:%s",
		dp.Namespace, dp.Name, m.controllerKind.GroupVersion(), m.controllerKind.Kind, m.Controller.GetName())
	deleteOwnerRefPatch := fmt.Sprintf(
		`{"metadata":{"ownerReferences":[{"$patch":"delete","uid":"%s"}],"uid":"%s"}}`,
		m.Controller.GetUID(), dp.UID)
	err := m.dpControl.PatchDeployment(dp.Namespace, dp.Name, []byte(deleteOwnerRefPatch))
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

func (m *BaseControllerRefManager) ClaimObject(obj metav1.Object, match func(metav1.Object) bool, adopt, release func(metav1.Object) error) (bool, error) {
	controllerRef := metav1.GetControllerOf(obj)
	if controllerRef != nil {
		if controllerRef.UID != m.Controller.GetUID() {
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
		if m.Controller.GetDeletionTimestamp() != nil {
			return false, nil
		}
		if err := release(obj); err != nil {
			// If the pod no longer exists, ignore the error.
			if errors.IsNotFound(err) {
				return false, nil
			}
			// Either someone else released it, or there was a transient error.
			// The controller should requeue and try again if it's still stale.
			return false, err
		}
		// Successfully released.
		return false, nil
	}

	// It's an orphan.
	if m.Controller.GetDeletionTimestamp() != nil || !match(obj) {
		// Ignore if we're being deleted or selector doesn't match.
		return false, nil
	}
	if obj.GetDeletionTimestamp() != nil {
		// Ignore if the object is being deleted
		return false, nil
	}
	// Selector matches. Try to adopt.
	if err := adopt(obj); err != nil {
		// If the pod no longer exists, ignore the error.
		if errors.IsNotFound(err) {
			return false, nil
		}
		// Either someone else claimed it first, or there was a transient error.
		// The controller should requeue and try again if it's still orphaned.
		return false, err
	}
	// Successfully adopted.
	return true, nil
}

type ServiceControllerRefManager struct {
	BaseControllerRefManager
	controllerKind schema.GroupVersionKind
	svcControl     SVCControlInterface
}

func NewServiceControllerRefManager(
	svcControl SVCControlInterface,
	controller metav1.Object,
	selector labels.Selector,
	controllerKind schema.GroupVersionKind,
	canAdopt func() error,
) *ServiceControllerRefManager {
	return &ServiceControllerRefManager{
		BaseControllerRefManager: BaseControllerRefManager{
			Controller:   controller,
			Selector:     selector,
			CanAdoptFunc: canAdopt,
		},
		controllerKind: controllerKind,
		svcControl:     svcControl,
	}
}

func (m *ServiceControllerRefManager) ClaimService(sets []*corev1.Service) ([]*corev1.Service, error) {
	var claimed []*corev1.Service
	var errlist []error

	match := func(obj metav1.Object) bool {
		return m.Selector.Matches(labels.Set(obj.GetLabels()))
	}
	adopt := func(obj metav1.Object) error {
		return m.AdoptService(obj.(*corev1.Service))
	}
	release := func(obj metav1.Object) error {
		return m.ReleaseService(obj.(*corev1.Service))
	}

	for _, service := range sets {
		ok, err := m.ClaimObject(service, match, adopt, release)
		if err != nil {
			errlist = append(errlist, err)
			continue
		}
		if ok {
			claimed = append(claimed, service)
		}
	}
	return claimed, utilerrors.NewAggregate(errlist)
}

func (m *ServiceControllerRefManager) AdoptService(svc *corev1.Service) error {
	if err := m.CanAdopt(); err != nil {
		return fmt.Errorf("can't adopt Service %v/%v (%v): %v", svc.Namespace, svc.Name, svc.UID, err)
	}
	addControllerPatch := fmt.Sprintf(
		`{"metadata":{"ownerReferences":[{"apiVersion":"%s","kind":"%s","name":"%s","uid":"%s","controller":true,"blockOwnerDeletion":true}],"uid":"%s"}}`,
		m.controllerKind.GroupVersion(), m.controllerKind.Kind,
		m.Controller.GetName(), m.Controller.GetUID(), svc.UID)
	return m.svcControl.PatchService(svc.Namespace, svc.Name, []byte(addControllerPatch))
}

func (m *ServiceControllerRefManager) CanAdopt() error {
	m.canAdoptOnce.Do(func() {
		if m.CanAdoptFunc != nil {
			m.canAdoptErr = m.CanAdoptFunc()
		}
	})
	return m.canAdoptErr
}

func (m *ServiceControllerRefManager) ReleaseService(svc *corev1.Service) error {
	klog.V(2).Infof("Patching Service %s_%s to remove its controllerRef to %s/%s:%s",
		svc.Namespace, svc.Name, m.controllerKind.GroupVersion(), m.controllerKind.Kind, m.Controller.GetName())
	deleteOwnerRefPatch := fmt.Sprintf(
		`{"metadata":{"ownerReferences":[{"$patch":"delete","uid":"%s"}],"uid":"%s"}}`,
		m.Controller.GetUID(), svc.UID)
	err := m.svcControl.PatchService(svc.Namespace, svc.Name, []byte(deleteOwnerRefPatch))
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
