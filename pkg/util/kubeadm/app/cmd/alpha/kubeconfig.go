/*
Copyright 2018 The Kubernetes Authors.

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

package alpha

import (
	"io"

	"github.com/spf13/cobra"

	kubeadmapiv1beta2 "github.com/superedge/superedge/pkg/util/kubeadm/app/apis/kubeadm/v1beta2"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/options"
	cmdutil "github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/util"
	kubeconfigphase "github.com/superedge/superedge/pkg/util/kubeadm/app/phases/kubeconfig"
	configutil "github.com/superedge/superedge/pkg/util/kubeadm/app/util/config"
)

var (
	kubeconfigLongDesc = cmdutil.LongDesc(`
	Kubeconfig file utilities.
	` + cmdutil.AlphaDisclaimer)

	userKubeconfigLongDesc = cmdutil.LongDesc(`
	Output a kubeconfig file for an additional user.
	` + cmdutil.AlphaDisclaimer)

	userKubeconfigExample = cmdutil.Examples(`
	# Output a kubeconfig file for an additional user named foo using a kubeadm config file bar
	kubeadm alpha kubeconfig user --client-name=foo --config=bar
	`)
)

// newCmdKubeConfigUtility returns main command for kubeconfig phase
func newCmdKubeConfigUtility(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubeconfig",
		Short: "Kubeconfig file utilities",
		Long:  kubeconfigLongDesc,
	}

	cmd.AddCommand(newCmdUserKubeConfig(out))
	return cmd
}

// newCmdUserKubeConfig returns sub commands for kubeconfig phase
func newCmdUserKubeConfig(out io.Writer) *cobra.Command {

	initCfg := cmdutil.DefaultInitConfiguration()
	clusterCfg := &kubeadmapiv1beta2.ClusterConfiguration{}

	var (
		token, clientName, cfgPath string
		organizations              []string
	)

	// Creates the UX Command
	cmd := &cobra.Command{
		Use:     "user",
		Short:   "Output a kubeconfig file for an additional user",
		Long:    userKubeconfigLongDesc,
		Example: userKubeconfigExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			// This call returns the ready-to-use configuration based on the defaults populated by flags
			internalCfg, err := configutil.LoadOrDefaultInitConfiguration(cfgPath, initCfg, clusterCfg)
			if err != nil {
				return err
			}

			// if the kubeconfig file for an additional user has to use a token, use it
			if token != "" {
				return kubeconfigphase.WriteKubeConfigWithToken(out, internalCfg, clientName, token)
			}

			// Otherwise, write a kubeconfig file with a generate client cert
			return kubeconfigphase.WriteKubeConfigWithClientCert(out, internalCfg, clientName, organizations)
		},
		Args: cobra.NoArgs,
	}

	options.AddConfigFlag(cmd.Flags(), &cfgPath)

	// Add command specific flags
	cmd.Flags().StringVar(&token, options.TokenStr, token, "The token that should be used as the authentication mechanism for this kubeconfig, instead of client certificates")
	cmd.Flags().StringVar(&clientName, "client-name", clientName, "The name of user. It will be used as the CN if client certificates are created")
	cmd.Flags().StringSliceVar(&organizations, "org", organizations, "The orgnizations of the client certificate. It will be used as the O if client certificates are created")

	cmd.MarkFlagRequired(options.CfgPath)
	cmd.MarkFlagRequired("client-name")
	return cmd
}
