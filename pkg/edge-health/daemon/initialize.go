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
	"os"
	"strings"

	"github.com/superedge/superedge/pkg/edge-health/common"
	"github.com/superedge/superedge/pkg/edge-health/data"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/klog/v2"
)

func initialize(masterUrl, kubeconfigPath, hostName string) {
	initClientSet(masterUrl, kubeconfigPath)
	initHostName(hostName)
	initLocalIp()
	initData()
	klog.InfoS("init common", "PodIP", common.PodIP, "PodName", common.PodName, "NodeIP", common.NodeIP, "NodeName", common.NodeName)
	if common.PodIP == "" || common.PodName == "" || common.NodeIP == "" || common.NodeName == "" {
		panic("need pod and node information through downward api env POD_IP,POD_NAME,NODE_IP,NODE_NAME")
	}
	if common.Namespace == "" {
		common.Namespace = common.DefaultNamespace
	}
}

func initClientSet(masterUrl, kubeconfigPath string) {
	var err error
	kubeconfig, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
	if err != nil {
		klog.Fatalf("Init: Error building kubeconfig: %s", err.Error())
	}
	common.MetadataClientSet = metadata.NewForConfigOrDie(kubeconfig)
	common.ClientSet = kubernetes.NewForConfigOrDie(kubeconfig)
}

func initHostName(hostName string) {
	if hostName == "" {
		common.NodeName = os.Getenv("NODE_NAME")
		common.NodeName = strings.Replace(common.NodeName, "\n", "", -1)
		common.NodeName = strings.Replace(common.NodeName, " ", "", -1)
		klog.V(2).Infof("Init: Host name is %s", common.NodeName)
	} else {
		common.NodeName = hostName
	}
	common.PodName = os.Getenv("POD_NAME")
	common.Namespace = os.Getenv("NAMESPACE")
}

func initLocalIp() {
	common.PodIP = os.Getenv("POD_IP")
	common.NodeIP = os.Getenv("NODE_IP")
}

func initData() {
	data.CheckInfoResult = data.NewCheckInfoData()
	data.Result = data.NewResultData()
	data.NodeList = data.NewNodeListData()
	data.ConfigMapList = data.NewConfigMapListData()
}
