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

package checkplugin

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/superedge/superedge/pkg/edge-health/common"
	"github.com/superedge/superedge/pkg/edge-health/data"
	"k8s.io/klog/v2"
)

type PingCheckPlugin struct {
	BasePlugin
}

func (p PingCheckPlugin) Name() string {
	return "PingCheck"
}

func (p *PingCheckPlugin) Set(s string) error {
	var (
		err error
	)

	for _, para := range strings.Split(s, ",") {
		if len(para) == 0 {
			continue
		}
		arr := strings.Split(para, "=")
		trimkey := strings.TrimSpace(arr[0])
		switch trimkey {
		case "timeout":
			timeout, _ := strconv.Atoi(strings.TrimSpace(arr[1]))
			(*p).HealthCheckoutTimeOut = timeout
		case "retrytime":
			retrytime, _ := strconv.Atoi(strings.TrimSpace(arr[1]))
			(*p).HealthCheckRetryTime = retrytime
		case "weight":
			weight, _ := strconv.ParseFloat(strings.TrimSpace(arr[1]), 64)
			(*p).Weight = weight
		case "port":
			port, _ := strconv.Atoi(strings.TrimSpace(arr[1]))
			(*p).Port = port
		}
		(*p).PluginName = p.Name()
	}

	PluginInfo = NewPluginInfo()
	PluginInfo.AddPlugin(p)
	klog.V(4).Infof("len of plugins is %d", len(PluginInfo.Plugins))

	return err
}

func (p *PingCheckPlugin) String() string {
	return fmt.Sprintf("%v", *p)
}

func (i *PingCheckPlugin) Type() string {
	return "PingCheckPlugin"
}

func (plugin PingCheckPlugin) CheckExecute(wg *sync.WaitGroup) {
	var err error
	execwg := sync.WaitGroup{}
	execwg.Add(len(data.CheckInfoResult.CheckInfo))
	for k := range data.CheckInfoResult.CopyCheckInfo() {
		temp := k
		go func(execwg *sync.WaitGroup) {
			for i := 0; i < plugin.HealthCheckRetryTime; i++ {
				if _, err = net.DialTimeout("tcp", temp+":"+strconv.Itoa(plugin.Port), time.Duration(plugin.HealthCheckoutTimeOut)*time.Second); err == nil {
					break
				}
			}
			if err == nil {
				klog.V(4).Infof("%s use %s plugin check %s successd", common.NodeIP, plugin.Name(), temp)
				data.CheckInfoResult.SetCheckInfo(temp, plugin.Name(), plugin.GetWeight(), 100)
			} else {
				klog.V(2).Infof("%s use %s plugin check %s failed, reason: %s", common.NodeIP, plugin.Name(), temp, err.Error())
				data.CheckInfoResult.SetCheckInfo(temp, plugin.Name(), plugin.GetWeight(), 0)
			}
			execwg.Done()
		}(&execwg)
	}
	execwg.Wait()
	wg.Done()
}
