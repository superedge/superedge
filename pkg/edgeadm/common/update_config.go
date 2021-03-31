package common

import (
	"context"
	"errors"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"strconv"
	"time"
)

func UpdateKubeProxyKubeconfig(kubeClient kubernetes.Interface) error {
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

func UpdateKubernetesEndpoint(clientSet kubernetes.Interface) error {
	endpoint, err := clientSet.CoreV1().Endpoints(
		constant.NAMESPACE_DEFAULT).Get(context.TODO(), "kubernetes", metav1.GetOptions{})
	if err != nil {
		return err
	}

	annotations := make(map[string]string)
	annotations[constant.EdgeLocalPort] = "51003"
	annotations[constant.EdgeLocalHost] = "127.0.0.1"
	endpoint.Annotations = annotations
	if _, err := clientSet.CoreV1().Endpoints(
		constant.NAMESPACE_DEFAULT).Update(context.TODO(), endpoint, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}
