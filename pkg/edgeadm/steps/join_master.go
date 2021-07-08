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
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	phases "k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/join"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"github.com/superedge/superedge/pkg/edgeadm/cmd"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeclient"
)

func NewJoinPreparePhase(config *cmd.EdgeadmConfig) workflow.Phase {
	return workflow.Phase{
		Name:  "join-prepare",
		Short: "join prepare for master or edge node",
		Run:   joinPreparePhase,
		RunIf: func(c workflow.RunData) (bool, error) {
			return config.IsEnableEdge, nil
		},
		InheritFlags: []string{},
	}
}

// joinMasterPreparePhase prepare join master logic.
func joinPreparePhase(c workflow.RunData) error {
	data, ok := c.(phases.JoinData)
	if !ok {
		return errors.New("installLiteAPIServer phase invoked with an invalid data struct")
	}

	if err := configControlPlaneInfo(data.Cfg()); err != nil {
		klog.Errorf("Config ControlPlaneInfo, error: %v")
		return err
	}

	tlsBootstrapCfg, err := data.TLSBootstrapCfg()
	if err != nil {
		return err
	}

	kubeClient, err := initKubeClient(data, tlsBootstrapCfg)
	if err != nil {
		klog.Errorf("Get kube client error: %v", err)
		return err
	}

	// Deletes the bootstrapKubeConfigFile, so the credential used for TLS bootstrap is removed from disk
	defer func() {
		os.Remove(kubeadmconstants.GetBootstrapKubeletKubeConfigPath())
		os.Remove(constant.KubeadmCertPath)
	}()

	// prepare join master node
	if data.Cfg().ControlPlane != nil {
		if err := prepareJoinMasterNode(kubeClient, data); err != nil {
			klog.Errorf("Prepare Join master node, error: %v", err)
			return nil
		}
	}

	// prepare join edge node
	if data.Cfg().ControlPlane == nil {
		if err := prepareJoinEdgeNode(kubeClient, data); err != nil {
			klog.Errorf("Prepare Join edge node, error: %v", err)
			return nil
		}
	}

	return nil
}

func configControlPlaneInfo(joinConfiguration *kubeadm.JoinConfiguration) error {
	endpoint := joinConfiguration.Discovery.BootstrapToken.APIServerEndpoint
	host, port, err := util.SplitHostPortIgnoreMissingPort(endpoint)
	if err != nil {
		return errors.Errorf("Invalid APIServerEndpoint: %s", endpoint)
	}
	if port != "" {
		endpoint = net.JoinHostPort(constant.AddonAPIServerDomain, port)
	} else {
		endpoint = constant.AddonAPIServerDomain
	}
	// if domain instead of ipv4 address was provided, we won't update control plane info
	if net.ParseIP(host) == nil {
		return nil
	}
	joinConfiguration.Discovery = kubeadm.Discovery{
		BootstrapToken: &kubeadm.BootstrapTokenDiscovery{
			APIServerEndpoint:        endpoint,
			Token:                    joinConfiguration.Discovery.BootstrapToken.Token,
			CACertHashes:             joinConfiguration.Discovery.BootstrapToken.CACertHashes,
			UnsafeSkipCAVerification: joinConfiguration.Discovery.BootstrapToken.UnsafeSkipCAVerification,
		},
		File:              joinConfiguration.Discovery.File,
		TLSBootstrapToken: joinConfiguration.Discovery.TLSBootstrapToken,
		Timeout:           joinConfiguration.Discovery.Timeout,
	}
	return ensureHostDNS(host)
}

func ensureHostDNS(publicIP string) error {
	cmds := []string{
		constant.ResetDNSCmd,
		fmt.Sprintf("cat << EOF >>%s \n%s\n%s\n%s\nEOF", constant.HostsFilePath, constant.HostDNSBeginMark, publicIP+" "+constant.AddonAPIServerDomain, constant.HostDNSEndMark),
	}
	for _, cmd := range cmds {
		if _, _, err := util.RunLinuxCommand(cmd); err != nil {
			klog.Errorf("Running linux command: %s error: %v", cmd, err)
			return err
		}
	}
	return nil
}

func prepareJoinEdgeNode(kubeClient *kubernetes.Clientset, data phases.JoinData) error {
	joinCfg, err := data.InitCfg()
	if err != nil {
		return err
	}

	// Set kubelet cluster-dns
	edgeInfoConfigMap, err := kubeClient.CoreV1().ConfigMaps(constant.NamespaceEdgeSystem).Get(context.TODO(), constant.EdgeCertCM, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Get configMap: %s, error: %v", constant.EdgeCertCM, err)
		return err
	}
	edgeCoreDNSClusterIP, ok := edgeInfoConfigMap.Data[constant.EdgeCoreDNSClusterIP]
	if !ok {
		return fmt.Errorf("Get lite-apiserver configMap %s value nil\n", constant.LiteAPIServerTLSJSON)
	}
	edgeCoreDNSClusterIP = strings.Replace(edgeCoreDNSClusterIP, "\n", "", -1)

	if joinCfg.NodeRegistration.KubeletExtraArgs == nil {
		joinCfg.NodeRegistration.KubeletExtraArgs = make(map[string]string)
	}
	joinCfg.NodeRegistration.KubeletExtraArgs["cluster-dns"] = edgeCoreDNSClusterIP
	klog.V(4).Infof("Get edge-coredns clusterIP %s", edgeCoreDNSClusterIP)

	return nil
}

func prepareJoinMasterNode(kubeClient *kubernetes.Clientset, data phases.JoinData) error {
	if err := setKubeAPIServerPatch(kubeClient, data.PatchesDir()); err != nil {
		klog.Errorf("Add kube-apiserver patch error: %v", err)
		return nil
	}
	return nil
}

func setKubeAPIServerPatch(kubeClient *kubernetes.Clientset, patchesDir string) error {
	edgeInfoConfigMap, err := kubeClient.CoreV1().ConfigMaps(constant.NamespaceEdgeSystem).Get(context.TODO(), constant.EdgeCertCM, metav1.GetOptions{})
	if err != nil {
		return err
	}

	tunnelCoreDNSClusterIP, ok := edgeInfoConfigMap.Data[constant.TunnelCoreDNSClusterIP]
	if !ok {
		return fmt.Errorf("Get tunnelCoreDNSClusterIP configMap %s value nil\n", constant.TunnelCoreDNSClusterIP)
	}

	option := map[string]interface{}{
		"TunnelCoreDNSClusterIP": strings.Replace(tunnelCoreDNSClusterIP, "\n", "", -1),
	}
	kubeAPIServerPatch, err := kubeclient.ParseString(constant.KubeAPIServerPatchYaml, option)
	if err != nil {
		klog.Errorf("Parse %s yaml: %s, option: %v, error: %v", constant.KubeAPIServerPatch, option, err)
		return err
	}

	if err := util.WriteFile(patchesDir+constant.KubeAPIServerPatch, string(kubeAPIServerPatch)); err != nil {
		klog.Errorf("Write file: %s, error: %v", constant.KubeAPIServerPatch, err)
		return err
	}

	return nil
}
