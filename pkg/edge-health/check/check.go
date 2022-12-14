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
	"sync"

	"github.com/superedge/superedge/pkg/edge-health/checkplugin"
	"github.com/superedge/superedge/pkg/edge-health/common"
	"github.com/superedge/superedge/pkg/edge-health/data"
	siteconst "github.com/superedge/superedge/pkg/site-manager/constant"
	"github.com/superedge/superedge/pkg/util"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/klog/v2"
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
	var host *metav1.PartialObjectMetadata

	masterSelector := labels.NewSelector()
	masterRequirement, err := labels.NewRequirement(common.MasterLabel, selection.DoesNotExist, []string{})
	if err != nil {
		klog.Errorf("can't new masterRequirement")
	}
	masterSelector = masterSelector.Add(*masterRequirement)

	hostObject, err := NodeMetaManager.NodeMetaILister.Get(common.NodeName)
	if err != nil {
		klog.Errorf("NodeMetaILister: can't get node with hostname %s, err: %v", common.NodeName, err)
		return
	}
	host = hostObject.(*metav1.PartialObjectMetadata)
	if config, err := ConfigMapManager.ConfigMapLister.ConfigMaps(util.PodNamespace).Get(common.TaintZoneConfig); err != nil { //multi-region cm not found
		if apierrors.IsNotFound(err) {
			if NodeList, err := NodeMetaManager.NodeMetaILister.List(masterSelector); err != nil {
				klog.Errorf("config not exist, get nodes err: %v", err)
				return
			} else {
				data.NodeList.SetNodeListDataByNodeSlice(NodeList)
			}
		} else {
			klog.Errorf("get ConfigMaps edge-health-zone-config err %v", err)
			return
		}
	} else { //node unit check cm found
		klog.V(4).Infof("cm value is %s", config.Data["TaintZoneAdmission"])
		if config.Data["TaintZoneAdmission"] == "false" { //close multi-region check
			if NodeList, err := NodeMetaManager.NodeMetaILister.List(masterSelector); err != nil {
				klog.Errorf("config exist, false, get nodes err : %v", err)
				return
			} else {
				data.NodeList.SetNodeListDataByNodeSlice(NodeList)
			}
		} else { //open node unit check
			// only check same unit node
			unitLabel := make(map[string]string, 1)
			for k, v := range host.Labels {
				if v == siteconst.NodeUnitSuperedge {
					unitLabel[k] = v
				}
			}
			klog.V(6).Infof("unitLabel is %s", unitLabel)

			if len(unitLabel) > 0 {
				labelSelector := &metav1.LabelSelector{
					MatchLabels: unitLabel,
				}
				selector, err := metav1.LabelSelectorAsSelector(labelSelector)
				if err != nil {
					klog.ErrorS(err, "metav1.LabelSelectorAsSelector error")
					return
				}

				if NodeList, err := NodeMetaManager.NodeMetaILister.List(selector); err != nil {
					klog.Errorf("config exist, true, host has zone label, get nodes err: %v", err)
					return
				} else {
					data.NodeList.SetNodeListDataByNodeSlice(NodeList)
				}
				klog.V(6).Infof("nodelist len is %d", data.NodeList.GetLenListData())

			} else {
				// could not find unit label, ondy check self
				data.NodeList.SetNodeListDataByNodeSlice([]runtime.Object{hostObject})
			}
		}
	}
	// TODO get ip list from node meta and pod hostIP
	iplist := make(map[string]bool)
	tempItems := data.NodeList.CopyNodeListData()
	for _, node := range tempItems {
		nodeIP, err := PodManager.GetNodeIPByNodeName(node.Name)
		if err != nil {
			klog.ErrorS(err, "GetNodeIPByNodeName error")
			continue
		}
		iplist[nodeIP] = true
		data.CheckInfoResult.SetCheckedIpCheckInfo(nodeIP)
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
			data.Result.SetResultFromCheckInfo(common.NodeIP, desip, data.ResultDetail{Normal: true})
		} else {
			data.Result.SetResultFromCheckInfo(common.NodeIP, desip, data.ResultDetail{Normal: false})
		}
	}
	klog.V(6).Infof("healthcheck: after health check, result is %v", data.Result.GetResultDataAll())
}

func (c CheckEdge) AddCheckPlugin(plugins []checkplugin.CheckPlugin) {
	for _, p := range plugins {
		c.CheckPlugins[p.Name()] = p
	}
}

func (c CheckEdge) GetCheckPlugins() map[string]checkplugin.CheckPlugin {
	return c.CheckPlugins
}
