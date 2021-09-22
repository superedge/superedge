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

package addon

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"
	"path/filepath"

	"github.com/superedge/superedge/pkg/util/kubeclient"
)

type addonAction struct {
	clientSet        *kubernetes.Clientset
	flags            *pflag.FlagSet
	manifestDir      string
	caKeyFile        string
	caCertFile       string
	masterPublicAddr string
	certSANs         []string
	kubeConfig       string

	app        bool
	core       bool
	device     bool
	support    bool
	sysmgmt    bool
	ui         bool
	completely bool
}

func NewAddonCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addon",
		Short: "Addon apps to Kubernetes cluster",
		Long:  cmdutil.MacroCommandLongDescription,
	}
	cmd.AddCommand(NewInstallEdgeAppsCMD())
	cmd.AddCommand(NewInstallEdgexCMD())
	cmd.AddCommand(NewInstallTopolvmCMD())
	return cmd
}

func NewDetachCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "detach",
		Short: "Detach apps from Kubernetes cluster",
		Long:  cmdutil.MacroCommandLongDescription,
	}
	cmd.AddCommand(NewDetachEdgeAppsCMD())
	cmd.AddCommand(NewDetachEdgexCMD())
	cmd.AddCommand(NewDetachTopolvmCMD())
	return cmd
}

func (a *addonAction) complete() error {
	configPath, err := a.flags.GetString("kubeconfig")
	if err != nil {
		klog.Errorf("Get kubeconfig flags error: %v", err)
	}
	if configPath == "~/.kube/config" {
		if home := homedir.HomeDir(); home != "" {
			configPath = filepath.Join(home, ".kube", "config")
		}
	}

	a.clientSet, err = kubeclient.GetClientSet(configPath)
	if err != nil {
		klog.Errorf("GetClientSet error: %v", err)
		return err
	}
	if a.clientSet == nil {
		return fmt.Errorf("Please set kubeconfig value!\n")
	}
	a.kubeConfig = configPath
	return nil
}
