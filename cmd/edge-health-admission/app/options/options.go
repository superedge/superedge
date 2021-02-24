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
	"net"
)

type Options struct {
	Master     string
	Kubeconfig string
	QPS        float32
	Burst      int
	SyncPeriod int
	KeyFile    string
	CertFile   string
	Addr       string
}

func NewAdmissionOptions() *Options {
	return &Options{
		QPS:        float32(1000),
		Burst:      1000,
		SyncPeriod: 30,
	}
}

func (o *Options) Validate() []error {
	var errs []error
	if net.ParseIP(o.Addr) == nil {
		errs = append(errs, fmt.Errorf("Invalid admission webhook listen address %s", o.Addr))
	}
	return errs
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Master, "master", o.Master, "apiserver master address")
	fs.StringVar(&o.Kubeconfig, "kubeconfig", o.Kubeconfig, "kubeconfig path, empty path means in cluster mode")
	fs.Float32Var(&o.QPS, "kube-qps", o.QPS, "kubeconfig qps setting")
	fs.IntVar(&o.Burst, "kube-burst", o.Burst, "kubeconfig burst setting")
	fs.IntVar(&o.SyncPeriod, "sync-period", o.SyncPeriod, "Period for syncing the objects")
	fs.StringVar(&o.CertFile, "admission-control-server-cert", o.CertFile, ""+
		"File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated "+
		"after server cert).")
	fs.StringVar(&o.KeyFile, "admission-control-server-key", o.KeyFile, ""+
		"File containing the default x509 private key matching --tls-cert-file.")
	fs.StringVar(&o.Addr, "adminssion-control-listen-addr", ":8443", "admission webhook listen address")
}

// CompletedOptions is a wrapper that enforces a call of Complete() before Run can be invoked.
type CompletedOptions struct {
	*Options
}

func Complete(o *Options) (CompletedOptions, error) {
	return CompletedOptions{o}, nil
}
