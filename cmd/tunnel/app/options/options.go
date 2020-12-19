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
	Debug      *bool
}

func NewTunnelOption() *TunnelOption {
	c := "../../conf/cloud_mode.toml"
	m := "cloud"
	d := false
	return &TunnelOption{
		TunnelMode: &m,
		TunnelConf: &c,
		Debug:      &d,
	}
}
func (option *TunnelOption) Addflag() (fsSet cliflag.NamedFlagSets) {
	fs := fsSet.FlagSet("tunnel")
	fs.StringVar(option.TunnelMode, "m", *option.TunnelMode, "Specify the edge proxy or cloud proxy")
	fs.StringVar(option.TunnelConf, "c", *option.TunnelConf, "Specify the configuration file path")
	return
}
