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
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	uuid "github.com/satori/go.uuid"
	"github.com/superedge/superedge/pkg/tunnel/context"
	"github.com/superedge/superedge/pkg/tunnel/proto"
	"github.com/superedge/superedge/pkg/tunnel/proxy/common"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"k8s.io/klog/v2"
)

func HttpProxyEdgeServer(conn net.Conn) {
	req, raw, err := util.GetRequestFromConn(conn)
	if err != nil {
		klog.Errorf("Failed to read httpRequest, error: %v", err)
		return
	}

	host, port, err := net.SplitHostPort(req.Host)
	if err != nil {
		if len(strings.Split(req.Host, ":")) < 2 {
			host = req.Host
			switch req.URL.Scheme {
			case "http":
				port = "80"
			case "https":
				port = "443"
			}
		} else {
			klog.Errorf("Failed to resolve host %s, error: %v", req.Host, err)
			return
		}

	}
	podIp := net.ParseIP(host)
	if podIp == nil {
		//校验serviceName的格式
		if len(strings.Split(host, ".")) < 2 {
			klog.Errorf("Please specify the nameSpace where the accessed service is located")
			return
		}
	}
	if req.Method == http.MethodConnect {
		uid := uuid.NewV4().String()
		node := context.GetContext().GetNode(os.Getenv(util.NODE_NAME_ENV))
		if node == nil {
			klog.Errorf("failed to get edge node: %s", os.Getenv(util.NODE_NAME_ENV))
			return
		}
		ch := context.GetContext().AddConn(uid)
		node.BindNode(uid)
		node.Send2Node(&proto.StreamMsg{
			Node:     os.Getenv(util.NODE_NAME_ENV),
			Category: util.HTTP_PROXY,
			Type:     util.HTTP_PROXY_ACCESS,
			Topic:    uid,
			Data:     raw.Bytes(),
			Addr:     net.JoinHostPort(host, port),
		})
		go common.Read(conn, node, util.HTTP_PROXY, util.TCP_BACKEND, uid, req.Host)
		common.Write(conn, ch)
	} else {
		uid := uuid.NewV4().String()
		node := context.GetContext().GetNode(os.Getenv(util.NODE_NAME_ENV))
		if node == nil {
			klog.Errorf("failed to get edge node: %s", os.Getenv(util.NODE_NAME_ENV))
			return
		}
		ch := context.GetContext().AddConn(uid)
		node.BindNode(uid)
		node.Send2Node(&proto.StreamMsg{
			Node:     os.Getenv(util.NODE_NAME_ENV),
			Category: util.HTTP_PROXY,
			Type:     util.HTTP_PROXY_ACCESS,
			Topic:    uid,
			Data:     []byte(fmt.Sprintf("CONNECT %s HTTP/1.1\r\n\r\n\r\n", net.JoinHostPort(host, port))),
			Addr:     net.JoinHostPort(host, port),
		})
		recv := <-ch.ConnRecv()
		resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(recv.Data)), nil)
		if err != nil {
			klog.Errorf("Failed to parse httpRequest, error: %v", err)
			conn.Close()
			return
		}
		if resp.StatusCode == http.StatusOK {
			node.Send2Node(&proto.StreamMsg{
				Node:     os.Getenv(util.NODE_NAME_ENV),
				Category: util.HTTP_PROXY,
				Type:     util.TCP_BACKEND,
				Topic:    uid,
				Data:     raw.Bytes(),
				Addr:     req.Host,
			})
			go common.Read(conn, node, util.HTTP_PROXY, util.TCP_BACKEND, uid, req.Host)
			common.Write(conn, ch)
		}

	}

}
