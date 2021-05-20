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
	"github.com/superedge/superedge/pkg/tunnel/context"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"k8s.io/klog"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
)

func HandleServerConn(conn net.Conn) {
	request, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		klog.Errorf("Failed to get http request, error: %v", err)
		return
	}
	if request.Method == util.HttpMethod {

		nodeinfo := strings.Split(request.Host, ":")
		node := context.GetContext().GetNode(nodeinfo[0])
		if node == nil {
			addrs, err := net.LookupHost(nodeinfo[0])
			if err != nil || len(addrs) == 0 {
				klog.Errorf("Node：%s is not connected", nodeinfo[0])
				_, err = conn.Write([]byte(util.BadGateway))
				if err != nil {
					klog.Errorf("")
				}
			}
			remoteConn, err := net.Dial("tcp", addrs[0]+":22")
			if err != nil {

			}
			remoteConn.Write(httputil.DumpRequest(request))

		}
	}
}
