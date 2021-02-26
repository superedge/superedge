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

// TODO: Deprecated in future since this is redundant?
type DeployClientInterface interface {
	PatchDeployment(namespace, name string, data []byte) error
	DeleteDeployment(namespace, name string) error
}

type SetClientInterface interface {
	PatchStatefulSet(namespace, name string, data []byte) error
	DeleteStatefulSet(namespace, name string) error
}

type SvcClientInterface interface {
	PatchService(namespace, name string, data []byte) error
	DeleteService(namespace, name string) error
}

type RealDeployClient struct {
	kubeClient clientset.Interface
}

type RealSetClient struct {
	kubeClient clientset.Interface
}

type RealSvcClient struct {
	kubeClient clientset.Interface
}

var _ DeployClientInterface = &RealDeployClient{}
var _ SetClientInterface = &RealSetClient{}
var _ SvcClientInterface = &RealSvcClient{}

func NewRealDeployClient(kubeClient clientset.Interface) *RealDeployClient {
	return &RealDeployClient{
		kubeClient: kubeClient,
	}
}

func (rdc RealDeployClient) PatchDeployment(namespace, name string, data []byte) error {
	_, err := rdc.kubeClient.AppsV1().Deployments(namespace).Patch(context.TODO(), name, types.StrategicMergePatchType, data, metav1.PatchOptions{})
	return err
}

func (rdc RealDeployClient) DeleteDeployment(namespace, name string) error {
	return rdc.kubeClient.AppsV1().Deployments(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func NewRealSvcClient(kubeClient clientset.Interface) *RealSvcClient {
	return &RealSvcClient{
		kubeClient: kubeClient,
	}
}

func (rsc *RealSvcClient) PatchService(namespace, name string, data []byte) error {
	_, err := rsc.kubeClient.CoreV1().Services(namespace).Patch(context.TODO(), name, types.StrategicMergePatchType, data, metav1.PatchOptions{})
	return err
}

func (rsc *RealSvcClient) DeleteService(namespace, name string) error {
	return rsc.kubeClient.CoreV1().Services(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func NewRealSetClient(kubeClient clientset.Interface) *RealSetClient {
	return &RealSetClient{
		kubeClient: kubeClient,
	}
}

func (rssc *RealSetClient) PatchStatefulSet(namespace, name string, data []byte) error {
	_, err := rssc.kubeClient.AppsV1().StatefulSets(namespace).Patch(context.TODO(), name, types.StrategicMergePatchType, data, metav1.PatchOptions{})
	return err
}

func (rssc *RealSetClient) DeleteStatefulSet(namespace, name string) error {
	return rssc.kubeClient.AppsV1().StatefulSets(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
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
