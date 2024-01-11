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

package common

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/superedge/superedge/pkg/tunnel/conf"
	"github.com/superedge/superedge/pkg/tunnel/proxy/modules/stream/streammng/connect"
	"github.com/superedge/superedge/pkg/tunnel/tunnelcontext"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"k8s.io/klog/v2"
)

type TargetType int

const (
	LocalPodType       TargetType = 0 //transfer through the tunnel of this tunnel-cloud pod
	RemotePodType      TargetType = 1 //transfer through the tunnel of other tunnel-cloud pod
	DisconnectNodeType TargetType = 2 //the target node is not registered in the route cache
)

func ForwardNode(nodename, host, port, category string, proxyConn net.Conn, ctx context.Context) error {
	node := tunnelcontext.GetContext().GetNode(nodename)

	//Direct forwarding edge nodes
	if node != nil {

		//If the edge node establishes a long connection with this pod, it will be forwarded directly
		conn, err := node.ConnectNode(category, net.JoinHostPort(host, port), ctx)
		if err == nil {
			_, err := proxyConn.Write([]byte(util.ConnectMsg))
			if err != nil {
				klog.ErrorS(err, "failed to write data to proxyConn", util.STREAM_TRACE_ID, ctx.Value(util.STREAM_TRACE_ID))
				return err
			}

			go Read(proxyConn, node, category, util.TCP_FORWARD, conn.GetUid())
			Write(proxyConn, conn)

		} else {
			klog.ErrorS(err, "failed to connect edge node", "nodeName", nodename, util.STREAM_TRACE_ID, ctx.Value(util.STREAM_TRACE_ID))
			respErr := util.InternalServerErrorMsg(proxyConn, err.Error(), ctx.Value(util.STREAM_TRACE_ID).(string))
			if respErr != nil {
				return respErr
			}
			return err
		}

	} else {

		//From tunnel-coredns, query the pods of tunnel-cloud where edge nodes establish long-term connections
		addr, ok := connect.Route.EdgeNode[nodename]

		//forward cloud node
		if !ok {
			_, cloudOk := connect.Route.CloudNode[nodename]
			if cloudOk {
				return DirectDial(host, port, category, proxyConn, ctx)
			}
		}

		//forward edge node
		/*
			todo Supports sending requests through nodes within nodeunit at the edge
		*/
		//You can only proxy once between tunnel-cloud pods
		if connect.IsEndpointIp(strings.Split(proxyConn.RemoteAddr().String(), ":")[0]) && !net.ParseIP(strings.Split(proxyConn.LocalAddr().String(), ":")[0]).IsLoopback() {
			klog.InfoS("loop forwarding", "remoteAddr", proxyConn.RemoteAddr().String(), "localAddr", proxyConn.LocalAddr().String(), util.STREAM_TRACE_ID, ctx.Value(util.STREAM_TRACE_ID))

			respErr := util.InternalServerErrorMsg(proxyConn, fmt.Sprintf("loop forwarding, remoteAddr:%s, localAddr:%s", proxyConn.RemoteAddr().String(), proxyConn.LocalAddr().String()), ctx.Value(util.STREAM_TRACE_ID).(string))
			if respErr != nil {
				return respErr
			}

			return fmt.Errorf("loop forwarding, remoteAddr:%s localAddr:%s, %s:%s", proxyConn.RemoteAddr().String(), proxyConn.LocalAddr().String(), util.STREAM_TRACE_ID, ctx.Value(util.STREAM_TRACE_ID))
		}
		remoteConn, err := GetRemoteConn(category, addr)
		if err != nil {
			klog.ErrorS(err, "failed to establish a connection between proxyServer and backendServer", util.STREAM_TRACE_ID, ctx.Value(util.STREAM_TRACE_ID))

			respErr := util.InternalServerErrorMsg(proxyConn, fmt.Sprintf("failed to establish a connection between proxyServer and backendServer, error: %v", err), ctx.Value(util.STREAM_TRACE_ID).(string))
			if respErr != nil {
				return respErr
			}

			return err
		}
		defer remoteConn.Close()

		//Forward HTTP_CONNECT request data
		remoteReq := &http.Request{
			Method: http.MethodConnect,
			URL:    &url.URL{Host: net.JoinHostPort(host, port)},
			Header: map[string][]string{
				util.STREAM_TRACE_ID: {ctx.Value(util.STREAM_TRACE_ID).(string)},
			},
		}
		proxyCtx, cancle := context.WithTimeout(ctx, 5*time.Second)
		defer cancle()
		remoteReq = remoteReq.WithContext(proxyCtx)
		err = remoteReq.Write(remoteConn)
		if err != nil {
			klog.ErrorS(err, "failed to write data to remoteConn", util.STREAM_TRACE_ID, ctx.Value(util.STREAM_TRACE_ID))

			respErr := util.InternalServerErrorMsg(proxyConn, fmt.Sprintf("failed to write data to remoteConn, error: %v", err), ctx.Value(util.STREAM_TRACE_ID).(string))
			if respErr != nil {
				return respErr
			}

			return err
		}

		// copy data and close conn
		return util.ConnCopyAndClose(remoteConn, proxyConn, ctx.Value(util.STREAM_TRACE_ID).(string))
	}

	return nil
}

