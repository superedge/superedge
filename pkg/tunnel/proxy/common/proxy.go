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
	"bytes"
	uuid "github.com/satori/go.uuid"
	"github.com/superedge/superedge/pkg/tunnel/conf"
	"github.com/superedge/superedge/pkg/tunnel/context"
	"github.com/superedge/superedge/pkg/tunnel/proxy/modules/stream/streammng/connect"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"io"
	"k8s.io/klog/v2"
	"net"
	"strings"
)

func ProxyEdgeNode(nodename, host, port, category string, proxyConn net.Conn, req *bytes.Buffer) {
	node := context.GetContext().GetNode(nodename)
	if node != nil {
		//If the edge node establishes a long connection with this pod, it will be forwarded directly
		uid := uuid.NewV4().String()
		ch := context.GetContext().AddConn(uid)
		node.BindNode(uid)
		_, err := proxyConn.Write([]byte(util.ConnectMsg))
		if err != nil {
			klog.Errorf("Failed to write data to proxyConn, error: %v", err)
			return
		}
		go Read(proxyConn, node, category, util.TCP_FRONTEND, uid, host+":"+port)
		Write(proxyConn, ch)
	} else {
		//From tunnel-coredns, query the pods of tunnel-cloud where edge nodes establish long-term connections
		var remoteConn net.Conn
		addrs, err := net.LookupHost(nodename)
		if err != nil {
			if dnsErr, ok := err.(*net.DNSError); ok {
				if dnsErr.IsNotFound {
					remoteConn, err = net.Dial("tcp", host+":"+port)
					if err != nil {
						klog.Errorf("Failed to send request from tunnel-cloud, error: %v", err)
						return
					}

					//Return 200 status code
					_, err = proxyConn.Write([]byte(util.ConnectMsg))
					if err != nil {
						klog.Errorf("Failed to write data to proxyConn, error: %v", err)
						return
					}
				}
			}
			if remoteConn == nil {
				klog.Errorf("DNS parsing error: %v", err)
				_, err = proxyConn.Write([]byte(util.InternalServerError))
				if err != nil {
					klog.Errorf("Failed to write data to proxyConn, error: %v", err)
				}
				return
			}
		} else {
			/*
				todo Supports sending requests through nodes within nodeunit at the edge
			*/

			//You can only proxy once between tunnel-cloud pods
			if connect.IsEndpointIp(strings.Split(proxyConn.RemoteAddr().String(), ":")[0]) {
				klog.Errorf("Loop forwarding")
				return
			}

			var addr string
			if category == util.EGRESS {
				addr = addrs[0] + ":" + conf.TunnelConf.TunnlMode.Cloud.Egress.EgressPort
			} else if category == util.SSH {
				addr = addrs[0] + ":22"
			}

			remoteConn, err = net.Dial("tcp", addr)
			if err != nil {
				klog.Errorf("Failed to establish a connection between proxyServer and backendServer, error: %v", err)
				return
			}

			//Forward HTTP_CONNECT request data
			_, err = remoteConn.Write(req.Bytes())
			if err != nil {
				klog.Errorf("Failed to write data to remoteConn, error: %v", err)
				return
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
		}
	}
}
