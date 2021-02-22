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
	"github.com/superedge/superedge/pkg/edge-health/config"
	"github.com/superedge/superedge/pkg/edge-health/daemon"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/version"
	"github.com/superedge/superedge/pkg/version/verflag"
	"k8s.io/klog"
	"os"
)

func NewEdgeHealthCommand(ctx context.Context) *cobra.Command {
	o := options.NewEdgeHealthOptions()
	cmd := &cobra.Command{
		Use: "edge-health",
		Run: func(cmd *cobra.Command, args []string) {
			verflag.PrintAndExitIfRequested()

			klog.Infof("Versions: %#v\n", version.Get())
			util.PrintFlags(cmd.Flags())

			// Set default options
			completedOptions, err := options.Complete(o)
			if err != nil {
				klog.Fatalf("Set default options err: %v", err)
				os.Exit(1)
			}

			// Validate options
			if errs := completedOptions.Validate(); len(errs) != 0 {
				klog.Fatalf("Validate options errs: %v", errs)
				os.Exit(1)
			}

			runEdgeHealth(ctx, completedOptions)
		},
	}
	fs := cmd.Flags()
	namedFlagSets := o.AddFlags()
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	return cmd
}

func runEdgeHealth(ctx context.Context, o options.CompletedOptions) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	edgeHealthConfig, err := config.NewEdgeHealthConfig(o)
	if err != nil {
		klog.Fatalf("Validate options errs: %v", err)
		return
	}

	go edgeHealthConfig.Run(ctx.Done())
	go daemon.NewEdgeHealthDaemon(edgeHealthConfig).Run(ctx.Done())
	<-ctx.Done()
}
