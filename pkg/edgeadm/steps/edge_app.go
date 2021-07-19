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
	"errors"
	"path/filepath"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/options"
	phases "k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/init"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"github.com/superedge/superedge/pkg/edgeadm/cmd"
	"github.com/superedge/superedge/pkg/edgeadm/common"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests"
	"github.com/superedge/superedge/pkg/util/kubeclient"
)

var (
	EdgeadmConf = new(cmd.EdgeadmConfig)
)

// NewEdgeAppsPhase returns the edge addon to edge Kubernetes cluster
func NewEdgeAppsPhase(config *cmd.EdgeadmConfig) workflow.Phase {
	EdgeadmConf = config
	return workflow.Phase{
		Name:  "edge-apps",
		Short: "Addon SuperEdge edge-apps to Kubernetes cluster",
		Long:  cmdutil.MacroCommandLongDescription,
		Phases: []workflow.Phase{
			{
				Name:           "all",
				Short:          "Install all the edge-apps addons to edge Kubernetes cluster",
				InheritFlags:   getAddonPhaseFlags("all"),
				RunAllSiblings: true,
			},
			{
				Name:         "tunnel",
				Short:        "Install the tunnel addon to edge Kubernetes cluster",
				InheritFlags: getAddonPhaseFlags("tunnel"),
				RunIf: func(data workflow.RunData) (bool, error) {
					return config.IsEnableEdge, nil
				},
				Run: runTunnelAddon,
			},
			{
				Name:         "edge-health",
				Short:        "Install the edge-health addon to edge Kubernetes cluster",
				InheritFlags: getAddonPhaseFlags("edge-health"),
				RunIf: func(data workflow.RunData) (bool, error) {
					return config.IsEnableEdge, nil
				},
				Run: runEdgeHealthAddon,
			},
			{
				Name:         "service-group",
				Short:        "Install the service-group addon to edge Kubernetes cluster",
				InheritFlags: getAddonPhaseFlags("service-group"),
				RunIf: func(data workflow.RunData) (bool, error) {
					return config.IsEnableEdge, nil
				},
				Run: runServiceGroupAddon,
			},
			{
				Name:         "edge-coredns",
				Short:        "Install the edge-coredns addon to edge Kubernetes cluster",
				InheritFlags: getAddonPhaseFlags("edge-coredns"),
				RunIf: func(data workflow.RunData) (bool, error) {
					return config.IsEnableEdge, nil
				},
				Run: runEdgeCorednsAddon,
			},
			{
				Name:         "join-prepare",
				Hidden:       true,
				Short:        "Prepare Config of join master or edge node",
				InheritFlags: getAddonPhaseFlags("join-prepare"),
				RunIf: func(data workflow.RunData) (bool, error) {
					return config.IsEnableEdge, nil
				},
				Run: joinNodePrepare,
			},
			{
				Name:         "update-config",
				Hidden:       true,
				Short:        "Update Kubernetes cluster config support marginal autonomy",
				InheritFlags: getAddonPhaseFlags("update-config"),
				RunIf: func(data workflow.RunData) (bool, error) {
					return config.IsEnableEdge, nil
				},
				Run: updateKubeConfig,
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
	}
	if name == "all" || name == "service-group" {
	}
	if name == "all" || name == "update-config" {
	}
	if name == "all" || name == "join-prepare" {
	}
	if name == "all" || name == "edge-coredns" {
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
	return data.Cfg(), EdgeadmConf, client, err
}

func runTunnelAddon(c workflow.RunData) error {
	cfg, edgeadmConf, client, err := getInitData(c)
	if err != nil {
		return err
	}

	if err := common.EnsureEdgeSystemNamespace(client); err != nil {
		return err
	}

	// Deploy tunnel-coredns
	option := map[string]interface{}{
		"Namespace":              constant.NamespaceEdgeSystem,
		"TunnelCoreDNSClusterIP": edgeadmConf.TunnelCoreDNSClusterIP,
	}
	klog.V(4).Infof("TunnelCoreDNSClusterIP: %s", edgeadmConf.TunnelCoreDNSClusterIP)

	userManifests := filepath.Join(edgeadmConf.ManifestsDir, manifests.APP_TUNNEL_CORDDNS)
	TunnelCoredns := common.ReadYaml(userManifests, manifests.TunnelCorednsYaml)
	err = kubeclient.CreateResourceWithFile(client, TunnelCoredns, option)
	if err != nil {
		return err
	}
	klog.Infof("Deploy %s success!", manifests.APP_TUNNEL_CORDDNS)

	// Deploy tunnel-cloud
	certSANs := cfg.APIServer.CertSANs
	caKeyFile := filepath.Join(cfg.CertificatesDir, kubeadmconstants.CAKeyName)
	caCertFile := filepath.Join(cfg.CertificatesDir, kubeadmconstants.CACertName)
	if err = common.DeployTunnelCloud(client, edgeadmConf.ManifestsDir,
		caCertFile, caKeyFile, edgeadmConf.TunnelCloudToken, certSANs); err != nil {
		klog.Errorf("Deploy tunnel-cloud, error: %v", err)
		return err
	}
	klog.Infof("Deploy %s success!", manifests.APP_TUNNEL_CLOUD)

	tunnelCloudNodeAddr := cfg.ControlPlaneEndpoint
	if len(certSANs) > 0 {
		tunnelCloudNodeAddr = certSANs[0]
	}
	// GetTunnelCloudPort
	tunnelCloudNodePort, err := common.GetTunnelCloudPort(client)
	if err != nil {
		klog.Errorf("Get tunnel-cloud port, error: %v", err)
		return err
	}

	// Deploy tunnel-edge
	if err = common.DeployTunnelEdge(client, edgeadmConf.ManifestsDir, caCertFile, caKeyFile,
		edgeadmConf.TunnelCloudToken, tunnelCloudNodeAddr, tunnelCloudNodePort); err != nil {
		klog.Errorf("Deploy tunnel-edge, error: %v", err)
		return err
	}
	klog.Infof("Deploy %s success!", manifests.APP_TUNNEL_EDGE)

	return err
}

func runEdgeHealthAddon(c workflow.RunData) error {
	_, edgeadmConf, client, err := getInitData(c)
	if err != nil {
		return err
	}

	if err := common.EnsureEdgeSystemNamespace(client); err != nil {
		return err
	}

	if err := common.DeployEdgeHealth(client, edgeadmConf.ManifestsDir); err != nil {
		klog.Errorf("Deploy edge health, error: %s", err)
		return err
	}

	return err
}

func runServiceGroupAddon(c workflow.RunData) error {
	_, edgeadmConf, client, err := getInitData(c)
	if err != nil {
		return err
	}

	if err := common.EnsureEdgeSystemNamespace(client); err != nil {
		return err
	}

	if err := common.DeployServiceGroup(client, edgeadmConf.ManifestsDir); err != nil {
		klog.Errorf("Deploy serivce group, error: %s", err)
		return err
	}

	klog.Infof("Deploy service-group success!")

	return err
}

func runEdgeCorednsAddon(c workflow.RunData) error {
	data, ok := c.(phases.InitData)
	if !ok {
		return errors.New("addon phase invoked with an invalid data struct")
	}

	client, err := data.Client()
	if err != nil {
		return err
	}

	//Add Label superedge.io.hostname to deploy edge-codedns service-group
	masterLabel := map[string]string{
		constant.EdgehostnameLabelKey: data.Cfg().NodeRegistration.Name,
	}
	if err := kubeclient.AddNodeLabel(client, data.Cfg().NodeRegistration.Name, masterLabel); err != nil {
		klog.Errorf("Add edged Node node label error: %v", err)
		return err
	}

	if err := common.DeployEdgeCorednsAddon(data.KubeConfigPath(), EdgeadmConf.ManifestsDir); err != nil {
		klog.Errorf("Deploy edge-coredns error: %v", err)
		return err
	}

	return nil
}

func updateKubeConfig(c workflow.RunData) error {
	initConfiguration, _, client, err := getInitData(c)
	if err != nil {
		return err
	}

	if err := common.UpdateKubeProxyKubeconfig(client); err != nil {
		klog.Errorf("Update kube-proxy config, error: %s", err)
		return err
	}

	if err := common.UpdateKubernetesEndpoint(client); err != nil {
		klog.Errorf("Update kubernetes endpoint, error: %s", err)
		return err
	}

	if len(initConfiguration.APIServer.CertSANs) > 0 {
		certSANs := initConfiguration.APIServer.CertSANs
		if err := common.UpdateClusterInfoKubeconfig(client, certSANs); err != nil {
			klog.Errorf("Update cluster-info config, error: %s", err)
			return err
		}
	}

	klog.Infof("Update Kubernetes cluster config support marginal autonomy success")

	return err
}

func joinNodePrepare(c workflow.RunData) error {
	cfg, egeadmConf, client, err := getInitData(c)
	if err != nil {
		return err
	}

	if err := common.EnsureEdgeSystemNamespace(client); err != nil {
		return err
	}

	// Prepare lite-apiserver config info
	caKeyFile := filepath.Join(cfg.CertificatesDir, kubeadmconstants.CAKeyName)
	caCertFile := filepath.Join(cfg.CertificatesDir, kubeadmconstants.CACertName)
	if err := common.JoinNodePrepare(client, egeadmConf.ManifestsDir, caCertFile, caKeyFile); err != nil {
		klog.Errorf("Prepare Config Join Node, error: %s", err)
		return err
	}
	klog.Infof("Prepare Config Join Node configMap success")

	return err
}

func deleteTunnelAddon(c workflow.RunData) error {
	cfg, edgeadmConf, client, err := getInitData(c)
	if err != nil {
		return err
	}

	if ok := common.CheckIfEdgeAppDeletable(client); !ok {
		klog.Info("Can not Delete Edge Apps, cluster has remaining edge nodes!")
		return nil
	}

	// GetTunnelCloudPort
	tunnelCloudNodePort, err := common.GetTunnelCloudPort(client)
	if err != nil {
		klog.Errorf("Get tunnel-cloud port, error: %v", err)
		return err
	}

	// Delete tunnel-edge
	certSANs := cfg.APIServer.CertSANs
	caKeyFile := filepath.Join(cfg.CertificatesDir, kubeadmconstants.CAKeyName)
	caCertFile := filepath.Join(cfg.CertificatesDir, kubeadmconstants.CACertName)
	tunnelCloudNodeAddr := cfg.ControlPlaneEndpoint
	if len(certSANs) > 0 {
		tunnelCloudNodeAddr = certSANs[0]
	}
	if err = common.DeleteTunnelEdge(client, edgeadmConf.ManifestsDir,
		caCertFile, caKeyFile, edgeadmConf.TunnelCloudToken, tunnelCloudNodeAddr, tunnelCloudNodePort); err != nil {
		klog.Errorf("Deploy tunnel-edge, error: %v", err)
		return err
	}
	klog.Infof("Delete %s success!", manifests.APP_TUNNEL_EDGE)

	// Delete tunnel-cloud
	if err = common.DeleteTunnelCloud(client, edgeadmConf.ManifestsDir,
		caCertFile, caKeyFile, edgeadmConf.TunnelCloudToken, certSANs); err != nil {
		klog.Errorf("Delete tunnel-cloud, error: %v", err)
		return err
	}
	klog.Infof("Delete %s success!", manifests.APP_TUNNEL_CLOUD)

	// Delete tunnel-coredns
	option := map[string]interface{}{
		"TunnelCoreDNSClusterIP": edgeadmConf.TunnelCoreDNSClusterIP,
	}
	klog.V(4).Infof("TunnelCoreDNSClusterIP: %s", edgeadmConf.TunnelCoreDNSClusterIP)

	userManifests := filepath.Join(edgeadmConf.ManifestsDir, manifests.APP_TUNNEL_CORDDNS)
	TunnelCoredns := common.ReadYaml(userManifests, manifests.TunnelCorednsYaml)
	err = kubeclient.DeleteResourceWithFile(client, TunnelCoredns, option)
	if err != nil {
		return err
	}
	klog.Infof("Delete %s success!", manifests.APP_TUNNEL_CORDDNS)

	return err

}

func deleteEdgeHealthAddon(c workflow.RunData) error {
	_, edgeadmConf, client, err := getInitData(c)
	if err != nil {
		return err
	}

	if ok := common.CheckIfEdgeAppDeletable(client); !ok {
		klog.Info("Can not Delete Edge Apps, cluster has remaining edge nodes!")
		return nil
	}

	if err := common.DeleteEdgeHealth(client, edgeadmConf.ManifestsDir); err != nil {
		klog.Errorf("Deploy edge health, error: %s", err)
		return err
	}

	return err
}

func deleteServiceGroupAddon(c workflow.RunData) error {
	_, edgeadmConf, client, err := getInitData(c)
	if err != nil {
		return err
	}

	if ok := common.CheckIfEdgeAppDeletable(client); !ok {
		klog.Info("Can not Delete Edge Apps, cluster has remaining edge nodes!")
		return nil
	}

	if err := common.DeleteServiceGroup(client, edgeadmConf.ManifestsDir); err != nil {
		klog.Errorf("Delete serivce group, error: %s", err)
		return err
	}

	klog.Infof("Delete service-group success!")

	return err
}

func recoverKubeConfig(c workflow.RunData) error {
	initConfiguration, _, client, err := getInitData(c)
	if err != nil {
		return err
	}

	if ok := common.CheckIfEdgeAppDeletable(client); !ok {
		klog.Info("Can not Delete Edge Apps, cluster has remaining edge nodes!")
		return nil
	}

	if err := common.RecoverKubeProxyKubeconfig(client); err != nil {
		klog.Errorf("Recover kube-proxy config, error: %s", err)
		return err
	}

	if err := common.RecoverKubernetesEndpoint(client); err != nil {
		klog.Errorf("Recover kubernetes endpoint, error: %s", err)
		return err
	}

	if len(initConfiguration.APIServer.CertSANs) > 0 {
		certSANs := initConfiguration.APIServer.CertSANs
		if err := common.RecoverClusterInfoKubeconfig(client, certSANs); err != nil {
			klog.Errorf("Recover cluster-info config, error: %s", err)
			return err
		}
	}

	klog.Infof("Recover Kubernetes cluster config support marginal autonomy success")

	return err
}
