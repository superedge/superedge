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

package steps

import (
	"github.com/pkg/errors"
	"github.com/superedge/superedge/pkg/edgeadm/cmd"
	"github.com/superedge/superedge/pkg/edgeadm/common"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests"
	kubeadmapi "github.com/superedge/superedge/pkg/util/kubeadm/app/apis/kubeadm"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/options"
	phases "github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/phases/init"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/phases/workflow"
	cmdutil "github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/util"
	kubeadmconstants "github.com/superedge/superedge/pkg/util/kubeadm/app/constants"
	"github.com/superedge/superedge/pkg/util/kubeclient"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"path/filepath"
)

var (
	coreDNSAddonLongDesc = cmdutil.LongDesc(`
		Install the CoreDNS addon components via the API server.
		Please note that although the DNS server is deployed, it will not be scheduled until CNI is installed.
		`)

	kubeProxyAddonLongDesc = cmdutil.LongDesc(`
		Install the kube-proxy addon components via the API server.
		`)
)

// NewAddonPhase returns the addon Cobra command
func NewEdgeAppsPhase() workflow.Phase { //todo 独立成一个独立的函数，为后面实现kube-*集群master无缝部署，加简便的添加边缘节点
	return workflow.Phase{
		Name:  "edge-apps",
		Short: "Install edge-apps",
		Long:  cmdutil.MacroCommandLongDescription, //todo
		Phases: []workflow.Phase{
			{
				Name:           "all",
				Short:          "Install all the edge-apps addons",
				InheritFlags:   getAddonPhaseFlags("all"),
				RunAllSiblings: true,
			},
			{
				Name:         "tunnel",
				Short:        "Install the tunnel addon to edge Kubernetes cluster",
				Long:         kubeProxyAddonLongDesc, //todo
				InheritFlags: getAddonPhaseFlags("tunnel"),
				Run:          runTunnelAddon,
			},
			{
				Name:         "edge-health",
				Short:        "Install the edge-health addon to edge Kubernetes cluster",
				Long:         coreDNSAddonLongDesc, //todo
				InheritFlags: getAddonPhaseFlags("edge-health"),
				Run:          runEdgeHealthAddon,
			},
			{
				Name:         "service-group",
				Short:        "Install the service-group addon to edge Kubernetes cluster",
				Long:         coreDNSAddonLongDesc, //todo
				InheritFlags: getAddonPhaseFlags("service-group"),
				Run:          runSerivceGroupAddon,
			},
			{
				Name:         "lite-apiserver",
				Short:        "Config the lite-apiserver configmap into edge Kubernetes cluster",
				Long:         coreDNSAddonLongDesc, //todo
				InheritFlags: getAddonPhaseFlags("lite-apiserver"),
				Run:          configLiteAPIServer,
			},
			{
				Name:         "update-config",
				Short:        "Update Kubernetes cluster config support marginal autonomy",
				Long:         coreDNSAddonLongDesc, //todo
				InheritFlags: getAddonPhaseFlags("update-config"),
				Run:          updateKubeConfig,
			},
		},
	}
}

func getAddonPhaseFlags(name string) []string {
	flags := []string{
		constant.ManifestsDir,
		options.KubeconfigPath,
	}
	if name == "all" || name == "tunnel" {
		flags = append(flags,
			options.CertificatesDir,
		)
	}
	if name == "all" || name == "edge-health" {
		//flags = append(flags,
		//	options.CertificatesDir,
		//)
	}
	if name == "all" || name == "service-group" {
		//flags = append(flags,
		//	options.CertificatesDir,
		//)
	}
	if name == "all" || name == "lite-apiserver" {
		//flags = append(flags,
		//	options.CertificatesDir,
		//)
	}
	if name == "all" || name == "update-config" {
		//flags = append(flags,
		//	options.CertificatesDir,
		//)
	}
	return flags
}

func getInitData(c workflow.RunData) (*kubeadmapi.InitConfiguration, *cmd.EdgeadmConfig, clientset.Interface, error) {
	data, ok := c.(phases.InitData)
	if !ok {
		return nil, nil, nil, errors.New("addon phase invoked with an invalid data struct")
	}

	client, err := data.Client()
	if err != nil {
		return nil, nil, nil, err
	}
	return data.Cfg(), data.EdgeadmConf(), client, err
}

