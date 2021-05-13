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

	crdClientset "github.com/superedge/superedge/pkg/application-grid-controller/generated/clientset/versioned"
	crdFactoryset "github.com/superedge/superedge/pkg/application-grid-controller/generated/informers/externalversions"
	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/generated/informers/externalversions/superedge.io/v1"
)

type ControllerConfig struct {
	ServiceGridInformer       crdv1.ServiceGridInformer
	DeploymentGridInformer    crdv1.DeploymentGridInformer
	StatefulSetGridInformer   crdv1.StatefulSetGridInformer
	ServiceInformer           corev1.ServiceInformer
	DeploymentInformer        appsv1.DeploymentInformer
	StatefulSetInformer       appsv1.StatefulSetInformer
	NodeInformer              corev1.NodeInformer
	NameSpaceInformer         corev1.NamespaceInformer
	FedDeploymentGridInformer crdv1.DeploymentGridInformer
	FedServiceGridInformer    crdv1.ServiceGridInformer
}

func NewControllerConfig(crdClient, fedCrdClient *crdClientset.Clientset, k8sClient *kubernetes.Clientset,
	resyncTime time.Duration, dedicatedNameSpace string) *ControllerConfig {
	crdFactory := crdFactoryset.NewSharedInformerFactory(crdClient, resyncTime)
	k8sFactory := informers.NewSharedInformerFactory(k8sClient, resyncTime)

	conf := &ControllerConfig{
		ServiceGridInformer:     crdFactory.Superedge().V1().ServiceGrids(),
		DeploymentGridInformer:  crdFactory.Superedge().V1().DeploymentGrids(),
		StatefulSetGridInformer: crdFactory.Superedge().V1().StatefulSetGrids(),
		ServiceInformer:         k8sFactory.Core().V1().Services(),
		DeploymentInformer:      k8sFactory.Apps().V1().Deployments(),
		StatefulSetInformer:     k8sFactory.Apps().V1().StatefulSets(),
		NodeInformer:            k8sFactory.Core().V1().Nodes(),
		NameSpaceInformer:       k8sFactory.Core().V1().Namespaces(),
	}

	if fedCrdClient != nil {
		fedCrdFactory := crdFactoryset.NewSharedInformerFactoryWithOptions(fedCrdClient, resyncTime, crdFactoryset.WithNamespace(dedicatedNameSpace))
		conf.FedDeploymentGridInformer = fedCrdFactory.Superedge().V1().DeploymentGrids()
		conf.FedServiceGridInformer = fedCrdFactory.Superedge().V1().ServiceGrids()
	}
	return conf
}

func (c *ControllerConfig) Run(stop <-chan struct{}) {
	go c.ServiceGridInformer.Informer().Run(stop)
	go c.DeploymentGridInformer.Informer().Run(stop)
	go c.StatefulSetGridInformer.Informer().Run(stop)
	go c.ServiceInformer.Informer().Run(stop)
	go c.DeploymentInformer.Informer().Run(stop)
	go c.StatefulSetInformer.Informer().Run(stop)
	go c.NodeInformer.Informer().Run(stop)
	go c.NameSpaceInformer.Informer().Run(stop)
	if c.FedDeploymentGridInformer != nil {
		go c.FedDeploymentGridInformer.Informer().Run(stop)
		go c.FedServiceGridInformer.Informer().Run(stop)
	}
}
