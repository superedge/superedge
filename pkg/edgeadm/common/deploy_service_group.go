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
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"path/filepath"
	"strings"
)

func DeployServiceGroup(clientSet kubernetes.Interface, manifestsDir string) error {
	advertiseAddress, err := GetKubeAPIServerAddr(clientSet)
	if err != nil {
		klog.Errorf("Get Kube-api-server add and port, error: %v", err)
		return err
	}

	option := map[string]interface{}{
		"AdvertiseAddress": advertiseAddress,
	}
	userGridWrapper := filepath.Join(manifestsDir, manifests.APP_APPLICATION_GRID_WRAPPER)
	gridWrapper := ReadYaml(userGridWrapper, manifests.ApplicationGridWrapperYaml)
	if err := kubeclient.CreateResourceWithFile(clientSet, gridWrapper, option); err != nil {
		return err
	}
	klog.V(4).Infof("Deploy %s success!", manifests.APP_APPLICATION_GRID_WRAPPER)

	userGridController := filepath.Join(manifestsDir, manifests.APP_APPLICATION_GRID_CONTROLLER)
	gridController := ReadYaml(userGridController, manifests.ApplicationGridControllerYaml)
	if err := CreateByYamlFile(clientSet, gridController); err != nil {
		klog.Errorf("Deploy %s error: %s", manifests.APP_APPLICATION_GRID_CONTROLLER, err)
		return err
	}

	klog.V(4).Infof("Create %s success!", manifests.APP_APPLICATION_GRID_CONTROLLER)

	return nil
}

func GetKubeAPIServerAddr(clientSet kubernetes.Interface) (string, error) {
	kubeClient := clientSet
	kubeProxyCM, err := kubeClient.CoreV1().ConfigMaps(
		constant.NamespcaeKubeSystem).Get(context.TODO(), "kube-proxy", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	proxyConfig, ok := kubeProxyCM.Data[constant.CMKubeConfig]
	if !ok {
		return "", errors.New("Get kube-proxy kubeconfig.conf nil\n")
	}
	klog.V(4).Infof("Get proxy-config: %s", util.ToJson(proxyConfig))

	config, err := clientcmd.Load([]byte(proxyConfig))
	if err != nil {
		return "", err
	}
	klog.V(4).Infof("Get proxy config: %s", util.ToJson(config))

	for key := range config.Clusters {
		KubeAPIServerAdrr := config.Clusters[key].Server
		if strings.Contains(KubeAPIServerAdrr, "http") {
			return KubeAPIServerAdrr, nil
		}
	}
	return "", errors.New("Get kube-api server addr nil\n")
}
