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
	"flag"
	"github.com/spf13/pflag"
	"time"
)

type Options struct {
	EnableAdmissionControl     bool
	AdmissionControlServerCert string
	AdmissionControlServerKey  string
	AdmissionControlListenAddr string
	QPS                        int
	Burst                      int
	LeaseDuration              time.Duration
	RenewDeadline              time.Duration
	RetryPeriod                time.Duration
	KubeConfig                 string
}

func NewOperatorOptions() *Options {
	return &Options{
		EnableAdmissionControl:     true,
		AdmissionControlListenAddr: "0.0.0.0:9000",
		QPS:                        30,
		Burst:                      60,
		LeaseDuration:              40 * time.Second,
		RenewDeadline:              30 * time.Second,
		RetryPeriod:                2 * time.Second,
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	flag.BoolVar(&o.EnableAdmissionControl, "enable-admission-control", o.EnableAdmissionControl, "")
	flag.StringVar(&o.AdmissionControlServerCert, "admission-control-server-cert", "", "")
	flag.StringVar(&o.AdmissionControlServerKey, "admission-control-server-key", "", "")
	flag.StringVar(&o.AdmissionControlListenAddr, "adminssion-control-listen-addr", o.AdmissionControlListenAddr, "")
	flag.IntVar(&o.QPS, "client-request-qps", o.QPS, "client request qps for client-go")
	flag.IntVar(&o.Burst, "client-request-burst", o.Burst, "client request burst for client-go")
	flag.DurationVar(&o.LeaseDuration, "leader-elcct-lease-duration", o.LeaseDuration, "leader elect lease duration")
	flag.DurationVar(&o.RenewDeadline, "leader-elect-renew-deadline", o.RenewDeadline, "leader elect renew deadline")
	flag.DurationVar(&o.RetryPeriod, "leader-elect-retry-period", o.RetryPeriod, "leader elect retry period")
	flag.StringVar(&o.KubeConfig, "kubeconfig", o.KubeConfig, "kubeconfig path, empty path means in cluster mode")
}
