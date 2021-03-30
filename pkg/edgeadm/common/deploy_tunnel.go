package common

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests"
	"github.com/superedge/superedge/pkg/util/kubeclient"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func DeployTunnelCloud(clientSet kubernetes.Interface, manifestsDir, caCertFile, caKeyFile, tunnelCloudToken string) error {
	nodes, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
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
	serviceCert, serviceKey, err := GetServiceCert("TunnelCloudService", caCertFile, caKeyFile, dns, masterIPs)
	if err != nil {
		return err
	}

	//todo: "TunnelCloudClient"
	tunnelProxyServerKey, tunnelProxyServerCrt, err := GetServiceCert("TunnelCloudClient", caCertFile, caKeyFile, []string{}, []string{})
	if err != nil {
		return err
	}

	option := map[string]interface{}{
		"TunnelCloudEdgeToken":                tunnelCloudToken,
		"TunnelPersistentConnectionServerKey": base64.StdEncoding.EncodeToString(serviceKey),
		"TunnelPersistentConnectionServerCrt": base64.StdEncoding.EncodeToString(serviceCert),
		"TunnelProxyServerKey":                base64.StdEncoding.EncodeToString(tunnelProxyServerKey),
		"TunnelProxyServerCrt":                base64.StdEncoding.EncodeToString(tunnelProxyServerCrt),
	}

	userManifests := filepath.Join(manifestsDir, manifests.APP_TUNNEL_CLOUD)
	tunnelCloudYaml := ReadYaml(userManifests, manifests.TunnelCloudYaml)
	err = kubeclient.CreateResourceWithFile(clientSet, tunnelCloudYaml, option)
	if err != nil {
		return err
	}

	fmt.Println("Create tunnel-cloud.yaml success!")

	return nil
}

func GetTunnelCloudPort(clientSet kubernetes.Interface) (int32, error) {
	var tunnelCloudNodePort int32 = 0
	for { //Make sure tunnel-cloud success created
		coredns, err := clientSet.CoreV1().Services(
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

func DeployTunnelEdge(clientSet kubernetes.Interface, manifestsDir,
	caCertFile, caKeyFile, tunnelCloudToken string, tunnelCloudNodePort int32) error {

	caCert, _, err := GetRootCartAndKey(caCertFile, caKeyFile)
	if err != nil {
		return err
	}

	caClientCert, caClientKey, err := GetClientCert(
		"TunnelCloudClient", caCertFile, caKeyFile)
	if err != nil {
		return err
	}

	masterIps, err := GetMasterIps(clientSet)
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

	userManifests := filepath.Join(manifestsDir, manifests.APP_TUNNEL_EDGE)
	tunnelEdgeYaml := ReadYaml(userManifests, manifests.TunnelEdgeYaml)
	err = kubeclient.CreateResourceWithFile(clientSet, tunnelEdgeYaml, option)
	if err != nil {
		return err
	}

	return nil
}
