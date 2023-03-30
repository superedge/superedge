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
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/superedge/superedge/cmd/application-grid-wrapper/app/options"
	"github.com/superedge/superedge/pkg/application-grid-wrapper/server"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/version/verflag"
)

func NewWrapperProxyCommand() *cobra.Command {
	o := options.NewGridWrapperOptions()

	cmd := &cobra.Command{
		Use:  "application-grid-wrapper",
		Long: `Wrapper proxy for kube-proxy.`,
		Run: func(cmd *cobra.Command, args []string) {
			verflag.PrintAndExitIfRequested()
			util.PrintFlags(cmd.Flags())
			restConfig, err := clientcmd.BuildConfigFromFlags("", o.KubeConfig)
			if err != nil {
				klog.Errorf("can't build rest config, %v", err)
				return
			}
			restConfig.QPS = o.QPS
			restConfig.Burst = o.Burst
			server := server.NewInterceptorServer(restConfig, o.HostName, o.WrapperInCluster,
				o.NotifyChannelSize, o.ServiceAutonomyEnhancementOption, o.SupportEndpointSlice)
			if server == nil {
				return
			}

			if err := server.Run(o.Debug, o.BindAddress, o.InsecureMode,
				o.CAFile, o.CertFile, o.KeyFile, o.ServiceAutonomyEnhancementOption); err != nil {
				klog.Errorf("fail to start server, %v", err)
				return
			}
		},
	}

	o.AddFlags(cmd.Flags())

	return cmd
}