// runCoreDNSAddon installs CoreDNS addon to a Kubernetes cluster
func runTunnelAddon(c workflow.RunData) error {
	cfg, edgeadmConf, client, err := getInitData(c)
	if err != nil {
		return err
	}

	// Deploy tunnel-coredns
	option := map[string]interface{}{
		"TunnelCoreDNSClusterIP": edgeadmConf.TunnelCoreDNSClusterIP,
	}
	klog.Infof("TunnelCoreDNSClusterIP: %s", edgeadmConf.TunnelCoreDNSClusterIP)
	userManifests := filepath.Join(edgeadmConf.ManifestsDir, manifests.APP_TUNNEL_CORDDNS)
	TunnelCoredns := common.ReadYaml(userManifests, manifests.TunnelCorednsYaml)
	err = kubeclient.CreateResourceWithFile(client, TunnelCoredns, option)
	if err != nil {
		return err
	}
	klog.Infof("Deploy %s success!", manifests.APP_TUNNEL_CORDDNS)

	// Deploy tunnel-cloud
	caKeyFile := filepath.Join(cfg.CertificatesDir, kubeadmconstants.CAKeyName)
	caCertFile := filepath.Join(cfg.CertificatesDir, kubeadmconstants.CACertName)
	if err = common.DeployTunnelCloud(client, edgeadmConf.ManifestsDir,
		caCertFile, caKeyFile, edgeadmConf.TunnelCloudToken); err != nil {
		klog.Errorf("Deploy tunnel-cloud, error: %v", err)
		return err
	}
	klog.Infof("Deploy %s success!", manifests.APP_TUNNEL_CLOUD)

	// GetTunnelCloudPort
	tunnelCloudNodePort, err := common.GetTunnelCloudPort(client)
	if err != nil {
		klog.Errorf("Get tunnel-cloud port, error: %v", err)
		return err
	}

	// Deploy tunnel-edge
	if err = common.DeployTunnelEdge(client, edgeadmConf.ManifestsDir,
		caCertFile, caKeyFile, edgeadmConf.TunnelCloudToken, tunnelCloudNodePort); err != nil {
		klog.Errorf("Deploy tunnel-edge, error: %v", err)
		return err
	}

	klog.Infof("Deploy %s success!", manifests.APP_TUNNEL_EDGE)

	return err
}

// runCoreDNSAddon installs CoreDNS addon to a Kubernetes cluster
func runEdgeHealthAddon(c workflow.RunData) error {
	_, edgeadmConf, client, err := getInitData(c)
	if err != nil {
		return err
	}

	if err := common.DeployEdgeHealth(client, edgeadmConf.ManifestsDir); err != nil {
		klog.Errorf("Deploy edge health, error: %s", err)
		return err
	}

	return err
}

// runCoreDNSAddon installs CoreDNS addon to a Kubernetes cluster
func runSerivceGroupAddon(c workflow.RunData) error {
	_, edgeadmConf, client, err := getInitData(c)
	if err != nil {
		return err
	}

	if err := common.DeploySerivceGroup(client, edgeadmConf.ManifestsDir); err != nil {
		klog.Errorf("Deploy serivce group, error: %s", err)
		return err
	}

	klog.Infof("Deploy service-group success!")

	return err
}

// runCoreDNSAddon installs CoreDNS addon to a Kubernetes cluster
func updateKubeConfig(c workflow.RunData) error {
	_, _, client, err := getInitData(c)
	if err != nil {
		return err
	}

	if err := common.UpdateKubeProxyKubeconfig(client); err != nil {
		klog.Errorf("Deploy serivce group, error: %s", err)
		return err
	}

	if err := common.UpdateKubernetesEndpoint(client); err != nil {
		klog.Errorf("Deploy serivce group, error: %s", err)
		return err
	}

	klog.Infof("Update Kubernetes cluster config support marginal autonomy success")

	return err
}

// runCoreDNSAddon installs CoreDNS addon to a Kubernetes cluster
func configLiteAPIServer(c workflow.RunData) error {
	cfg, edgeadmConf, client, err := getInitData(c)
	if err != nil {
		return err
	}

	caKeyFile := filepath.Join(cfg.CertificatesDir, kubeadmconstants.CAKeyName)
	caCertFile := filepath.Join(cfg.CertificatesDir, kubeadmconstants.CACertName)
	if err := common.CreateLiteApiServerCert(client, edgeadmConf.ManifestsDir, caCertFile, caKeyFile); err != nil {
		klog.Errorf("Config lite-apiserver, error: %s", err)
		return err
	}

	klog.Infof("Config lite-apiserver configMap success")

	return err
}
