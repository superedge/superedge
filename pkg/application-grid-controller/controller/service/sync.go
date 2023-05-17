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

package service

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	commonutil "github.com/superedge/superedge/pkg/application-grid-controller/util"
	"k8s.io/apimachinery/pkg/api/errors"
	klog "k8s.io/klog/v2"

	"github.com/hashicorp/go-multierror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes/scheme"

	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/service/util"
)

func (sgc *ServiceGridController) reconcile(g *crdv1.ServiceGrid, svcList []*corev1.Service) error {
	var (
		adds    []*corev1.Service
		updates []*corev1.Service
		deletes []*corev1.Service
	)

	sgTargetSvcName := util.GetServiceName(g)
	isExistingSvc := false
	for _, svc := range svcList {
		if svc.Name == sgTargetSvcName {
			isExistingSvc = true

			SvcToUpdate := util.CreateService(g)

			scheme := scheme.Scheme
			scheme.Default(SvcToUpdate)

			if !commonutil.DeepContains(svc.Spec, SvcToUpdate.Spec) {
				klog.Infof("svc %s template changed", svc.Name)
				out, _ := json.Marshal(SvcToUpdate.Spec)
				klog.V(5).Infof("svcToUpdate is %s", string(out))
				out, _ = json.Marshal(svc.Spec)
				klog.V(5).Infof("svc is %s", string(out))
				updates = append(updates, SvcToUpdate)
				continue
			}
		} else {
			deletes = append(deletes, svc)
		}
	}

	if !isExistingSvc {
		adds = append(adds, util.CreateService(g))
	}

	return sgc.syncService(g, adds, updates, deletes)
}

func (sgc *ServiceGridController) syncService(sg *crdv1.ServiceGrid, adds, updates, deletes []*corev1.Service) error {
	wg := sync.WaitGroup{}
	totalSize := len(adds) + len(updates) + len(deletes)
	wg.Add(totalSize)
	errCh := make(chan error, totalSize)

	for i := range adds {
		go func(svc *corev1.Service) {
			defer wg.Done()
			klog.V(4).Infof("Creating service %s/%s by syncService", svc.Namespace, svc.Name)
			sgCopy := sg.DeepCopy()
			_, err := sgc.kubeClient.CoreV1().Services(svc.Namespace).Create(context.TODO(), svc, metav1.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				errCh <- err
				sgCopy.Status.Conditions = []metav1.Condition{{
					Type:               common.CreateError,
					Status:             metav1.ConditionTrue,
					Message:            err.Error(),
					LastTransitionTime: metav1.NewTime(time.Now()),
					Reason:             common.CreateError,
				}}

				sgc.eventRecorder.Eventf(sg, corev1.EventTypeWarning, common.CreateError, err.Error())

				_, err := sgc.crdClient.SuperedgeV1().ServiceGrids(sgCopy.Namespace).UpdateStatus(context.TODO(), sgCopy, metav1.UpdateOptions{})
				if err != nil {
					klog.Errorf("Updating add services %d when error occurred %v", svc.Name, err)
				}
			}
		}(adds[i])
	}

	for i := range updates {
		go func(svc *corev1.Service) {
			defer wg.Done()
			klog.V(4).Infof("Updating service %s/%s by syncService", svc.Namespace, svc.Name)
			svcCurrent, err := sgc.kubeClient.CoreV1().Services(svc.Namespace).Get(context.TODO(), svc.Name, metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Get service: %s, error: %v", svc.Name, err)
				return
			}
			if svc.ResourceVersion == "" {
				svc.ResourceVersion = svcCurrent.ResourceVersion
			}
			if svc.Spec.ClusterIP == "" {
				svc.Spec.ClusterIP = svcCurrent.Spec.ClusterIP
			}
			sgCopy := sg.DeepCopy()
			_, err = sgc.kubeClient.CoreV1().Services(svc.Namespace).Update(context.TODO(), svc, metav1.UpdateOptions{})
			if err != nil {
				errCh <- err
				sgCopy.Status.Conditions = []metav1.Condition{{
					Type:               common.UpdateError,
					Status:             metav1.ConditionTrue,
					Message:            err.Error(),
					LastTransitionTime: metav1.NewTime(time.Now()),
					Reason:             common.UpdateError,
				}}

				sgc.eventRecorder.Eventf(sg, corev1.EventTypeWarning, common.UpdateError, err.Error())

				_, err := sgc.crdClient.SuperedgeV1().ServiceGrids(sgCopy.Namespace).UpdateStatus(context.TODO(), sgCopy, metav1.UpdateOptions{})
				if err != nil {
					klog.Errorf("Updating update services %s when error occurred %v", svc.Name, err)
				}
			}
		}(updates[i])
	}

	for i := range deletes {
		go func(svc *corev1.Service) {
			defer wg.Done()
			klog.V(4).Infof("Deleting service %s/%s by syncService", svc.Namespace, svc.Name)
			sgCopy := sg.DeepCopy()
			err := sgc.kubeClient.CoreV1().Services(svc.Namespace).Delete(context.TODO(), svc.Name, metav1.DeleteOptions{})
			if err != nil && !errors.IsNotFound(err) {
				errCh <- err
				sgCopy.Status.Conditions = []metav1.Condition{{
					Type:               common.DeleteError,
					Status:             metav1.ConditionTrue,
					Message:            err.Error(),
					LastTransitionTime: metav1.NewTime(time.Now()),
					Reason:             common.DeleteError,
				}}

				sgc.eventRecorder.Eventf(sg, corev1.EventTypeWarning, common.DeleteError, err.Error())

				_, err := sgc.crdClient.SuperedgeV1().ServiceGrids(sgCopy.Namespace).UpdateStatus(context.TODO(), sgCopy, metav1.UpdateOptions{})
				if err != nil {
					klog.Errorf("Updating delete services %d when error occurred %v", svc.Name, err)
				}
			}
		}(deletes[i])
	}

	wg.Wait()

	var err error
	for len(errCh) != 0 {
		select {
		case e := <-errCh:
			err = multierror.Append(err, e)
		default:
		}
	}

	return err
}

