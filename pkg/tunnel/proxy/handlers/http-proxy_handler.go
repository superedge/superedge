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
	"fmt"
	"github.com/superedge/superedge/pkg/tunnel/context"
	"github.com/superedge/superedge/pkg/tunnel/proto"
	"github.com/superedge/superedge/pkg/tunnel/proxy/common"
	"github.com/superedge/superedge/pkg/tunnel/proxy/common/indexers"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"net/http"
	"os"

	"k8s.io/klog/v2"
	"net"
)

func AccessHandler(msg *proto.StreamMsg) error {

	/*
			1. Forward to remote proxyServer
		 	2. Forward to cloud pod
	*/
	connCh := context.GetContext().GetConn(msg.GetTopic())
	if connCh != nil {
		connCh.Send2Conn(msg)
		return nil
	}

	//return the message that httpConnect connection establishment failed
	errMsg := func(node context.Node) {
		if node != nil {
			node.Send2Node(&proto.StreamMsg{
				Node:     msg.GetNode(),
				Category: util.HTTP_PROXY,
				Type:     util.TCP_BACKEND,
				Topic:    msg.GetTopic(),
				Data:     []byte(util.InternalServerError),
			})
		}

	}

	//returns the message that the httpConnect connection was established successfully
	successMsg := func(node context.Node) {
		if node != nil {
			node.Send2Node(&proto.StreamMsg{
				Node:     msg.GetNode(),
				Category: util.HTTP_PROXY,
				Type:     util.TCP_BACKEND,
				Topic:    msg.GetTopic(),
				Data:     []byte(util.ConnectMsg),
			})
		}

	}

	localNode := context.GetContext().GetNode(msg.GetNode())
	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(msg.Data)))
	if err != nil {
		klog.Errorf("Failed to parse httpRequest, error: %v", err)
		errMsg(localNode)
		return err
	}
	if req.Method != http.MethodConnect {
		klog.Errorf("Does not handle non-httpConnect requests, error: %v", err)
		errMsg(localNode)
		return err
	}
	//与集群外的server建立连接
	connectOutcluster := func(outClusterHost, outClusterPort, nodeName string) error {

		node := context.GetContext().GetNode(nodeName)
		if node == nil {
			return fmt.Errorf("the node's tunnel is broken, nodeName: %s", nodeName)
		}

		remoteConn, err := net.Dial("tcp", net.JoinHostPort(outClusterHost, outClusterPort))
		if err != nil {
			klog.Errorf("Failed to establish connection with cloud pod, error: %v", err)
			errMsg(node)
			return err
		}

		//Return 200 status code
		successMsg(node)
		remoteCh := context.GetContext().AddConn(msg.GetTopic())
		node.BindNode(msg.GetTopic())
		go common.Read(remoteConn, node, util.HTTP_PROXY, util.TCP_BACKEND, msg.GetTopic(), msg.GetAddr())
		go common.Write(remoteConn, remoteCh)
		return nil
	}

	if os.Getenv(util.EdgeNoProxy) != "" {
		if !util.NewHttpProxyConfig(os.Getenv(util.EdgeNoProxy)).UseProxy(msg.Addr) {
			return fmt.Errorf("edge nodes are prohibited from accessing the host %s", msg.Addr)
		}
	}

	host, port, err := net.SplitHostPort(msg.Addr)
	if err != nil {
		klog.Errorf("Failed to resolve host, error: %v", err)
		errMsg(localNode)
		return err
	}
	var podIP string
	if net.ParseIP(host) != nil {
		podIP = host
	} else {
		//Handling access to domain names outside the cluster
		domain, err := common.GetDomainFromHost(host)
		if err != nil {
			errMsg(localNode)
			return err
		}
		if domain != "" {
			return connectOutcluster(domain, port, msg.GetNode())
		}
		//Handling access to services in the cluster
		podIP, port, err = common.GetPodIpFromService(req.Host)
		if err != nil {
			klog.Errorf("Failed to get podIp through service, error: %v", err)
			errMsg(localNode)
			return err
		}
	}
	nodeName, err := indexers.GetNodeByPodIP(podIP)
	if err != nil {
		//Handle access to ip outside the cluster
		pingErr := util.Ping(podIP)
		if pingErr == nil {
			return connectOutcluster(podIP, port, msg.GetNode())
		}
		klog.Errorf("Error in ping ip %s outside the cluster, error: %v", podIP, err)
		errMsg(localNode)
		klog.Errorf("Failed to get the node where the pod is located from podIp %s, error: %v", podIP, err)
		return err
	}

	switch common.GetTargetType(nodeName) {
	case common.LocalPodType:
		//Return 200 status code
		successMsg(localNode)

		remoteNode := context.GetContext().GetNode(nodeName)
		remoteNode.AddPairNode(msg.GetTopic(), localNode.GetName())
		localNode.AddPairNode(msg.GetTopic(), remoteNode.GetName())
		remoteNode.Send2Node(&proto.StreamMsg{
			Node:     nodeName,
			Category: util.HTTP_PROXY,
			Type:     util.TCP_FRONTEND,
			Topic:    msg.GetTopic(),
			Data:     []byte{},
			Addr:     net.JoinHostPort(podIP, port),
		})
	case common.RemotePodType:
		//Establish a connection with the remote proxyServer
		getRemoteConn := func() (net.Conn, error) {
			remoteConn, err := common.GetRemoteConn(nodeName, msg.GetCategory())
			if err != nil {
				klog.Errorf("Failed to establish connection with remote server, error: %v", err)
				return nil, err
			}

			//Send httpConnect request to remote proxyServer
			_, err = fmt.Fprintf(remoteConn, fmt.Sprintf("CONNECT %s HTTP/1.1\r\n\r\n\r\n", net.JoinHostPort(podIP, port)))
			if err != nil {
				klog.Errorf("Failed to send httpConnect request to remote proxyServer, error: %v", err)
				return nil, err
			}

			return remoteConn, nil
		}

		remoteConn, err := getRemoteConn()
		if err != nil {
			errMsg(localNode)
			return err
		}

		//Return 200 status code
		successMsg(localNode)

		remoteCh := context.GetContext().AddConn(msg.GetTopic())
		localNode.BindNode(msg.GetTopic())
		go common.Read(remoteConn, localNode, util.HTTP_PROXY, util.TCP_BACKEND, msg.GetTopic(), msg.GetAddr())
		go common.Write(remoteConn, remoteCh)

	//CloudNodeType and EdgeNodeType are just to break down the reasons why they are not accessible
	case common.CloudNodeType:
		return connectOutcluster(podIP, port, msg.GetNode())
	case common.EdgeNodeType:
		klog.Errorf("The tunnel connection of the edge node %s is disconnected", nodeName)
		errMsg(localNode)
		return fmt.Errorf("The tunnel connection of the edge node is disconnected")
	}
	return nil
}
