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
	"github.com/spf13/cobra"
	"github.com/superedge/superedge/cmd/penetrator-job/app/options"
	"github.com/superedge/superedge/pkg/penetrator/job"
	"github.com/superedge/superedge/pkg/penetrator/job/conf"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/version"
	"github.com/superedge/superedge/pkg/version/verflag"
	"k8s.io/klog/v2"
)

func NewJobCommand() *cobra.Command {
	o := options.NewJobOptions()

	cmd := &cobra.Command{
		Use: "AddNodeJob",
		Run: func(cmd *cobra.Command, args []string) {
			verflag.PrintAndExitIfRequested()

			klog.Infof("Versions: %#v\n", version.Get())
			util.PrintFlags(cmd.Flags())
			err := conf.InitJobConfig(o.JobConfPath, o.SecretPath)
			if err != nil {
				klog.Fatalf("failed to init job config, error: %v", err)
			}
			job.AddNodes(o.Nodes)
		},
	}

	fs := cmd.Flags()
	o.AddFlags(fs)

	return cmd
}
