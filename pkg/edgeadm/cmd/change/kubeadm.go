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
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	k8scert "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"

	"superedge/pkg/edgeadm/common"
	"superedge/pkg/edgeadm/constant"
	"superedge/pkg/edgeadm/constant/manifests"
	"superedge/pkg/util"
	"superedge/pkg/util/kubeclient"
)

func (c *changeAction) runKubeamdChange() error {
	// Create APPs that do not affect the use of the original cluster
	affectYamls := map[string]string{
		manifests.APP_TUNNEL_CORDDNS: common.ReadYaml(c.manifests+"/"+manifests.APP_TUNNEL_CORDDNS, manifests.TunnelCorednsYaml),
	}
	for appName, yamlFile := range affectYamls {
		if err := common.CreateByYamlFile(c.clientSet, yamlFile); err != nil {
			return err
		}
		fmt.Printf("Create %s success!\n", appName)
	}

	// Create tunnel-cloud
	tunnelCloudToken, err := c.deployTunnelCloud()
	if err != nil {
		return err
	}

	tunnelCloudNodePort, err := c.waitTunnelCloudReady()
	if err != nil {
		return err
	}

	// Create tunnel-edge
	if err := c.deployTunnelEdge(tunnelCloudNodePort, tunnelCloudToken); err != nil {
		return err
	}

	// Deploy lite-api-server
	if err := c.createLiteApiServerCert(); err != nil {
		return err
	}

	helperJobYaml := common.ReadYaml(c.manifests+"/"+manifests.APP_HELPER_JOB, manifests.HelperJobYaml)
	if err := common.DeployHelperJob(c.clientSet, helperJobYaml, constant.ACTION_CHANGE, constant.NODE_ROLE_NODE); err != nil {
		return err
	}

	yamlMap := map[string]string{
		manifests.APP_EDGE_HEALTH_ADMISSION:       common.ReadYaml(c.manifests+"/"+manifests.APP_EDGE_HEALTH_ADMISSION, manifests.EdgeHealthAdmissionYaml),
		manifests.APP_EDGE_HEALTH_WEBHOOK:         common.ReadYaml(c.manifests+"/"+manifests.APP_EDGE_HEALTH_WEBHOOK, manifests.EdgeHealthWebhookConfigYaml),
		manifests.APP_APPLICATION_GRID_WRAPPER:    common.ReadYaml(c.manifests+"/"+manifests.APP_APPLICATION_GRID_WRAPPER, manifests.ApplicationGridWrapperYaml),
		manifests.APP_APPLICATION_GRID_CONTROLLER: common.ReadYaml(c.manifests+"/"+manifests.APP_APPLICATION_GRID_CONTROLLER, manifests.ApplicationGridControllerYaml),
	}
	for appName, yamlFile := range yamlMap {
		if err := common.CreateByYamlFile(c.clientSet, yamlFile); err != nil {
			return err
		}
		fmt.Printf("Create %s success!\n", appName)
	}

	// apply tunnel-health
	if err := c.deployEdgeHealth(); err != nil {
		return err
	}

	if err := c.updateKubeProxyKubeconfig(); err != nil {
		return err
	}

	if err := c.updateKubernetesEndpoint(); err != nil {
		return err
	}

	if err := common.DeployHelperJob(c.clientSet, helperJobYaml, constant.ACTION_CHANGE, constant.NODE_ROLE_MASTER); err != nil {
		return err
	}

	util.OutSuccessMessage("Deploy Kubeadm Cluster Change To Edge cluster Success!")

	return nil
}

func (c *changeAction) createLiteApiServerCert() error {
	c.clientSet.CoreV1().ConfigMaps("kube-system").Delete(
		context.TODO(), constant.EDGE_CERT_CM, metav1.DeleteOptions{})

	kubeService, err := c.clientSet.CoreV1().Services(
		constant.NAMESPACE_DEFAULT).Get(context.TODO(), constant.SERVICE_KUBERNETES, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if kubeService.Spec.ClusterIP == "" {
		return errors.New("Get kubernetes service clusterIP nil\n")
	}

	liteApiServerCrt, liteApiServerKey, err :=
		c.getServiceCert("LiteApiServer", []string{"127.0.0.1"}, []string{kubeService.Spec.ClusterIP})
	if err != nil {
		return err
	}

	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: constant.EDGE_CERT_CM,
		},
		Data: map[string]string{
			constant.LITE_API_SERVER_CRT:     string(liteApiServerCrt),
			constant.LITE_API_SERVER_KEY:     string(liteApiServerKey),
			constant.LITE_API_SERVER_TLS_CFG: constant.LiteApiServerTlsCfg,
			manifests.APP_lITE_APISERVER:     common.ReadYaml(c.manifests+"/"+manifests.APP_lITE_APISERVER, manifests.LiteApiServerYaml),
		},
	}

	if _, err := c.clientSet.CoreV1().ConfigMaps("kube-system").
		Create(context.TODO(), configMap, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}

