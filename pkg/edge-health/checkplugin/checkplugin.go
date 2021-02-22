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
	"github.com/superedge/superedge/pkg/edge-health/metadata"
	"k8s.io/klog"
	"sync"
)

type CheckPlugin interface {
	CheckExecute(*metadata.CheckMetadata)
	Name() string
	SetWeight(float64)
	GetWeight() float64
}

type BasePlugin struct {
	HealthCheckoutTimeOut int
	HealthCheckRetries    int
	Weight                float64 // eg:0.3
	Port                  int
}

func (p BasePlugin) SetWeight(weight float64) {
	p.Weight = weight
}

func (p BasePlugin) GetWeight() float64 {
	return p.Weight
}

var (
	PluginOnce sync.Once
	PluginInfo Plugin
)

type Plugin struct {
	Plugins []CheckPlugin
}

func NewPlugin() Plugin {
	PluginOnce.Do(func() {
		PluginInfo = Plugin{
			Plugins: []CheckPlugin{},
		}
	})
	return PluginInfo
}

func (p *Plugin) Register(plugin CheckPlugin) {
	p.Plugins = append(p.Plugins, plugin)
	klog.V(4).Info("Register check plugin: %v", plugin)
}
