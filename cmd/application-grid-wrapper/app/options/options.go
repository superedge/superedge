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
)

type Options struct {
	QPS                              float32
	Burst                            int
	CAFile                           string
	KeyFile                          string
	CertFile                         string
	KubeConfig                       string
	BindAddress                      string
	InsecureMode                     bool
	HostName                         string
	WrapperInCluster                 bool
	NotifyChannelSize                int
	Debug                            bool
	ServiceAutonomyEnhancementOption ServiceAutonomyEnhancementOptions
	SupportEndpointSlice             bool
}

func NewGridWrapperOptions() *Options {
	return &Options{
		QPS:               float32(1000),
		Burst:             1000,
		BindAddress:       "localhost:51006",
		InsecureMode:      true,
		NotifyChannelSize: 100,
		WrapperInCluster:  true,
		ServiceAutonomyEnhancementOption: ServiceAutonomyEnhancementOptions{
			Enabled:           false,
			UpdateInterval:    30,
			NeighborStatusSvc: "http://localhost:51005/localinfo",
		},
		SupportEndpointSlice: false,
	}
}

func (sc *Options) AddFlags(fs *pflag.FlagSet) {
	fs.Float32Var(&sc.QPS, "kube-qps", sc.QPS, "kubeconfig qps setting")
	fs.IntVar(&sc.Burst, "kube-burst", sc.Burst, "kubeconfig burst setting")
	fs.StringVar(&sc.CAFile, "ca-file", sc.CAFile,
		"Certificate Authority file for communication between wrapper and kube-proxy")
	fs.StringVar(&sc.KeyFile, "key-file", sc.KeyFile, "Private key file for communication between wrapper and kube-proxy")
	fs.StringVar(&sc.CertFile, "cert-file", sc.CertFile, "Certificate file for communication between wrapper and kube-proxy")
	fs.StringVar(&sc.KubeConfig, "kubeconfig", sc.KubeConfig, "kubeconfig for wrapper to communicate to apiserver")
	fs.StringVar(&sc.BindAddress, "bind-address", sc.BindAddress, "wrapper bind address, ip:port")
	fs.BoolVar(&sc.InsecureMode, "insecure", sc.InsecureMode,
		"if true, disable tls communication between wrapper and kube-proxy")
	fs.StringVar(&sc.HostName, "hostname", sc.HostName, "hostname for this wrapper")
	fs.BoolVar(&sc.WrapperInCluster, "wrapper-in-cluster", sc.WrapperInCluster, "wrapper k8s in cluster config")
	fs.IntVar(&sc.NotifyChannelSize, "notify-channel-size", sc.NotifyChannelSize,
		"channel size for service and endpoints sent")
	fs.BoolVar(&sc.Debug, "debug", sc.Debug, "enable pprof handler")
	fs.Var(&sc.ServiceAutonomyEnhancementOption, "service-autonomy-enhancement", "service-autonomy-enhancement")
	fs.BoolVar(&sc.SupportEndpointSlice, "support-endpointslice", sc.SupportEndpointSlice, "support EndpointSlice")
}
