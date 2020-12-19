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

	"superedge/pkg/edgeadm/cmd"
	"superedge/pkg/edgeadm/cmd/change"
	"superedge/pkg/edgeadm/cmd/revert"
)

func NewEdgeadmCommand() *cobra.Command {
	cmds := &cobra.Command{
		Use:   "edgeadm COMMAND [arg...]",
		Short: "edgeadm use to manage edge cluster",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Help()
		},
	}

	// add kubeconfig to persistent flags
	cmds.PersistentFlags().String("kubeconfig", "", "The path to the kubeconfig file")
	cmds.AddCommand(cmd.NewManifestsCMD())
	cmds.AddCommand(change.NewChangeCMD())
	cmds.AddCommand(revert.NewRevertCMD())
	cmds.AddCommand(cmd.NewVersionCMD())
	return cmds
}
