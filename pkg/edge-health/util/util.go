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

package util

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/superedge/superedge/pkg/edge-health/check"
	"github.com/superedge/superedge/pkg/edge-health/common"
	"github.com/superedge/superedge/pkg/edge-health/data"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"golang.org/x/sys/unix"
	"io"
	"k8s.io/api/core/v1"
	"os"
	"os/signal"
)

func GenerateHmac(communicatedata data.CommunicateData) (string, error) {
	part1byte, _ := json.Marshal(communicatedata.SourceIP)
	part2byte, _ := json.Marshal(communicatedata.ResultDetail)
	hmacBefore := string(part1byte) + string(part2byte)
	if hmacconf, err := check.ConfigMapManager.ConfigMapLister.ConfigMaps(constant.NamespaceEdgeSystem).Get(common.HmacConfig); err != nil {
		return "", err
	} else {
		return GetHmacCode(hmacBefore, hmacconf.Data[common.HmacKey])
	}
}

func GetHmacCode(s, key string) (string, error) {
	h := hmac.New(sha256.New, []byte(key))
	if _, err := io.WriteString(h, s); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func GetNodeNameByIp(nodes []v1.Node, Ip string) string {
	for _, v := range nodes {
		for _, i := range v.Status.Addresses {
			if i.Type == v1.NodeInternalIP {
				if i.Address == Ip {
					return v.Name
				}
			}
		}
	}
	return ""
}

func SignalWatch() (context.Context, context.CancelFunc) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, unix.SIGTERM, unix.SIGINT)
	go func() {
		for range signals {
			cancelFunc()
		}
	}()
	return ctx, cancelFunc
}
