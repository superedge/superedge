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

package handlers

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/superedge/superedge/pkg/tunnel/proto"
	"github.com/superedge/superedge/pkg/tunnel/proxy/common"
	"github.com/superedge/superedge/pkg/tunnel/proxy/modules/stream/streammng/connect"
	"github.com/superedge/superedge/pkg/tunnel/tunnelcontext"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"k8s.io/klog/v2"
)

func AccessHandler(msg *proto.StreamMsg) error {

	/*
			1. Forward to remote proxyServer
		 	2. Forward to cloud pod
	*/
	klog.V(2).InfoS("receive access msg", "nodeName", msg.GetNode(),
		"category", msg.GetCategory(), "type", msg.GetType(), util.STREAM_TRACE_ID, msg.GetTopic())
	// return the message that httpConnect connection establishment failed
	errMsg := func(node tunnelcontext.Node, respErr error) {
		if node != nil {
			node.Send2Node(&proto.StreamMsg{
				Node:     msg.GetNode(),
				Category: msg.GetCategory(),
				Type:     tunnelcontext.CONNECT_FAILED,
				Topic:    msg.GetTopic(),
				Data:     []byte(respErr.Error()),
			})
		}

	}

	// returns the message that the httpConnect connection was established successfully
	successMsg := func(node tunnelcontext.Node) {
		if node != nil {
			node.Send2Node(&proto.StreamMsg{
				Node:     msg.GetNode(),
				Category: msg.Category,
				Type:     tunnelcontext.CONNECT_SUCCESSED,
				Topic:    msg.GetTopic(),
				Data:     []byte(util.ConnectMsg),
			})
		}

	}

	localNode := tunnelcontext.GetContext().GetNode(msg.GetNode())
	if localNode == nil {
		nodeErr := fmt.Errorf("the  tunnel of node %s is broken", msg.GetNode())
		klog.ErrorS(nodeErr, "the edge node sending the request is disconnected from the cloud", util.STREAM_TRACE_ID, msg.GetTopic())
		return nodeErr
	}
	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(msg.Data)))
	if err != nil {
		klog.ErrorS(err, "failed to parse httpRequest", util.STREAM_TRACE_ID, msg.GetTopic())
		errMsg(localNode, fmt.Errorf("failed to parse httpRequest, error:%v,rawRequest:%s", err, msg.GetData()))
		return err
	}
	req = req.WithContext(context.WithValue(req.Context(), util.STREAM_TRACE_ID, msg.GetTopic()))
	if req.Method != http.MethodConnect {
		klog.InfoS("Does not handle non-httpConnect requests", util.STREAM_TRACE_ID, msg.GetTopic())
		errMsg(localNode, fmt.Errorf("does not handle non-httpConnect requests"))
		return err
	}

	directDial := func(ip, port, nodeName string) error {
		node := tunnelcontext.GetContext().GetNode(nodeName)
		if node == nil {
			nodeErr := fmt.Errorf("the  tunnel of node %s is broken", nodeName)
			klog.ErrorS(nodeErr, "the edge node sending the request is disconnected from the cloud", util.STREAM_TRACE_ID, msg.GetTopic())
			return nodeErr
		}
		targetServer := net.JoinHostPort(ip, port)
		remoteConn, err := net.Dial("tcp", targetServer)
		if err != nil {
			klog.ErrorS(err, "tunnel-cloud failed to establish a connection with the target server", "targetServer", targetServer, util.STREAM_TRACE_ID, msg.GetTopic())
			errMsg(node, fmt.Errorf("tunnel-cloud failed to establish a connection with the target server %s, error:%v", targetServer, err))
			return err
		}

		// Return 200 status code
		successMsg(node)
		remoteCh := tunnelcontext.GetContext().AddConn(msg.GetTopic())
		node.BindNode(msg.GetTopic())
		go common.Read(remoteConn, node, msg.Category, util.TCP_FORWARD, msg.GetTopic())
		go common.Write(remoteConn, remoteCh)
		return nil
	}

	if os.Getenv(util.EdgeNoProxy) != "" {
		if !util.NewHttpProxyConfig(os.Getenv(util.EdgeNoProxy)).UseProxy(msg.Addr) {
			return fmt.Errorf("edge nodes are prohibited from accessing the addr %s, %s:%s", msg.Addr, util.STREAM_TRACE_ID, msg.GetTopic())
		}
	}

	host, port, err := net.SplitHostPort(msg.Addr)
	if err != nil {
		klog.ErrorS(err, "failed to resolve host", util.STREAM_TRACE_ID, msg.GetTopic())
		errMsg(localNode, fmt.Errorf("failed to resolve host, error: %v", err))
		return err
	}

	info, directDialFlag, err := getForwardInfo(host, port)
	if err != nil {
		errMsg(localNode, fmt.Errorf("failed to get forwarding info, error:%v", err))
		return err
	}
	if directDialFlag {
		if info != nil {
			host = info.podIp
			port = info.port
		}
		return directDial(host, port, msg.GetNode())
	}

	switch common.GetTargetType(info.nodeName) {
	case common.LocalPodType:
		remoteNode := tunnelcontext.GetContext().GetNode(info.nodeName)
		if remoteNode != nil {
			remoteNode.AddPairNode(msg.GetTopic(), localNode.GetName())
			localNode.AddPairNode(msg.GetTopic(), remoteNode.GetName())
			msg.Addr = net.JoinHostPort(info.podIp, info.port)
			msg.Node = remoteNode.GetName()
			remoteNode.Send2Node(msg)
		} else {
			remoteErr := fmt.Errorf("the remote edge node %s is not connected to the cloud", info.nodeName)
			klog.ErrorS(remoteErr, "the remote node sending the request is disconnected from the cloud", util.STREAM_TRACE_ID, msg.GetTopic())
			return remoteErr
		}

	case common.RemotePodType:
		// Establish a connection with the remote proxyServer
		if remoteIp, ok := connect.Route.EdgeNode[info.nodeName]; ok {
			remoteConn, err := common.GetRemoteConn(msg.GetCategory(), remoteIp)
			if err != nil {
				errMsg(localNode, err)
				return err
			}

			remoteReq := &http.Request{
				Method: http.MethodConnect,
				URL:    &url.URL{Host: net.JoinHostPort(info.podIp, info.port)},
				Header: map[string][]string{
					util.STREAM_TRACE_ID: {msg.GetTopic()},
				},
			}

			err = remoteReq.Write(remoteConn)
			if err != nil {
				errMsg(localNode, err)
				return err
			}
			resp, respBuffer, err := util.GetRespFromConn(remoteConn, nil)
			if err != nil {
				errMsg(localNode, err)
				return err
			}
			if resp.StatusCode != http.StatusOK {
				klog.InfoS("failed to connect edge node", "nodeName", info.nodeName, "response", string(respBuffer.Bytes()), util.STREAM_TRACE_ID, msg.GetTopic())
				respErr := fmt.Errorf("failed to connect edge node %s, error:%v", info.nodeName, string(respBuffer.Bytes()))
				errMsg(localNode, respErr)
				return respErr
			}
			// Return 200 status code
			successMsg(localNode)
			remoteCh := tunnelcontext.GetContext().AddConn(msg.GetTopic())
			localNode.BindNode(msg.GetTopic())
			go common.Read(remoteConn, localNode, msg.GetCategory(), util.TCP_FORWARD, msg.GetTopic())
			go common.Write(remoteConn, remoteCh)
		} else {
			klog.InfoS("the edge node  disconnected", "nodeName", info.nodeName, util.STREAM_TRACE_ID, msg.GetTopic())
			nodeErr := fmt.Errorf("the edge node %s disconnected", info.nodeName)
			errMsg(localNode, nodeErr)
			return nodeErr
		}
	case common.DisconnectNodeType:
		klog.InfoS("the target node is not registered in the route cache", "nodeName", info.nodeName, util.STREAM_TRACE_ID, msg.GetTopic())
		nodeErr := fmt.Errorf("the target node %s is not registered in the route cache", info.nodeName)
		errMsg(localNode, nodeErr)
		return nodeErr
	}
	return nil
}
