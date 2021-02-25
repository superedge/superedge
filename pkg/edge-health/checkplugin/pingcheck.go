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
	"strconv"
	"strings"
	"time"
)

type PingCheckPlugin struct {
	BasePlugin
}

func (pcp *PingCheckPlugin) Name() string {
	return "PingCheck"
}

// TODO: handle flag parse errors
func (pcp *PingCheckPlugin) Set(s string) error {
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
			pcp.HealthCheckoutTimeOut = timeout
		case "retries":
			retries, _ := strconv.Atoi(strings.TrimSpace(arr[1]))
			pcp.HealthCheckRetries = retries
		case "weight":
			weight, _ := strconv.ParseFloat(strings.TrimSpace(arr[1]), 64)
			pcp.Weight = weight
		case "port":
			port, _ := strconv.Atoi(strings.TrimSpace(arr[1]))
			pcp.Port = port
		}
	}
	PluginInfo = NewPlugin()
	PluginInfo.Register(pcp)
	return err
}

func (pcp *PingCheckPlugin) String() string {
	return fmt.Sprintf("%+v", pcp)
}

func (pcp *PingCheckPlugin) Type() string {
	return "PingCheckPlugin"
}

func (pcp *PingCheckPlugin) CheckExecute(checkMetadata *metadata.CheckMetadata) {
	copyCheckedIp := checkMetadata.CopyCheckedIp()
	util.ParallelizeUntil(context.TODO(), 16, len(copyCheckedIp), func(index int) {
		checkedIp := copyCheckedIp[index]
		var err error
		for i := 0; i < pcp.HealthCheckRetries; i++ {
			if _, err := net.DialTimeout("tcp", checkedIp+":"+strconv.Itoa(pcp.Port), time.Duration(pcp.HealthCheckoutTimeOut)*time.Second); err == nil {
				break
			}
		}
		if err == nil {
			klog.V(4).Infof("Edge ping health check plugin %s for ip %s succeed", pcp.Name(), checkedIp)
			checkMetadata.SetByPluginScore(checkedIp, pcp.Name(), pcp.GetWeight(), common.CheckScoreMax)
		} else {
			klog.Warning("Edge ping health check plugin %s for ip %s failed, possible reason %s", pcp.Name(), checkedIp, err.Error())
			checkMetadata.SetByPluginScore(checkedIp, pcp.Name(), pcp.GetWeight(), common.CheckScoreMin)
		}
	})
}