func GetRemoteAddr(category, host string) string {
	switch category {
	case util.SSH:
		return fmt.Sprintf("%s:%d", host, conf.TunnelConf.TunnlMode.Cloud.SSH.SSHPort)
	case util.EGRESS, util.HTTP_PROXY:
		return fmt.Sprintf("%s:%d", host, conf.TunnelConf.TunnlMode.Cloud.HttpProxy.ProxyPort)
	}
	return host
}

func GetTargetType(nodeName string) TargetType {
	node := tunnelcontext.GetContext().GetNode(nodeName)
	if node != nil {
		return LocalPodType
	}

	if _, ok := connect.Route.EdgeNode[nodeName]; ok {
		return RemotePodType
	}
	return DisconnectNodeType
}

func GetRemoteConn(category, addr string) (net.Conn, error) {
	addr = GetRemoteAddr(category, addr)
	remoteConn, err := net.Dial(util.TCP, addr)
	return remoteConn, err
}

func DirectDial(host, port, category string, proxyConn net.Conn, ctx context.Context) error {
	//Handling access to out-of-cluster ip
	pingErr := util.Ping(host)
	if pingErr == nil {
		remoteConn, err := net.Dial("tcp", net.JoinHostPort(host, port))
		if err != nil {
			klog.ErrorS(err, "failed to establish tcp connection with server outside the cluster", util.STREAM_TRACE_ID, ctx.Value(util.STREAM_TRACE_ID))

			respErr := util.InternalServerErrorMsg(proxyConn, fmt.Sprintf("failed to establish tcp connection with server outside the cluster, error: %v", err), ctx.Value(util.STREAM_TRACE_ID).(string))
			if respErr != nil {
				return respErr
			}

			return err
		}
		defer remoteConn.Close()

		_, err = proxyConn.Write([]byte(util.ConnectMsg))
		if err != nil {
			klog.ErrorS(err, "failed to write data to proxyConn", util.STREAM_TRACE_ID, ctx.Value(util.STREAM_TRACE_ID))
			return err
		}

		// copy data and close conn
		return util.ConnCopyAndClose(remoteConn, proxyConn, ctx.Value(util.STREAM_TRACE_ID).(string))
	} else {
		klog.ErrorS(pingErr, "failed to get the node where the pod is located", "category", category, util.STREAM_TRACE_ID, ctx.Value(util.STREAM_TRACE_ID))

		respErr := util.InternalServerErrorMsg(proxyConn, fmt.Sprintf("failed to get the node where the pod is located, category: %s, error: %v", category, pingErr), ctx.Value(util.STREAM_TRACE_ID).(string))
		if respErr != nil {
			return respErr
		}
		return pingErr
	}
}
