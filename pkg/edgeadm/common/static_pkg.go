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
	"fmt"
	"os"
	"path"
	"strings"

	"k8s.io/klog/v2"

	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
)

func UnzipPackage(srcPackage, dstPath string) error {
	if strings.Contains(srcPackage, "http") {
		klog.Info("Downloading install package...\n")
		downloadPackage := fmt.Sprintf("rm -rf %s && wget -k --progress=dot:giga %s -O %s", constant.TMPPackgePath, srcPackage, constant.TMPPackgePath)
		if _, _, err := util.RunLinuxCommand(downloadPackage); err != nil {
			return err
		}
		srcPackage = constant.TMPPackgePath
	}

	tarUnzipCmd := fmt.Sprintf("tar -xzvf %s -C %s", srcPackage, dstPath)
	if _, _, err := util.RunLinuxCommand(tarUnzipCmd); err != nil {
		return err
	}
	return nil
}

func SetPackagePath(workerPath string) error {
	moveBin := fmt.Sprintf("mv -f %s/* /usr/bin/", workerPath+constant.InstallBin)
	if _, _, err := util.RunLinuxCommand(moveBin); err != nil {
		return err
	}

	os.MkdirAll(path.Dir(constant.CNIDir), 0755)
	if err := UnzipPackage(workerPath+constant.CNIPluginsPKG, constant.CNIDir); err != nil {
		return err
	}
	klog.V(4).Infof("Install cni plugins success")

	os.MkdirAll(path.Dir(constant.KubeletServiceFile), 0755)
	if err := util.WriteWithBufio(constant.KubeletServiceFile, constant.KubeletService); err != nil {
		return err
	}

	os.MkdirAll(path.Dir(constant.KubeadmConfFile), 0755)
	if err := util.WriteWithBufio(constant.KubeadmConfFile, constant.KubeadmConfig); err != nil {
		return err
	}

	os.MkdirAll(path.Dir(constant.KubeletSysConf), 0755)
	if err := util.WriteWithBufio(constant.KubeletSysConf, constant.KubeletSys); err != nil {
		return err
	}

	os.MkdirAll(path.Dir(constant.KubectlBashCompletion), 0755)
	kubectlBash := fmt.Sprintf("kubectl completion bash > %s", constant.KubectlBashCompletion)
	if _, _, err := util.RunLinuxCommand(kubectlBash); err != nil {
		return err
	}
	return nil
}
