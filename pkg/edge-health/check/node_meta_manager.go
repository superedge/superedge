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

	"github.com/superedge/superedge/pkg/edge-health/data"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/client-go/metadata"
	"k8s.io/client-go/metadata/metadatainformer"
	"k8s.io/client-go/tools/cache"

	"time"

	"k8s.io/klog/v2"
)

var NodeMetaManager *NodeMetaController

func NewNodeMetaController(metaClientset metadata.Interface) *NodeMetaController {
	sif := metadatainformer.NewSharedInformerFactory(
		metaClientset,
		10*time.Minute,
	)
	// only watch node metadata
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"}
	genericInformer := sif.ForResource(gvr)

	nodeMetaInformer := genericInformer.Informer()
	n := &NodeMetaController{}
	nodeMetaInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: n.handleNodeAddUpdate,
		UpdateFunc: func(old, cur interface{}) {
			n.handleNodeAddUpdate(cur)
		},
		DeleteFunc: n.handleNodeDelete,
	})
	n.MetaClientset = metaClientset
	n.NodeMetaInformer = nodeMetaInformer
	n.NodeMetaILister = genericInformer.Lister()
	n.NodeMetaInformerSynced = nodeMetaInformer.HasSynced
	NodeMetaManager = n
	return n
}

type NodeMetaController struct {
	MetaClientset          metadata.Interface
	NodeMetaInformer       cache.SharedIndexInformer
	NodeMetaILister        cache.GenericLister
	NodeMetaInformerSynced cache.InformerSynced
}

func (n *NodeMetaController) handleNodeAddUpdate(obj interface{}) {
	node, ok := obj.(*metav1.PartialObjectMetadata)
	if !ok {
		return
	}
	klog.V(4).Infof("Add/Update Node meta %s", node.Name)
	data.NodeList.SetNodeListDataByNode(*node)
}

func (n *NodeMetaController) handleNodeDelete(obj interface{}) {
	node, ok := obj.(*metav1.PartialObjectMetadata)
	if !ok {
		return
	}
	klog.V(4).Infof("Delete Node meta %s", node.Name)
	data.NodeList.DeleteNodeListDataByNode(*node)
}

func (n *NodeMetaController) Run(ctx context.Context) {
	defer runtimeutil.HandleCrash()

	go n.NodeMetaInformer.Run(ctx.Done())

	if ok := cache.WaitForCacheSync(
		ctx.Done(),
		n.NodeMetaInformerSynced,
	); !ok {
		klog.Fatal("failed to wait for caches to sync")
	}

	<-ctx.Done()
}
