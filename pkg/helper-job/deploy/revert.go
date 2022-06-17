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

package deploy

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/superedge/superedge/pkg/helper-job/constant"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeclient"
)

func revertMasterJob(kubeClient *kubernetes.Clientset, nodeName string) error {
	klog.Infof("Maste %s Start Revert TkeEdge.", nodeName)
	if err := os.Rename(constant.KubeAPIServerBackUPYamlPath, constant.KubeAPIServerSrcYamlPath); err != nil {
		return err
	}

	if err := isRunningMaster(); err != nil {
		klog.Errorf("Check master is running error: %v", err)
		return err
	}

	nodeLabel := map[string]string{
		util.EdgeMasterLabelKey: util.EdgeMasterLabelValueEnable,
		util.EdgeChangeLabelKey: util.EdgeChangeLabelValueEnable,
	}
	if err := kubeclient.DeleteNodeLabel(kubeClient, nodeName, nodeLabel); err != nil {
		return err
	}
	return nil
}

func isRunningMaster() error {
	return wait.PollImmediate(time.Second, 3*time.Minute, func() (bool, error) {
		clientSet, err := kubeclient.GetClientSet("")
		if err != nil {
			return false, nil
		}

		healthStatus := 0
		clientSet.Discovery().RESTClient().Get().AbsPath("/healthz").Do(context.TODO()).StatusCode(&healthStatus)
		if healthStatus != http.StatusOK {
			return false, nil
		}
		return true, nil
	})
}

func revertNodeJob(kubeClient *kubernetes.Clientset, nodeName string) error {
	if err := util.WriteFile(constant.KubeletStartEnvFile, ""); err != nil {
		return err
	}

	if err := runLinuxCommand(constant.KUBELET_RESTART_CMD); err != nil {
		klog.Errorf("Running linux command: %s error: %v", constant.KUBELET_RESTART_CMD, err)
		return err
	}
	klog.Infof("Node: %s Restart kubelet config success.", nodeName)

	if err := runLinuxCommand(constant.KUBELET_STATUS_CMD); err != nil {
		klog.Errorf("Running linux command: %s error: %v", constant.KUBELET_RESTART_CMD, err)
		return err
	}
	klog.Infof("Node: %s Status kubelet config success.", nodeName)

	if err := checkKubeletHealthz(); err != nil {
		return fmt.Errorf("Node: %s iCheck kubelet status error: %v\n", nodeName, err)
	}

	if err := util.RemoveFile(constant.LiteAPIServerYamlPath); err != nil {
		return err
	}

	nodeLabel := map[string]string{
		util.EdgeNodeLabelKey:   util.EdgeNodeLabelValueEnable,
		util.EdgeChangeLabelKey: util.EdgeChangeLabelValueEnable,
	}
	if err := kubeclient.DeleteNodeLabel(kubeClient, nodeName, nodeLabel); err != nil {
		return err
	}

	return nil
}
