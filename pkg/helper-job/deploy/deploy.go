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
	"errors"
	"os"

	"k8s.io/klog/v2"

	edgeadmConstant "github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util/kubeclient"
)

func Run() error {
	kubeClient, err := kubeclient.GetClientSet("")
	if err != nil {
		klog.Errorf("Get kube client error: %v", err)
		return err
	}

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		return errors.New("Get ENV NODE_NAME nil\n")
	}

	nodeRole := os.Getenv("NODE_ROLE")
	if nodeRole == "" {
		return errors.New("Get ENV NODE_ROLE nil\n")
	}

	action := os.Getenv("ACTION")
	if nodeRole == "" {
		return errors.New("Get ENV ACTION nil\n")
	}

	klog.Infof("Node: %s Start Running %s", nodeName, action)

	switch nodeRole {
	case edgeadmConstant.NodeRoleMaster:
		switch action {
		case edgeadmConstant.ActionChange:
			if err := changeMasterJob(kubeClient, nodeName); err != nil {
				klog.Errorf("Change master: %s job running error: %v", nodeName, err)
				return err
			}

		case edgeadmConstant.ActionRevert:
			if err := revertMasterJob(kubeClient, nodeName); err != nil {
				klog.Errorf("Revert master: %s job running error: %v", nodeName, err)
				return err
			}
		}

	case edgeadmConstant.NodeRoleNode:
		switch action {
		case edgeadmConstant.ActionChange:
			if err := changeNodeJob(kubeClient, nodeName); err != nil {
				klog.Errorf("Change node: %s job running error: %v", nodeName, err)
				return err
			}

		case edgeadmConstant.ActionRevert:
			if err := revertNodeJob(kubeClient, nodeName); err != nil {
				klog.Errorf("Revert node: %s job running error: %v", nodeName, err)
				return err
			}
		}
	}

	return nil
}
