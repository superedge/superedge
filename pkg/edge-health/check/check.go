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
	"superedge/pkg/edge-health/checkplugin"
	"superedge/pkg/edge-health/common"
	"superedge/pkg/edge-health/data"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/klog"
	"sync"
)

type Check interface {
	GetNodeList()
	Check()
	AddCheckPlugin(plugins []checkplugin.CheckPlugin)
	CheckPluginsLen() int
	GetHealthCheckPeriod() int
}

type CheckEdge struct {
	HealthCheckPeriod    int
	CheckPlugins         map[string]checkplugin.CheckPlugin
	HealthCheckScoreLine float64
}

func NewCheckEdge(checkplugins []checkplugin.CheckPlugin, healthcheckperiod int, healthCheckScoreLine float64) Check {
	m := make(map[string]checkplugin.CheckPlugin)
	for _, v := range checkplugins {
		m[v.Name()] = v
	}
	return CheckEdge{
		HealthCheckPeriod:    healthcheckperiod,
		HealthCheckScoreLine: healthCheckScoreLine,
		CheckPlugins:         m,
	}

}

func (c CheckEdge) GetNodeList() {
	var hostzone string
	var host *v1.Node

	masterSelector := labels.NewSelector()
	masterRequirement, err := labels.NewRequirement(common.MasterLabel, selection.DoesNotExist, []string{})
	if err != nil {
		klog.Errorf("can't new masterRequirement")
	}
	masterSelector = masterSelector.Add(*masterRequirement)

	if host, err = NodeManager.NodeLister.Get(common.HostName); err != nil {
		klog.Errorf("GetNodeList: can't get node with hostname %s, err: %v", common.HostName, err)
		return
	}

	if config, err := ConfigMapManager.ConfigMapLister.ConfigMaps("kube-system").Get(common.TaintZoneConfig); err != nil { //multi-region cm not found
		if apierrors.IsNotFound(err) {
			if NodeList, err := NodeManager.NodeLister.List(masterSelector); err != nil {
				klog.Errorf("config not exist, get nodes err: %v", err)
				return
			} else {
				data.NodeList.SetNodeListDataByNodeSlice(NodeList)
			}
		} else {
			klog.Errorf("get ConfigMaps edge-health-zone-config err %v", err)
			return
		}
	} else { //multi-region cm found
		klog.V(4).Infof("cm value is %s", config.Data["TaintZoneAdmission"])
		if config.Data["TaintZoneAdmission"] == "false" { //close multi-region check
			if NodeList, err := NodeManager.NodeLister.List(masterSelector); err != nil {
				klog.Errorf("config exist, false, get nodes err : %v", err)
				return
			} else {
				data.NodeList.SetNodeListDataByNodeSlice(NodeList)
			}
		} else { //open multi-region check
			if _, ok := host.Labels[common.TopologyZone]; ok {
				hostzone = host.Labels[common.TopologyZone]
				klog.V(4).Infof("hostzone is %s", hostzone)

				masterzoneSelector := labels.NewSelector()
				zoneRequirement, err := labels.NewRequirement(common.TopologyZone, selection.Equals, []string{hostzone})
				if err != nil {
					klog.Errorf("can't new zoneRequirement")
				}
				masterzoneSelector = masterzoneSelector.Add(*masterRequirement, *zoneRequirement)
				if NodeList, err := NodeManager.NodeLister.List(masterzoneSelector); err != nil {
					klog.Errorf("config exist, true, host has zone label, get nodes err: %v", err)
					return
				} else {
					data.NodeList.SetNodeListDataByNodeSlice(NodeList)
				}
				klog.V(4).Infof("nodelist len is %d", data.NodeList.GetLenListData())
			} else {
				data.NodeList.SetNodeListDataByNodeSlice([]*v1.Node{host})
			}
		}
	}

	iplist := make(map[string]bool)
	tempItems := data.NodeList.CopyNodeListData()
	for _, v := range tempItems {
		for _, i := range v.Status.Addresses {
			if i.Type == v1.NodeInternalIP {
				iplist[i.Address] = true
				data.CheckInfoResult.SetCheckedIpCheckInfo(i.Address)
			}
		}
	}

	for _, v := range data.CheckInfoResult.TraverseCheckedIpCheckInfo() {
		if _, ok := iplist[v]; !ok {
			data.CheckInfoResult.DeleteCheckedIpCheckInfo(v)
		}
	}

	for k := range data.Result.CopyResultDataAll() {
		if _, ok := iplist[k]; !ok {
			data.Result.DeleteResultData(k)
		}
	}

	klog.V(4).Infof("GetNodeList: checkinfo is %v", data.CheckInfoResult)
}

func (c CheckEdge) CheckPluginsLen() int {
	return len(c.CheckPlugins)
}

func (c CheckEdge) GetHealthCheckPeriod() int {
	return c.HealthCheckPeriod
}

func (c CheckEdge) Check() {
	wg := sync.WaitGroup{}
	wg.Add(c.CheckPluginsLen())
	for _, plugin := range c.GetCheckPlugins() {
		go plugin.CheckExecute(&wg)
	}
	wg.Wait()
	klog.V(4).Info("check finished")
	klog.V(4).Infof("healthcheck: after health check, checkinfo is %v", data.CheckInfoResult.CheckInfo)

	calculatetemp := data.CheckInfoResult.CopyCheckInfo()
	for desip, plugins := range calculatetemp {
		totalscore := 0.0
		for _, score := range plugins {
			totalscore += score
		}
		if totalscore >= c.HealthCheckScoreLine {
			data.Result.SetResultFromCheckInfo(common.LocalIp, desip, data.ResultDetail{Normal: true})
		} else {
			data.Result.SetResultFromCheckInfo(common.LocalIp, desip, data.ResultDetail{Normal: false})
		}
	}
	klog.V(4).Infof("healthcheck: after health check, result is %v", data.Result.Result)
}

func (c CheckEdge) AddCheckPlugin(plugins []checkplugin.CheckPlugin) {
	for _, p := range plugins {
		c.CheckPlugins[p.Name()] = p
	}
}

func (c CheckEdge) GetCheckPlugins() map[string]checkplugin.CheckPlugin {
	return c.CheckPlugins
}
