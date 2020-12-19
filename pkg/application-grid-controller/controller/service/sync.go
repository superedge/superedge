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
	"sync"

	"github.com/hashicorp/go-multierror"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	crdv1 "superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	"superedge/pkg/application-grid-controller/controller/service/util"
)

func (sgc *ServiceGridController) reconcile(g *crdv1.ServiceGrid, svcList []*corev1.Service) error {
	existedSvcsMap := make(map[string]*corev1.Service)

	for _, svc := range svcList {
		existedSvcsMap[svc.Name] = svc
	}

	var (
		adds    []*corev1.Service
		updates []*corev1.Service
		deletes []*corev1.Service
	)

	name := util.GetServiceName(g)
	for _, svc := range svcList {
		if name == svc.Name {
			template := util.KeepConsistence(g, svc)
			if !apiequality.Semantic.DeepEqual(template, svc) {
				updates = append(updates, template)
				continue
			}

			updates = append(updates, svc)
			continue
		}
		deletes = append(deletes, svc)
	}

	if len(updates) == 0 {
		adds = append(adds, util.CreateService(g))
	}

	return sgc.syncService(adds, updates, deletes)
}

func (sgc *ServiceGridController) syncService(adds, updates, deletes []*corev1.Service) error {
	wg := sync.WaitGroup{}
	totalSize := len(adds) + len(updates) + len(deletes)
	wg.Add(totalSize)
	errCh := make(chan error, totalSize)

	for i := range adds {
		go func(d *corev1.Service) {
			defer wg.Done()
			_, err := sgc.kubeClient.CoreV1().Services(d.Namespace).Create(context.TODO(), d, metav1.CreateOptions{})
			if err != nil {
				errCh <- err
			}
		}(adds[i])
	}

	for i := range updates {
		go func(d *corev1.Service) {
			defer wg.Done()
			_, err := sgc.kubeClient.CoreV1().Services(d.Namespace).Update(context.TODO(), d, metav1.UpdateOptions{})
			if err != nil {
				errCh <- err
			}
		}(updates[i])
	}

	for i := range deletes {
		go func(d *corev1.Service) {
			defer wg.Done()
			err := sgc.kubeClient.CoreV1().Services(d.Namespace).Delete(context.TODO(), d.Name, metav1.DeleteOptions{})
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
			err = multierror.Append(err, e)
		default:
		}
	}

	return err
}
