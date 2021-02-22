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
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	"os"
	"strings"
)

type Options struct {
	Check  *CheckOptions
	Commun *CommunOptions
	Node   *NodeOptions
	Vote   *VoteOptions
}

func NewEdgeHealthOptions() *Options {
	options := &Options{
		NewCheckOptions(),
		NewCommunOptions(),
		NewNodeOptions(),
		NewVoteOptions(),
	}
	options.Check.Default()
	options.Commun.Default()
	options.Node.Default()
	options.Vote.Default()
	return options
}

func (o *Options) AddFlags() (fsSet cliflag.NamedFlagSets) {
	o.Check.AddFlags(fsSet.FlagSet("check"))
	o.Commun.AddFlags(fsSet.FlagSet("commun"))
	o.Node.AddFlags(fsSet.FlagSet("node"))
	o.Vote.AddFlags(fsSet.FlagSet("vote"))
	return
}

func (o *Options) Validate() []error {
	var errs []error

	errs = append(errs, o.Check.Validate()...)
	errs = append(errs, o.Commun.Validate()...)
	errs = append(errs, o.Node.Validate()...)
	errs = append(errs, o.Vote.Validate()...)

	return errs
}

// CompletedOptions is a wrapper that enforces a call of Complete() before Run can be invoked.
type CompletedOptions struct {
	*Options
}

func Complete(o *Options) (CompletedOptions, error) {
	var options CompletedOptions
	options.Options = o
	if o.Node.HostName == "" {
		o.Node.HostName = os.Getenv("NODE_NAME")
		o.Node.HostName = strings.Replace(o.Node.HostName, "\n", "", -1)
		o.Node.HostName = strings.Replace(o.Node.HostName, " ", "", -1)
		klog.V(2).Infof("Host name is %s", o.Node.HostName)
	}

	return options, nil
}
