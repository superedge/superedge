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

package join

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/superedge/superedge/pkg/edgeadm/cmd"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/edgecluster"
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
	"time"
)

type joinNodeData struct {
	JoinOptions joinOptions             `json:"joinOptions"`
	Cluster     edgecluster.EdgeCluster `json:"cluster"`
	steps       []Handler
	Step        int `json:"step"`
	Progress    ClusterProgress
}

func newJoinNode() joinNodeData {
	return joinNodeData{}
}

func NewJoinNodeCMD(edgeConfig *cmd.EdgeadmConfig) *cobra.Command {
	action := newJoinNode()
	joinOptions := &action.JoinOptions
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Join a edge node into cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if err := action.complete(edgeConfig); err != nil {
				util.OutPutMessage(err.Error())
				return
			}

			if err := action.validate(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}

			if err := action.runJoin(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}
		},
	}

	AddEdgeConfigFlags(cmd.Flags(), &joinOptions.EdgeJoinConfig)
	AddKubeadmConfigFlags(cmd.Flags(), &joinOptions.KubeadmConfig)
	return cmd
}

func (e *joinNodeData) preInstallHook() error {
	klog.Infof("=========, preInstallHook")
	return e.execHook(constant.PreInstallHook)
}

func (e *joinNodeData) execHook(filename string) error {
	klog.V(5).Info("Execute hook script %s", filename)
	f, err := os.OpenFile(constant.EdgeClusterLogFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0744)
	if err != nil {
		klog.Error(err.Error())
	}
	defer f.Close()

	cmd := exec.Command(filename)
	cmd.Stdout = f
	cmd.Stderr = f
	err = cmd.Run()
	if err != nil {
		return err
	}
	return err
}

func (e *joinNodeData) tarInstallMovePackage() error {
	workerPath := e.JoinOptions.EdgeJoinConfig.WorkerPath
	tarInstallCmd := fmt.Sprintf("tar -xzvf %s -C %s",
		e.JoinOptions.EdgeJoinConfig.InstallPkgPath, workerPath+constant.EdgeamdDir)
	if err := util.RunLinuxCommand(tarInstallCmd); err != nil {
		return err
	}
	return nil
}

func (e *joinNodeData) kubeadmJoinNode() error {
	cmds := []string{
		//kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash %s --experimental-control-plane --certificate-key %s
		fmt.Sprintf("kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash %s --experimental-control-plane --certificate-key %s", e.JoinOptions.MasterIp, e.JoinOptions.JoinToken, e.JoinOptions.TokenCaCertHash, e.JoinOptions.CertificateKey),
	}
	for _, cmd := range cmds {
		if err := util.RunLinuxCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

func (e *joinNodeData) config() error {
	return nil
}

func (e *joinNodeData) complete(edgeConfig *cmd.EdgeadmConfig) error {
	e.JoinOptions.EdgeJoinConfig.WorkerPath = edgeConfig.WorkerPath
	return nil
}

func (e *joinNodeData) validate() error {
	return nil
}

func (e *joinNodeData) backup() error {
	klog.V(4).Infof("===>starting install backup()")
	data, _ := json.MarshalIndent(e, "", " ")
	return ioutil.WriteFile(e.JoinOptions.EdgeJoinConfig.WorkerPath+constant.EdgeClusterFile, data, 0777)
}

func (e *joinNodeData) runJoin() error {
	start := time.Now()
	e.joinSteps()
	defer e.backup()

	if e.Step == 0 {
		klog.V(4).Infof("===>starting install task")
		e.Progress.Status = constant.StatusDoing
	}

	for e.Step < len(e.steps) {
		klog.V(4).Infof("%d.%s doing", e.Step, e.steps[e.Step].Name)

		start := time.Now()
		err := e.steps[e.Step].Func()
		if err != nil {
			e.Progress.Status = constant.StatusFailed
			klog.V(4).Infof("%d.%s [Failed] [%fs] error %s", e.Step, e.steps[e.Step].Name, time.Since(start).Seconds(), err)
			return nil
		}
		klog.V(4).Infof("%d.%s [Success] [%fs]", e.Step, e.steps[e.Step].Name, time.Since(start).Seconds())

		e.Step++
		e.backup()
	}

	klog.V(1).Info("===>install task [Sucesss] [%fs]", time.Since(start).Seconds())
	return nil
}

func (e *joinNodeData) joinSteps() error {

	// tar -xzvf install-package
	e.steps = append(e.steps, []Handler{
		{
			Name: "tar -xzvf install.tar.gz",
			Func: e.tarInstallMovePackage,
		},
		{
			Name: "set /root/.bashrc",
			Func: e.tarInstallMovePackage,
		},
	}...)

	// init node
	e.steps = append(e.steps, []Handler{
		{
			Name: "check node",
			Func: e.tarInstallMovePackage,
		},
		{
			Name: "init node",
			Func: e.tarInstallMovePackage,
		},
	}...)

	// install container runtime
	e.steps = append(e.steps, []Handler{
		{
			Name: "install docker",
			Func: e.tarInstallMovePackage,
		},
	}...)

	// install lite-apiserver
	e.steps = append(e.steps, []Handler{
		{
			Name: "install docker",
			Func: e.tarInstallMovePackage,
		},
	}...)

	// kubeadm join
	e.steps = append(e.steps, []Handler{
		{
			Name: "kubeadm join",
			Func: e.kubeadmJoinNode,
		},
	}...)

	// check edge cluster health
	e.steps = append(e.steps, []Handler{
		{
			Name: "check edge cluster",
			Func: e.tarInstallMovePackage,
		},
	}...)

	return nil
}
