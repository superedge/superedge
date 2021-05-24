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
	uuid "github.com/satori/go.uuid"
	"github.com/superedge/superedge/pkg/tunnel/context"
	"github.com/superedge/superedge/pkg/tunnel/proxy/common"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"io"
	"k8s.io/klog"
	"net"
	"net/http"
	"strings"
)

const (
	RequestCache = 10 * 1024
)

func HandleServerConn(proxyConn net.Conn) {
	rawRequest := bytes.NewBuffer(make([]byte, RequestCache))
	rawRequest.Reset()
	reqReader := bufio.NewReader(io.TeeReader(proxyConn, rawRequest))
	request, err := http.ReadRequest(reqReader)
	if err != nil {
		klog.Errorf("Failed to get http request, error: %v", err)
		return
	}
	if request.Method == util.HttpMethod {
		nodeinfo := strings.Split(request.Host, ":")
		node := context.GetContext().GetNode(nodeinfo[0])
		if node == nil {
			addrs, err := net.LookupHost(nodeinfo[0])
			if err != nil {
				klog.Errorf("DNS parsing error: %v", err)
				return
			}

			if len(addrs) == 0 {
				klog.Errorf("Node：%s is not connected", nodeinfo[0])
				_, err = proxyConn.Write([]byte(util.BadGateway))
				if err != nil {
					klog.Errorf("Failed to write data to proxyConn, error: %v", err)
					return
				}
			}

			remoteConn, err := net.Dial("tcp", addrs[0]+":22")
			if err != nil {
				klog.Errorf("Failed to establish a connection between proxyServer and backendServer, error: %v", err)
				return
			}
			_, err = remoteConn.Write(rawRequest.Bytes())
			if err != nil {
				klog.Errorf("Failed to write data to remoteConn, error: %v", err)
				return
			}
			go func() {
				_, writeErr := io.Copy(remoteConn, proxyConn)
				if writeErr != nil {
					klog.Errorf("Failed to copy data to remoteConn, error: %v", err)
				}
			}()
			_, err = io.Copy(proxyConn, remoteConn)
			if err != nil {
				klog.Errorf("Failed to read data from remoteConn, error: %v", err)
				return
			}
		} else {
			uid := uuid.NewV4().String()
			ch := context.GetContext().AddConn(uid)
			go common.Read(proxyConn, node, util.TCP_FRONTEND, uid)
			common.Write(proxyConn, ch)
		}
	}
}
