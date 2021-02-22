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
	"context"
	"fmt"
	"github.com/superedge/superedge/pkg/edge-health/common"
	"github.com/superedge/superedge/pkg/edge-health/metadata"
	"github.com/superedge/superedge/pkg/edge-health/util"
	"k8s.io/klog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type KubeletCheckPlugin struct {
	BasePlugin
	client *http.Client
}

func (kcp *KubeletCheckPlugin) Name() string {
	return "KubeletCheck"
}

// TODO: handle flag parse errors
func (kcp *KubeletCheckPlugin) Set(s string) error {
	var err error
	for _, para := range strings.Split(s, ",") {
		if len(para) == 0 {
			continue
		}
		arr := strings.Split(para, "=")
		trimKey := strings.TrimSpace(arr[0])
		switch trimKey {
		case "timeout":
			timeout, _ := strconv.Atoi(strings.TrimSpace(arr[1]))
			kcp.HealthCheckoutTimeOut = timeout
		case "retries":
			retries, _ := strconv.Atoi(strings.TrimSpace(arr[1]))
			kcp.HealthCheckRetries = retries
		case "weight":
			weight, _ := strconv.ParseFloat(strings.TrimSpace(arr[1]), 64)
			kcp.Weight = weight
		case "port":
			port, _ := strconv.Atoi(strings.TrimSpace(arr[1]))
			kcp.Port = port
		}
	}
	// Init http client for later usage
	kcp.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: time.Duration(kcp.HealthCheckoutTimeOut) * time.Second,
	}
	PluginInfo = NewPlugin()
	PluginInfo.Register(kcp)
	return err
}

func (kcp *KubeletCheckPlugin) String() string {
	return fmt.Sprintf("%v", kcp)
}

func (kcp *KubeletCheckPlugin) Type() string {
	return "KubeletCheckPlugin"
}

func (kcp *KubeletCheckPlugin) CheckExecute(checkMetadata *metadata.CheckMetadata) {
	copyCheckedIp := checkMetadata.CopyCheckedIp()
	util.ParallelizeUntil(context.TODO(), 16, len(copyCheckedIp), func(index int) {
		checkedIp := copyCheckedIp[index]
		if err := kubeletPing(kcp.client, checkedIp, kcp.Port, kcp.HealthCheckRetries); err != nil {
			klog.V(4).Infof("Edge kubelet health check plugin %s for ip %s succeed", kcp.Name(), checkedIp)
			checkMetadata.SetByPluginScore(checkedIp, kcp.Name(), kcp.GetWeight(), common.CheckScoreMax)
		} else {
			klog.Warning("Edge kubelet health check plugin %s for ip %s failed, possible reason %s", kcp.Name(), checkedIp, err.Error())
			checkMetadata.SetByPluginScore(checkedIp, kcp.Name(), kcp.GetWeight(), common.CheckScoreMin)
		}
	})
}

func kubeletPing(client *http.Client, checkedIp string, port, retries int) error {
	var (
		err error
		req *http.Request
	)
	for i := 0; i < retries; i++ {
		// Construct kubelet healthz http request
		url := "http://" + checkedIp + ":" + strconv.Itoa(port) + "/healthz"
		klog.V(4).Infof("Url is %s", url)
		req, err = http.NewRequest("HEAD", url, nil)
		if err != nil {
			klog.Errorf("New kubelet ping request failed %v", err)
			continue
		}
		if err = util.DoRequestAndDiscard(client, req); err != nil {
			klog.Errorf("DoRequestAndDiscard kubelet ping request failed %v", err)
		} else {
			break
		}
	}
	return err
}
