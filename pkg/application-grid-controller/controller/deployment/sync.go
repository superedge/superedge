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
	"k8s.io/klog"
	"sync"

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
			states[util.GetGridValueFromName(dg, dp.Name)] = dp.Status
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
		if dgc.templateHasher.IsTemplateHashChanged(dg, v, dp) {
			klog.Infof("deployment %s template hash changed", dp.Name)
			updates = append(updates, DeploymentToUpdate)
			continue
		} else {
			scheme := scheme.Scheme
			scheme.Default(DeploymentToUpdate)
			DeploymentToUpdate.Spec.Replicas = dp.Spec.Replicas
			if !commonutil.DeepContains(dp.Spec, DeploymentToUpdate.Spec){
				klog.Infof("deployment %s template changed", dp.Name)
				out, _ := json.Marshal(DeploymentToUpdate.Spec)
				klog.V(5).Infof("deploymentToUpdate is %s", string(out))
				out, _ = json.Marshal(dp.Spec)
				klog.V(5).Info("deployment is %s", string(out))
				updates = append(updates, DeploymentToUpdate)
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
		return err
	}

	return dgc.syncStatus(dg, dpList, gridValues)
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
			if err != nil {
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
			if err != nil {
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
