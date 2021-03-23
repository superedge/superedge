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

package join

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/superedge/superedge/pkg/edgeadm/cmd"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
)

/*
添加master的一些注意事项：
- 需要把添加maste的hostname，写入第一个master的/etc/host: masterIP hostname, 这样在其他master节点才能用其他master的hostname访问
*/

type ClusterProgress struct {
	Status     string   `json:"status"`
	Data       string   `json:"data"`
	URL        string   `json:"url,omitempty"`
	Username   string   `json:"username,omitempty"`
	Password   []byte   `json:"password,omitempty"`
	CACert     []byte   `json:"caCert,omitempty"`
	Hosts      []string `json:"hosts,omitempty"`
	Servers    []string `json:"servers,omitempty"`
	Kubeconfig []byte   `json:"kubeconfig,omitempty"`
}

type Handler struct {
	Name string
	Func func() error
}

type joinOptions struct {
	EdgeJoinConfig             edgeJoinConfig
	KubeadmConfig              kubeadmConfig
	MasterIp                   string
	JoinToken                  string
	TokenCaCertHash            string
	CertificateKey             string
	KubernetesServiceClusterIP string
}

type edgeJoinConfig struct {
	WorkerPath     string `yaml:"workerPath"`
	InstallPkgPath string `yaml:"InstallPkgPath"`
}

type kubeadmConfig struct {
	KubeadmConfPath string `yaml:"kubeadmConfPath"`
}

func NewJoinCMD(edgeConfig *cmd.EdgeadmConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join",
		Short: "Join a node into cluster",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Help()
		},
	}
	cmd.AddCommand(NewJoinMasterCMD(edgeConfig))
	cmd.AddCommand(NewJoinNodeCMD(edgeConfig))
	return cmd
}

func AddEdgeConfigFlags(flagSet *pflag.FlagSet, cfg *edgeJoinConfig) {
	flagSet.StringVar(
		&cfg.InstallPkgPath, constant.InstallPkgPath, "./edge-v0.3.0-kube-v1.18.2-install-pkg.tar.gz",
		"Install static package path of edge kubernetes cluster.",
	)
}

func AddKubeadmConfigFlags(flagSet *pflag.FlagSet, cfg *kubeadmConfig) {
	flagSet.StringVar(
		&cfg.KubeadmConfPath, constant.KubeadmConfig, "/root/.edgeadm/kubeadm.config",
		"Install static package path of edge kubernetes cluster.",
	)
}
