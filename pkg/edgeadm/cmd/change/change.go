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

package change

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"superedge/pkg/edgeadm/constant"
	"superedge/pkg/util"
	"superedge/pkg/util/kubeclient"
)

type changeAction struct {
	deployName string
	clientSet  *kubernetes.Clientset

	flags      *pflag.FlagSet
	caCertFile string
	caKeyFile  string

	manifests string
}

func newChange() changeAction {
	return changeAction{}
}

func NewChangeCMD() *cobra.Command {
	action := newChange()
	cmd := &cobra.Command{
		Use:   "change -p DeployName",
		Short: "Change kubernetes cluster to edge cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if err := action.complete(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}

			if err := action.validate(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}

			if err := action.runChange(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}
		},
	}

	action.flags = cmd.Flags()
	cmd.Flags().StringVarP(&action.deployName, "deploy", "p",
		"kubeadm", "The mode about deploy k8s cluster, support value: kubeadm.")

	cmd.Flags().StringVar(&action.caCertFile, "ca.cert", "",
		"The root certificate file for cluster.")

	cmd.Flags().StringVar(&action.caKeyFile, "ca.key", "",
		"The root certificate key file for cluster.")

	cmd.Flags().StringVarP(&action.manifests, "manifests-dir", "m",
		"./manifests/", "Change yaml folder of edge cluster.")

	return cmd
}

func (c *changeAction) complete() error {
	configPath, err := c.flags.GetString("kubeconfig")
	if err != nil {
		klog.Errorf("Get kubeconfig flags error: %v", err)
	}

	c.clientSet, err = kubeclient.GetClientSet(configPath)
	if err != nil {
		klog.Errorf("GetClientSet error: %v", err)
		return err
	}
	if c.clientSet == nil {
		return fmt.Errorf("Please set kubeconfig value!\n")
	}

	return nil
}

func (c *changeAction) validate() error {
	return nil
}

func (c *changeAction) runChange() error {
	switch c.deployName {
	case constant.DeployModeKubeadm:
		return c.runKubeamdChange()
	default:
		return fmt.Errorf("Not support %s change to edge cluster\n", c.deployName)
	}

	return nil
}
