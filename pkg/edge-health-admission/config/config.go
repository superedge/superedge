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

package config

import (
	"crypto/tls"
	"flag"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// Config contains the server (the webhook) cert and key.
type Config struct {
	CertFile       string
	KeyFile        string
	KubeconfigPath string
	MasterUrl      string
}

var Kubeclient clientset.Interface
var NodeAlwaysReachable bool

func (c *Config) AddFlags() {
	flag.StringVar(&c.CertFile, "admission-control-server-cert", c.CertFile, ""+
		"File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated "+
		"after server cert).")
	flag.StringVar(&c.KeyFile, "admission-control-server-key", c.KeyFile, ""+
		"File containing the default x509 private key matching --tls-cert-file.")
	flag.StringVar(&c.KubeconfigPath, "kubeconfigpath", c.KubeconfigPath, "")
	flag.StringVar(&c.MasterUrl, "masterurl", c.MasterUrl, "")
}

func ConfigTLS(config Config) *tls.Config {
	sCert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		klog.Fatal(err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{sCert},
	}
}
