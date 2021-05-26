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
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"

	"github.com/superedge/superedge/pkg/edgeadm/common"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
)

var (
	dockerExample = cmdutil.Examples(`
		# Install docker container runtime.
		  kubeadm init phase container`)
)

//install container runtime (docker | containerd | CRI-O)
func NewContainerPhase() workflow.Phase {
	return workflow.Phase{
		Name:         "container",
		Short:        "Install docker container runtime",
		Long:         "Install docker container runtime",
		Example:      dockerExample,
		Run:          installContainer,
		InheritFlags: []string{},
	}
}

func installContainer(c workflow.RunData) error {
	err := installDocker()
	if err != nil {
		return err
	}
	klog.Info("Installed docker container runtime successfully")

	return nil
}

func installDocker() error {
	klog.V(4).Infof("Start install docker container runtime")
	//unzip Docker Package
	if err := common.UnzipPackage(EdgeadmConf.WorkerPath+constant.ZipContainerPath, EdgeadmConf.WorkerPath+constant.UnZipContainerDstPath); err != nil {
		klog.Errorf("Unzip Docker container runtime Package: %s, error: %v", EdgeadmConf.WorkerPath+constant.UnZipContainerDstPath, err)
		return err
	}

	if _, _, err := util.RunLinuxCommand(EdgeadmConf.WorkerPath + constant.DockerInstallShell); err != nil {
		klog.Errorf("Run Docker install shell: %s, error: %v",
			EdgeadmConf.WorkerPath+constant.UnZipContainerDstPath, err)
		return err
	}

	klog.V(4).Infof("Install docker container runtime success")
	return nil
}
