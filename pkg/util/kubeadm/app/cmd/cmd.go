/*
Copyright 2016 The Kubernetes Authors.

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

package cmd

import (
	"io"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/alpha"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/options"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/upgrade"
	kubeadmutil "github.com/superedge/superedge/pkg/util/kubeadm/app/util"
	// Register the kubeadm configuration types because CLI flag generation
	// depends on the generated defaults.
)

// NewKubeadmCommand returns cobra.Command to run kubeadm command
func NewKubeadmCommand(in io.Reader, out, err io.Writer) *cobra.Command {
	var rootfsPath string

	cmds := &cobra.Command{
		Use:   "kubeadm",
		Short: "kubeadm: easily bootstrap a secure Kubernetes cluster",
		Long: dedent.Dedent(`

			    ┌──────────────────────────────────────────────────────────┐
			    │ KUBEADM                                                  │
			    │ Easily bootstrap a secure Kubernetes cluster             │
			    │                                                          │
			    │ Please give us feedback at:                              │
			    │ https://github.com/kubernetes/kubeadm/issues             │
			    └──────────────────────────────────────────────────────────┘

			Example usage:

			    Create a two-machine cluster with one control-plane node
			    (which controls the cluster), and one worker node
			    (where your workloads, like Pods and Deployments run).

			    ┌──────────────────────────────────────────────────────────┐
			    │ On the first machine:                                    │
			    ├──────────────────────────────────────────────────────────┤
			    │ control-plane# kubeadm init                              │
			    └──────────────────────────────────────────────────────────┘

			    ┌──────────────────────────────────────────────────────────┐
			    │ On the second machine:                                   │
			    ├──────────────────────────────────────────────────────────┤
			    │ worker# kubeadm join <arguments-returned-from-init>      │
			    └──────────────────────────────────────────────────────────┘

			    You can then repeat the second step on as many other machines as you like.

		`),
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if rootfsPath != "" {
				if err := kubeadmutil.Chroot(rootfsPath); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmds.ResetFlags()

	cmds.AddCommand(newCmdCompletion(out, ""))
	cmds.AddCommand(newCmdConfig(out))
	cmds.AddCommand(newCmdInit(out, nil))
	cmds.AddCommand(newCmdJoin(out, nil))
	cmds.AddCommand(newCmdReset(in, out, nil))
	cmds.AddCommand(newCmdVersion(out))
	cmds.AddCommand(newCmdToken(out, err))
	cmds.AddCommand(upgrade.NewCmdUpgrade(out))
	cmds.AddCommand(alpha.NewCmdAlpha(in, out))
	options.AddKubeadmOtherFlags(cmds.PersistentFlags(), &rootfsPath)

	// TODO: remove "certs" from "alpha"
	// https://github.com/kubernetes/kubeadm/issues/2291
	cmds.AddCommand(alpha.NewCmdCertsUtility(out))

	return cmds
}
