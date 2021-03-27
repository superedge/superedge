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
	addNodeFlags(cmd.Flags(), joinOptions)
	return cmd
}

func addNodeFlags(flagSet *pflag.FlagSet, option *joinOptions) {
	flagSet.StringVar(
		&option.JoinToken, constant.TokenStr, "",
		"The token to use for establishing bidirectional trust between nodes and control-plane nodes. The format is [a-z0-9]{6}\\\\.[a-z0-9]{16} - e.g. abcdef.0123456789abcdef",
	)
	flagSet.StringVar(
		&option.TokenCaCertHash, constant.TokenDiscoveryCAHash, "",
		"For token-based discovery, validate that the root CA public key matches this hash (format: \\\"<type>:<value>\\\").",
	)
	flagSet.StringVar(
		&option.KubernetesServiceClusterIP, constant.KubernetesServiceClusterIP, "",
		"Cluster IP of kubernetes service in default namespace, using command 'kubectl get service kubernetes' to get cluster IP",
	)
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
	if _, _, err := util.RunLinuxCommand(tarInstallCmd); err != nil {
		return err
	}
	return nil
}

func (e *joinNodeData) generateLiteApiserverKey() error {
	cmds := []string{
		fmt.Sprintf("mkdir -p /etc/kubernetes/edge/ && openssl genrsa -out /etc/kubernetes/edge/lite-apiserver.key 2048"),
		fmt.Sprintf("cp /etc/kubernetes/edge/lite-apiserver.key /etc/kubernetes/pki/lite-apiserver.key"),
	}
	for _, cmd := range cmds {
		if _, _, err := util.RunLinuxCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

func (e *joinNodeData) generateLiteApiserverCsr() error {
	cmds := []string{
		fmt.Sprintf("mkdir -p /etc/kubernetes/edge/ && cat << EOF >/etc/kubernetes/edge/lite-apiserver.conf\n[req]\ndistinguished_name = req_distinguished_name\nreq_extensions = v3_req\n[req_distinguished_name]\nCN = lite-apiserver\n[v3_req]\nbasicConstraints = CA:FALSE\nkeyUsage = nonRepudiation, digitalSignature, keyEncipherment\nsubjectAltName = @alt_names\n[alt_names]\nDNS.1 = localhost\nIP.1 = 127.0.0.1\nIP.1 = %s\nEOF", e.JoinOptions.KubernetesServiceClusterIP),
		fmt.Sprintf("cd /etc/kubernetes/edge/ && openssl req -new -key lite-apiserver.key -subj \"/CN=lite-apiserver\" -config lite-apiserver.conf -out lite-apiserver.csr"),
		fmt.Sprintf("cd /etc/kubernetes/edge/ && openssl x509 -req -in lite-apiserver.csr -CA /etc/kubernetes/pki/ca.crt -CAkey /etc/kubernetes/pki/ca.key -CAcreateserial -days 5000 -extensions v3_req -extfile lite-apiserver.conf -out lite-apiserver.crt"),
		fmt.Sprintf("cp /etc/kubernetes/edge/lite-apiserver.crt /etc/kubernetes/pki/lite-apiserver.crt"),
	}
	for _, cmd := range cmds {
		if _, _, err := util.RunLinuxCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

func (e *joinNodeData) generateLiteApiserverTlsJson() error {
	cmds := []string{
		fmt.Sprintf("mkdir -p /etc/kubernetes/edge/ && cat << EOF >/etc/kubernetes/edge/tls.json\n[\n    {\n        \"key\":\"/var/lib/kubelet/pki/kubelet-client-current.pem\",\n        \"cert\":\"/var/lib/kubelet/pki/kubelet-client-current.pem\"\n    }\n]\nEOF"),
		fmt.Sprintf("cd /etc/kubernetes/edge/ && openssl req -new -key lite-apiserver.key -subj \"/CN=lite-apiserver\" -config lite-apiserver.conf -out lite-apiserver.csr"),
		fmt.Sprintf("cd /etc/kubernetes/edge/ && openssl x509 -req -in lite-apiserver.csr -CA /etc/kubernetes/pki/ca.crt -CAkey /etc/kubernetes/pki/ca.key -CAcreateserial -days 5000 -extensions v3_req -extfile lite-apiserver.conf -out lite-apiserver.crt"),
	}
	for _, cmd := range cmds {
		if _, _, err := util.RunLinuxCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

func (e *joinNodeData) startLiteApiserver() error {
	workerPath := e.JoinOptions.EdgeJoinConfig.WorkerPath
	cmd := fmt.Sprintf("%s --ca-file=/etc/kubernetes/pki/ca.crt "+
		"--tls-cert-file=/etc/kubernetes/edge/lite-apiserver.crt "+
		"--tls-private-key-file=/etc/kubernetes/edge/lite-apiserver.key "+
		"--kube-apiserver-url=%s "+
		"--kube-apiserver-port=6443 "+
		"--port=51003 --tls-config-file=/etc/kubernetes/edge/tls.json "+
		"--v=4 "+
		"--file-cache-path=/data/lite-apiserver/cache "+
		"--sync-duration=120 "+
		"--timeout=3",
		workerPath+constant.InstallBin+"lite-apiserver", e.JoinOptions.MasterIp)
	if _, _, err := util.RunLinuxCommand(cmd); err != nil {
		return err
	}
	return nil
}

func (e *joinNodeData) kubeadmJoinNode() error {
	cmds := []string{
		//kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash %s --experimental-control-plane --certificate-key %s
		fmt.Sprintf("kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash %s", e.JoinOptions.MasterIp, e.JoinOptions.JoinToken, e.JoinOptions.TokenCaCertHash),
	}
	for _, cmd := range cmds {
		if _, _, err := util.RunLinuxCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

func (e *joinNodeData) checkEdgeCluster() error {
	return nil
}

func (e *joinNodeData) config() error {
	return nil
}

func (e *joinNodeData) complete(edgeConfig *cmd.EdgeadmConfig, args []string, option *joinOptions) error {
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
			Name: "generate lite-apiserver.key",
			Func: e.generateLiteApiserverKey,
		},
		{
			Name: "generate lite-apiserver.csr",
			Func: e.generateLiteApiserverCsr,
		},
		{
			Name: "generate tls.json",
			Func: e.generateLiteApiserverTlsJson,
		},
		{
			Name: "start lite-apiserver",
			Func: e.startLiteApiserver,
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
			Func: e.checkEdgeCluster,
		},
	}...)

	return nil
}
