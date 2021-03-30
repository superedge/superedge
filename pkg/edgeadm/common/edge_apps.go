package common

import (
	"context"
	"encoding/base64"
	"fmt"

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

	tunnelCloudYaml := ReadYaml(manifestsDir+"/"+manifests.APP_TUNNEL_CLOUD, manifests.TunnelCloudYaml)
	err = kubeclient.CreateResourceWithFile(clientSet, tunnelCloudYaml, option)
	if err != nil {
		return err
	}

	fmt.Println("Create tunnel-cloud.yaml success!")

	return nil
}
