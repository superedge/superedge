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

// Package options contains all the primary arguments for a kubelet.
package options

import (
	cliflag "k8s.io/component-base/cli/flag"

	"github.com/superedge/superedge/pkg/lite-apiserver/config"
	"github.com/superedge/superedge/pkg/lite-apiserver/options"
)

// ServerRunOptions runs a lite api server.
type ServerRunOptions struct {
	TLSOptions       *options.TLSOptions
	RunServerOptions *options.RunServerOptions
}

// NewServerRunOptions creates a new ServerRunOptions object with default parameters
func NewServerRunOptions() *ServerRunOptions {
	s := ServerRunOptions{
		TLSOptions:       options.NewTLSOptions(),
		RunServerOptions: options.NewRunServerOptions(),
	}
	return &s
}

// Flags returns flag set
func (s *ServerRunOptions) Flags() (fsSet cliflag.NamedFlagSets) {
	s.TLSOptions.AddFlags(fsSet.FlagSet("tls"))
	s.RunServerOptions.AddFlags(fsSet.FlagSet("server"))
	return
}

// Validate checks ServerRunOptions and return a slice of found errs.
func (s *ServerRunOptions) Validate() []error {
	var errs []error

	errs = append(errs, s.TLSOptions.Validate()...)
	errs = append(errs, s.RunServerOptions.Validate()...)

	return errs
}

// Complete set default ServerRunOptions.
// Should be called after flags parsed.
func (s *ServerRunOptions) Complete() error {
	return nil
}

// ApplyTo applies the storage options to the method receiver and returns self
func (s *ServerRunOptions) ApplyTo() (*config.LiteServerConfig, error) {
	liteServerConfig := &config.LiteServerConfig{}
	err := s.RunServerOptions.ApplyTo(liteServerConfig)
	if err != nil {
		return nil, err
	}

	err = s.TLSOptions.ApplyTo(liteServerConfig)
	if err != nil {
		return nil, err
	}
	return liteServerConfig, nil
}
