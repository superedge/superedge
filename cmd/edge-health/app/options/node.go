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
	"runtime"
)

type NodeOptions struct {
	MasterUrl      string
	KubeconfigPath string
	HostName       string // Node name
	SyncPeriod     int
	// qps controls the number of queries per second allowed for this connection.
	QPS float32
	// burst allows extra queries to accumulate when a client is exceeding its rate.
	Burst  int
	Worker int
}

func NewNodeOptions() *NodeOptions {
	return &NodeOptions{}
}

func (o *NodeOptions) Default() {
	o.QPS = float32(1000)
	o.Burst = 1000
	o.SyncPeriod = 30
	o.Worker = runtime.NumCPU()
}

func (o *NodeOptions) Validate() []error {
	var errs []error
	if o.HostName == "" {
		errs = append(errs, fmt.Errorf("Invalid hostname %s", o.HostName))
	}
	return errs
}

func (o *NodeOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.StringVar(&o.MasterUrl, "master", o.MasterUrl, "Kube-apiserver master url")
	fs.StringVar(&o.KubeconfigPath, "kubeconfig", o.KubeconfigPath, "Kubeconfig path")
	fs.StringVar(&o.HostName, "hostname", o.HostName, "Host name")
	fs.IntVar(&o.SyncPeriod, "sync-period", o.SyncPeriod, "Period for syncing the objects")
	fs.Float32Var(&o.QPS, "kube-api-qps", o.QPS, "QPS to use while talking with kubernetes apiserver.")
	fs.IntVar(&o.Burst, "kube-api-burst", o.Burst, "Burst to use while talking with kubernetes apiserver.")
	fs.IntVar(&o.Worker, "worker", o.Worker, "Worker number of controller")
}