func (c *changeAction) deployTunnelCloud() (string, error) {
	nodes, err := c.clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	var masterIPs []string
	for _, node := range nodes.Items {
		if _, ok := node.Labels[constant.KubernetesDefaultRoleLabel]; ok {
			for _, address := range node.Status.Addresses {
				if address.Type == v1.NodeInternalIP || address.Type == v1.NodeExternalIP {
					masterIPs = append(masterIPs, address.Address)
				}
			}
		}
	}

	dns := []string{
		"tunnel.cloud.io",
	}
	serviceCert, serviceKey, err := c.getServiceCert("TunnelCloudService", dns, masterIPs)
	if err != nil {
		return "", err
	}

	//todo: "TunnelCloudClient"
	kubeletCert, kubeletKey, err := c.getServiceCert("TunnelCloudClient", []string{}, []string{})
	if err != nil {
		return "", err
	}

	tunnelCloudToken := util.GetRandToken(32)
	option := map[string]interface{}{
		"TunnelCloudEdgeToken":                tunnelCloudToken,
		"TunnelProxyServerKey":                base64.StdEncoding.EncodeToString(kubeletKey),
		"TunnelProxyServerCrt":                base64.StdEncoding.EncodeToString(kubeletCert),
		"TunnelPersistentConnectionServerKey": base64.StdEncoding.EncodeToString(serviceKey),
		"TunnelPersistentConnectionServerCrt": base64.StdEncoding.EncodeToString(serviceCert),
	}

	tunnelCloudYaml := common.ReadYaml(c.manifests+"/"+manifests.APP_TUNNEL_CLOUD, manifests.TunnelCloudYaml)
	err = kubeclient.CreateResourceWithFile(c.clientSet, tunnelCloudYaml, option)
	if err != nil {
		return "", err
	}

	fmt.Println("Create tunnel-cloud.yaml success!")

	return tunnelCloudToken, nil
}

func (c *changeAction) waitTunnelCloudReady() (int32, error) {
	var tunnelCloudNodePort int32 = 0
	for { //Make sure tunnel-cloud success created
		coredns, err := c.clientSet.CoreV1().Services(
			"kube-system").Get(context.TODO(), constant.SERVICE_TUNNEL_CLOUD, metav1.GetOptions{})
		if err == nil {
			for _, port := range coredns.Spec.Ports {
				tunnelCloudNodePort = port.NodePort
			}
			break
		}
		time.Sleep(time.Second)
	}

	if tunnelCloudNodePort == 0 {
		return tunnelCloudNodePort, errors.New("Get tunnel-cloud nodePort nil\n")
	}

	return tunnelCloudNodePort, nil
}

func (c *changeAction) deployTunnelEdge(tunnelCloudNodePort int32, tunnelCloudToken string) error {
	caCert, _, err := common.GetRootCartAndKey(c.caCertFile, c.caKeyFile)
	if err != nil {
		return err
	}

	caClientCert, caClientKey, err := common.GetClientCert(
		"TunnelCloudClient", c.caCertFile, c.caKeyFile)
	if err != nil {
		return err
	}

	masterIps, err := common.GetMasterIps(c.clientSet)
	if err != nil {
		return err
	}

	option := map[string]interface{}{
		"MasterIP":                       masterIps[0],
		"KubernetesCaCert":               base64.StdEncoding.EncodeToString(caCert),
		"KubeletClientKey":               base64.StdEncoding.EncodeToString(caClientKey),
		"KubeletClientCrt":               base64.StdEncoding.EncodeToString(caClientCert),
		"TunnelCloudEdgeToken":           tunnelCloudToken,
		"TunnelPersistentConnectionPort": tunnelCloudNodePort,
	}

	tunnelEdgeYaml := common.ReadYaml(c.manifests+"/"+manifests.APP_TUNNEL_EDGE, manifests.TunnelEdgeYaml)
	err = kubeclient.CreateResourceWithFile(c.clientSet, tunnelEdgeYaml, option)
	if err != nil {
		return err
	}

	fmt.Println("Create tunnel-edge.yaml success!")

	return nil
}

