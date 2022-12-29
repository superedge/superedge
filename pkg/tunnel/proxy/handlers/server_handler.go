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
	"fmt"
	"github.com/superedge/superedge/pkg/tunnel/proxy/common"
	"github.com/superedge/superedge/pkg/tunnel/proxy/common/indexers"
	"github.com/superedge/superedge/pkg/tunnel/proxy/modules/stream/streammng/connect"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"io"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"net"
	"net/http"
)

func HandleServerConn(proxyConn net.Conn, category string, noAccess func(host string) error) {
	defer proxyConn.Close()
	req, raw, err := util.GetRequestFromConn(proxyConn)
	if err != nil {
		klog.V(8).Infof("Failed to get http request, error: %v", err)
		return
	}
	if req.Method != http.MethodConnect {
		klog.V(8).Infof("Only HTTP CONNECT requests are supported, request = %v", req)
		return
	}

	if noAccess != nil {
		err = noAccess(req.Host)
		if err != nil {
			klog.Errorf("host %s is forbidden to access", req.Host)
			return
		}
	}

	host, port, err := net.SplitHostPort(req.Host)
	if err != nil {
		klog.Errorf("Failed to get host and port, module: %s, error: %v", util.EGRESS, err)
		err = util.WriteMsg(proxyConn, util.InternalServerError)
		klog.Errorf("Error writing return message, error: %v", err)
		return
	}
	ip := net.ParseIP(host)
	if ip == nil {
		/*
		 1. nodeName(kubectl logs/exec; ssh)
		 2. serviceName()
		 3. domain
		*/

		// nodeName
		node, err := indexers.NodeLister.Get(host)
		if err == nil {
			//cloud
			if v, ok := connect.Route.CloudNode[node.Name]; ok {
				err = dailDirect(v, port, category, proxyConn)
				if err != nil {
					klog.Errorf("Failed to forward the request of the user-defined cloud service, error:%v", node.Name, err)
				}
				return
			} else {
				if _, ok := connect.Route.EdgeNode[node.Name]; ok {
					var interIp string
					for _, addr := range node.Status.Addresses {
						if addr.Type == "InternalIP" {
							interIp = addr.Address
						}
					}
					err = common.ProxyEdgeNode(node.Name, interIp, port, category, proxyConn, raw)
					if err != nil {
						klog.Errorf("The request forwarded to the edge node %s failed, error:%v", node.Name, err)
					}
				}
				return
			}
		}

		//serviceName
		if v, ok := connect.Route.UserServicesMap[host]; ok {
			if v == util.CLOUD {
				err = dailDirect(host, port, category, proxyConn)
				if err != nil {
					klog.Errorf("Failed to forward user-defined cloud service %s request, error:%v", host, err)
				}
				return
			}
			if v == util.EDGE {
				host, port, nodeName, err := getNodeName(req.Host)
				if err != nil {
					klog.Errorf("Failed to obtain the backend pod instance through the user-defined edge service %s, error:%v", req.Host, err)
					return
				}
				err = common.ProxyEdgeNode(nodeName, host, port, category, proxyConn, raw)
				if err != nil {
					klog.Errorf("Failed to forward user-defined cloud service %s requestt, error:%v", host, err)
				}
				return
			}
		}

		if v, ok := connect.Route.ServicesMap[host]; ok {
			if v == util.CLOUD {
				err = dailDirect(host, port, category, proxyConn)
				if err != nil {
					klog.Errorf("Failed to forward  cloud service %s request, error:%v", host, err)
					return
				}
			}
			if v == util.EDGE {
				host, port, nodeName, err := getNodeName(req.Host)
				if err != nil {
					klog.Errorf("Failed to obtain the backend pod instance through the edge service %s, error:%v", req.Host, err)
					return
				}
				common.ProxyEdgeNode(nodeName, host, port, category, proxyConn, raw)
				return
			}
		}

		//domain
		err = dailDirect(host, port, category, proxyConn)
		if err != nil {
			klog.Errorf("Forward request failed, host:%s, error:%v", host, err)
			return
		}

	} else {
		/*
		  1. clusterIp
		  2. podIp
		*/

		//clusterIp
		svc, err := indexers.GetServiceByClusterIP(host)
		if err != nil && !apierrors.IsNotFound(err) {
			klog.Errorf("Failed to get servcie by clusterip, error: %v", err)
		}

		if svc != nil {
			// cloud service
			if v, ok := connect.Route.UserServicesMap[fmt.Sprintf("%s.%s", svc.Name, svc.Namespace)]; ok {
				if v == util.CLOUD {
					err = dailDirect(host, port, category, proxyConn)
					if err != nil {
						klog.Errorf("Failed to forward user-defined cloud service %s request, error:%v", host, err)
						return
					}
				}
				if v == util.EDGE {
					host, port, _, err = getNodeName(fmt.Sprintf("%s.%s:%s", svc.Name, svc.Namespace, port))
					if err != nil {
						klog.Errorf("Failed to obtain the backend pod instance through the user-defined edge service %s, error:%v", req.Host, err)
						return
					}
				}
			}

			if v, ok := connect.Route.ServicesMap[fmt.Sprintf("%s.%s", svc.Name, svc.Namespace)]; ok {
				if v == util.CLOUD {
					err = dailDirect(host, port, category, proxyConn)
					if err != nil {
						klog.Errorf("Failed to forward  cloud service %s request, error:%v", host, err)
						return
					}
				}
				if v == util.EDGE {
					host, port, _, err = getNodeName(fmt.Sprintf("%s.%s:%s", svc.Name, svc.Namespace, port))
					if err != nil {
						klog.Errorf("Failed to obtain the backend pod instance through the edge service %s, error:%v", req.Host, err)
						return
					}
				}
			}

		}

		//Request pods on edge nodes
		node, err := indexers.GetNodeByPodIP(host)
		if err != nil {
			//Handling access to out-of-cluster ip
			err = dailDirect(host, port, category, proxyConn)
			if err != nil {
				klog.Errorf("Failed to forward IP %s requests outside the cluster, error: %v", host, err)
			}
			return
		}

		common.ProxyEdgeNode(node, host, port, category, proxyConn, raw)
	}

}

func getNodeName(service string) (podIp, port, nodeName string, err error) {
	podIp, port, err = common.GetPodIpFromService(service)
	if err != nil {
		klog.Errorf("Failed to get podIp, error: %v", err)
		return
	}

	//Only handle access within the cluster
	nodeName, err = indexers.GetNodeByPodIP(podIp)
	if err != nil {
		klog.Errorf("Failed to get the node name where the pod is located, error: %v", err)
		return
	}
	return
}

func dailDirect(host, port, category string, proxyConn net.Conn) error {
	//Handling access to out-of-cluster ip
	pingErr := util.Ping(host)
	if pingErr == nil {
		remoteConn, err := net.Dial("tcp", net.JoinHostPort(host, port))
		if err != nil {
			klog.Errorf("Failed to establish tcp connection with server outside the cluster, error: %v", err)
			return err
		}
		err = util.WriteMsg(proxyConn, util.ConnectMsg)
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
		klog.Errorf("Failed to get the node where the pod is located, module: %s, error: %v", category, pingErr)
		return pingErr
	}
	return nil
}
