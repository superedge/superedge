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

type Host struct {
	IP     string
	Domain string
}

type initOptions struct {
	// base config
	WorkerPath      string `yaml:"workerPath"`
	InstallPkgPath  string `yaml:"InstallPkgPath"`
	KubeadmConfPath string `yaml:"kubeadmConfPath"` //todo: if need ?

	// kube-api config
	VIP         string   `yaml:"vip"` //todo: default value
	PodCIDR     string   `yaml:"podCidr"`
	ServiceCIDR string   `yaml:"serviceCIDR"`
	Registry    string   `yaml:"registry"` //container registry to pull control plane images
	CertSANS    []string `yaml:"certSans"`
	MasterIP    string   `yaml:"masterIP"`
	ApiServer   string   `yaml:"apiServer"` //apiserver domain name
	K8sVersion  string   `yaml:"k8sVersion"`

	// other
	Hosts []Host `yaml:"k8sVersion"`
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

	AddEdgeConfigFlags(cmd.Flags(), initOptions)
	AddKubeConfigFlags(cmd.Flags(), initOptions)
	//AddKubeadmConfigFlags(cmd.Flags(), &initOptions.KubeadmConfig)
	// parse-config

	return cmd
}

func AddEdgeConfigFlags(flagSet *pflag.FlagSet, cfg *initOptions) {
	flagSet.StringVar(
		&cfg.InstallPkgPath, constant.InstallPkgPath, "./edge-v0.3.0-kube-v1.18.2-install-pkg.tar.gz",
		"Install static package path of edge kubernetes cluster.",
	)
}

func AddKubeConfigFlags(flagSet *pflag.FlagSet, cfg *initOptions) {
	flagSet.StringVar(
		&cfg.VIP, constant.VIP, "10.10.0.2", "Virtual ip.",
	)
	flagSet.StringVar(
		&cfg.PodCIDR, constant.PodCIDR, "192.168.0.0/18",
		"Specify range of IP addresses for the pod network. If set, the control plane will automatically allocate CIDRs for every node.",
	)
	flagSet.StringVar(
		&cfg.ServiceCIDR, constant.ServiceCIDR, "10.96.0.0/12",
		"Use alternative range of IP address for service VIPs.",
	)

	flagSet.StringVar(
		&cfg.Registry, constant.Registry, "ccr.ccs.tencentyun.com/eck-private", //todo: ccr.ccs.tencentyun.com/eck-private only using test
		"Choose a container registry to pull control plane images from.",
	)
	flagSet.StringSliceVar(
		&cfg.CertSANS, constant.CertSANS, []string{"apiserver.cluster.local,apiserver.edge.com"},
		"Optional extra Subject Alternative Names (SANs) to use for the API Server serving certificate. Can be both IP addresses and DNS names.",
	)
	flagSet.StringVar(
		&cfg.MasterIP, constant.MasterIP, "", "First Master node IP address.",
	)
	flagSet.StringVar(
		&cfg.ApiServer, constant.APIServer, "apiserver.cluster.local",
		"The IP address the API Server will advertise it's listening on. If not set the default network interface will be used.",
	)
	flagSet.StringVar(
		&cfg.K8sVersion, constant.K8sVersion, "v1.18.2", "Choose a specific Kubernetes version for the control plane.",
	)
}

//func AddKubeadmConfigFlags(flagSet *pflag.FlagSet, cfg *kubeadmConfig) {
//	flagSet.StringVar(
//		&cfg.KubeadmConfPath, constant.KubeadmConfig, "/root/.edgeadm/kubeadm.config",
//		"Install static package path of edge kubernetes cluster.",
//	)
//}
func (e *initData) config() error {
	return nil
}

func (e *initData) complete(edgeConfig *cmd.EdgeadmConfig) error {
	e.InitOptions.WorkerPath = edgeConfig.WorkerPath
	localIP, err := util.GetLocalIP() //todo: private default Interface to choose loadIP
	if err != nil {
		return err
	}
	e.InitOptions.MasterIP = localIP

	publicIP, err := util.GetHostPublicIP()
	if err != nil {
		return err
	}
	e.InitOptions.CertSANS = append(e.InitOptions.CertSANS,
		"127.0.0.1", localIP, publicIP, e.InitOptions.VIP)

	// default init
	localHosts := []Host{
		{
			IP:     "127.0.0.1",
			Domain: constant.EdgeClusterKubeAPI,
		},
		{
			IP:     localIP,
			Domain: constant.EdgeClusterKubeAPI,
		},
		{
			IP:     publicIP,
			Domain: constant.EdgeClusterKubeAPI,
		},
	}
	e.InitOptions.Hosts = localHosts

	return nil
}

func (e *initData) validate() error {
	return nil
}

func (e *initData) backup() error {
	klog.V(4).Infof("Install backup()")
	data, _ := json.MarshalIndent(e, "", " ")
	return ioutil.WriteFile(e.InitOptions.WorkerPath+constant.EdgeClusterFile, data, 0777)
}

