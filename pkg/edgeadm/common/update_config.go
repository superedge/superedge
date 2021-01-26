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
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"strconv"
	"time"
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
		constant.NamespcaeKubeSystem).Get(context.TODO(), "kube-proxy", metav1.GetOptions{})
	if err != nil {
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

func UpdateKubernetesEndpoint(clientSet kubernetes.Interface) error {
	endpoint, err := clientSet.CoreV1().Endpoints(
		constant.NamespaceDefault).Get(context.TODO(), "kubernetes", metav1.GetOptions{})
	if err != nil {
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
