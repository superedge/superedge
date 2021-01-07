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

package options

import (
	"github.com/spf13/pflag"
	"github.com/superedge/superedge/pkg/edge-health/checkplugin"
)

type CheckOptions struct {
	Checks               []string //check method
	HealthCheckPeriod    int
	HealthCheckScoreLine float64
}

func NewCheckOptions() *CheckOptions {
	return &CheckOptions{}
}

func (o *CheckOptions) Default() {
	o.HealthCheckScoreLine = 100.0
	o.HealthCheckPeriod = 10
}

func (o *CheckOptions) Validate() []error {
	return nil
}

func (o *CheckOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	var kubeletcheckplugin checkplugin.KubeletCheckPlugin
	fs.Var(&kubeletcheckplugin, "kubeletplugin", "kubelet plugin")

	var kubeletauthcheckplugin checkplugin.KubeletAuthCheckPlugin
	fs.Var(&kubeletauthcheckplugin, "kubeletauthplugin", "kubelet token plugin")

	var pingcheckplugin checkplugin.PingCheckPlugin
	fs.Var(&pingcheckplugin, "pingplugin", "ping plugin")

	fs.IntVar(&o.HealthCheckPeriod, "healthcheckperiod", o.HealthCheckPeriod, "health check period")
	fs.Float64Var(&o.HealthCheckScoreLine, "healthcheckscoreline", o.HealthCheckScoreLine, "health check score line")
}
