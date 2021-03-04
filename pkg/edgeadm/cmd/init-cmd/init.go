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
	"github.com/superedge/superedge/pkg/edgeadm/cmd/init-cmd/config"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/edgecluster"
	"io"
	"io/ioutil"
	log "k8s.io/klog/v2"
	"time"
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

type initData struct {
	Config   config.Config           `json:"config"`
	Cluster  edgecluster.EdgeCluster `json:"cluster"`
	steps    []Handler
	Step     int `json:"step"`
	progress ClusterProgress
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

func NewInitCMD(out io.Writer) *cobra.Command {
	action := newInit()
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create first master node of edge kubernetes cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if err := action.complete(); err != nil {
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

	AddEdgeConfigFlags(cmd.Flags(), &action.Config.EdgeConfig)
	AddKubeadmConfigFlags(cmd.Flags(), &action.Config.KubeadmConfig)
	// parse-config
	//

	return cmd
}

func AddEdgeConfigFlags(flagSet *pflag.FlagSet, cfg *config.EdgeConfig) {
	flagSet.StringVar(
		&cfg.InstallPkgPath, constant.InstallPkgPath, "/root/install-pkg.tar.gz",
		"Install static package path of edge kubernetes cluster.",
	)
}

func AddKubeadmConfigFlags(flagSet *pflag.FlagSet, cfg *config.KubeadmConfig) {
	flagSet.StringVar(
		&cfg.KubeadmConfPath, constant.KubeadmConfig, "/root/.edgeadm/kubeadm.config",
		"Install static package path of edge kubernetes cluster.",
	)
}

func (e *initData) config() error {
	return nil
}

func (e *initData) complete() error {
	return nil
}

func (e *initData) validate() error {
	return nil
}

func (e *initData) backup() error {
	log.V(4).Infof("===>starting install backup()")
	data, _ := json.MarshalIndent(e, "", " ")
	return ioutil.WriteFile(constant.EdgeClusterFile, data, 0777)
}

func (e *initData) runInit() error {
	start := time.Now()
	e.initSteps()
	defer e.backup()

	if e.Step == 0 {
		log.V(4).Infof("===>starting install task")
		e.progress.Status = constant.StatusDoing
	}

	for e.Step < len(e.steps) {
		log.V(4).Infof("%d.%s doing", e.Step, e.steps[e.Step].Name)

		start := time.Now()
		err := e.steps[e.Step].Func()
		if err != nil {
			e.progress.Status = constant.StatusFailed
			log.V(4).Infof("%d.%s [Failed] [%fs] error %s", e.Step, e.steps[e.Step].Name, time.Since(start).Seconds(), err)
			return nil
		}
		log.V(4).Infof("%d.%s [Success] [%fs]", e.Step, e.steps[e.Step].Name, time.Since(start).Seconds())

		e.Step++
		e.backup()
	}

	log.V(5).Info("===>install task [Sucesss] [%fs]", time.Since(start).Seconds())
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
			Name: "Execute pre install hook",
			Func: e.tarInstallMovePackage,
		},
	}...)
	return nil
}
