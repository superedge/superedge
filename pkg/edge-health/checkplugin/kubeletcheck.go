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

package checkplugin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/superedge/superedge/pkg/edge-health/common"
	"github.com/superedge/superedge/pkg/edge-health/data"
	"k8s.io/klog/v2"
)

type KubeletCheckPlugin struct {
	BasePlugin
}

func (p KubeletCheckPlugin) Name() string {
	return "KubeletCheck"
}

func (p *KubeletCheckPlugin) Set(s string) error {
	var (
		err error
	)

	for _, para := range strings.Split(s, ",") {
		if len(para) == 0 {
			continue
		}
		arr := strings.Split(para, "=")
		trimkey := strings.TrimSpace(arr[0])
		switch trimkey {
		case "timeout":
			timeout, _ := strconv.Atoi(strings.TrimSpace(arr[1]))
			(*p).HealthCheckoutTimeOut = timeout
		case "retrytime":
			retrytime, _ := strconv.Atoi(strings.TrimSpace(arr[1]))
			(*p).HealthCheckRetryTime = retrytime
		case "weight":
			weight, _ := strconv.ParseFloat(strings.TrimSpace(arr[1]), 64)
			(*p).Weight = weight
		case "port":
			port, _ := strconv.Atoi(strings.TrimSpace(arr[1]))
			(*p).Port = port
		}
		(*p).PluginName = p.Name()
	}
	PluginInfo = NewPluginInfo()
	PluginInfo.AddPlugin(p)
	klog.V(4).Infof("len of plugins is %d", len(PluginInfo.Plugins))

	return err
}

func (p *KubeletCheckPlugin) String() string {
	return fmt.Sprintf("%v", *p)
}

func (i *KubeletCheckPlugin) Type() string {
	return "KubeletCheckPlugin"
}

func (plugin KubeletCheckPlugin) CheckExecute(wg *sync.WaitGroup) {
	execwg := sync.WaitGroup{}
	execwg.Add(len(data.CheckInfoResult.CheckInfo))
	for k := range data.CheckInfoResult.CopyCheckInfo() {
		temp := k
		go func(execwg *sync.WaitGroup) {
			checkOk, err := ping(plugin.HealthCheckoutTimeOut, plugin.HealthCheckRetryTime, temp, plugin.Port)
			if checkOk {
				klog.V(4).Infof("%s use %s plugin check %s successd", common.NodeIP, plugin.Name(), temp)
				data.CheckInfoResult.SetCheckInfo(temp, plugin.Name(), plugin.GetWeight(), 100)
			} else {
				klog.V(2).Infof("%s use %s plugin check %s failed, reason: %s", common.NodeIP, plugin.Name(), temp, err.Error())
				data.CheckInfoResult.SetCheckInfo(temp, plugin.Name(), plugin.GetWeight(), 0)
			}
			execwg.Done()
		}(&execwg)
	}
	execwg.Wait()
	wg.Done()
}

func ping(timeout, retryTime int, checkedIp string, port int) (bool, error) {
	var (
		err error
		ok  bool
	)
	client := http.Client{Timeout: time.Duration(timeout) * time.Second}
	url := "http://" + checkedIp + ":" + strconv.Itoa(port) + "/healthz"
	klog.V(4).Infof("url is %s", url)
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false, fmt.Errorf("error new ping request: %w", err)
	}
	for i := 0; i < retryTime; i++ {
		if ok, err = PingDo(client, req); ok {
			return true, nil
		}
	}
	klog.Error("kubelet ping failed")
	return false, err
}