func (e *initData) runInit() error {
	start := time.Now()
	e.initSteps()
	defer e.backup()

	if e.Step == 0 {
		klog.V(4).Infof("starting install task")
		e.Progress.Status = constant.StatusDoing
	}

	for e.Step < len(e.steps) {
		klog.V(4).Infof("===> %d.%s doing", e.Step, e.steps[e.Step].Name)

		start := time.Now()
		err := e.steps[e.Step].Func()
		if err != nil {
			e.Progress.Status = constant.StatusFailed
			klog.V(4).Infof("%d.%s [Failed] [%fs] error %s", e.Step, e.steps[e.Step].Name, time.Since(start).Seconds(), err)
			return nil
		}
		klog.V(4).Infof("Running %s [Success] [%fs]", e.steps[e.Step].Name, time.Since(start).Seconds())

		e.Step++
		e.backup()
	}

	klog.V(1).Infof("install task [Sucesss] [%fs]", time.Since(start).Seconds())
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
			Func: func() error {
				return TarInstallMovePackage(e) //todo：Abstract into general functions
			},
		},
		{
			Name: "init shell before install",
			Func: func() error {
				return NilFunc(e) //todo step test
				return InitShellPreInstall(e)
			},
		},
		{
			Name: "set bash_complete", //todo: set kubectl bash_complete
			Func: func() error {
				return SetBinExport(e)
			},
		},
	}...)

	// init node
	e.steps = append(e.steps, []Handler{
		{
			Name: "check node",
			Func: func() error {
				return NilFunc(e)
			},
		},
		{
			Name: "set node host",
			Func: func() error {
				return NilFunc(e) //todo step test
				return SetNodeHost(e)
			},
		},
		{
			Name: "set kernel module",
			Func: func() error {
				return NilFunc(e) //todo step test
				return SetKernelModule(e)
			},
		},
		{
			Name: "set sysctl",
			Func: func() error {
				return NilFunc(e) //todo step test
				return SetSysctl(e)
			},
		},
	}...)

	// install container runtime
	e.steps = append(e.steps, []Handler{
		{
			Name: "install docker",
			Func: func() error {
				return NilFunc(e)
			},
		},
		{
			Name: "load images",
			Func: func() error {
				return NilFunc(e)
			},
		},
	}...)

	// create kubeadm config
	e.steps = append(e.steps, []Handler{
		{
			Name: "create kubeadm config",
			Func: func() error {
				return NilFunc(e) //todo step test
				return CreateKubeadmConfig(e)
			},
		},
	}...)

	// create ca
	e.steps = append(e.steps, []Handler{
		{
			Name: "create etcd ca",
			Func: func() error {
				return NilFunc(e)
			},
		},
		{
			Name: "create kube-api-service ca",
			Func: func() error {
				return NilFunc(e)
			},
		},
		{
			Name: "create kube-controller-manager ca",
			Func: func() error {
				return NilFunc(e)
			},
		},
		{
			Name: "create kube-scheduler ca",
			Func: func() error {
				return NilFunc(e)
			},
		},
	}...)

	// config && create kube-* yaml
	e.steps = append(e.steps, []Handler{
		{
			Name: "create kubeadm config",
			Func: func() error {
				return NilFunc(e)
			},
		},
		{
			Name: "config etcd",
			Func: func() error {
				return NilFunc(e)
			},
		},
		{
			Name: "config kube-api-service",
			Func: func() error {
				return NilFunc(e)
			},
		},
		{
			Name: "config kube-controller-manager",
			Func: func() error {
				return NilFunc(e)
			},
		},
		{
			Name: "config kube-scheduler",
			Func: func() error {
				return NilFunc(e)
			},
		},
	}...)

	// kubeadm init
	e.steps = append(e.steps, []Handler{
		{
			Name: "kubeadm init",
			Func: func() error {
				return NilFunc(e)
			},
		},
	}...)

	// check kubernetes cluster health
	e.steps = append(e.steps, []Handler{
		{
			Name: "check kubernetes cluster",
			Func: func() error {
				return NilFunc(e)
			},
		},
	}...)

	// install cloud edge-apps
	e.steps = append(e.steps, []Handler{
		{
			Name: "deploy tunnel-coredns",
			Func: func() error {
				return NilFunc(e)
			},
		},
		{
			Name: "deploy tunnel-cloud",
			Func: func() error {
				return NilFunc(e)
			},
		},
		{
			Name: "deploy application-grid controller",
			Func: func() error {
				return NilFunc(e)
			},
		},
		{
			Name: "deploy application-grid wrapper",
			Func: func() error {
				return NilFunc(e)
			},
		},
		{
			Name: "deploy edge-health admission",
			Func: func() error {
				return NilFunc(e)
			},
		},
	}...)

	// install edge edge-apps
	e.steps = append(e.steps, []Handler{
		{
			Name: "daemonset flannel",
			Func: func() error {
				return NilFunc(e)
			},
		},
		{
			Name: "daemonset tunnel edge",
			Func: func() error {
				return NilFunc(e)
			},
		},
		{
			Name: "daemonset coredns",
			Func: func() error {
				return NilFunc(e)
			},
		},
		{
			Name: "daemonset kube-proxy",
			Func: func() error {
				return NilFunc(e)
			},
		},
		{
			Name: "daemonset edge-health",
			Func: func() error {
				return NilFunc(e)
			},
		},
	}...)

	// check edge cluster health
	e.steps = append(e.steps, []Handler{
		{
			Name: "check edge cluster",
			Func: func() error {
				return NilFunc(e)
			},
		},
	}...)

	return nil
}
