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
	appv1 "k8s.io/client-go/informers/apps/v1"
	corev1 "k8s.io/client-go/informers/core/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	crdClientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	crdFactory "github.com/superedge/superedge/pkg/site-manager/generated/informers/externalversions"
	crdv1 "github.com/superedge/superedge/pkg/site-manager/generated/informers/externalversions/site.superedge.io/v1alpha2"
)

type ControllerConfig struct {
	NodeInformer      corev1.NodeInformer
	DaemonSetInformer appv1.DaemonSetInformer
	NodeUnitInformer  crdv1.NodeUnitInformer
	NodeGroupInformer crdv1.NodeGroupInformer
}

func NewControllerConfig(k8sClient *kubernetes.Clientset, crdClient *crdClientset.Clientset, resyncTime time.Duration) *ControllerConfig {
	crdFactory := crdFactory.NewSharedInformerFactory(crdClient, resyncTime)
	k8sFactory := informers.NewSharedInformerFactory(k8sClient, resyncTime)

	return &ControllerConfig{
		NodeInformer:      k8sFactory.Core().V1().Nodes(),
		DaemonSetInformer: k8sFactory.Apps().V1().DaemonSets(),
		NodeUnitInformer:  crdFactory.Site().V1alpha2().NodeUnits(),
		NodeGroupInformer: crdFactory.Site().V1alpha2().NodeGroups(),
	}
}

func (c *ControllerConfig) Run(stop <-chan struct{}) {
	// TODO need run factory and wait cache sync in site controller run not here
	go c.NodeInformer.Informer().Run(stop)
	go c.DaemonSetInformer.Informer().Run(stop)
	go c.NodeUnitInformer.Informer().Run(stop)
	go c.NodeGroupInformer.Informer().Run(stop)

	klog.V(4).Infof("Site-manager set informer success")
}
