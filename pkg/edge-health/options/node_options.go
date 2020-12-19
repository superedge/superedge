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

import "github.com/spf13/pflag"

type NodeOptions struct {
	MasterUrl      string //apiserver url
	KubeconfigPath string
	HostName       string //node name
}

func NewNodeOptions() *NodeOptions {
	return &NodeOptions{}
}

func (o *NodeOptions) Validate() []error {
	return nil
}

func (o *NodeOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.StringVar(&o.MasterUrl, "masterurl", o.MasterUrl, "master url")
	fs.StringVar(&o.KubeconfigPath, "kubeconfig", o.KubeconfigPath, "kubeconfig path")
	fs.StringVar(&o.HostName, "hostname", o.HostName, "host name")
}
