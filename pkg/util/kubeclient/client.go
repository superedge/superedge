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

package kubeclient

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"

	"github.com/superedge/superedge/pkg/util"
)

func GetClientSet(kubeconfigFile string) (*kubernetes.Clientset, error) {
	if !util.IsFileExist(kubeconfigFile) {
		kubeconfigFile = ""
	}

	if kubeconfigFile == "" {
		kubeconfigFile = os.Getenv("KUBECONFIG")
	}

	if kubeconfigFile == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfigFile = filepath.Join(home, ".kube", "config")
		}
	}

	if !util.IsFileExist(kubeconfigFile) {
		kubeconfigFile = ""
	}

	if kubeconfigFile == "" {
		kubeconfigFile = CustomConfig()
	}

	if kubeconfigFile == "" {
		return nil, fmt.Errorf("kubeconfig nil, Please appoint --kubeconfig, KUBECONFIG or ~/.kube/config")
	}

	os.Setenv("KUBECONF", kubeconfigFile)
	os.Setenv("KUBECONFIG", kubeconfigFile)
	restCfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigFile)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		klog.Errorf("Get kube client error: %v", err)
		return nil, err
	}

	return kubeClient, nil
}

func CustomConfig() string {
	kubeConf, err := base64.StdEncoding.DecodeString(os.Getenv("KUBECONFIG"))
	if err != nil {
		klog.Errorf("Get KUBECONF error: %v", err)
		return ""
	}

	klog.V(4).Infof("Get KUBECONF: \n %s", string(kubeConf))
	if string(kubeConf) == "" {
		return ""
	}

	if err := util.WriteWithBufio("/tmp/kubeconf", string(kubeConf)); err != nil {
		klog.Errorf("Write KUBECONF error: %v", err)
		return ""
	}

	return "/tmp/kubeconf"
}

func IsOverK8sVersion(baseK8sVersion, k8sVersion string) (bool, error) {
	drtK8sVerion, err := k8sVerisonInt(k8sVersion)
	if err != nil {
		return false, err
	}
	srcK8sVerion, err := k8sVerisonInt(baseK8sVersion)
	if err != nil {
		return false, err
	}
	return srcK8sVerion >= drtK8sVerion, nil
}

func k8sVerisonInt(version string) (int, error) {
	if strings.Contains(version, "-") {
		v := strings.Split(version, "-")[0]
		version = v
	}
	version = strings.Replace(version, "v", "", -1)
	versionSlice := strings.Split(version, ".")

	versionStr := ""
	for index, value := range versionSlice {
		if 0 == len(value) {
			versionStr += "00"
		}
		if 1 == len(value) {
			versionStr += "0" + value
		}
		if 2 == len(value) {
			versionStr += value
		}
		if index == 2 {
			break
		}
	}

	return strconv.Atoi(versionStr)
}

func AddNodeLabel(kubeClient kubernetes.Interface, nodeName string, labels map[string]string) error {
	return wait.PollImmediate(time.Second, 3*time.Minute, func() (bool, error) {
		node, err := kubeClient.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Get node: %s error: %v", nodeName, err)
			return false, nil
		}

		for key, _ := range labels {
			node.ObjectMeta.Labels[key] = labels[key]
		}
		klog.V(4).Infof("Add edge node label: %s", util.ToJson(node.ObjectMeta.Labels))
		if _, err := kubeClient.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{}); err != nil {
			klog.Errorf("Update node: %s labels error: %v", nodeName, err)
			return false, nil
		}
		return true, nil
	})
}

func DeleteNodeLabel(kubeClient *kubernetes.Clientset, nodeName string, labels map[string]string) error {
	return wait.PollImmediate(time.Second, 3*time.Minute, func() (bool, error) {
		node, err := kubeClient.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Get node: %s error: %v", nodeName, err)
			return false, nil
		}

		for key, srcValue := range labels {
			if dstVal, ok := node.Labels[key]; ok && dstVal == srcValue {
				delete(node.Labels, key)
			}
		}

		if _, err := kubeClient.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{}); err != nil {
			klog.Errorf("Update node: %s labels error: %v", nodeName, err)
			return false, err
		}
		return true, nil
	})
}

func CheckNodeLabel(kubeClient *kubernetes.Clientset, nodeName string, labels map[string]string) (bool, error) {
	node, err := kubeClient.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Get node: %s infos error: %v", nodeName, err)
		return false, err
	}
	for key, srcValue := range labels {
		if dstValue, ok := node.Labels[key]; !ok || dstValue != srcValue {
			return false, nil
		}
	}
	return true, nil
}

func GetClusterInfo(kubeconfigFile string) (*api.Cluster, error) {
	if !util.IsFileExist(kubeconfigFile) {
		kubeconfigFile = ""
	}

	if kubeconfigFile == "" {
		kubeconfigFile = os.Getenv("KUBECONFIG")
	}

	if kubeconfigFile == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfigFile = filepath.Join(home, ".kube", "config")
		}
	}

	if !util.IsFileExist(kubeconfigFile) {
		kubeconfigFile = ""
	}

	if kubeconfigFile == "" {
		kubeconfigFile = CustomConfig()
	}

	if kubeconfigFile == "" {
		return nil, fmt.Errorf("kubeconfig nil, Please appoint --kubeconfig, KUBECONFIG or ~/.kube/config")
	}

	os.Setenv("KUBECONF", kubeconfigFile)
	os.Setenv("KUBECONFIG", kubeconfigFile)

	// load the kubeconfig file to get the CA certificate and endpoint
	config, err := clientcmd.LoadFromFile(kubeconfigFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load kubeconfig")
	}

	// load the default cluster config
	clusterConfig := kubeconfigutil.GetClusterFromKubeConfig(config)
	if clusterConfig == nil {
		return nil, errors.New("failed to get default cluster config")
	}
	return clusterConfig, nil
}
