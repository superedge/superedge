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

package daemon

import (
	"fmt"
	"github.com/superedge/superedge/pkg/edge-health/checkplugin"
	"github.com/superedge/superedge/pkg/edge-health/commun"
	"github.com/superedge/superedge/pkg/edge-health/config"
	"github.com/superedge/superedge/pkg/edge-health/metadata"
	"github.com/superedge/superedge/pkg/edge-health/vote"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

type EdgeHealthDaemon struct {
	cfg         *config.EdgeHealthConfig
	metadata    *metadata.EdgeHealthMetadata
	checkPlugin checkplugin.Plugin
	nodeLister  corelisters.NodeLister
	cmLister    corelisters.ConfigMapLister
}

func NewEdgeHealthDaemon(c *config.EdgeHealthConfig) *EdgeHealthDaemon {
	ehd := &EdgeHealthDaemon{
		cfg:         c,
		metadata:    metadata.NewEdgeHealthMetadata(),
		checkPlugin: checkplugin.NewPlugin(),
		nodeLister:  c.NodeInformer.Lister(),
		cmLister:    c.ConfigMapInformer.Lister(),
	}
	ehd.cfg.NodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node := obj.(*corev1.Node)
			klog.V(4).Infof("Add Node %s", node.Name)
			ehd.metadata.NodeMetadata.AddOrUpdateByNode(node)
		},
		UpdateFunc: func(old, new interface{}) {
			newNode := new.(*corev1.Node)
			oldNode := old.(*corev1.Node)
			klog.V(4).Infof("Update Node %s", newNode.Name)
			if newNode.ResourceVersion == oldNode.ResourceVersion {
				// Periodic resync will send update events for all known Nodes.
				// Two different versions of the same Pod will always have different RVs.
				return
			}
			ehd.metadata.NodeMetadata.AddOrUpdateByNode(newNode)
		},
		DeleteFunc: func(obj interface{}) {
			node, ok := obj.(*corev1.Node)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
					return
				}
				node, ok = tombstone.Obj.(*corev1.Node)
				if !ok {
					utilruntime.HandleError(fmt.Errorf("Tombstone contained object is not a Node %#v", obj))
					return
				}
			}
			ehd.metadata.NodeMetadata.DeleteByNode(node)
		},
	})
	return ehd
}

func (ehd *EdgeHealthDaemon) Run(stopCh <-chan struct{}) {
	// Execute edge health prepare and check
	ehd.PrepareAndCheck(stopCh)

	// Execute vote
	vote := vote.NewVoteEdge(&ehd.cfg.Vote)
	go vote.Vote(ehd.metadata, ehd.cfg.Kubeclient, ehd.cfg.Node.LocalIp, stopCh)

	// Execute communication
	communEdge := commun.NewCommunEdge(&ehd.cfg.Commun)
	communEdge.Commun(ehd.metadata.CheckMetadata, ehd.cmLister, ehd.cfg.Node.LocalIp, stopCh)

	<-stopCh
}
