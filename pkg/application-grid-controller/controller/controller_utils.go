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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var (
	KeyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc
)

type DPControlInterface interface {
	PatchDeployment(namespace, name string, data []byte) error
}

type SVCControlInterface interface {
	PatchService(namespace, name string, data []byte) error
}

type RealDPControl struct {
	KubeClient clientset.Interface
}

type RealSVCControl struct {
	KubeClient clientset.Interface
}

var _ DPControlInterface = &RealDPControl{}
var _ SVCControlInterface = &RealSVCControl{}

func (r RealDPControl) PatchDeployment(namespace, name string, data []byte) error {
	_, err := r.KubeClient.AppsV1().Deployments(namespace).Patch(context.TODO(), name, types.StrategicMergePatchType, data, metav1.PatchOptions{})
	return err
}

func (r RealSVCControl) PatchService(namespace, name string, data []byte) error {
	_, err := r.KubeClient.CoreV1().Services(namespace).Patch(context.TODO(), name, types.StrategicMergePatchType, data, metav1.PatchOptions{})
	return err
}

func RecheckDeletionTimestamp(getObject func() (metav1.Object, error)) func() error {
	return func() error {
		obj, err := getObject()
		if err != nil {
			return fmt.Errorf("can't recheck DeletionTimestamp: %v", err)
		}
		if obj.GetDeletionTimestamp() != nil {
			return fmt.Errorf("%v/%v has just been deleted at %v", obj.GetNamespace(), obj.GetName(),
				obj.GetDeletionTimestamp())
		}
		return nil
	}
}
