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
	"flag"
	"github.com/spf13/pflag"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"io"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/superedge/superedge/pkg/edgeadm/cmd/check"
	"github.com/superedge/superedge/pkg/edgeadm/cmd/clean"
	initCmd "github.com/superedge/superedge/pkg/edgeadm/cmd/init-cmd"
	"github.com/superedge/superedge/pkg/edgeadm/cmd/install"
	"github.com/superedge/superedge/pkg/edgeadm/cmd/join"
	"github.com/superedge/superedge/pkg/edgeadm/cmd/token"

	"github.com/superedge/superedge/pkg/edgeadm/cmd"
	"github.com/superedge/superedge/pkg/edgeadm/cmd/change"
	"github.com/superedge/superedge/pkg/edgeadm/cmd/revert"
)

var (
	edgeadmConf = cmd.EdgeadmConfig{
		WorkerPath: "/tmp",
		Kubeconfig: "~/.kube/config",
	}
)

func NewEdgeadmCommand(in io.Reader, out, err io.Writer) *cobra.Command {
	cmds := &cobra.Command{
		Use:   "edgeadm COMMAND [arg...]",
		Short: "edgeadm use to manage edge cluster",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Help()
		},
	}

	// add kubeconfig to persistent flags
	globalFlagSet(nil)
	cmds.ResetFlags()

	// edgeadm about change cluster
	cmds.AddCommand(cmd.NewManifestsCMD())
	cmds.AddCommand(change.NewChangeCMD())
	cmds.AddCommand(revert.NewRevertCMD())
	cmds.AddCommand(cmd.NewVersionCMD())

	// edgeadm create edge cluster
	cmds.AddCommand(check.NewCheckCMD())
	cmds.AddCommand(install.NewInstallCMD())
	cmds.AddCommand(initCmd.NewInitCMD(os.Stdout, &edgeadmConf))
	cmds.AddCommand(join.NewJoinCMD())
	cmds.AddCommand(clean.NewCleanCMD())
	cmds.AddCommand(token.NewTokenCMD())

	return cmds
}

func globalFlagSet(flagset *flag.FlagSet) {
	if flagset == nil {
		flagset = flag.CommandLine
	}

	flagset.StringVar(&edgeadmConf.WorkerPath, "worker-path", "/tmp", "Worker path of install edge kubernetes cluster.")
	flagset.StringVar(&edgeadmConf.Kubeconfig, "kubeconfig", "~/.kube/config", "The path to the kubeconfig file.")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	os.MkdirAll(path.Dir(edgeadmConf.WorkerPath+constant.EdgeClusterLogFile), 0755)
	pflag.Set("log_file", edgeadmConf.WorkerPath+constant.EdgeClusterLogFile)
	flag.Parse()
}
