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
	"fmt"
	"github.com/spf13/pflag"
	"github.com/superedge/superedge/pkg/edge-health/checkplugin"
)

type CheckOptions struct {
	CheckPeriod            int
	CheckScoreLine         float64
	PingCheckPlugin        checkplugin.PingCheckPlugin
	KubeletCheckPlugin     checkplugin.KubeletCheckPlugin
	KubeletAuthCheckPlugin checkplugin.KubeletAuthCheckPlugin
}

func NewCheckOptions() *CheckOptions {
	return &CheckOptions{}
}

func (o *CheckOptions) Default() {
	o.CheckScoreLine = 100.0
	o.CheckPeriod = 10
}

func (o *CheckOptions) Validate() []error {
	var errs []error
	if o.CheckPeriod <= 0 {
		errs = append(errs, fmt.Errorf("Invalid health check period %d", o.CheckPeriod))
	}
	if o.CheckScoreLine <= 0 || o.CheckScoreLine > 100 {
		errs = append(errs, fmt.Errorf("Invalid health check score line %f", o.CheckScoreLine))
	}
	// Total weight of edge health check plugins must be 1 since CheckScoreLine is in the range of (0, 100]
	totalWeight := o.PingCheckPlugin.GetWeight() + o.KubeletCheckPlugin.GetWeight() + o.KubeletAuthCheckPlugin.GetWeight()
	if totalWeight != 1 {
		errs = append(errs, fmt.Errorf("Invalid health check plugins total weight %f", totalWeight))
	}
	return errs
}

func (o *CheckOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.Var(&o.KubeletCheckPlugin, "kubelet-plugin", "Kubelet check plugin")
	fs.Var(&o.PingCheckPlugin, "ping-plugin", "Ping check plugin")
	fs.Var(&o.KubeletAuthCheckPlugin, "kubelet-auth-plugin", "Kubelet token check plugin")
	fs.IntVar(&o.CheckPeriod, "health-check-period", o.CheckPeriod, "Period of Health check")
	fs.Float64Var(&o.CheckScoreLine, "health-check-scoreline", o.CheckScoreLine, "Health check score line(in the range of (0, 100])")
}
