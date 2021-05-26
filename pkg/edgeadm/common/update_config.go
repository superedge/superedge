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

package common

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	bootstrapapi "k8s.io/cluster-bootstrap/token/api"
	"k8s.io/klog/v2"

	"github.com/superedge/superedge/pkg/edgeadm/constant"
)

// runCoreDNSAddon installs CoreDNS addon to a Kubernetes cluster
func UpdateKubeConfig(client *kubernetes.Clientset) error {
	if err := UpdateKubeProxyKubeconfig(client); err != nil {
		klog.Errorf("Deploy serivce group, error: %s", err)
		return err
	}

	if err := UpdateKubernetesEndpoint(client); err != nil {
		klog.Errorf("Deploy serivce group, error: %s", err)
		return err
	}

	klog.Infof("Update Kubernetes cluster config support marginal autonomy success")

	return nil
}

func UpdateKubeProxyKubeconfig(kubeClient kubernetes.Interface) error {
	kubeProxyCM, err := kubeClient.CoreV1().ConfigMaps(
		constant.NamespcaeKubeSystem).Get(context.TODO(), constant.CMKubeProxy, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// backup original ConfigMap
	oldKubeProxyCM := kubeProxyCM.DeepCopy()
	oldKubeProxyCM.Name = constant.CMKubeProxyNoEdge
	oldKubeProxyCM.ResourceVersion = ""
	if _, err := kubeClient.CoreV1().ConfigMaps(
		constant.NamespcaeKubeSystem).Create(context.TODO(), oldKubeProxyCM, metav1.CreateOptions{}); err != nil {
		return err
	}

	proxyConfig, ok := kubeProxyCM.Data[constant.CMKubeConfig]
	if !ok {
		return errors.New("Get kube-proxy kubeconfig.conf nil\n")
	}

	config, err := clientcmd.Load([]byte(proxyConfig))
	if err != nil {
		return err
	}

	for key := range config.Clusters {
		config.Clusters[key].Server = constant.ApplicationGridWrapperServiceAddr
	}

	content, err := clientcmd.Write(*config)
	if err != nil {
		return err
	}
	kubeProxyCM.Data[constant.CMKubeConfig] = string(content)

	if _, err := kubeClient.CoreV1().ConfigMaps(
		constant.NamespcaeKubeSystem).Update(context.TODO(), kubeProxyCM, metav1.UpdateOptions{}); err != nil {
		return err
	}

	daemonSets, err := kubeClient.AppsV1().DaemonSets(
		constant.NamespcaeKubeSystem).Get(context.TODO(), "kube-proxy", metav1.GetOptions{})
	if err != nil {
		return err
	}

	if len(daemonSets.Spec.Template.Annotations) == 0 {
		daemonSets.Spec.Template.Annotations = make(map[string]string)
	}
	daemonSets.Spec.Template.Annotations[constant.UpdateKubeProxyTime] = strconv.FormatInt(time.Now().Unix(), 10)

	if _, err := kubeClient.AppsV1().DaemonSets(
		constant.NamespcaeKubeSystem).Update(context.TODO(), daemonSets, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func UpdateClusterInfoKubeconfig(kubeClient kubernetes.Interface, certSANs []string) error {
	if len(certSANs) <= 0 {
		return nil
	}
	clusterInfoCM, err := kubeClient.CoreV1().ConfigMaps(
		metav1.NamespacePublic).Get(context.TODO(), bootstrapapi.ConfigMapClusterInfo, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// backup original ConfigMap
	oldClusterInfoCM := clusterInfoCM.DeepCopy()
	oldClusterInfoCM.Name = constant.ConfigMapClusterInfoNoEdge
	oldClusterInfoCM.ResourceVersion = ""
	if _, err := kubeClient.CoreV1().ConfigMaps(
		constant.NamespcaeKubeSystem).Create(context.TODO(), oldClusterInfoCM, metav1.CreateOptions{}); err != nil {
		return err
	}

	kubeconfig, ok := clusterInfoCM.Data[bootstrapapi.KubeConfigKey]
	if !ok {
		return errors.New("Get cluster-info kubeconfig nil\n")
	}

	config, err := clientcmd.Load([]byte(kubeconfig))
	if err != nil {
		return err
	}

	kubeAPIServerPublicAddr := certSANs[0]
	for _, certSAN := range certSANs {
		address := net.ParseIP(certSAN)
		if address != nil {
			kubeAPIServerPublicAddr = address.String()
			break
		}
	}

	for key := range config.Clusters {
		srcKubeAPIServerAddr := config.Clusters[key].Server
		kubeAPIServerAddr := strings.TrimPrefix(srcKubeAPIServerAddr, "https://")
		index := strings.Index(kubeAPIServerAddr, ":")
		kubeAPIServerAddr = kubeAPIServerAddr[:index]
		dstKubeAPIServerAddr := strings.Replace(srcKubeAPIServerAddr, kubeAPIServerAddr, kubeAPIServerPublicAddr, -1)
		config.Clusters[key].Server = dstKubeAPIServerAddr
	}

	content, err := clientcmd.Write(*config)
	if err != nil {
		return err
	}
	clusterInfoCM.Data[bootstrapapi.KubeConfigKey] = string(content)

	if _, err := kubeClient.CoreV1().ConfigMaps(
		metav1.NamespacePublic).Update(context.TODO(), clusterInfoCM, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func UpdateKubernetesEndpoint(clientSet kubernetes.Interface) error {
	endpoint, err := clientSet.CoreV1().Endpoints(
		constant.NamespaceDefault).Get(context.TODO(), constant.KubernetesEndpoint, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// backup original ConfigMap
	oldEndpoint := endpoint.DeepCopy()
	oldEndpoint.Name = constant.KubernetesEndpointNoEdge
	oldEndpoint.ResourceVersion = ""
	if _, err := clientSet.CoreV1().Endpoints(
		constant.NamespaceDefault).Create(context.TODO(), oldEndpoint, metav1.CreateOptions{}); err != nil {
		return err
	}

	annotations := make(map[string]string)
	annotations[constant.EdgeLocalPort] = "51003"
	annotations[constant.EdgeLocalHost] = "127.0.0.1"
	endpoint.Annotations = annotations
	if _, err := clientSet.CoreV1().Endpoints(
		constant.NamespaceDefault).Update(context.TODO(), endpoint, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func PatchKubeProxy(clientSet kubernetes.Interface) error {
	patch := fmt.Sprintf(constant.KubeProxyPatchJson, constant.EdgeNodeLabelKey, constant.EdgeNodeLabelValueEnable)
	if _, err := clientSet.AppsV1().DaemonSets(constant.NamespcaeKubeSystem).Patch(
		context.TODO(), constant.ModeKubeProxy, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{}); err != nil {
		return fmt.Errorf("Patching daemonset: %s, error: %v\n", constant.ModeKubeProxy, err)
	}
	return nil
}

func RecoverKubeConfig(client *kubernetes.Clientset) error {
	if err := RecoverKubeProxyKubeconfig(client); err != nil {
		klog.Errorf("Delete serivce group, error: %s", err)
		return err
	}

	if err := RecoverKubernetesEndpoint(client); err != nil {
		klog.Errorf("Delete serivce group, error: %s", err)
		return err
	}

	klog.Infof("Recover Kubernetes cluster config support marginal autonomy success")

	return nil
}

func RecoverKubeProxyKubeconfig(kubeClient kubernetes.Interface) error {
	kubeProxyCM, err := kubeClient.CoreV1().ConfigMaps(
		constant.NamespcaeKubeSystem).Get(context.TODO(), constant.CMKubeProxyNoEdge, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// recover backup ConfigMap
	if err := kubeClient.CoreV1().ConfigMaps(
		constant.NamespcaeKubeSystem).Delete(context.TODO(), constant.CMKubeProxy, metav1.DeleteOptions{}); err != nil {
		return err
	}
	oldKubeProxyCM := kubeProxyCM.DeepCopy()
	oldKubeProxyCM.Name = constant.CMKubeProxy
	oldKubeProxyCM.ResourceVersion = ""
	if _, err := kubeClient.CoreV1().ConfigMaps(
		constant.NamespcaeKubeSystem).Create(context.TODO(), oldKubeProxyCM, metav1.CreateOptions{}); err != nil {
		return err
	}

	if err := kubeClient.CoreV1().ConfigMaps(
		constant.NamespcaeKubeSystem).Delete(context.TODO(), constant.CMKubeProxyNoEdge, metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}

func RecoverKubernetesEndpoint(clientSet kubernetes.Interface) error {
	endpoint, err := clientSet.CoreV1().Endpoints(
		constant.NamespaceDefault).Get(context.TODO(), constant.KubernetesEndpointNoEdge, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// recover backup ConfigMap
	if err := clientSet.CoreV1().Endpoints(
		constant.NamespaceDefault).Delete(context.TODO(), constant.KubernetesEndpoint, metav1.DeleteOptions{}); err != nil {
		return err
	}
	oldEndpoint := endpoint.DeepCopy()
	oldEndpoint.Name = constant.KubernetesEndpoint
	oldEndpoint.ResourceVersion = ""
	if _, err := clientSet.CoreV1().Endpoints(
		constant.NamespaceDefault).Create(context.TODO(), oldEndpoint, metav1.CreateOptions{}); err != nil {
		return err
	}

	if err := clientSet.CoreV1().Endpoints(
		constant.NamespaceDefault).Delete(context.TODO(), constant.KubernetesEndpointNoEdge, metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}

func RecoverClusterInfoKubeconfig(kubeClient kubernetes.Interface, certSANs []string) error {
	if len(certSANs) <= 0 {
		return nil
	}
	clusterInfoCM, err := kubeClient.CoreV1().ConfigMaps(
		metav1.NamespacePublic).Get(context.TODO(), constant.ConfigMapClusterInfoNoEdge, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// recover backup ConfigMap
	if err := kubeClient.CoreV1().ConfigMaps(
		constant.NamespcaeKubeSystem).Delete(context.TODO(), bootstrapapi.ConfigMapClusterInfo, metav1.DeleteOptions{}); err != nil {
		return err
	}
	oldClusterInfoCM := clusterInfoCM.DeepCopy()
	oldClusterInfoCM.Name = bootstrapapi.ConfigMapClusterInfo
	oldClusterInfoCM.ResourceVersion = ""
	if _, err := kubeClient.CoreV1().ConfigMaps(
		constant.NamespcaeKubeSystem).Create(context.TODO(), oldClusterInfoCM, metav1.CreateOptions{}); err != nil {
		return err
	}
	return nil
}
