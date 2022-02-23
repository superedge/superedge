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

package revert

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/superedge/superedge/pkg/edgeadm/common"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *revertAction) runKubeamdRevert() error {
	if err := common.DeployHelperJob(r.clientSet,
		manifests.HelperJobYaml, constant.ActionRevert, constant.NodeRoleNode); err != nil {
		return err
	}

	// Delete tunnel-edge
	if err := r.deleteTunnelEdge(); err != nil {
		return err
	}

	yamlMap := map[string]string{
		manifests.APP_TUNNEL_CORDDNS:              manifests.TunnelCorednsYaml,
		manifests.APP_EDGE_HEALTH_ADMISSION:       manifests.EdgeHealthAdmissionYaml,
		manifests.APP_EDGE_HEALTH_WEBHOOK:         manifests.EdgeHealthWebhookConfigYaml,
		manifests.APP_APPLICATION_GRID_WRAPPER:    manifests.ApplicationGridWrapperYaml,
		manifests.APP_APPLICATION_GRID_CONTROLLER: manifests.ApplicationGridControllerYaml,
	}
	option := map[string]interface{}{
		"Namespace": constant.NamespaceEdgeSystem,
		"CABundle":  "",
		"ServerCrt": "",
		"ServerKey": "",
	}
	for appName, yamlFile := range yamlMap {
		if err := kubeclient.DeleteResourceWithFile(r.clientSet, yamlFile, option); err != nil {
			return err
		}
		fmt.Printf("Revert %s success!\n", appName)
	}

	// Delete tunnel-health
	if err := r.deleteEdgeHealth(); err != nil {
		return err
	}

	if err := r.revertKubernetesEndpoint(); err != nil {
		return err
	}

	if err := r.revertKubeProxyKubeconfig(); err != nil {
		return err
	}

	// Delete tunnel-cloud
	if err := r.deleteTunnelCloud(); err != nil {
		return err
	}

	if err := r.deleteLiteApiServerCert(); err != nil {
		return err
	}

	if err := common.DeployHelperJob(r.clientSet,
		manifests.HelperJobYaml, constant.ActionRevert, constant.NodeRoleMaster); err != nil {
		return err
	}

	if err := r.deleteNodeLabel(); err != nil {
		return err
	}

	if err := r.deleteEdgeSystemNamespace(); err != nil {
		return err
	}

	util.OutSuccessMessage("Deploy Kubeadm Cluster Revert To Edge Cluster Success!")

	return nil
}

func (r *revertAction) deleteLiteApiServerCert() error {
	return r.clientSet.CoreV1().ConfigMaps(constant.NamespaceEdgeSystem).
		Delete(context.TODO(), constant.EdgeCertCM, metav1.DeleteOptions{})
}

func (r *revertAction) deleteTunnelCloud() error {
	option := map[string]interface{}{
		"Namespace":                           constant.NamespaceEdgeSystem,
		"TunnelCloudEdgeToken":                "tunnelCloudEdgeToken",
		"TunnelProxyServerKey":                base64.StdEncoding.EncodeToString([]byte("tunnelProxyServerKey")),
		"TunnelProxyServerCrt":                base64.StdEncoding.EncodeToString([]byte("tunnelProxyServerCrt")),
		"TunnelPersistentConnectionServerKey": base64.StdEncoding.EncodeToString([]byte("tunnelPersistentConnectionServerKey")),
		"TunnelPersistentConnectionServerCrt": base64.StdEncoding.EncodeToString([]byte("tunnelPersistentConnectionServerCrt")),
	}

	return kubeclient.DeleteResourceWithFile(r.clientSet, manifests.TunnelCloudYaml, option)
}

func (r *revertAction) deleteTunnelEdge() error {
	option := map[string]interface{}{
		"Namespace":                      constant.NamespaceEdgeSystem,
		"MasterIP":                       "127.0.0.1",
		"TunnelCloudEdgeToken":           "tunnelCloudEdgeToken",
		"KubernetesCaCert":               base64.StdEncoding.EncodeToString([]byte("kubernetesCaCert")),
		"KubeletClientKey":               base64.StdEncoding.EncodeToString([]byte("kubeletClientKey")),
		"KubeletClientCrt":               base64.StdEncoding.EncodeToString([]byte("kubeletClientCrt")),
		"TunnelPersistentConnectionPort": "tunnelPersistentConnectionPort",
	}

	if err := kubeclient.DeleteResourceWithFile(r.clientSet, manifests.TunnelEdgeYaml, option); err != nil {
		return err
	}

	fmt.Println("Revert tunnel-edge success!")

	return nil
}

func (r *revertAction) deleteEdgeHealth() error {
	option := map[string]interface{}{
		"Namespace": constant.NamespaceEdgeSystem,
		"HmacKey":   "token",
	}

	if err := kubeclient.DeleteResourceWithFile(r.clientSet, manifests.EdgeHealthYaml, option); err != nil {
		return err
	}

	fmt.Println("Revert edge-health success!")

	return nil
}

func (r *revertAction) deleteNodeLabel() error {
	kubeclient := r.clientSet

	labelSelector := fmt.Sprintf("%s=%s", constant.EdgeChangeLabelKey, constant.EdgeChangeLabelValueEnable)
	nodes, err := kubeclient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return err
	}

	for _, node := range nodes.Items {
		if _, ok := node.Labels[constant.EdgeChangeLabelKey]; ok {
			delete(node.Labels, constant.EdgeChangeLabelKey)
		}
		_, err = kubeclient.CoreV1().Nodes().Update(context.TODO(), &node, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *revertAction) revertKubeProxyKubeconfig() error {
	kubeClient := r.clientSet

	if err := kubeClient.AppsV1().DaemonSets(constant.NamespaceEdgeSystem).Delete(context.TODO(), constant.EdgeKubeProxy, metav1.DeleteOptions{}); err != nil {
		return err
	}

	if err := kubeClient.CoreV1().ConfigMaps(constant.NamespaceEdgeSystem).Delete(context.TODO(), constant.EdgeKubeProxy, metav1.DeleteOptions{}); err != nil {
		return err
	}

	if err := kubeClient.CoreV1().ServiceAccounts(constant.NamespaceEdgeSystem).Delete(context.TODO(), constant.KubeProxy, metav1.DeleteOptions{}); err != nil {
		return err
	}

	if err := common.RecoverKubeProxy(kubeClient); err != nil {
		return err
	}

	return nil
}

func (r *revertAction) revertKubernetesEndpoint() error {
	endpoint, err := r.clientSet.CoreV1().Endpoints("default").Get(context.TODO(), "kubernetes", metav1.GetOptions{})
	if err != nil {
		return err
	}

	delete(endpoint.Annotations, constant.EdgeLocalPort)
	delete(endpoint.Annotations, constant.EdgeLocalHost)
	if _, err := r.clientSet.CoreV1().Endpoints("default").Update(context.TODO(), endpoint, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func (r *revertAction) deleteEdgeSystemNamespace() error {
	if err := r.clientSet.CoreV1().Namespaces().Delete(context.TODO(), constant.NamespaceEdgeSystem, metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}
