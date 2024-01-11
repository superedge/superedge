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
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	uuid "github.com/satori/go.uuid"
	"k8s.io/klog/v2"

	"github.com/superedge/superedge/pkg/tunnel/proto"
	"github.com/superedge/superedge/pkg/tunnel/proxy/common"
	"github.com/superedge/superedge/pkg/tunnel/tunnelcontext"
	"github.com/superedge/superedge/pkg/tunnel/util"
)

func HttpProxyEdgeServer(conn net.Conn) {
	req, raw, err := util.GetRequestFromConn(conn)
	if err != nil {
		klog.V(2).ErrorS(err, "failed to get http request", "remoteAddr", conn.RemoteAddr(),
			"localAddr", conn.LocalAddr())
		return
	}
	uuid := uuid.NewV4().String()
	req = req.WithContext(context.WithValue(req.Context(), util.STREAM_TRACE_ID, uuid))
	klog.V(2).InfoS("receive request", "method", req.Method, "host", req.Host,
		"remoteAddr", conn.RemoteAddr(), "localAddr", conn.LocalAddr(), "req", req, util.STREAM_TRACE_ID, uuid)

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
		// 校验serviceName的格式
		if len(strings.Split(host, ".")) < 2 {
			klog.Errorf("the service format is incorrect, the supported format: serviceName.nameSpace")
			writeErr := util.InternalServerErrorMsg(conn,
				"the service format is incorrect, supported format: serviceName.nameSpace", uuid)
			if writeErr != nil {
				klog.Error(writeErr)
			}
			return
		}
	}
	if req.Method == http.MethodConnect {
		node := tunnelcontext.GetContext().GetNode(os.Getenv(util.NODE_NAME_ENV))
		if node != nil {
			tunnelConn, err := node.ConnectNode(util.HTTP_PROXY, net.JoinHostPort(host, port), req.Context())
			if err != nil {
				err := util.InternalServerErrorMsg(conn, err.Error(), uuid)
				if err != nil {
					klog.ErrorS(err, "failed to write resp msg", util.STREAM_TRACE_ID, uuid)
					return
				}
			}

			// Return 200 status code
			_, err = conn.Write([]byte(util.ConnectMsg))
			if err != nil {
				klog.ErrorS(err, "failed to write data to proxyConn", util.STREAM_TRACE_ID, uuid)
				return
			}

			go common.Read(conn, node, util.HTTP_PROXY, util.TCP_FORWARD, tunnelConn.GetUid())
			common.Write(conn, tunnelConn)
		} else {
			err := util.InternalServerErrorMsg(conn,
				fmt.Sprintf("failed to get edge node %s", os.Getenv(util.NODE_NAME_ENV)), uuid)
			if err != nil {
				klog.Errorf("failed to write resp msg, error:%v", err)
				return
			}
		}

	} else {
		node := tunnelcontext.GetContext().GetNode(os.Getenv(util.NODE_NAME_ENV))
		if node == nil {
			err := util.InternalServerErrorMsg(conn,
				fmt.Sprintf("failed to get edge node %s", os.Getenv(util.NODE_NAME_ENV)), uuid)
			if err != nil {
				klog.ErrorS(err, "failed to write resp msg", uuid)
				return
			}
		} else {
			tunnelConn, err := node.ConnectNode(util.HTTP_PROXY, net.JoinHostPort(host, port), req.Context())
			if err != nil {
				err := util.InternalServerErrorMsg(conn, err.Error(), uuid)
				if err != nil {
					klog.ErrorS(err, "failed to write resp msg", util.STREAM_TRACE_ID, uuid)
					return
				}
			} else {
				node.Send2Node(&proto.StreamMsg{
					Node:     os.Getenv(util.NODE_NAME_ENV),
					Category: util.HTTP_PROXY,
					Type:     util.TCP_FORWARD,
					Topic:    uuid,
					Data:     raw.Bytes(),
					Addr:     req.Host,
				})
				go common.Read(conn, node, util.HTTP_PROXY, util.TCP_FORWARD, tunnelConn.GetUid())
				common.Write(conn, tunnelConn)
			}
		}
	}

}
