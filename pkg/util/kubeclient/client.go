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
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog"

	"superedge/pkg/util"
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
		return nil, fmt.Errorf("kubeconfig nil, Please appoint --kubeconfig, KUBECONFIG or ~/kube/config")
	}

	os.Setenv("KUBECONF", kubeconfigFile)
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
	kubeConf, err := base64.StdEncoding.DecodeString(os.Getenv("KUBECONF"))
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
