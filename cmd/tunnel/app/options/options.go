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
	cliflag "k8s.io/component-base/cli/flag"
)

type TunnelOption struct {
	TunnelMode *string
	TunnelConf *string
	Kubeconfig *string
	Debug      *bool
	QPS        float32
	Burst      int
}

func NewTunnelOption() *TunnelOption {
	c := "../../conf/cloud_mode.toml"
	m := "cloud"
	k := ""
	d := false
	return &TunnelOption{
		TunnelMode: &m,
		TunnelConf: &c,
		Kubeconfig: &k,
		Debug:      &d,
		QPS:        1000,
		Burst:      1000,
	}
}
func (option *TunnelOption) Addflag() (fsSet cliflag.NamedFlagSets) {
	fs := fsSet.FlagSet("tunnel")
	fs.StringVar(option.TunnelMode, "m", *option.TunnelMode, "Specify the edge proxy or cloud proxy")
	fs.StringVar(option.TunnelConf, "c", *option.TunnelConf, "Specify the configuration file path")
	fs.StringVar(option.Kubeconfig, "k", *option.Kubeconfig, "Specify the kubeconfig")
	fs.Float32Var(&option.QPS, "kube-qps", option.QPS, "kubeconfig qps setting")
	fs.IntVar(&option.Burst, "kube-burst", option.Burst, "kubeconfig burst setting")
	return
}
