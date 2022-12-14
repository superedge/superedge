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

package check

import (
	"context"

	"github.com/superedge/superedge/pkg/edge-health/common"
	"github.com/superedge/superedge/pkg/edge-health/data"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	informcorev1 "k8s.io/client-go/informers/core/v1"
	listerv1 "k8s.io/client-go/listers/core/v1"

	"time"

	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

var ConfigMapManager *ConfigMapController

func NewConfigMapController(clientset kubernetes.Interface) *ConfigMapController {
	labelSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{EdgeHealthPodLabelKey: EdgeHealthPodLabelValue},
	}
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		panic("edge-health pod label selector error")
	}

	SharedInformerFactory := informers.NewSharedInformerFactory(clientset, 10*time.Minute)
	configMapInformer := SharedInformerFactory.InformerFor(&v1.ConfigMap{}, func(k kubernetes.Interface, duration time.Duration) cache.SharedIndexInformer {
		return informcorev1.NewFilteredConfigMapInformer(k, common.Namespace, duration,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
			func(options *metav1.ListOptions) {
				options.LabelSelector = selector.String()
			},
		)
	})

	n := &ConfigMapController{}
	configMapInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: n.handleConfigMapAddUpdate,
		UpdateFunc: func(old, cur interface{}) {
			n.handleConfigMapAddUpdate(cur)
		},
		DeleteFunc: n.handleConfigMapDelete,
	})
	n.clientset = clientset
	n.ConfigMapInformer = configMapInformer
	n.ConfigMapLister = listerv1.NewConfigMapLister(configMapInformer.GetIndexer())
	n.ConfigMapListerSynced = configMapInformer.HasSynced
	ConfigMapManager = n
	return n
}

type ConfigMapController struct {
	clientset             kubernetes.Interface
	ConfigMapInformer     cache.SharedIndexInformer
	ConfigMapLister       corelisters.ConfigMapLister
	ConfigMapListerSynced cache.InformerSynced
}

func (n *ConfigMapController) handleConfigMapAddUpdate(obj interface{}) {
	config, ok := obj.(*v1.ConfigMap)
	if !ok {
		return
	}
	klog.V(4).Infof("Add/Update ConfigMap %s", config.Name)
	data.ConfigMapList.SetConfigListData(*config)
}

func (n *ConfigMapController) handleConfigMapDelete(obj interface{}) {
	config, ok := obj.(*v1.ConfigMap)
	if !ok {
		return
	}
	klog.V(4).Infof("Delete ConfigMap %s", config.Name)
	data.ConfigMapList.DeleteConfigListData(*config)
}

func (n *ConfigMapController) Run(ctx context.Context) {
	defer runtimeutil.HandleCrash()

	go n.ConfigMapInformer.Run(ctx.Done())

	if ok := cache.WaitForCacheSync(
		ctx.Done(),
		n.ConfigMapListerSynced,
	); !ok {
		klog.Fatal("failed to wait for caches to sync")
	}

	//go wait.Until(func() { n.enqueueAll() }, common.ReListTime, ctx.Done())

	<-ctx.Done()
}
