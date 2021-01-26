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
	"github.com/superedge/superedge/pkg/edgeadm/common"
	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests"
	"github.com/superedge/superedge/pkg/util/kubeclient"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/options"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"
	"path/filepath"
)

var (
	networkAddonLongDesc = cmdutil.LongDesc(`
		Install network designed for Kubernetes.
		`)
)

func NewCNINetworkAppsPhase() workflow.Phase {
	return workflow.Phase{
		Name:  "cni",
		Short: "Install network designed for Kubernetes",
		Long:  networkAddonLongDesc,
		Phases: []workflow.Phase{
			{
				Name:         "flannel",
				Short:        "Install the flannel addon to Kubernetes cluster",
				InheritFlags: getNetworkAddonPhaseFlags("flannel"),
				Run:          runFlannelAddon,
			},
		},
	}
}

func getNetworkAddonPhaseFlags(name string) []string {
	var flags = make([]string, 0)
	if name == "flannel" {
		flags = append(flags,
			options.NetworkingPodSubnet,
		)
	}
	return flags
}

func runFlannelAddon(c workflow.RunData) error {
	cfg, edgeadmConf, client, err := getInitData(c)
	if err != nil {
		return err
	}
	// Deploy flannel
	option := map[string]interface{}{
		"PodNetworkCidr": cfg.Networking.PodSubnet,
	}
	klog.V(4).Infof("pod network cidr: %s", edgeadmConf.TunnelCoreDNSClusterIP)

	userManifests := filepath.Join(edgeadmConf.ManifestsDir, manifests.KUBE_FlANNEl)
	flannelYaml := common.ReadYaml(userManifests, manifests.KubeFlannelYaml)
	err = kubeclient.CreateResourceWithFile(client, flannelYaml, option)
	if err != nil {
		return err
	}

	klog.Infof("Deploy %s success!", manifests.KUBE_FlANNEl)
	return err
}
