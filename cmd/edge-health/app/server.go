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

package app

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/superedge/superedge/cmd/edge-health/app/options"
	"github.com/superedge/superedge/pkg/edge-health/common"
	"github.com/superedge/superedge/pkg/edge-health/daemon"
	"github.com/superedge/superedge/pkg/edge-health/registry"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/version"
	"github.com/superedge/superedge/pkg/version/verflag"
	"k8s.io/klog/v2"
)

func NewEdgeHealthCommand(ctx context.Context, registryOptions ...registry.ExtendOptions) *cobra.Command {
	o := options.NewEdgeHealthOptions()
	cmd := &cobra.Command{
		Use: common.CmdName,
		Run: func(cmd *cobra.Command, args []string) {
			verflag.PrintAndExitIfRequested()

			klog.Infof("Versions: %#v\n", version.Get())
			util.PrintFlags(cmd.Flags())

			completedOptions := options.Complete(o)
			if errs := completedOptions.Validate(); len(errs) != 0 {
				klog.Fatalf("options validate err: %v", errs)
			}

			daemon.NewEdgeHealthDaemon(completedOptions, registryOptions...).Run(ctx)
		},
	}
	fs := cmd.Flags()
	namedFlagSets := o.AddFlags()
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	return cmd
}
