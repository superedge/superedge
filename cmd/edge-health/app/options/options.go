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
	"github.com/superedge/superedge/pkg/edge-health/options"
	cliflag "k8s.io/component-base/cli/flag"
)

type Options struct {
	CheckOptions  *options.CheckOptions
	CommunOptions *options.CommunOptions
	NodeOptions   *options.NodeOptions
	VoteOptions   *options.VoteOptions
}

func NewEdgeHealthOptions() *Options {
	return &Options{
		options.NewCheckOptions(),
		options.NewCommunOptions(),
		options.NewNodeOptions(),
		options.NewVoteOptions(),
	}
}

func (s *Options) Complete() error {
	return nil
}

func (o *Options) AddFlags() (fsSet cliflag.NamedFlagSets) {
	o.NodeOptions.AddFlags(fsSet.FlagSet("node"))
	o.CheckOptions.AddFlags(fsSet.FlagSet("check"))
	o.CommunOptions.AddFlags(fsSet.FlagSet("communicate"))
	o.VoteOptions.AddFlags(fsSet.FlagSet("vote"))
	return
}

func (o *Options) Validate() []error {
	var errs []error
	errs = append(errs, o.VoteOptions.Validate()...)
	errs = append(errs, o.CommunOptions.Validate()...)
	errs = append(errs, o.CheckOptions.Validate()...)
	errs = append(errs, o.NodeOptions.Validate()...)
	return errs
}

type CompletedOptions struct {
	*Options
}

func Complete(o *Options) CompletedOptions {
	var completeoptions CompletedOptions
	o.CheckOptions.Default()
	o.CommunOptions.Default()
	o.VoteOptions.Default()
	completeoptions.Options = o
	return completeoptions
}
