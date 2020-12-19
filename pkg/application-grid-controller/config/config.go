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

package config

import (
	"time"

	"k8s.io/client-go/informers"
	appsv1 "k8s.io/client-go/informers/apps/v1"
	corev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"

	crdClientset "superedge/pkg/application-grid-controller/generated/clientset/versioned"
	crdFactory "superedge/pkg/application-grid-controller/generated/informers/externalversions"
	crdv1 "superedge/pkg/application-grid-controller/generated/informers/externalversions/superedge.io/v1"
)

type ControllerConfig struct {
	ServiceGridInformer    crdv1.ServiceGridInformer
	DeploymentGridInformer crdv1.DeploymentGridInformer
	ServiceInformer        corev1.ServiceInformer
	DeploymentInformer     appsv1.DeploymentInformer
	NodeInformer           corev1.NodeInformer
}

func NewControllerConfig(crdClient *crdClientset.Clientset, k8sClient *kubernetes.Clientset, resyncTime time.Duration) *ControllerConfig {
	crdFactory := crdFactory.NewSharedInformerFactory(crdClient, resyncTime)
	k8sFactory := informers.NewSharedInformerFactory(k8sClient, resyncTime)

	return &ControllerConfig{
		ServiceGridInformer:    crdFactory.Superedge().V1().ServiceGrids(),
		DeploymentGridInformer: crdFactory.Superedge().V1().DeploymentGrids(),
		ServiceInformer:        k8sFactory.Core().V1().Services(),
		DeploymentInformer:     k8sFactory.Apps().V1().Deployments(),
		NodeInformer:           k8sFactory.Core().V1().Nodes(),
	}
}

func (c *ControllerConfig) Run(stop <-chan struct{}) {
	go c.ServiceGridInformer.Informer().Run(stop)
	go c.DeploymentGridInformer.Informer().Run(stop)
	go c.ServiceInformer.Informer().Run(stop)
	go c.DeploymentInformer.Informer().Run(stop)
	go c.NodeInformer.Informer().Run(stop)
}