func (sgc *ServiceGridController) reconcileFed(sg *crdv1.ServiceGrid, sgList []*crdv1.ServiceGrid, disNsList []string) error {
	existedDisSgMap := make(map[string]*crdv1.ServiceGrid)

	for _, fedsg := range sgList {
		existedDisSgMap[fedsg.Namespace] = fedsg
	}

	wanted := sets.NewString()
	for _, v := range disNsList {
		wanted.Insert(v)
	}

	var (
		adds    []*crdv1.ServiceGrid
		updates []*crdv1.ServiceGrid
		deletes []*crdv1.ServiceGrid
	)

	for _, ns := range disNsList {
		fedsg, found := existedDisSgMap[ns]
		if !found {
			ServiceGridToAdd := util.CreateServiceGrid(sg, ns)
			adds = append(adds, ServiceGridToAdd)
			continue
		}

		ServiceGridToUpdate := util.UpdateServiceGrid(sg, fedsg)

		scheme := scheme.Scheme
		scheme.Default(ServiceGridToUpdate)
		if !commonutil.DeepContains(fedsg.Spec, ServiceGridToUpdate.Spec) {
			klog.Infof("serviceGrid %s template changed", fedsg.Name)
			out, _ := json.Marshal(ServiceGridToUpdate.Spec)
			klog.V(5).Infof("ServiceGridToUpdate is %s", string(out))
			out, _ = json.Marshal(fedsg.Spec)
			klog.V(5).Infof("existedServiceGrid is %s", string(out))
			updates = append(updates, ServiceGridToUpdate)
			continue
		}
	}

	// If deployment's name is not matched with grid value but has the same selector, we remove it.
	for _, sgexist := range sgList {
		if !wanted.Has(sgexist.Namespace) {
			deletes = append(deletes, sgexist)
		}
	}

	if err := sgc.syncDisServiceGrid(adds, updates, deletes); err != nil {
		return err
	}

	return nil
}

func (sgc *ServiceGridController) syncDisServiceGrid(adds, updates, deletes []*crdv1.ServiceGrid) error {
	wg := sync.WaitGroup{}
	totalSize := len(adds) + len(updates) + len(deletes)
	wg.Add(totalSize)
	errCh := make(chan error, totalSize)

	for i := range adds {
		go func(s *crdv1.ServiceGrid) {
			defer wg.Done()
			klog.V(4).Infof("Creating DisServiceGrid %s/%s by syncDisServiceGrid", s.Namespace, s.Name)
			_, err := sgc.crdClient.SuperedgeV1().ServiceGrids(s.Namespace).Create(context.TODO(), s, metav1.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				errCh <- err
			}
		}(adds[i])
	}

	for i := range updates {
		go func(s *crdv1.ServiceGrid) {
			defer wg.Done()
			klog.V(4).Infof("Updating DisServiceGrid %s/%s by syncServiceGrid", s.Namespace, s.Name)
			_, err := sgc.crdClient.SuperedgeV1().ServiceGrids(s.Namespace).Update(context.TODO(), s, metav1.UpdateOptions{})
			if err != nil {
				errCh <- err
			}
		}(updates[i])
	}

	for i := range deletes {
		go func(s *crdv1.ServiceGrid) {
			defer wg.Done()
			klog.V(4).Infof("Deleting DisServiceGrid %s/%s by syncDisServiceGrid", s.Namespace, s.Name)
			err := sgc.crdClient.SuperedgeV1().ServiceGrids(s.Namespace).Delete(context.TODO(), s.Name, metav1.DeleteOptions{})
			if err != nil && !errors.IsNotFound(err) {
				errCh <- err
			}
		}(deletes[i])
	}

	wg.Wait()

	var err error
	for len(errCh) != 0 {
		select {
		case e := <-errCh:
			if !errors.IsConflict(e) {
				err = multierror.Append(err, e)
			}
		default:
		}
	}

	return err
}
