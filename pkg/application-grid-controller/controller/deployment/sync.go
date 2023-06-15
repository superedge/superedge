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

package deployment

import (
	"context"
	"encoding/json"
	"sync"

	corev1 "k8s.io/api/core/v1"
	klog "k8s.io/klog/v2"

	"github.com/hashicorp/go-multierror"
	appsv1 "k8s.io/api/apps/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes/scheme"

	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/deployment/util"
	commonutil "github.com/superedge/superedge/pkg/application-grid-controller/util"
)

func (dgc *DeploymentGridController) syncStatus(dg *crdv1.DeploymentGrid, dpList []*appsv1.Deployment, gridValues []string) error {
	wanted := sets.NewString()
	for _, v := range gridValues {
		wanted.Insert(util.GetDeploymentName(dg, v))
	}

	states := make(map[string]appsv1.DeploymentStatus)
	for _, dp := range dpList {
		if wanted.Has(dp.Name) {
			states[util.GetGridValueFromName(dg, dp.Name)] = commonutil.FiltDgStatus(dp.Status)
		}
	}
	if !apiequality.Semantic.DeepEqual(dg.Status.States, states) {
		// NEVER modify objects from the store. It's a read-only, local cache.
		// You can use DeepCopy() to make a deep copy of original object and modify this copy
		// Or create a copy manually for better performance
		dgCopy := dg.DeepCopy()
		dgCopy.Status.States = states
		klog.V(4).Infof("old status is %#v", dg.Status.States)
		klog.V(4).Infof("Updating deployment grid %s/%s to status %#v", dgCopy.Namespace, dgCopy.Name, states)
		_, err := dgc.crdClient.SuperedgeV1().DeploymentGrids(dgCopy.Namespace).UpdateStatus(context.TODO(), dgCopy, metav1.UpdateOptions{})
		if err != nil && errors.IsConflict(err) {
			return nil
		}
		return err
	}
	return nil
}

func (dgc *DeploymentGridController) reconcile(dg *crdv1.DeploymentGrid, dpList []*appsv1.Deployment, gridValues []string) error {
	existedDPMap := make(map[string]*appsv1.Deployment)

	for _, dp := range dpList {
		existedDPMap[dp.Name] = dp
	}

	wanted := sets.NewString()
	for _, v := range gridValues {
		wanted.Insert(util.GetDeploymentName(dg, v))
	}

	var (
		adds    []*appsv1.Deployment
		updates []*appsv1.Deployment
		deletes []*appsv1.Deployment
	)

	for _, v := range gridValues {
		name := util.GetDeploymentName(dg, v)

		dp, found := existedDPMap[name]
		if !found {
			DeploymentToAdd, err := util.CreateDeployment(dg, v, dgc.templateHasher)
			if err != nil {
				return err
			}
			adds = append(adds, DeploymentToAdd)
			continue
		}

		DeploymentToUpdate, err := util.CreateDeployment(dg, v, dgc.templateHasher)
		if err != nil {
			return err
		}
		IsTemplateHashChanged, IsReplicasChanged := dgc.templateHasher.IsTemplateHashChanged(dg, v, dp), dgc.templateHasher.IsReplicasChanged(dg, v, dp)
		klog.V(5).InfoS("deploymentgrid template change status", "IsTemplateHashChanged", IsTemplateHashChanged, "IsReplicasChanged", IsReplicasChanged)
		if IsTemplateHashChanged || IsReplicasChanged {
			klog.InfoS("deployment template changed", "dp name", dp.Name)
			updates = append(updates, DeploymentToUpdate)
			continue
		} else {
			scheme := scheme.Scheme
			scheme.Default(DeploymentToUpdate)
			DeploymentToUpdate.Spec.Replicas = dp.Spec.Replicas
			if !commonutil.DeepContains(dp.Spec, DeploymentToUpdate.Spec) {
				klog.Infof("deployment %s template changed", dp.Name)
				out, _ := json.Marshal(DeploymentToUpdate.Spec)
				klog.V(5).Infof("deploymentToUpdate is %s", string(out))
				out, _ = json.Marshal(dp.Spec)
				klog.V(5).Infof("deployment is %s", string(out))
				updates = append(updates, DeploymentToUpdate)
				continue
			}
		}
	}

	// If deployment's name is not matched with grid value but has the same selector, we remove it.
	for _, dp := range dpList {
		if !wanted.Has(dp.Name) {
			deletes = append(deletes, dp)
		}
	}

	if err := dgc.syncDeployment(adds, updates, deletes); err != nil {
		dgc.eventRecorder.Eventf(dg, corev1.EventTypeWarning, "SyncDeploymentGridFailed",
			"sync deploymentGrid %s/%s failed because of %v", dg.Name, dg.Namespace, err)
		return err
	}

	if len(dpList) != 0 {
		err := dgc.syncStatus(dg, dpList, gridValues)
		if err != nil {
			dgc.eventRecorder.Eventf(dg, corev1.EventTypeWarning, "SyncDeploymentGridStatusFailed",
				"sync deploymentGridStatus %s/%s failed because of %v", dg.Name, dg.Namespace, err)
		}
		return err
	}

	return nil
}

