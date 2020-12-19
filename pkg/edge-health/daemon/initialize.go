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

package daemon

import (
	"context"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"os"
	"strings"
	"superedge/pkg/edge-health/common"
	"superedge/pkg/edge-health/data"
)

func initialize(masterUrl, kubeconfigPath, hostName string) {
	initClientSet(masterUrl, kubeconfigPath)
	initHostName(hostName)
	initLocalIp()
	initData()
}

func initClientSet(masterUrl, kubeconfigPath string) {
	var err error
	kubeconfig, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
	if err != nil {
		klog.Fatalf("Init: Error building kubeconfig: %s", err.Error())
	}
	common.ClientSet, err = kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		klog.Fatalf("Init: Error building clientset: %s", err.Error())
	}
}

func initHostName(hostName string) {
	if hostName == "" {
		common.HostName = os.Getenv("NODE_NAME")
		common.HostName = strings.Replace(common.HostName, "\n", "", -1)
		common.HostName = strings.Replace(common.HostName, " ", "", -1)
		klog.V(2).Infof("Init: Host name is %s", common.HostName)
	} else {
		common.HostName = hostName
	}
}

func initLocalIp() {
	klog.V(2).Infof("common.hostname is %s", common.HostName)
	if host, err := common.ClientSet.CoreV1().Nodes().Get(context.TODO(), common.HostName, metav1.GetOptions{}); err != nil {
		klog.Fatalf("Init: Error getting hostname node: %s", err.Error())
	} else {
		for _, v := range host.Status.Addresses {
			if v.Type == v1.NodeInternalIP {
				common.LocalIp = v.Address
				klog.V(2).Infof("Init: host ip is %s", common.LocalIp)
			}
		}
	}
}

func initData() {
	data.CheckInfoResult = data.NewCheckInfoData()
	data.Result = data.NewResultData()
	data.NodeList = data.NewNodeListData()
	data.ConfigMapList = data.NewConfigMapListData()
}
