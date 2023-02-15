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
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/superedge/superedge/pkg/tunnel/context"
	"github.com/superedge/superedge/pkg/tunnel/proto"
	"github.com/superedge/superedge/pkg/tunnel/proxy/common"
	"github.com/superedge/superedge/pkg/tunnel/proxy/common/indexers"
	"github.com/superedge/superedge/pkg/tunnel/proxy/modules/stream/streammng/connect"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"io"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"net"
	"net/http"
	"strings"
)

func HttpProxyCloudServer(proxyConn net.Conn) {
	req, raw, err := util.GetRequestFromConn(proxyConn)
	if err != nil {
		klog.V(8).Infof("Failed to read httpRequest, error: %v", err)
		return
	}

	errMsg := func(conn net.Conn) {
		_, err := proxyConn.Write([]byte(util.InternalServerError))
		if err != nil {
			klog.Errorf("Failed to write data to proxyConn, error: %v", err)

		}
	}

	successMsg := func(conn net.Conn) error {
		_, err := proxyConn.Write([]byte(util.ConnectMsg))
		if err != nil {
			klog.Errorf("Failed to write data to proxyConn, error: %v", err)
			return err
		}
		return nil
	}

	host, port, err := net.SplitHostPort(req.Host)
	if err != nil {
		klog.Errorf("Failed to resolve host, error: %v", err)
		errMsg(proxyConn)
		return
	}

	var podIp, nodeName string
	if net.ParseIP(host) != nil {
		node, err := indexers.GetNodeByPodIP(host)
		if err != nil {
			klog.Errorf("Only process the request that the destination ip forwarded by the destination is podIp, error: %v", err)
			return
		}
		podIp = host
		nodeName = node
	} else {
		internalIp, err := indexers.GetNodeIPByName(host)
		if err != nil {
			if apierrors.IsNotFound(err) {
				err = dailOutCluster(host, port, proxyConn)
				if err != nil {
					klog.Errorf("Failed to forward request, host: %s, error: %v", host, err)
					return
				}
			} else {
				klog.Errorf("Failed to get internalIp of node, error: %v", err)
				common.ProxyEdgeNode(host, "127.0.0.1", port, util.EGRESS, proxyConn, raw)
			}
		} else {
			common.ProxyEdgeNode(host, internalIp, port, util.EGRESS, proxyConn, raw)
		}
		getNodeName := func(service string) (string, error) {
			podIp, port, err = common.GetPodIpFromService(service)
			if err != nil {
				klog.Errorf("Failed to get podIp, error: %v", err)
				return "", err
			}

			//Only handle access within the cluster
			nodeName, err := indexers.GetNodeByPodIP(podIp)
			if err != nil {
				klog.Errorf("Failed to get the node name where the pod is located, error: %v", err)
				return "", err
			}
			return nodeName, nil
		}
		nodeName, err = getNodeName(req.Host)
		if err != nil {
			errMsg(proxyConn)
			return
		}
	}

	switch common.GetTargetType(nodeName) {
	case common.LocalPodType:
		uid := uuid.NewV4().String()
		ch := context.GetContext().AddConn(uid)
		remoteNode := context.GetContext().GetNode(nodeName)
		remoteNode.BindNode(uid)
		if req.Method == http.MethodConnect {
			err := successMsg(proxyConn)
			if err != nil {
				klog.Errorf("Return httpConnect connection establishment failed, error: %v", err)
				return
			}
		} else {
			remoteNode.Send2Node(&proto.StreamMsg{
				Node:     nodeName,
				Category: util.HTTP_PROXY,
				Type:     util.TCP_FRONTEND,
				Topic:    uid,
				Data:     raw.Bytes(),
				Addr:     podIp + ":" + port,
			})
		}
		go common.Read(proxyConn, remoteNode, util.HTTP_PROXY, util.TCP_FRONTEND, uid, podIp+":"+port)
		common.Write(proxyConn, ch)
	case common.RemotePodType:
		//获取与远端proxyServer的连接
		getRemoteConn := func() (net.Conn, error) {
			if connect.IsEndpointIp(strings.Split(proxyConn.RemoteAddr().String(), ":")[0]) {
				klog.Errorf("Only one request can be forwarded")
				return nil, fmt.Errorf("Only one request can be forwarded")
			}
			remoteConn, err := common.GetRemoteConn(nodeName, util.HTTP_PROXY)
			if err != nil {
				klog.Errorf("Failed to establish connection with proxyServer of next hop, error: %v", err)
				return nil, err
			}
			_, err = fmt.Fprintf(remoteConn, fmt.Sprintf("CONNECT %s HTTP/1.1\r\n\r\n\r\n", podIp+":"+port))
			if err != nil {
				klog.Errorf("Failed to write httpConnect request to remote proxyServer, error: %v", err)
				return nil, err
			}
			return remoteConn, nil
		}

		remoteConn, err := getRemoteConn()
		if err != nil {
			errMsg(proxyConn)
			return
		}

		if req.Method != http.MethodConnect {
			_, err = remoteConn.Write(raw.Bytes())
			if err != nil {
				klog.Errorf("Failed to forward request to remote proxyServer, error: %v", err)
				errMsg(proxyConn)
				return
			}
		}

		go func() {
			_, err = io.Copy(proxyConn, remoteConn)
			if err != nil {
				klog.Errorf("Failed to read data from proxyServer in cloud, error: %v", err)
			}
		}()
		_, err = io.Copy(remoteConn, proxyConn)
		if err != nil {
			klog.Errorf("Failed to write data to remote proxyServer, error: %v", err)
		}
	case common.CloudNodeType:
		proxyConn.Write([]byte("Do not forward requests to the cloud server"))
		proxyConn.Close()
	case common.EdgeNodeType:
		proxyConn.Write([]byte(fmt.Sprintf("The tunnel connection of the edge node %s is disconnected", nodeName)))
		proxyConn.Close()
	}

}

func dailOutCluster(host, port string, proxyConn net.Conn) error {
	//Handling access to out-of-cluster ip
	pingErr := util.Ping(host)
	if pingErr == nil {
		remoteConn, err := net.Dial("tcp", net.JoinHostPort(host, port))
		if err != nil {
			klog.Errorf("Failed to establish tcp connection with server outside the cluster, error: %v", err)
			return err
		}
		_, err = proxyConn.Write([]byte(util.ConnectMsg))
		if err != nil {
			if err != nil {
				klog.Errorf("Failed to write data to proxyConn, error: %v", err)
				return err
			}
		}
		defer remoteConn.Close()
		go func() {
			_, writeErr := io.Copy(remoteConn, proxyConn)
			if writeErr != nil {
				klog.Errorf("Failed to copy data to remoteConn, error: %v", writeErr)
			}
		}()
		_, err = io.Copy(proxyConn, remoteConn)
		if err != nil {
			klog.Errorf("Failed to read data from remoteConn, error: %v", err)
			return err
		}

	} else {
		klog.Errorf("Failed to get the node where the pod is located, module: %s, error: %v", util.EGRESS, pingErr)
		return pingErr
	}
	return nil
}
