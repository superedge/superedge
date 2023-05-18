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
	uuid "github.com/satori/go.uuid"
	"github.com/superedge/superedge/pkg/tunnel/proto"
	"github.com/superedge/superedge/pkg/tunnel/proxy/common"
	"github.com/superedge/superedge/pkg/tunnel/tunnelcontext"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"k8s.io/klog/v2"
	"net"
	"net/http"
	"os"
	"strings"
)

func HttpProxyEdgeServer(conn net.Conn) {
	req, raw, err := util.GetRequestFromConn(conn)
	if err != nil {
		klog.V(2).ErrorS(err, "failed to get http request", "remoteAddr", conn.RemoteAddr(), "localAddr", conn.LocalAddr())
		return
	}
	req = req.WithContext(context.WithValue(req.Context(), util.STREAM_TRACE_ID, uuid.NewV4().String()))
	klog.V(3).InfoS("receive request", "method", req.Method, "host", req.Host, "remoteAddr", conn.RemoteAddr(), "localAddr", conn.LocalAddr(), "req", req)

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
			klog.Errorf("the service format is incorrect, the supported format: serviceName.nameSpace")
			writeErr := util.InternalServerErrorMsg(conn, fmt.Sprintf("the service format is incorrect, the supported format: serviceName.nameSpace"), req.Context().Value(util.STREAM_TRACE_ID).(string))
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
				err := util.InternalServerErrorMsg(conn, err.Error(), req.Context().Value(util.STREAM_TRACE_ID).(string))
				if err != nil {
					klog.ErrorS(err, "failed to write resp msg", util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
				}
			}
			go common.Read(conn, node, util.HTTP_PROXY, util.TCP_FORWARD, tunnelConn.GetUid())
			common.Write(conn, tunnelConn)
		} else {
			err := util.InternalServerErrorMsg(conn, fmt.Sprintf("failed to get edge node %s", os.Getenv(util.NODE_NAME_ENV)), req.Context().Value(util.STREAM_TRACE_ID).(string))
			if err != nil {
				klog.Errorf("failed to write resp msg, error:%v", err)
			}
		}

	} else {
		node := tunnelcontext.GetContext().GetNode(os.Getenv(util.NODE_NAME_ENV))
		if node == nil {
			err := util.InternalServerErrorMsg(conn, fmt.Sprintf("failed to get edge node %s", os.Getenv(util.NODE_NAME_ENV)), req.Context().Value(util.STREAM_TRACE_ID).(string))
			if err != nil {
				klog.ErrorS(err, "failed to write resp msg", req.Context().Value(util.STREAM_TRACE_ID).(string))
			}
		} else {
			tunnelConn, err := node.ConnectNode(util.HTTP_PROXY, net.JoinHostPort(host, port), req.Context())
			if err != nil {
				err := util.InternalServerErrorMsg(conn, err.Error(), req.Context().Value(util.STREAM_TRACE_ID).(string))
				if err != nil {
					klog.ErrorS(err, "failed to write resp msg", util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
					return
				}
			} else {
				node.Send2Node(&proto.StreamMsg{
					Node:     os.Getenv(util.NODE_NAME_ENV),
					Category: util.HTTP_PROXY,
					Type:     util.TCP_FORWARD,
					Topic:    req.Context().Value(util.STREAM_TRACE_ID).(string),
					Data:     raw.Bytes(),
					Addr:     req.Host,
				})
				go common.Read(conn, node, util.HTTP_PROXY, util.TCP_FORWARD, tunnelConn.GetUid())
				common.Write(conn, tunnelConn)
			}
		}
	}

}
