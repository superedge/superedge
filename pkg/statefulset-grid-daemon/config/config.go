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
	crdClientset "github.com/superedge/superedge/pkg/application-grid-controller/generated/clientset/versioned"
	crdFactory "github.com/superedge/superedge/pkg/application-grid-controller/generated/informers/externalversions"
	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/generated/informers/externalversions/superedge.io/v1"
	"time"

	"k8s.io/client-go/informers"
	appsv1 "k8s.io/client-go/informers/apps/v1"
	corev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
)

type ControllerConfig struct {
	StatefulSetGridInformer crdv1.StatefulSetGridInformer
	StatefulSetInformer     appsv1.StatefulSetInformer
	NodeInformer            corev1.NodeInformer
	PodInformer             corev1.PodInformer
	ServiceInformer         corev1.ServiceInformer
}

func NewControllerConfig(k8sClient *kubernetes.Clientset, crdClient *crdClientset.Clientset, resyncTime time.Duration) *ControllerConfig {
	crdFactory := crdFactory.NewSharedInformerFactory(crdClient, resyncTime)
	k8sFactory := informers.NewSharedInformerFactory(k8sClient, resyncTime)

	return &ControllerConfig{
		StatefulSetGridInformer: crdFactory.Superedge().V1().StatefulSetGrids(),
		StatefulSetInformer:     k8sFactory.Apps().V1().StatefulSets(),
		NodeInformer:            k8sFactory.Core().V1().Nodes(),
		PodInformer:             k8sFactory.Core().V1().Pods(),
		ServiceInformer:         k8sFactory.Core().V1().Services(),
	}
}

func (c *ControllerConfig) Run(stop <-chan struct{}) {
	go c.StatefulSetGridInformer.Informer().Run(stop)
	go c.NodeInformer.Informer().Run(stop)
	go c.StatefulSetInformer.Informer().Run(stop)
	go c.PodInformer.Informer().Run(stop)
	go c.ServiceInformer.Informer().Run(stop)
}