func (dgc *DeploymentGridController) syncDeployment(adds, updates, deletes []*appsv1.Deployment) error {
	wg := sync.WaitGroup{}
	totalSize := len(adds) + len(updates) + len(deletes)
	wg.Add(totalSize)
	errCh := make(chan error, totalSize)

	for i := range adds {
		go func(d *appsv1.Deployment) {
			defer wg.Done()
			klog.V(4).Infof("Creating deployment %s/%s by syncDeployment", d.Namespace, d.Name)
			_, err := dgc.kubeClient.AppsV1().Deployments(d.Namespace).Create(context.TODO(), d, metav1.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				errCh <- err
			}
		}(adds[i])
	}

	for i := range updates {
		go func(d *appsv1.Deployment) {
			defer wg.Done()
			klog.V(4).Infof("Updating deployment %s/%s by syncDeployment", d.Namespace, d.Name)
			_, err := dgc.kubeClient.AppsV1().Deployments(d.Namespace).Update(context.TODO(), d, metav1.UpdateOptions{})
			if err != nil {
				errCh <- err
			}
		}(updates[i])
	}

	for i := range deletes {
		go func(d *appsv1.Deployment) {
			defer wg.Done()
			klog.V(4).Infof("Deleting deployment %s/%s by syncDeployment", d.Namespace, d.Name)
			err := dgc.kubeClient.AppsV1().Deployments(d.Namespace).Delete(context.TODO(), d.Name, metav1.DeleteOptions{})
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

func (dgc *DeploymentGridController) reconcileFed(dg *crdv1.DeploymentGrid, dgList []*crdv1.DeploymentGrid, disNsList []string) error {
	existedDisDgMap := make(map[string]*crdv1.DeploymentGrid)

	for _, feddg := range dgList {
		existedDisDgMap[feddg.Namespace] = feddg
	}

	wanted := sets.NewString()
	for _, v := range disNsList {
		wanted.Insert(v)
	}

	var (
		adds    []*crdv1.DeploymentGrid
		updates []*crdv1.DeploymentGrid
		deletes []*crdv1.DeploymentGrid
	)

	for _, ns := range disNsList {
		feddg, found := existedDisDgMap[ns]
		if !found {
			DeploymentGridToAdd := util.CreateDeploymentGrid(dg, ns)
			adds = append(adds, DeploymentGridToAdd)
			continue
		}

		DeploymentGridToUpdate := util.UpdateDeploymentGrid(dg, feddg)

		scheme := scheme.Scheme
		scheme.Default(DeploymentGridToUpdate)
		if !commonutil.DeepContains(feddg.Spec, DeploymentGridToUpdate.Spec) {
			klog.Infof("deployment %s template changed", feddg.Name)
			out, _ := json.Marshal(DeploymentGridToUpdate.Spec)
			klog.V(5).Infof("deploymentGridToUpdate is %s", string(out))
			out, _ = json.Marshal(feddg.Spec)
			klog.V(5).Infof("existedDeploymentGrid is %s", string(out))
			updates = append(updates, DeploymentGridToUpdate)
			continue
		}
	}

	// If deployment's name is not matched with grid value but has the same selector, we remove it.
	for _, dg := range dgList {
		if !wanted.Has(dg.Namespace) {
			deletes = append(deletes, dg)
		}
	}

	if err := dgc.syncDisDeploymentGrid(adds, updates, deletes); err != nil {
		return err
	}

	if len(dgList) != 0 && len(disNsList) != 0 {
		return dgc.syncDisStatus(dg, dgList, disNsList)
	}
	return nil
}

func (dgc *DeploymentGridController) syncDisDeploymentGrid(adds, updates, deletes []*crdv1.DeploymentGrid) error {
	wg := sync.WaitGroup{}
	totalSize := len(adds) + len(updates) + len(deletes)
	wg.Add(totalSize)
	errCh := make(chan error, totalSize)

	for i := range adds {
		go func(d *crdv1.DeploymentGrid) {
			defer wg.Done()
			klog.V(4).Infof("Creating DisDeploymentGrid %s/%s by syncDisDeployment", d.Namespace, d.Name)
			_, err := dgc.crdClient.SuperedgeV1().DeploymentGrids(d.Namespace).Create(context.TODO(), d, metav1.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				errCh <- err
			}
		}(adds[i])
	}

	for i := range updates {
		go func(d *crdv1.DeploymentGrid) {
			defer wg.Done()
			klog.V(4).Infof("Updating DisDeploymentGrid %s/%s by syncDisDeployment", d.Namespace, d.Name)
			_, err := dgc.crdClient.SuperedgeV1().DeploymentGrids(d.Namespace).Update(context.TODO(), d, metav1.UpdateOptions{})
			if err != nil {
				errCh <- err
			}
		}(updates[i])
	}

	for i := range deletes {
		go func(d *crdv1.DeploymentGrid) {
			defer wg.Done()
			klog.V(4).Infof("Deleting DisDeploymentGrid %s/%s by syncDisDeployment", d.Namespace, d.Name)
			err := dgc.crdClient.SuperedgeV1().DeploymentGrids(d.Namespace).Delete(context.TODO(), d.Name, metav1.DeleteOptions{})
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

func (dgc *DeploymentGridController) syncDisStatus(dg *crdv1.DeploymentGrid, dgList []*crdv1.DeploymentGrid, disNsList []string) error {
	wanted := sets.NewString()
	for _, v := range disNsList {
		wanted.Insert(v)
	}

	states := make(map[string]appsv1.DeploymentStatus)
	for _, dg := range dgList {
		if wanted.Has(dg.Namespace) {
			for k, v := range dg.Status.States {
				states[k] = v
			}
		}
	}
	if !apiequality.Semantic.DeepEqual(dg.Status.States, states) {
		// NEVER modify objects from the store. It's a read-only, local cache.
		// You can use DeepCopy() to make a deep copy of original object and modify this copy
		// Or create a copy manually for better performance
		dgCopy := dg.DeepCopy()
		dgCopy.Status.States = states
		klog.V(4).Infof("Updating deployment grid %s/%s status %#v", dgCopy.Namespace, dgCopy.Name, states)
		_, err := dgc.crdClient.SuperedgeV1().DeploymentGrids(dgCopy.Namespace).UpdateStatus(context.TODO(), dgCopy, metav1.UpdateOptions{})
		if err != nil && errors.IsConflict(err) {
			return nil
		}
		return err
	}
	return nil
}
