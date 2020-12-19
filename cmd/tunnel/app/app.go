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
	"superedge/cmd/tunnel/app/options"
	"superedge/pkg/tunnel/conf"
	. "superedge/pkg/tunnel/model"
	"superedge/pkg/tunnel/proxy/https"
	"superedge/pkg/tunnel/proxy/stream"
	"superedge/pkg/tunnel/proxy/tcp"
	"superedge/pkg/util"
	"superedge/pkg/version"
	"superedge/pkg/version/verflag"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func NewTunnelCommand() *cobra.Command {
	option := options.NewTunnelOption()
	cmd := &cobra.Command{
		Use: "tunnel",
		Run: func(cmd *cobra.Command, args []string) {
			verflag.PrintAndExitIfRequested()

			klog.Infof("Versions: %#v\n", version.Get())
			util.PrintFlags(cmd.Flags())

			err := conf.InitConf(*option.TunnelMode, *option.TunnelConf)
			if err != nil {
				klog.Info("tunnel failed to load configuration file !")
				return
			}
			InitModules(*option.TunnelMode)
			stream.InitStream(*option.TunnelMode)
			tcp.InitTcp()
			https.InitHttps()
			LoadModules(*option.TunnelMode)
			ShutDown()
		},
	}
	fs := cmd.Flags()
	namedFlagSets := option.Addflag()
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}
	return cmd
}
