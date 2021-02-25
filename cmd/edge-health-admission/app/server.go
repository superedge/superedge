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
	"github.com/superedge/superedge/cmd/edge-health-admission/app/options"
	"github.com/superedge/superedge/pkg/edge-health-admission/admission"
	"github.com/superedge/superedge/pkg/edge-health-admission/config"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/version"
	"github.com/superedge/superedge/pkg/version/verflag"
	"k8s.io/klog"
	"os"
)

func NewEdgeHealthAdmissionCommand(ctx context.Context) *cobra.Command {
	o := options.NewAdmissionOptions()
	cmd := &cobra.Command{
		Use: "edge-health-admission",
		Run: func(cmd *cobra.Command, args []string) {
			verflag.PrintAndExitIfRequested()

			klog.Infof("Versions: %#v\n", version.Get())
			util.PrintFlags(cmd.Flags())

			// Complete options
			completedOptions, err := options.Complete(o)
			if err != nil {
				klog.Fatalf("Complete options err: %+v", err)
				os.Exit(1)
			}

			// Validate options
			if errs := completedOptions.Validate(); len(errs) != 0 {
				klog.Fatalf("Validate options errs: %+v", errs)
				os.Exit(1)
			}

			runCommand(ctx, completedOptions)
		},
	}

	fs := cmd.Flags()
	o.AddFlags(fs)

	return cmd
}

func runCommand(ctx context.Context, o options.CompletedOptions) {
	edgeHealthAdmissionConfig, err := config.NewEdgeHealthAdmissionConfig(o)
	if err != nil {
		klog.Fatalf("NewEdgeHealthAdmissionConfig err: %+v", err)
		return
	}
	go edgeHealthAdmissionConfig.Run(ctx.Done())
	go admission.NewEdgeHealthAdmission(edgeHealthAdmissionConfig).Run(ctx.Done())
	<-ctx.Done()
}
