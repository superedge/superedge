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
	"github.com/superedge/superedge/pkg/tunnel/proxy/common/indexers"
	"github.com/superedge/superedge/pkg/tunnel/proxy/modules/stream/streammng/connect"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"io"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/util/proxy"
	"k8s.io/klog/v2"
	"net"
	"strconv"
	"strings"
)

type TargetType int

const (
	LocalPodType  TargetType = 0 //transfer through the tunnel of this tunnel-cloud pod
	RemotePodType TargetType = 1 //transfer through the tunnel of other tunnel-cloud pod
	CloudNodeType TargetType = 2 //send requests directly in this tunnel-cloud pod
	EdgeNodeType  TargetType = 3 //the target node is on the edge and cannot send requests directly
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
					remoteConn, err = net.Dial(util.TCP, host+":"+port)
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

			remoteConn, err = net.Dial(util.TCP, addr)
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

func GetPodIpFromService(service string) (string, error) {
	/*
	   1. Directly forward to the node where the pod is located according to the podip(The received proxy request needs to be guaranteed to be in the form of podip, so as to avoid making another service-to-endpoint selection)
	   2. serviceName first checks whether it is in the format of serviceName.nameSpace
	   3. Support service types: ClusterIP, LoadBalancer, NodePort and externalName
	*/
	host, port, err := net.SplitHostPort(service)
	if err != nil {
		klog.Errorf("Failed to resolve host, error: %v", err)
		return "", err
	}
	podIp := net.ParseIP(host)
	if podIp == nil {
		services := strings.Split(host, ".")
		portInt32, err := strconv.ParseInt(port, 10, 32)
		if err != nil {
			klog.Errorf("Failed to resolve port, error: %v", err)
			return "", err
		}
		podUrl, err := proxy.ResolveEndpoint(indexers.ServiceLister, indexers.EndpointLister, services[1], services[0], int32(portInt32))
		if err != nil {
			klog.Errorf("Failed to get podIp from service, error: %v", err)
			return "", err
		}
		return podUrl.Hostname(), nil
	}
	return host, nil
}

func GetDomainFromHost(host string) (string, error) {
	services := strings.Split(host, ".")
	if len(services) > 2 {
		return host, nil
	}
	targetService, err := indexers.ServiceLister.Services(services[1]).Get(services[0])
	if err != nil {
		if apierrors.IsNotFound(err) {
			return host, nil
		}
		klog.Errorf("Failed to get service %s from cluster, error: %v", host, err)
		return "", err
	}
	if targetService.Spec.Type == v1.ServiceTypeExternalName {
		return targetService.Spec.ExternalName, nil
	}
	return "", nil
}

func GetTargetType(nodeName string) TargetType {
	node := context.GetContext().GetNode(nodeName)
	if node != nil {
		return LocalPodType
	}

	_, err := net.LookupHost(nodeName)
	if err == nil {
		return RemotePodType
	}

	if dnsErr, ok := err.(*net.DNSError); ok {
		if dnsErr.IsNotFound {
			/*
				todo 需要判断节节点点是否为没有建立云边隧道的边缘节点
			*/

			return CloudNodeType
		}
	}
	return LocalPodType
}

func GetRemoteProxyServerPort(category string) string {
	switch category {
	case util.SSH:
		return "22"
	case util.EGRESS:
		return conf.TunnelConf.TunnlMode.Cloud.Egress.EgressPort
	case util.HTTP_PROXY:
		return conf.TunnelConf.TunnlMode.Cloud.HttpProxy.ProxyPort
	}
	return "10250"
}

func GetRemoteConn(nodeName, category string) (net.Conn, error) {
	addrs, err := net.LookupHost(nodeName)
	if err != nil {
		return nil, err
	}
	remoteConn, err := net.Dial("tcp", addrs[0]+":"+GetRemoteProxyServerPort(category))
	if err != nil {
		klog.Errorf("Failed to establish a connection between proxyServer and backendServer, error: %v", err)
		return nil, err
	}
	return remoteConn, nil
}
