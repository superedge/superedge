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
	"context"
	"flag"
	"testing"

	"k8s.io/klog/v2"

	appsv1 "k8s.io/api/apps/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/fake"

	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/testutil"
)

func init() {
	flagSets := flag.NewFlagSet("test", flag.ExitOnError)
	klog.InitFlags(flagSets)
	_ = flagSets.Set("v", "4")
	_ = flagSets.Set("logtostderr", "true")
	_ = flagSets.Parse(nil)
}

func TestDeploymentControllerRefManager(t *testing.T) {
	controller := &crdv1.DeploymentGrid{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-1",
			Namespace: "test",
			UID:       "uid-test-1",
		},
	}
	selector := labels.SelectorFromSet(map[string]string{
		"foo": "bar",
	})
	controllerKind := crdv1.SchemeGroupVersion.WithKind("DeploymentGrid")
	testObjects := []*appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj-owned-by-someone",
				Namespace: "test",
				Labels: map[string]string{
					"foo": "bar",
				},
				UID: "obj-owned-by-someone",
				OwnerReferences: []metav1.OwnerReference{
					{
						Controller: boolPtr(true),
						UID:        "uid-test-2",
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj-owned-not-match",
				Namespace: "test",
				Labels: map[string]string{
					"a": "b",
				},
				UID: "obj-owned-not-match",
				OwnerReferences: []metav1.OwnerReference{
					{
						Controller: boolPtr(true),
						UID:        "uid-test-1",
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj-owned",
				Namespace: "test",
				Labels: map[string]string{
					"foo": "bar",
				},
				UID: "obj-owned",
				OwnerReferences: []metav1.OwnerReference{
					{
						Controller: boolPtr(true),
						UID:        "uid-test-1",
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj-orphan-delete",
				Namespace: "test",
				Labels: map[string]string{
					"foo": "bar",
				},
				UID:               "obj-orphan-delete",
				DeletionTimestamp: timePtr(metav1.Now()),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj-adopt",
				Namespace: "test",
				Labels: map[string]string{
					"foo": "bar",
				},
				UID: "obj-adopt",
			},
		},
	}

	kubeClient := fake.NewSimpleClientset()
	for i, obj := range testObjects {
		_, err := kubeClient.AppsV1().Deployments(obj.Namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
		if err != nil {
			t.Errorf("can't create testobject %d, %v", i, err)
			return
		}
	}

	claimedObjects := []*appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj-owned",
				Namespace: "test",
				Labels: map[string]string{
					"foo": "bar",
				},
				UID: "obj-owned",
				OwnerReferences: []metav1.OwnerReference{
					{
						Controller: boolPtr(true),
						UID:        "uid-test-1",
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj-adopt",
				Namespace: "test",
				Labels: map[string]string{
					"foo": "bar",
				},
				UID: "obj-adopt",
			},
		},
	}

	changedObjects := []*appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj-adopt",
				Namespace: "test",
				Labels: map[string]string{
					"foo": "bar",
				},
				UID: "obj-adopt",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion:         controllerKind.GroupVersion().String(),
						Kind:               controllerKind.Kind,
						Name:               controller.Name,
						UID:                controller.UID,
						Controller:         boolPtr(true),
						BlockOwnerDeletion: boolPtr(true),
					},
				},
			},
		},
	}

	manager := NewDeploymentControllerRefManager(&RealDeployClient{
		kubeClient: kubeClient,
	}, controller, selector, controllerKind, func() error { return nil })

	claimed, err := manager.ClaimDeployment(testObjects)
	if err != nil {
		t.Errorf("clamin got error %v", err)
		return
	}

	if len(claimed) != len(claimedObjects) {
		t.Errorf("claimed objects length expect %d got %d", len(claimed), len(claimedObjects))
		return
	}

	for i := range claimed {
		if !apiequality.Semantic.DeepEqual(claimed[i], claimedObjects[i]) {
			t.Errorf("%d expect claim object %s to be %s", i, testutil.JsonStringfy(claimed[i]), testutil.JsonStringfy(claimedObjects[i]))
			return
		}
	}

	for i := range changedObjects {
		obj, err := kubeClient.AppsV1().Deployments(changedObjects[i].Namespace).Get(context.TODO(), changedObjects[i].Name, metav1.GetOptions{})
		if err != nil {
			t.Errorf("can't get released object, %v", err)
			return
		}

		if !apiequality.Semantic.DeepEqual(obj, changedObjects[i]) {
			t.Errorf("%d expect changed object %s to be %s", i, testutil.JsonStringfy(obj), testutil.JsonStringfy(changedObjects[i]))
			return
		}
	}
}

func boolPtr(val bool) *bool {
	v := val
	return &v
}

func timePtr(val metav1.Time) *metav1.Time {
	v := val
	return &v
}
