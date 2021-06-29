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
	"errors"
	"fmt"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	phases "k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/join"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"github.com/superedge/superedge/pkg/edgeadm/cmd"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeclient"
)

func NewJoinMasterPreparePhase(config *cmd.EdgeadmConfig) workflow.Phase {
	return workflow.Phase{
		Name:  "join-master-prepare",
		Short: "join master prepare for master node",
		Run:   joinMasterPreparePhase,
		RunIf: func(c workflow.RunData) (bool, error) {
			return config.IsEnableEdge, nil
		},
		InheritFlags: []string{},
	}
}

// joinMasterPreparePhase prepare join master logic.
func joinMasterPreparePhase(c workflow.RunData) error {
	data, ok := c.(phases.JoinData)
	if !ok {
		return errors.New("installLiteAPIServer phase invoked with an invalid data struct")
	}

	if data.Cfg().ControlPlane == nil {
		return nil
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
