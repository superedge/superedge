package revert

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	"superedge/pkg/edgeadm/common"
	"superedge/pkg/edgeadm/constant"
	"superedge/pkg/edgeadm/constant/manifests"
	"superedge/pkg/util"
	"superedge/pkg/util/kubeclient"
)

func (r *revertAction) runKubeamdRevert() error {
	if err := common.DeployHelperJob(r.clientSet,
		manifests.HelperJobYaml, constant.ACTION_REVERT, constant.NODE_ROLE_NODE); err != nil {
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
	for appName, yamlFile := range yamlMap {
		if err := common.DeleteByYamlFile(r.clientSet, yamlFile); err != nil {
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
		manifests.HelperJobYaml, constant.ACTION_REVERT, constant.NODE_ROLE_MASTER); err != nil {
		return err
	}
	//if err := r.waitingKubeAPIRevert(30); err != nil {
	//	return err
	//}

	util.OutSuccessMessage("Deploy Kubeadm Cluster Revert To Edge Cluster Success!")

	return nil
}

func (r *revertAction) deleteLiteApiServerCert() error {
	return r.clientSet.CoreV1().ConfigMaps("kube-system").
		Delete(context.TODO(), constant.EDGE_CERT_CM, metav1.DeleteOptions{})
}

func (r *revertAction) deleteTunnelCloud() error {
	option := map[string]interface{}{
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
		"HmacKey": "token",
	}

	if err := kubeclient.DeleteResourceWithFile(r.clientSet, manifests.EdgeHealthYaml, option); err != nil {
		return err
	}

	fmt.Println("Revert edge-health success!")

	return nil
}

func (r *revertAction) revertKubeProxyKubeconfig() error {
	kubeClient := r.clientSet

	kubeProxyCM, err := kubeClient.CoreV1().ConfigMaps("kube-system").Get(context.TODO(), "kube-proxy", metav1.GetOptions{})
	if err != nil {
		return err
	}

	proxyConfig, ok := kubeProxyCM.Data["kubeconfig.conf"]
	if !ok {
		return errors.New("Get kube-proxy kubeconfig.conf nil\n")
	}

	config, err := clientcmd.Load([]byte(proxyConfig))
	if err != nil {
		return err
	}

	masterIps, err := common.GetMasterIps(r.clientSet)
	if err != nil {
		return err
	}
	if len(masterIps) < 1 {
		return errors.New("Get master ip nil\n")
	}

	for key := range config.Clusters {
		config.Clusters[key].Server = "https://" + masterIps[0] + ":6443"
	}

	content, err := clientcmd.Write(*config)
	if err != nil {
		return err
	}
	kubeProxyCM.Data["kubeconfig.conf"] = string(content)

	if _, err := kubeClient.CoreV1().ConfigMaps("kube-system").Update(context.TODO(), kubeProxyCM, metav1.UpdateOptions{}); err != nil {
		return err
	}

	daemonSets, err := kubeClient.AppsV1().DaemonSets("kube-system").Get(context.TODO(), "kube-proxy", metav1.GetOptions{})
	if err != nil {
		return err
	}

	delete(daemonSets.Spec.Template.Annotations, constant.UpdateKubeProxyTime)
	if _, err := kubeClient.AppsV1().DaemonSets("kube-system").Update(context.TODO(), daemonSets, metav1.UpdateOptions{}); err != nil {
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
