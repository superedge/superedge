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

package steps

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/options"
	phasesinit "k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/init"
	phasesjoin "k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/join"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"

	"github.com/superedge/superedge/pkg/edgeadm/cmd"
	"github.com/superedge/superedge/pkg/edgeadm/common"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests"
	"github.com/superedge/superedge/pkg/util"
)

var (
	kubeVIPInitExample = cmdutil.Examples(`
		# Install kube-vip.
		  kubeadm init phase kube-vip`)
)

//install kube-vip
func NewKubeVIPInitPhase(config *cmd.EdgeadmConfig) workflow.Phase {
	EdgeadmConf = config
	return workflow.Phase{
		Name:         "kube-vip",
		Short:        "Install kube-vip",
		Long:         "Install kube-vip",
		Example:      kubeVIPInitExample,
		Run:          installKubeVIP,
		InheritFlags: getKubeVIPPhaseFlags(true),
	}
}

func NewKubeVIPJoinPhase(config *cmd.EdgeadmConfig) workflow.Phase {
	EdgeadmConf = config
	return workflow.Phase{
		Name:         "kube-vip",
		Short:        "Install kube-vip",
		Long:         "Install kube-vip",
		Run:          installKubeVIP,
		Hidden:       true,
		InheritFlags: getKubeVIPPhaseFlags(false),
	}
}

func getKubeVIPPhaseFlags(isInitPhase bool) []string {
	flags := []string{
		constant.HANetworkInterface,
		constant.DefaultHA,
	}
	if isInitPhase {
		flags = append(flags, options.ControlPlaneEndpoint)
	}
	return flags
}

func installKubeVIP(c workflow.RunData) error {
	switch data := c.(type) {
	case phasesinit.InitData:
		if data.Cfg().ControlPlaneEndpoint != "" && EdgeadmConf.DefaultHA != "" {
			vip, _, err := util.SplitHostPortIgnoreMissingPort(data.Cfg().ControlPlaneEndpoint)
			if err != nil {
				return errors.Wrapf(err, "--control-plane-endpoint format invalid")
			}
			switch EdgeadmConf.DefaultHA {
			case constant.DefaultHAKubeVIP:
				if err := deployKubeVIP(vip, EdgeadmConf.KubeVIPInterface); err != nil {
					return errors.Wrapf(err, "failed to deploy kube-vip")
				}
				klog.Info("Installed kube-vip successfully")
			default:
				errors.Errorf("HA: %s is not supported, please use option `--default-ha=kube-vip` instead", EdgeadmConf.DefaultHA)
			}
		}
	case phasesjoin.JoinData:
		if data.Cfg().ControlPlane != nil && EdgeadmConf.DefaultHA != "" {
			vip, _, err := util.SplitHostPortIgnoreMissingPort(data.Cfg().Discovery.BootstrapToken.APIServerEndpoint)
			if err != nil {
				return errors.Wrapf(err, "invalid apiserver endpoint: %s", data.Cfg().Discovery.BootstrapToken.APIServerEndpoint)
			}
			switch EdgeadmConf.DefaultHA {
			case constant.DefaultHAKubeVIP:
				if err := deployKubeVIP(vip, EdgeadmConf.KubeVIPInterface); err != nil {
					return errors.Wrapf(err, "failed to deploy kube-vip")
				}
				klog.Info("Installed kube-vip successfully")
			default:
				errors.Errorf("HA: %s is not supported, please use option `--default-ha=kube-vip` instead", EdgeadmConf.DefaultHA)
			}
		}
	default:
		return errors.New("install kube-vip phase invoked with an invalid data struct")
	}
	return nil
}

func deployKubeVIP(vip, networkInterface string) error {
	kubeVIPYaml := common.ReadYaml(manifests.KUBE_VIP, manifests.KubeVIPYaml)
	kubeVIPYaml = strings.ReplaceAll(kubeVIPYaml, "{{.VIP}}", vip)
	kubeVIPYaml = strings.ReplaceAll(kubeVIPYaml, "{{.INTERFACE}}", networkInterface)
	cmd := fmt.Sprintf("mkdir -p %s && cat << EOF >%s \n%s\nEOF", constant.KubeManifestsPath, constant.KubeVIPPath, kubeVIPYaml)
	if _, _, err := util.RunLinuxCommand(cmd); err != nil {
		klog.Errorf("Deploy kube-vip error: %v", err)
		return err
	}
	return nil
}
