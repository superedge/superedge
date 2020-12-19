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
	"k8s.io/klog"

	"superedge/cmd/application-grid-wrapper/app/options"
	"superedge/pkg/application-grid-wrapper/server"
	"superedge/pkg/util"
	"superedge/pkg/version/verflag"

	"github.com/spf13/cobra"
)

func NewWrapperProxyCommand() *cobra.Command {
	o := options.NewGridWrapperOptions()

	cmd := &cobra.Command{
		Use:  "application-grid-wrapper",
		Long: `Wrapper proxy for kube-proxy.`,
		Run: func(cmd *cobra.Command, args []string) {
			verflag.PrintAndExitIfRequested()
			util.PrintFlags(cmd.Flags())

			server := server.NewInterceptorServer(o.KubeConfig, o.HostName, o.WrapperInCluster, o.NotifyChannelSize)
			if server == nil {
				return
			}

			if err := server.Run(o.Debug, o.BindAddress, o.InsecureMode,
				o.CAFile, o.CertFile, o.KeyFile); err != nil {
				klog.Errorf("fail to start server, %v", err)
				return
			}
		},
	}

	o.AddFlags(cmd.Flags())

	return cmd
}
