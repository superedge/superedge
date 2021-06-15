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

package steps

import (
	"fmt"

	"k8s.io/klog/v2"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"

	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
)

var (
	initNodeLongDesc = cmdutil.LongDesc(`Init node before install node or master of Kubernetes.`)
)

// Init node before install node or master of Kubernetes
func NewInitNodePhase() workflow.Phase {
	return workflow.Phase{
		Name:  "init-node",
		Short: "Init node before install node or master of Kubernetes",
		Long:  initNodeLongDesc,
		Run:   initNode,
	}
}

func initNode(c workflow.RunData) error {
	klog.V(4).Infof("Start init node")
	if _, _, err := util.RunLinuxCommand(EdgeadmConf.WorkerPath + constant.InitNodeShell); err != nil {
		klog.Errorf("Run init node shell: %s, error: %v",
			EdgeadmConf.WorkerPath+constant.InitNodeShell, err)
		return err
	}

	klog.V(4).Infof("Init node success")
	return nil
}

// set hostname about node or master of Kubernetes
func setHostname(c workflow.RunData) error {
	loadIP, err := util.GetLocalIP()
	if err != nil {
		return err
	}
	steHostname := fmt.Sprintf("hostnamectl set-hostname node-%s", loadIP)
	if _, _, err := util.RunLinuxCommand(steHostname); err != nil {
		return err
	}
	return err
}

// stop firewalld
func stopFirewall(c workflow.RunData) error {
	if _, _, err := util.RunLinuxCommand(constant.StopFireWall); err != nil {
		klog.Errorf("Run off stop firewall: %v", err)
	}
	return nil
}
