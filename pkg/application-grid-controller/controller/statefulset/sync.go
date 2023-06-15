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
	"encoding/json"
	"sync"

	"github.com/hashicorp/go-multierror"
	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/statefulset/util"
	commonutil "github.com/superedge/superedge/pkg/application-grid-controller/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
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
	if !apiequality.Semantic.DeepEqual(ssg.Status.States, states) {
		// NEVER modify objects from the store. It's a read-only, local cache.
		// You can use DeepCopy() to make a deep copy of original object and modify this copy
		// Or create a copy manually for better performance
		ssgCopy := ssg.DeepCopy()
		ssgCopy.Status.States = states
		klog.V(4).Infof("Updating statefulset grid %s/%s status %#v", ssgCopy.Namespace, ssgCopy.Name, states)
		_, err := ssgc.crdClient.SuperedgeV1().StatefulSetGrids(ssgCopy.Namespace).UpdateStatus(context.TODO(), ssgCopy, metav1.UpdateOptions{})
		if err != nil && errors.IsConflict(err) {
			return nil
		}
		return err
	}
	return nil
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
			StatefulSetToAdd, err := util.CreateStatefulSet(ssg, v, ssgc.templateHasher)
			if err != nil {
				return err
			}
			adds = append(adds, StatefulSetToAdd)
			continue
		}

		StatefulSetToUpdate, err := util.CreateStatefulSet(ssg, v, ssgc.templateHasher)
		if err != nil {
			return err
		}
		IsTemplateHashChanged, IsReplicasChanged := ssgc.templateHasher.IsTemplateHashChanged(ssg, v, set), ssgc.templateHasher.IsReplicasChanged(ssg, v, set)
		klog.V(5).InfoS("statefulsetgrid template change status", "IsTemplateHashChanged", IsTemplateHashChanged, "IsReplicasChanged", IsReplicasChanged)

		if IsTemplateHashChanged || IsReplicasChanged {
			klog.InfoS("statefulset template hash changed", "sts name", set.Name)
			updates = append(updates, StatefulSetToUpdate)
			continue
		} else {
			scheme := scheme.Scheme
			scheme.Default(StatefulSetToUpdate)
			StatefulSetToUpdate.Spec.Replicas = set.Spec.Replicas
			if !commonutil.DeepContains(set.Spec, StatefulSetToUpdate.Spec) {
				klog.Infof("statefulset %s template changed", set.Name)
				out, _ := json.Marshal(StatefulSetToUpdate.Spec)
				klog.V(5).Infof("StatefulSetToUpdate is %s", string(out))
				out, _ = json.Marshal(set.Spec)
				klog.Infof("StatefulSet is %s", string(out))
				updates = append(updates, StatefulSetToUpdate)
				continue
			}
		}
	}

	// If statefulset's name is not matched with grid value but has the same selector, we remove it.
	for _, set := range setList {
		if !wanted.Has(set.Name) {
			deletes = append(deletes, set)
		}
	}

	if err := ssgc.syncStatefulSet(adds, updates, deletes); err != nil {
		ssgc.eventRecorder.Eventf(ssg, corev1.EventTypeWarning, "SyncStatefulSetGridFailed",
			"sync statefulSetGrid %s/%s failed because of %v", ssg.Name, ssg.Namespace, err)
		return err
	}

	err := ssgc.syncStatus(ssg, setList, gridValues)
	if err != nil {
		ssgc.eventRecorder.Eventf(ssg, corev1.EventTypeWarning, "SyncStatefulSetGridStatusFailed",
			"sync statefulSetGridStatus %s/%s failed because of %v", ssg.Name, ssg.Namespace, err)
	}
	return err
}

func (ssgc *StatefulSetGridController) syncStatefulSet(adds, updates, deletes []*appsv1.StatefulSet) error {
	wg := sync.WaitGroup{}
	totalSize := len(adds) + len(updates) + len(deletes)
	wg.Add(totalSize)
	errCh := make(chan error, totalSize)

	for i := range adds {
		go func(set *appsv1.StatefulSet) {
			defer wg.Done()
			klog.V(4).Infof("Creating statefulset %s/%s by syncStatefulSet", set.Namespace, set.Name)
			_, err := ssgc.kubeClient.AppsV1().StatefulSets(set.Namespace).Create(context.TODO(), set, metav1.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				errCh <- err
			}
		}(adds[i])
	}

	for i := range updates {
		go func(set *appsv1.StatefulSet) {
			defer wg.Done()
			klog.V(4).Infof("Updating statefulset %s/%s by syncStatefulSet", set.Namespace, set.Name)
			_, err := ssgc.kubeClient.AppsV1().StatefulSets(set.Namespace).Update(context.TODO(), set, metav1.UpdateOptions{})
			if err != nil {
				errCh <- err
			}
		}(updates[i])
	}

	for i := range deletes {
		go func(set *appsv1.StatefulSet) {
			defer wg.Done()
			klog.V(4).Infof("Deleting statefulset %s/%s by syncStatefulSet", set.Namespace, set.Name)
			err := ssgc.kubeClient.AppsV1().StatefulSets(set.Namespace).Delete(context.TODO(), set.Name, metav1.DeleteOptions{})
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
