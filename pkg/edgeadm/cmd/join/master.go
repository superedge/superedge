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
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

type joinMasterData struct {
	JoinOptions joinOptions             `json:"joinOptions"`
	Cluster     edgecluster.EdgeCluster `json:"cluster"`
	steps       []Handler
	Step        int `json:"step"`
	Progress    ClusterProgress
}

func newJoinMaster() joinMasterData {
	return joinMasterData{}
}

func NewJoinMasterCMD(edgeConfig *cmd.EdgeadmConfig) *cobra.Command {
	action := newJoinMaster()
	joinOptions := &action.JoinOptions
	cmd := &cobra.Command{
		Use:   "master [api-server-endpoint]",
		Short: "Join a master node into cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if err := action.complete(edgeConfig, args, joinOptions); err != nil {
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
		// We accept the control-plane location as an optional positional argument
		Args: cobra.MaximumNArgs(1),
	}

	AddEdgeConfigFlags(cmd.Flags(), &joinOptions.EdgeJoinConfig)
	AddKubeadmConfigFlags(cmd.Flags(), &joinOptions.KubeadmConfig)
	addMasterFlags(cmd.Flags(), joinOptions)
	return cmd
}

func addMasterFlags(flagSet *pflag.FlagSet, option *joinOptions) {
	flagSet.StringVar(
		&option.JoinToken, constant.TokenStr, "",
		"The token to use for establishing bidirectional trust between nodes and control-plane nodes. The format is [a-z0-9]{6}\\\\.[a-z0-9]{16} - e.g. abcdef.0123456789abcdef",
	)
	flagSet.StringVar(
		&option.TokenCaCertHash, constant.TokenDiscoveryCAHash, "",
		"For token-based discovery, validate that the root CA public key matches this hash (format: \\\"<type>:<value>\\\").",
	)
	flagSet.StringVar(
		&option.CertificateKey, constant.CertificateKey, "",
		"Key used to encrypt the control-plane certificates in the kubeadm-certs Secret.",
	)
}

func (e *joinMasterData) preInstallHook() error {
	klog.Infof("=========, preInstallHook")
	return e.execHook(constant.PreInstallHook)
}

func (e *joinMasterData) execHook(filename string) error {
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

func (e *joinMasterData) tarInstallMovePackage() error {
	workerPath := e.JoinOptions.EdgeJoinConfig.WorkerPath
	tarInstallCmd := fmt.Sprintf("tar -xzvf %s -C %s",
		e.JoinOptions.EdgeJoinConfig.InstallPkgPath, workerPath+constant.EdgeamdDir)
	if err := util.RunLinuxCommand(tarInstallCmd); err != nil {
		return err
	}
	return nil
}

func (e *joinMasterData) kubeadmJoinMaster() error {
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

func (e *joinMasterData) config() error {
	return nil
}

func (e *joinMasterData) complete(edgeConfig *cmd.EdgeadmConfig, args []string, option *joinOptions) error {
	e.JoinOptions.EdgeJoinConfig.WorkerPath = edgeConfig.WorkerPath
	if len(args) == 1 {
		option.MasterIp = args[0]
	} else if len(args) > 1 {
		klog.Warningf("[WARNING] More than one API server endpoint supplied on command line %v. Using the first one.", args)
		option.MasterIp = args[0]
	} else {
		return errors.New("[Error] need an API server endpoint as control plane to join")
	}
	return nil
}

func (e *joinMasterData) validate() error {
	return nil
}

func (e *joinMasterData) backup() error {
	klog.V(4).Infof("===>starting install backup()")
	data, _ := json.MarshalIndent(e, "", " ")
	return ioutil.WriteFile(e.JoinOptions.EdgeJoinConfig.WorkerPath+constant.EdgeClusterFile, data, 0777)
}

func (e *joinMasterData) runJoin() error {
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

func (e *joinMasterData) joinSteps() error {

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

	// create ca
	e.steps = append(e.steps, []Handler{
		{
			Name: "create etcd ca",
			Func: e.tarInstallMovePackage,
		},
		{
			Name: "create kube-api-service ca",
			Func: e.tarInstallMovePackage,
		},
		{
			Name: "create kube-controller-manager ca",
			Func: e.tarInstallMovePackage,
		},
		{
			Name: "create kube-scheduler ca",
			Func: e.tarInstallMovePackage,
		},
	}...)

	// kubeadm join
	e.steps = append(e.steps, []Handler{
		{
			Name: "kubeadm join",
			Func: e.kubeadmJoinMaster,
		},
	}...)

	// check kubernetes cluster health
	e.steps = append(e.steps, []Handler{
		{
			Name: "check kubernetes cluster",
			Func: e.tarInstallMovePackage,
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
