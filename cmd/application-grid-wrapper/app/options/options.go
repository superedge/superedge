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
	CAFile            string
	KeyFile           string
	CertFile          string
	KubeConfig        string
	BindAddress       string
	InsecureMode      bool
	HostName          string
	WrapperInCluster  bool
	NotifyChannelSize int
	Debug             bool
}

func NewGridWrapperOptions() *Options {
	return &Options{
		BindAddress:       "localhost:51006",
		InsecureMode:      true,
		NotifyChannelSize: 100,
		WrapperInCluster:  true,
	}
}

func (sc *Options) AddFlags(fs *pflag.FlagSet) {
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
}
