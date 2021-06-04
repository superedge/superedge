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

// Package options contains all of the primary arguments for a kubelet.
package options

import (
	"encoding/json"
	"io/ioutil"

	"github.com/spf13/pflag"

	"k8s.io/klog/v2"

	"github.com/superedge/superedge/pkg/lite-apiserver/config"
)

type TLSOptions struct {
	CertConfig string
}

func NewTLSOptions() *TLSOptions {
	return &TLSOptions{}
}

// ApplyTo applies the storage options to the method receiver and returns self
func (s *TLSOptions) ApplyTo(c *config.LiteServerConfig) error {

	if len(s.CertConfig) == 0 {
		return nil
	}

	data, err := ioutil.ReadFile(s.CertConfig)
	if err != nil {
		klog.Errorf("cannot open tls config file %v", err)
		return err
	}

	err = json.Unmarshal(data, &c.TLSConfig)
	if err != nil {
		klog.Errorf("cannot open unmarshal config file %v", err)
		return err
	}

	return nil
}

// Validate checks validation of ServerRunOptions
func (s *TLSOptions) Validate() []error {
	var errors []error

	return errors
}

// AddUniversalFlags adds flags for a specific APIServer to the specified FlagSet
func (s *TLSOptions) AddFlags(fs *pflag.FlagSet) {
	// Note: the weird ""+ in below lines seems to be the only way to get gofmt to
	// arrange these text blocks sensibly. Grrr.

	fs.StringVar(&s.CertConfig, "tls-config-file", "", "")
}
