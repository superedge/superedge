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
	"sync"

	"github.com/hashicorp/go-multierror"
	appsv1 "k8s.io/api/apps/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	crdv1 "superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	"superedge/pkg/application-grid-controller/controller/deployment/util"
)

func (dgc *DeploymentGridController) syncStatus(g *crdv1.DeploymentGrid, dpList []*appsv1.Deployment, gridValues []string) error {
	wanted := sets.NewString()
	for _, v := range gridValues {
		wanted.Insert(util.GetDeploymentName(g, v))
	}

	states := make(map[string]appsv1.DeploymentStatus)
	for _, dp := range dpList {
		if wanted.Has(dp.Name) {
			states[util.GetGridValueFromName(g, dp.Name)] = dp.Status
		}
	}
	g.Status.States = states
	_, err := dgc.crdClient.SuperedgeV1().DeploymentGrids(g.Namespace).UpdateStatus(context.TODO(), g, metav1.UpdateOptions{})
	if err != nil && errors.IsConflict(err) {
		return nil
	}
	return err
}

func (dgc *DeploymentGridController) reconcile(g *crdv1.DeploymentGrid, dpList []*appsv1.Deployment, gridValues []string) error {
	existedDPsMap := make(map[string]*appsv1.Deployment)

	for _, dp := range dpList {
		existedDPsMap[dp.Name] = dp
	}

	wanted := sets.NewString()
	for _, v := range gridValues {
		/* nginx-zone1
		 */
		wanted.Insert(util.GetDeploymentName(g, v))
	}

	var (
		adds    []*appsv1.Deployment
		updates []*appsv1.Deployment
		deletes []*appsv1.Deployment
	)

	for _, v := range gridValues {
		name := util.GetDeploymentName(g, v)
		spec := *g.Spec.Template.DeepCopy()
		spec.Template.Spec.NodeSelector = map[string]string{
			g.Spec.GridUniqKey: v,
		}

		dp, found := existedDPsMap[name]
		if !found {
			adds = append(adds, util.CreateDeployment(g, v))
			continue
		}

		template := util.KeepConsistence(g, dp, v)
		if !apiequality.Semantic.DeepEqual(template, dp) {
			updates = append(updates, template)
		} else {
			updates = append(updates, dp)
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

	return dgc.syncStatus(g, updates, gridValues)
}

func (dgc *DeploymentGridController) syncDeployment(adds, updates, deletes []*appsv1.Deployment) error {
	wg := sync.WaitGroup{}
	totalSize := len(adds) + len(updates) + len(deletes)
	wg.Add(totalSize)
	errCh := make(chan error, totalSize)

	for i := range adds {
		go func(d *appsv1.Deployment) {
			defer wg.Done()
			_, err := dgc.kubeClient.AppsV1().Deployments(d.Namespace).Create(context.TODO(), d, metav1.CreateOptions{})
			if err != nil {
				errCh <- err
			}
		}(adds[i])
	}

	for i := range updates {
		go func(d *appsv1.Deployment) {
			defer wg.Done()
			_, err := dgc.kubeClient.AppsV1().Deployments(d.Namespace).Update(context.TODO(), d, metav1.UpdateOptions{})
			if err != nil {
				errCh <- err
			}
		}(updates[i])
	}

	for i := range deletes {
		go func(d *appsv1.Deployment) {
			defer wg.Done()
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