func (c *changeAction) deployEdgeHealth() error {
	option := map[string]interface{}{
		"HmacKey": util.GetRandToken(16),
	}

	edgeHealthYaml := common.ReadYaml(c.manifests+"/"+manifests.APP_EDGE_HEALTH, manifests.EdgeHealthYaml)
	if err := kubeclient.CreateResourceWithFile(c.clientSet, edgeHealthYaml, option); err != nil {
		return err
	}

	fmt.Println("Create edge-health.yaml success!")

	return nil
}

func (c *changeAction) getServiceCert(commonName string, dns []string, ips []string) ([]byte, []byte, error) {
	caCert, caKey, err := c.getCertAndKey()
	if err != nil {
		return nil, nil, err
	}

	certIps := []net.IP{net.ParseIP("127.0.0.1")}
	for _, ip := range ips {
		certIps = append(certIps, net.ParseIP(ip))
	}
	serverCert, serverKey, err := util.GenerateCertAndKeyConfig(caCert, caKey, &k8scert.Config{
		CommonName:   commonName,
		Organization: []string{"superedge"},
		AltNames: k8scert.AltNames{
			DNSNames: dns,
			IPs:      certIps,
		},
		Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	})

	serverCertData := util.EncodeCertPEM(serverCert)
	serverKeyData, err := keyutil.MarshalPrivateKeyToPEM(serverKey)
	if err != nil {
		return nil, nil, err
	}

	return serverCertData, serverKeyData, err
}

func (c *changeAction) getCertAndKey() (*x509.Certificate, *rsa.PrivateKey, error) {
	caCert, caKey, err := common.GetRootCartAndKey(c.caCertFile, c.caKeyFile)
	if err != nil {
		return nil, nil, err
	}

	cert, key, err := common.ParseCertAndKey(caCert, caKey)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}

func (c *changeAction) updateKubeProxyKubeconfig() error {
	kubeClient := c.clientSet
	kubeProxyCM, err := kubeClient.CoreV1().ConfigMaps(
		constant.NAMESPACE_KUBE_SYSTEM).Get(context.TODO(), "kube-proxy", metav1.GetOptions{})
	if err != nil {
		return err
	}

	proxyConfig, ok := kubeProxyCM.Data[constant.CM_KUBECONFIG_CONF]
	if !ok {
		return errors.New("Get kube-proxy kubeconfig.conf nil\n")
	}

	config, err := clientcmd.Load([]byte(proxyConfig))
	if err != nil {
		return err
	}

	for key := range config.Clusters {
		config.Clusters[key].Server = constant.APPLICAION_GRID_WRAPPER_SERVICE_ADDR
	}

	content, err := clientcmd.Write(*config)
	if err != nil {
		return err
	}
	kubeProxyCM.Data[constant.CM_KUBECONFIG_CONF] = string(content)

	if _, err := kubeClient.CoreV1().ConfigMaps(
		constant.NAMESPACE_KUBE_SYSTEM).Update(context.TODO(), kubeProxyCM, metav1.UpdateOptions{}); err != nil {
		return err
	}

	daemonSets, err := kubeClient.AppsV1().DaemonSets(
		constant.NAMESPACE_KUBE_SYSTEM).Get(context.TODO(), "kube-proxy", metav1.GetOptions{})
	if err != nil {
		return err
	}

	if len(daemonSets.Spec.Template.Annotations) == 0 {
		daemonSets.Spec.Template.Annotations = make(map[string]string)
	}
	daemonSets.Spec.Template.Annotations[constant.UpdateKubeProxyTime] = strconv.FormatInt(time.Now().Unix(), 10)

	if _, err := kubeClient.AppsV1().DaemonSets(
		constant.NAMESPACE_KUBE_SYSTEM).Update(context.TODO(), daemonSets, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func (c *changeAction) updateKubernetesEndpoint() error {
	endpoint, err := c.clientSet.CoreV1().Endpoints(
		constant.NAMESPACE_DEFAULT).Get(context.TODO(), "kubernetes", metav1.GetOptions{})
	if err != nil {
		return err
	}

	annotations := make(map[string]string)
	annotations[constant.EdgeLocalPort] = "51003"
	annotations[constant.EdgeLocalHost] = "127.0.0.1"
	endpoint.Annotations = annotations
	if _, err := c.clientSet.CoreV1().Endpoints(
		constant.NAMESPACE_DEFAULT).Update(context.TODO(), endpoint, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}
