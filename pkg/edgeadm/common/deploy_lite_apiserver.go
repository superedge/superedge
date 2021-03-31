package common

import (
	"context"
	"errors"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"path/filepath"
)

func CreateLiteApiServerCert(clientSet kubernetes.Interface, manifestsDir, caCertFile, caKeyFile string) error {
	clientSet.CoreV1().ConfigMaps("kube-system").Delete(
		context.TODO(), constant.EDGE_CERT_CM, metav1.DeleteOptions{})

	kubeService, err := clientSet.CoreV1().Services(
		constant.NAMESPACE_DEFAULT).Get(context.TODO(), constant.SERVICE_KUBERNETES, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if kubeService.Spec.ClusterIP == "" {
		return errors.New("Get kubernetes service clusterIP nil\n")
	}

	liteApiServerCrt, liteApiServerKey, err :=
		GetServiceCert("LiteApiServer", caCertFile, caKeyFile, []string{"127.0.0.1"}, []string{kubeService.Spec.ClusterIP})
	if err != nil {
		return err
	}

	userLiteAPIServer := filepath.Join(manifestsDir, manifests.APP_lITE_APISERVER)
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: constant.EDGE_CERT_CM,
		},
		Data: map[string]string{
			constant.LITE_API_SERVER_CRT:     string(liteApiServerCrt),
			constant.LITE_API_SERVER_KEY:     string(liteApiServerKey),
			constant.LITE_API_SERVER_TLS_CFG: constant.LiteApiServerTlsCfg,
			manifests.APP_lITE_APISERVER:     ReadYaml(userLiteAPIServer, manifests.LiteApiServerYaml),
		},
	}

	if _, err := clientSet.CoreV1().ConfigMaps("kube-system").
		Create(context.TODO(), configMap, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}
