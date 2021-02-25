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
	"crypto/tls"
	"fmt"
	"github.com/superedge/superedge/pkg/edge-health/common"
	"github.com/superedge/superedge/pkg/edge-health/metadata"
	"github.com/superedge/superedge/pkg/edge-health/util"
	"io/ioutil"
	"k8s.io/klog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type KubeletAuthCheckPlugin struct {
	BasePlugin
	client *http.Client
}

func (kacp *KubeletAuthCheckPlugin) Name() string {
	return "KubeletAuthCheck"
}

func (kacp *KubeletAuthCheckPlugin) Set(s string) error {
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
			kacp.HealthCheckoutTimeOut = timeout
		case "retries":
			retries, _ := strconv.Atoi(strings.TrimSpace(arr[1]))
			kacp.HealthCheckRetries = retries
		case "weight":
			weight, _ := strconv.ParseFloat(strings.TrimSpace(arr[1]), 64)
			kacp.Weight = weight
		case "port":
			port, _ := strconv.Atoi(strings.TrimSpace(arr[1]))
			kacp.Port = port
		}
	}
	// Init http client for later usage
	kacp.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
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
		Timeout: time.Duration(kacp.HealthCheckoutTimeOut) * time.Second,
	}
	PluginInfo = NewPlugin()
	PluginInfo.Register(kacp)
	return err
}

func (kacp *KubeletAuthCheckPlugin) String() string {
	return fmt.Sprintf("%+v", kacp)
}

func (kacp *KubeletAuthCheckPlugin) Type() string {
	return "KubeletAuthCheckPlugin"
}

func (kacp *KubeletAuthCheckPlugin) CheckExecute(checkMetadata *metadata.CheckMetadata) {
	copyCheckedIp := checkMetadata.CopyCheckedIp()
	util.ParallelizeUntil(context.TODO(), 16, len(copyCheckedIp), func(index int) {
		checkedIp := copyCheckedIp[index]
		if err := kubeletAuthPing(kacp.client, checkedIp, kacp.Port, kacp.HealthCheckRetries); err != nil {
			klog.V(4).Infof("Edge health check plugin %s for ip %s succeed", kacp.Name(), checkedIp)
			checkMetadata.SetByPluginScore(checkedIp, kacp.Name(), kacp.GetWeight(), common.CheckScoreMax)
		} else {
			klog.Warning("Edge health check plugin %s for ip %s failed, possible reason %s", kacp.Name(), checkedIp, err.Error())
			checkMetadata.SetByPluginScore(checkedIp, kacp.Name(), kacp.GetWeight(), common.CheckScoreMin)
		}
	})
}

func kubeletAuthPing(client *http.Client, checkedIp string, port, retries int) error {
	var (
		err   error
		req   *http.Request
		token []byte
	)
	for i := 0; i < retries; i++ {
		// Read kubelet token file
		token, err = ioutil.ReadFile(common.TokenFile)
		if err != nil {
			klog.Errorf("ReadTokenFile failed %+v", err)
			continue
		}
		// Construct kubelet healthz http request with token
		url := "https://" + checkedIp + ":" + strconv.Itoa(port) + "/healthz"
		klog.V(4).Infof("Url is %s", url)
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			klog.Errorf("New kubelet auth ping request failed %+v", err)
			continue
		}
		req.Header.Add("Authorization", "Bearer "+string(token))
		if err = util.DoRequestAndDiscard(client, req); err != nil {
			klog.Errorf("DoRequestAndDiscard kubelet auth ping request failed %+v", err)
		} else {
			break
		}
	}
	return err
}
