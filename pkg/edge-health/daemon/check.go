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
	"context"
	"github.com/superedge/superedge/pkg/edge-health/common"
	"github.com/superedge/superedge/pkg/edge-health/metadata"
	"github.com/superedge/superedge/pkg/edge-health/util"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/klog"
)

func (ehd *EdgeHealthDaemon) SyncNodeList() {
	// Only sync nodes when self-located found
	var host *v1.Node
	if host = ehd.metadata.GetNodeByName(ehd.cfg.Node.HostName); host == nil {
		klog.Errorf("Self-hostname %s not found", ehd.cfg.Node.HostName)
		return
	}

	// Filter cloud nodes and retain edge ones
	masterRequirement, err := labels.NewRequirement(common.MasterLabel, selection.DoesNotExist, []string{})
	if err != nil {
		klog.Errorf("New masterRequirement failed %v", err)
		return
	}
	masterSelector := labels.NewSelector()
	masterSelector = masterSelector.Add(*masterRequirement)

	if mrc, err := ehd.cmLister.ConfigMaps(metav1.NamespaceSystem).Get(common.TaintZoneConfigMap); err != nil {
		if apierrors.IsNotFound(err) { // multi-region configmap not found
			if NodeList, err := ehd.nodeLister.List(masterSelector); err != nil {
				klog.Errorf("Multi-region configmap not found and get nodes err %v", err)
				return
			} else {
				ehd.metadata.SetByNodeList(NodeList)
			}
		} else {
			klog.Errorf("Get multi-region configmap err %v", err)
			return
		}
	} else { // multi-region configmap found
		mrcv := mrc.Data[common.TaintZoneConfigMapKey]
		klog.V(4).Infof("Multi-region value is %s", mrcv)
		if mrcv == "false" { // close multi-region check
			if NodeList, err := ehd.nodeLister.List(masterSelector); err != nil {
				klog.Errorf("Multi-region configmap exist but disabled and get nodes err %v", err)
				return
			} else {
				ehd.metadata.SetByNodeList(NodeList)
			}
		} else { // open multi-region check
			if hostZone, existed := host.Labels[common.TopologyZone]; existed {
				klog.V(4).Infof("Host %s has HostZone %s", host.Name, hostZone)
				zoneRequirement, err := labels.NewRequirement(common.TopologyZone, selection.Equals, []string{hostZone})
				if err != nil {
					klog.Errorf("New masterZoneRequirement failed: %v", err)
					return
				}
				masterZoneSelector := labels.NewSelector()
				masterZoneSelector = masterZoneSelector.Add(*masterRequirement, *zoneRequirement)
				if nodeList, err := ehd.nodeLister.List(masterZoneSelector); err != nil {
					klog.Errorf("TopologyZone label for hostname %s but get nodes err: %v", host.Name, err)
					return
				} else {
					ehd.metadata.SetByNodeList(nodeList)
				}
			} else { // Only check itself if there is no TopologyZone label
				klog.V(4).Infof("Only check itself since there is no TopologyZone label for hostname %s", host.Name)
				ehd.metadata.SetByNodeList([]*v1.Node{host})
			}
		}
	}

	// Init check plugin score
	ipList := make(map[string]struct{})
	for _, node := range ehd.metadata.Copy() {
		for _, addr := range node.Status.Addresses {
			if addr.Type == v1.NodeInternalIP {
				ipList[addr.Address] = struct{}{}
				ehd.metadata.InitCheckPluginScore(addr.Address)
			}
		}
	}

	// Delete redundant check plugin score
	for _, checkedIp := range ehd.metadata.CopyCheckedIp() {
		if _, existed := ipList[checkedIp]; !existed {
			ehd.metadata.DeleteCheckPluginScore(checkedIp)
		}
	}

	// Delete redundant check info
	for checkerIp := range ehd.metadata.CopyAll() {
		if _, existed := ipList[checkerIp]; !existed {
			ehd.metadata.DeleteByIp(ehd.cfg.Node.LocalIp, checkerIp)
		}
	}

	klog.V(4).Infof("SyncNodeList check info %+v successfully", ehd.metadata)
}

func (ehd *EdgeHealthDaemon) ExecuteCheck() {
	util.ParallelizeUntil(context.TODO(), 16, len(ehd.checkPlugin.Plugins), func(index int) {
		ehd.checkPlugin.Plugins[index].CheckExecute(ehd.metadata.CheckMetadata)
	})
	klog.V(4).Infof("CheckPluginScoreInfo is %v after health check", ehd.metadata.CheckPluginScoreInfo)

	for checkedIp, pluginScores := range ehd.metadata.CopyCheckPluginScore() {
		totalScore := 0.0
		for _, score := range pluginScores {
			totalScore += score
		}
		if totalScore >= ehd.cfg.Check.HealthCheckScoreLine {
			ehd.metadata.SetByCheckDetail(ehd.cfg.Node.LocalIp, checkedIp, metadata.CheckDetail{Normal: true})
		} else {
			ehd.metadata.SetByCheckDetail(ehd.cfg.Node.LocalIp, checkedIp, metadata.CheckDetail{Normal: false})
		}
	}
	klog.V(4).Infof("CheckInfo is %v after health check", ehd.metadata.CheckInfo)
}
