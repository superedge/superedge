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

package connect

import (
	"bufio"
	"bytes"
	"github.com/superedge/superedge/pkg/tunnel/proxy/common"
	"github.com/superedge/superedge/pkg/tunnel/proxy/common/indexers"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"io"
	"k8s.io/klog/v2"
	"net"
	"net/http"
)

const (
	RequestCache = 10 * 1024
)

func HandleEgressConn(proxyConn net.Conn) {
	defer proxyConn.Close()
	rawRequest := bytes.NewBuffer(make([]byte, RequestCache))
	rawRequest.Reset()
	reqReader := bufio.NewReader(io.TeeReader(proxyConn, rawRequest))
	request, err := http.ReadRequest(reqReader)
	if err != nil {
		klog.Errorf("Failed to get http request, error: %v", err)
		return
	}
	if request.Method == util.HttpMethod {
		host, port, err := net.SplitHostPort(request.Host)
		if err != nil {
			klog.Errorf("Failed to get host and port, module: %s, error: %v", util.EGRESS, err)
			proxyConn.Write([]byte("Failed to get host and port"))
			return
		}
		ip := net.ParseIP(host)
		if ip == nil {
			internalIp, err := indexers.GetNodeIPByName(host)
			if err != nil {
				klog.Errorf("Failed to get internalIp of node, error: %v", err)
				common.ProxyEdgeNode(host, "127.0.0.1", port, util.EGRESS, proxyConn, rawRequest)
			} else {
				common.ProxyEdgeNode(host, internalIp, port, util.EGRESS, proxyConn, rawRequest)
			}

		} else {
			//Request pods on edge nodes
			node, err := indexers.GetNodeByPodIP(host)
			if err != nil {
				klog.Errorf("Failed to get the node where the pod is located, module: %s, error: %v", util.EGRESS, err)
				return
			}
			common.ProxyEdgeNode(node, host, port, util.EGRESS, proxyConn, rawRequest)
		}
	}
}
