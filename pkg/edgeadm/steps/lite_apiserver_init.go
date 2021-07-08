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
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
	phases "k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/join"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"

	"github.com/superedge/superedge/pkg/edgeadm/cmd"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeclient"
)

func NewLiteApiServerInitPhase(config *cmd.EdgeadmConfig) workflow.Phase {
	return workflow.Phase{
		Name:  "lite-apiserver init",
		Short: "Install lite-apiserver on edge node",
		Run:   installLiteAPIServer,
		RunIf: func(c workflow.RunData) (bool, error) {
			return config.IsEnableEdge, nil
		},
		InheritFlags: []string{},
	}
}

// installLiteAPIServer executes install lite-apiserver logic.
func installLiteAPIServer(c workflow.RunData) error {
	data, ok := c.(phases.JoinData)
	if !ok {
		return errors.New("installLiteAPIServer phase invoked with an invalid data struct")
	}

	if data.Cfg().ControlPlane != nil {
		return nil
	}

	// Deploy LiteAPIServer
	isDeploy, err := isRunningLiteAPIServer()
	if isDeploy || err != nil {
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

	if err := deployLiteAPIServer(kubeClient, data); err != nil {
		klog.Errorf("Deploy LiteAPIServer error: %v", err)
		return err
	}

	//node kube-api-server addr set lite-api-server addr if deploy lite-api-server success.
	for _, cluster := range tlsBootstrapCfg.Clusters {
		cluster.Server = constant.LiteAPIServerAddr
	}
	return nil
}

func isRunningLiteAPIServer() (bool, error) {
	cmdRun := fmt.Sprintf(constant.LiteAPIServerStatusCmd)
	if _, _, err := util.RunLinuxCommand(cmdRun); err != nil {
		klog.Warningf("Running linux command: %s error: %v", cmdRun, err)
		return false, nil
	}
	return true, nil
}

func initKubeClient(data phases.JoinData, tlsBootstrapCfg *clientcmdapi.Config) (*kubernetes.Clientset, error) {
	// Write the bootstrap kubelet config file or the TLS-Bootstrapped kubelet config file down to disk
	klog.V(1).Infof("[kubelet-start] writing bootstrap kubelet config file at %s", kubeadmconstants.GetBootstrapKubeletKubeConfigPath())
	for _, c := range tlsBootstrapCfg.Clusters {
		server := c.Server
		address, err := url.Parse(server)
		if err != nil {
			return nil, err
		}
		if net.ParseIP(address.Hostname()) == nil {
			c.Server = server
		} else if address.Port() != "" {
			c.Server = fmt.Sprintf("%s://%s:%s", address.Scheme, constant.AddonAPIServerDomain, address.Port())
		} else {
			c.Server = fmt.Sprintf("%s://%s", address.Scheme, constant.AddonAPIServerDomain)

		}
	}
	if err := kubeconfigutil.WriteToDisk(kubeadmconstants.GetBootstrapKubeletKubeConfigPath(), tlsBootstrapCfg); err != nil {
		return nil, errors.Wrap(err, "couldn't save bootstrap-kubelet.conf to disk")
	}
	bootstrapClient, err := kubeconfigutil.ClientSetFromFile(kubeadmconstants.GetBootstrapKubeletKubeConfigPath())
	if err != nil {
		return nil, errors.Errorf("couldn't create client from kubeconfig file %q", kubeadmconstants.GetBootstrapKubeletKubeConfigPath())
	}
	return bootstrapClient, nil
}

func deployLiteAPIServer(kubeClient *kubernetes.Clientset, data phases.JoinData) error {
	liteApiServerConfigMap, err := kubeClient.CoreV1().ConfigMaps(constant.NamespaceEdgeSystem).Get(context.TODO(), constant.EdgeCertCM, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if err := generateLiteAPIServerCert(liteApiServerConfigMap.Data); err != nil {
		klog.Errorf("Generate lite-apiserver cert, error: %v", err)
		return err
	}
	if err := createLiteAPIServerConfig(data); err != nil {
		klog.Errorf("Create lite-apiserver config, error: %v", err)
		return err
	}

	if err := startLiteAPIServer(); err != nil {
		klog.Errorf("Start lite-apiserver, error: %v", err)
		return err
	}
	klog.Infof("Deploy lite-apiserver success!")
	return nil
}

func generateLiteAPIServerCert(liteApiServerConfigMap map[string]string) error {
	ca, ok := liteApiServerConfigMap[constant.KubeAPICACrt]
	if !ok {
		return fmt.Errorf("Get lite-apiserver configMap %s value nil\n", constant.KubeAPICACrt)
	}
	key, ok := liteApiServerConfigMap[constant.LiteAPIServerKey]
	if !ok {
		return fmt.Errorf("Get lite-apiserver configMap %s value nil\n", constant.LiteAPIServerKey)
	}
	crt, ok := liteApiServerConfigMap[constant.LiteAPIServerCrt]
	if !ok {
		return fmt.Errorf("Get lite-apiserver configMap %s value nil\n", constant.LiteAPIServerCrt)
	}
	tls, ok := liteApiServerConfigMap[constant.LiteAPIServerTLSJSON]
	if !ok {
		return fmt.Errorf("Get lite-apiserver configMap %s value nil\n", constant.LiteAPIServerTLSJSON)
	}

	cmds := []string{
		fmt.Sprintf("mkdir -p %s", constant.KubeEdgePath),
		fmt.Sprintf("mkdir -p %s", constant.KubePkiPath),
		fmt.Sprintf("cat << EOF >%s \n%s\nEOF", constant.LiteAPIServerCACertPath, ca),
		fmt.Sprintf("cat << EOF >%s \n%s\nEOF", constant.LiteAPIServerKeyPath, key),
		fmt.Sprintf("cat << EOF >%s \n%s\nEOF", constant.LiteAPIServerCrtPath, crt),
		fmt.Sprintf("cat << EOF >%s \n%s\nEOF", constant.LiteAPIServerTLSPath, tls),
	}
	for _, cmd := range cmds {
		if _, _, err := util.RunLinuxCommand(cmd); err != nil {
			klog.Errorf("Running linux command: %s error: %v", cmd, err)
			return err
		}
	}
	return nil
}

func createLiteAPIServerConfig(data phases.JoinData) error {
	if data.Cfg().Discovery.BootstrapToken == nil {
		return errors.New("Get bootstrap token nil")
	}
	addr, port, err := util.SplitHostPortWithDefaultPort(data.Cfg().Discovery.BootstrapToken.APIServerEndpoint, "443")
	if err != nil {
		return fmt.Errorf("Get kube-api addr: %s, port: %s, form bootstrap token error: %v\n", addr, port, err)
	}
	klog.V(4).Infof("Get kube-api-server addr: %v", addr)

	liteAIPServerConfigTemplate := strings.ReplaceAll(constant.LiteAPIServerTemplate, "${MASTER_PORT}", port)
	if net.ParseIP(addr) == nil {
		liteAIPServerConfigTemplate = strings.ReplaceAll(liteAIPServerConfigTemplate, "${MASTER_IP}", addr)
	} else {
		liteAIPServerConfigTemplate = strings.ReplaceAll(liteAIPServerConfigTemplate, "${MASTER_IP}", constant.AddonAPIServerDomain)

	}
	cmds := []string{
		fmt.Sprintf(`echo "%s" > %s`, liteAIPServerConfigTemplate, constant.LiteAPIServerConfFile),
	}
	for _, cmd := range cmds {
		if _, _, err := util.RunLinuxCommand(cmd); err != nil {
			klog.Errorf("Running linux command: %s error: %v", cmd, err)
			return err
		}
	}
	return nil
}

func startLiteAPIServer() error {
	if _, _, err := util.RunLinuxCommand(constant.LiteAPIServerRestartCmd); err != nil {
		klog.Errorf("Running linux command: %s error: %v", constant.LiteAPIServerRestartCmd, err)
		return err
	}
	return nil
}

func NewAddEdgeNodeLabelPhase(config *cmd.EdgeadmConfig) workflow.Phase {
	return workflow.Phase{
		Name:   "lite-apiserver init",
		Short:  "Install lite-apiserver on edge node",
		Run:    addEdgeNodeLabel,
		Hidden: true,
		RunIf: func(c workflow.RunData) (bool, error) {
			data, ok := c.(phases.JoinData)
			if !ok {
				return false, errors.New("installLiteAPIServer phase invoked with an invalid data struct")
			}
			return config.IsEnableEdge && data.Cfg().ControlPlane == nil, nil
		},
		InheritFlags: []string{},
	}
}

func addEdgeNodeLabel(c workflow.RunData) error {
	data, ok := c.(phases.JoinData)
	if !ok {
		return errors.New("installLiteAPIServer phase invoked with an invalid data struct")
	}

	if data.Cfg().ControlPlane != nil {
		return nil
	}
	kubeletConf := filepath.Join(kubeadmconstants.KubernetesDir, kubeadmconstants.KubeletKubeConfigFileName)
	clientSet, err := kubeclient.GetClientSet(kubeletConf)
	if err != nil {
		return err
	}
	masterLabel := map[string]string{
		constant.EdgeNodeLabelKey:     constant.EdgeNodeLabelValueEnable,
		constant.EdgehostnameLabelKey: data.Cfg().NodeRegistration.Name,
	}

	if err := kubeclient.AddNodeLabel(clientSet, data.Cfg().NodeRegistration.Name, masterLabel); err != nil {
		klog.Errorf("Add edged Node node label error: %v", err)
		return err
	}
	return nil
}
