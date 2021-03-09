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

package init_cmd

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/superedge/superedge/pkg/edgeadm/cmd"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/edgecluster"
	"io"
	"io/ioutil"
	"k8s.io/klog/v2"
	"time"
)

var (
	WorkerPath string = "/tmp"
)

type ClusterProgress struct {
	Status     string   `json:"status"`
	Data       string   `json:"data"`
	URL        string   `json:"url,omitempty"`
	Username   string   `json:"username,omitempty"`
	Password   []byte   `json:"password,omitempty"`
	CACert     []byte   `json:"caCert,omitempty"`
	Hosts      []string `json:"hosts,omitempty"`
	Servers    []string `json:"servers,omitempty"`
	Kubeconfig []byte   `json:"kubeconfig,omitempty"`
}

type Handler struct {
	Name string
	Func func() error
}

type initOptions struct {
	EdgeInitConfig edgeInitConfig
	KubeadmConfig  kubeadmConfig
}
type edgeInitConfig struct {
	WorkerPath     string `yaml:"workerPath"`
	InstallPkgPath string `yaml:"InstallPkgPath"`
}

type kubeadmConfig struct {
	KubeadmConfPath string `yaml:"kubeadmConfPath"`
}

type initData struct {
	InitOptions initOptions             `json:"initOptions"`
	Cluster     edgecluster.EdgeCluster `json:"cluster"`
	steps       []Handler
	Step        int `json:"step"`
	Progress    ClusterProgress
	//strategy        *clusterstrategy.Strategy
	//clusterProvider clusterprovider.Provider
	//isFromRestore   bool
	//docker *docker.Docker
	//globalClient kubernetes.Interface
	//servers      []string
	//namespace    string
	//Para    *types.CreateClusterPara `json:"para"`
}

func newInit() initData {
	return initData{}
}

func NewInitCMD(out io.Writer, edgeConfig *cmd.EdgeadmConfig) *cobra.Command {
	action := newInit()
	initOptions := &action.InitOptions
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create first master node of edge kubernetes cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if err := action.complete(edgeConfig); err != nil {
				util.OutPutMessage(err.Error())
				return
			}

			if err := action.validate(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}

			if err := action.runInit(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}
		},
	}

	AddEdgeConfigFlags(cmd.Flags(), &initOptions.EdgeInitConfig)
	AddKubeadmConfigFlags(cmd.Flags(), &initOptions.KubeadmConfig)
	// parse-config

	return cmd
}

func AddEdgeConfigFlags(flagSet *pflag.FlagSet, cfg *edgeInitConfig) {
	flagSet.StringVar(
		&cfg.InstallPkgPath, constant.InstallPkgPath, "./edge-v0.3.0-kube-v1.18.2-install-pkg.tar.gz",
		"Install static package path of edge kubernetes cluster.",
	)
}

func AddKubeadmConfigFlags(flagSet *pflag.FlagSet, cfg *kubeadmConfig) {
	flagSet.StringVar(
		&cfg.KubeadmConfPath, constant.KubeadmConfig, "/root/.edgeadm/kubeadm.config",
		"Install static package path of edge kubernetes cluster.",
	)
}

func (e *initData) config() error {
	return nil
}

func (e *initData) complete(edgeConfig *cmd.EdgeadmConfig) error {
	e.InitOptions.EdgeInitConfig.WorkerPath = edgeConfig.WorkerPath
	return nil
}

func (e *initData) validate() error {
	return nil
}

func (e *initData) backup() error {
	klog.V(4).Infof("===>starting install backup()")
	data, _ := json.MarshalIndent(e, "", " ")
	return ioutil.WriteFile(e.InitOptions.EdgeInitConfig.WorkerPath+constant.EdgeClusterFile, data, 0777)
}

func (e *initData) runInit() error {
	start := time.Now()
	e.initSteps()
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

func (e *initData) initSteps() error {
	//e.steps = append(e.steps, []Handler{
	//	{
	//		Name: "Execute pre install hook",
	//		Func: e.preInstallHook,
	//	},
	//}...)

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

	// config && create kube-* yaml
	e.steps = append(e.steps, []Handler{
		{
			Name: "create kubeadm config",
			Func: e.tarInstallMovePackage,
		},
		{
			Name: "config etcd",
			Func: e.tarInstallMovePackage,
		},
		{
			Name: "config kube-api-service",
			Func: e.tarInstallMovePackage,
		},
		{
			Name: "config kube-controller-manager",
			Func: e.tarInstallMovePackage,
		},
		{
			Name: "config kube-scheduler",
			Func: e.tarInstallMovePackage,
		},
	}...)

	// kubeadm init
	e.steps = append(e.steps, []Handler{
		{
			Name: "kubeadm init",
			Func: e.tarInstallMovePackage,
		},
	}...)

	// check kubernetes cluster health
	e.steps = append(e.steps, []Handler{
		{
			Name: "check kubernetes cluster",
			Func: e.tarInstallMovePackage,
		},
	}...)

	// install cloud edge-apps
	e.steps = append(e.steps, []Handler{
		{
			Name: "deploy tunnel-coredns",
			Func: e.tarInstallMovePackage,
		},
		{
			Name: "deploy tunnel-cloud",
			Func: e.tarInstallMovePackage,
		},
		{
			Name: "deploy application-grid controller",
			Func: e.tarInstallMovePackage,
		},
		{
			Name: "deploy application-grid wrapper",
			Func: e.tarInstallMovePackage,
		},
		{
			Name: "deploy edge-health admission",
			Func: e.tarInstallMovePackage,
		},
	}...)

	// install edge edge-apps
	e.steps = append(e.steps, []Handler{
		{
			Name: "daemonset flannel",
			Func: e.tarInstallMovePackage,
		},
		{
			Name: "daemonset tunnel edge",
			Func: e.tarInstallMovePackage,
		},
		{
			Name: "daemonset coredns",
			Func: e.tarInstallMovePackage,
		},
		{
			Name: "daemonset kube-proxy",
			Func: e.tarInstallMovePackage,
		},
		{
			Name: "daemonset edge-health",
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
