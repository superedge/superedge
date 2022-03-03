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
	"github.com/superedge/superedge/pkg/tunnel/util"
	"io"
	"k8s.io/klog/v2"
	"net"
	"net/http"
)

const (
	RequestCache = 10 * 1024
)

func HandleServerConn(proxyConn net.Conn) {
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
			klog.Errorf("Failed to obtain the login destination node and SSH server port, module: %s, error: %v", util.SSH, err)
			proxyConn.Write([]byte("Failed to obtain the login destination node and SSH server port"))
			return
		}
		common.ProxyEdgeNode(host, "127.0.0.1", port, util.SSH, proxyConn, rawRequest)
	}
}
