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
	"io/ioutil"
	"net/http"
	"sync"

	"k8s.io/klog/v2"
)

type CheckPlugin interface {
	CheckExecute(wg *sync.WaitGroup)
	Name() string
	SetWeight(float64)
	GetWeight() float64
}
type PluginFactory = func() (CheckPlugin, error)

type Registry map[string]PluginFactory

func Merge(outOfTree Registry) error {
	for name, factory := range outOfTree {
		checkplugin, err := factory()
		if err != nil {
			return err
		}
		PluginInfo = NewPluginInfo()
		PluginInfo.AddPlugin(checkplugin)
		klog.V(4).Info("add plugin success", name)
	}
	return nil
}

type BasePlugin struct {
	PluginName            string
	HealthCheckoutTimeOut int
	HealthCheckRetryTime  int
	Weight                float64 //ex:0.3
	Port                  int
}

func (p BasePlugin) SetWeight(weight float64) {
	p.Weight = weight
}

func (p BasePlugin) GetWeight() float64 {
	return p.Weight
}

func NewBasePlugin(healthCheckoutTimeOut, retrytime, port int, weight float64, pluginname string) BasePlugin {
	return BasePlugin{
		PluginName:            pluginname,
		HealthCheckoutTimeOut: healthCheckoutTimeOut,
		HealthCheckRetryTime:  retrytime,
		Weight:                weight,
		Port:                  port,
	}
}

var (
	PluginOnce sync.Once
	PluginMu   sync.Mutex
	PluginInfo Plugin
)

type Plugin struct {
	Plugins []CheckPlugin
}

func NewPluginInfo() Plugin {
	PluginOnce.Do(func() {
		PluginInfo = Plugin{
			Plugins: []CheckPlugin{},
		}
	})
	return PluginInfo
}

func (p *Plugin) AddPlugin(plugin CheckPlugin) {
	PluginMu.Lock()
	defer PluginMu.Unlock()
	// TODO: should check plugin is it exist or not, now have a logic bug
	// plugins can be add many times since it is only a slice, you could check on the ut TestMerge
	p.Plugins = append(p.Plugins, plugin)
	klog.V(4).Info("add ok")
}

func PingDo(client http.Client, req *http.Request) (bool, error) {
	var (
		response []byte
		err      error
	)
	resp, err := client.Do(req)
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		if response, err = ioutil.ReadAll(resp.Body); err == nil {
			err = fmt.Errorf("ping failed, StatusCode is %d, resp body is %s", resp.StatusCode, string(response))
		}
		return false, err
	}
	klog.V(4).Info("ping succeed")
	return true, nil
}
