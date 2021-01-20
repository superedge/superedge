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

package statefulset

import (
	"context"
	"sync"

	"github.com/hashicorp/go-multierror"
	appsv1 "k8s.io/api/apps/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/statefulset/util"
)

func (ssgc *StatefulSetGridController) syncStatus(ssg *crdv1.StatefulSetGrid, setList []*appsv1.StatefulSet, gridValues []string) error {
	wanted := sets.NewString()
	for _, v := range gridValues {
		wanted.Insert(util.GetStatefulSetName(ssg, v))
	}

	states := make(map[string]appsv1.StatefulSetStatus)
	for _, set := range setList {
		if wanted.Has(set.Name) {
			states[util.GetGridValueFromName(ssg, set.Name)] = set.Status
		}
	}
	ssgCopy := ssg.DeepCopy()
	ssgCopy.Status.States = states
	_, err := ssgc.crdClient.SuperedgeV1().StatefulSetGrids(ssg.Namespace).UpdateStatus(context.TODO(), ssgCopy, metav1.UpdateOptions{})
	if err != nil && errors.IsConflict(err) {
		return nil
	}
	return err
}

func (ssgc *StatefulSetGridController) reconcile(ssg *crdv1.StatefulSetGrid, setList []*appsv1.StatefulSet, gridValues []string) error {
	existedSetMap := make(map[string]*appsv1.StatefulSet)

	for _, set := range setList {
		existedSetMap[set.Name] = set
	}

	wanted := sets.NewString()
	for _, v := range gridValues {
		wanted.Insert(util.GetStatefulSetName(ssg, v))
	}

	var (
		adds    []*appsv1.StatefulSet
		updates []*appsv1.StatefulSet
		deletes []*appsv1.StatefulSet
	)

	for _, v := range gridValues {
		name := util.GetStatefulSetName(ssg, v)

		set, found := existedSetMap[name]
		if !found {
			adds = append(adds, util.CreateStatefulSet(ssg, v))
			continue
		}

		template := util.KeepConsistence(ssg, set, v)
		if !apiequality.Semantic.DeepEqual(template, set) {
			updates = append(updates, template)
		}
	}

	// If statefulset's name is not matched with grid value but has the same selector, we remove it.
	for _, set := range setList {
		if !wanted.Has(set.Name) {
			deletes = append(deletes, set)
		}
	}

	if err := ssgc.syncStatefulSet(adds, updates, deletes); err != nil {
		return err
	}

	return ssgc.syncStatus(ssg, setList, gridValues)
}

func (ssgc *StatefulSetGridController) syncStatefulSet(adds, updates, deletes []*appsv1.StatefulSet) error {
	wg := sync.WaitGroup{}
	totalSize := len(adds) + len(updates) + len(deletes)
	wg.Add(totalSize)
	errCh := make(chan error, totalSize)

	for i := range adds {
		go func(set *appsv1.StatefulSet) {
			defer wg.Done()
			_, err := ssgc.kubeClient.AppsV1().StatefulSets(set.Namespace).Create(context.TODO(), set, metav1.CreateOptions{})
			if err != nil {
				errCh <- err
			}
		}(adds[i])
	}

	for i := range updates {
		go func(set *appsv1.StatefulSet) {
			defer wg.Done()
			_, err := ssgc.kubeClient.AppsV1().StatefulSets(set.Namespace).Update(context.TODO(), set, metav1.UpdateOptions{})
			if err != nil {
				errCh <- err
			}
		}(updates[i])
	}

	for i := range deletes {
		go func(set *appsv1.StatefulSet) {
			defer wg.Done()
			err := ssgc.kubeClient.AppsV1().StatefulSets(set.Namespace).Delete(context.TODO(), set.Name, metav1.DeleteOptions{})
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
