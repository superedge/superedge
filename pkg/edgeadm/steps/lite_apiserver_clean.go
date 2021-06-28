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
	"os"
	"path/filepath"

	"k8s.io/klog/v2"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/initsystem"

	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
)

func NewCleanupLiteApiServerPhase() workflow.Phase {
	return workflow.Phase{
		Name:         "lite-apiserver clean up",
		Short:        "Clean up lite-apiserver on edge node",
		Run:          runCleanupLiteAPIServer,
		InheritFlags: []string{},
	}
}

// runCleanupLiteAPIServer executes cleanup lite-apiserver logic.
func runCleanupLiteAPIServer(c workflow.RunData) error {
	// Try to stop the lite-apiserver service
	klog.V(1).Infof("[reset] Getting init system")
	initSystem, err := initsystem.GetInitSystem()
	if err != nil {
		klog.Warningln("[reset] The lite-apiserver service could not be stopped by edgeadm. Unable to detect a supported init system!")
		klog.Warningln("[reset] Please ensure lite-apiserver is stopped manually")
	} else {
		klog.V(1).Infof("[reset] Stopping the lite-apiserver service")
		if err := initSystem.ServiceStop("lite-apiserver"); err != nil {
			klog.Warningf("[reset] The lite-apiserver service could not be stopped by edgeadm: [%v]\n", err)
			klog.Warningln("[reset] Please ensure lite-apiserver is stopped manually")
		}
	}
	resetHostsFile()
	resetConfigDir(constant.KubeEdgePath, constant.LiteAPIServerCACertPath)
	return nil
}

// resetConfigDir is used to cleanup the files edgeadm writes in /etc/Kubernetes/.
func resetConfigDir(configPathDir, pkiPathDir string) {
	filesToClean := []string{
		filepath.Join(configPathDir, constant.LiteAPIServerKey),
		filepath.Join(configPathDir, constant.LiteAPIServerCrt),
		filepath.Join(configPathDir, constant.LiteAPIServerTLSPath),
		pkiPathDir,
	}
	klog.V(1).Infof("[reset] Deleting files: %v\n", filesToClean)
	for _, path := range filesToClean {
		if err := os.RemoveAll(path); err != nil {
			klog.Warningf("[reset] Failed to remove file: %q [%v]\n", path, err)
		}
	}
}

func resetHostsFile() error {
	klog.V(1).Infof("[reset] Resetting file: %s\n", constant.HostsFilePath)
	if _, _, err := util.RunLinuxCommand(constant.ResetDNSCmd); err != nil {
		klog.Errorf("[reset] Failed to reset file: %s error: %v\n", constant.ResetDNSCmd, err)
		return err
	}
	return nil
}
