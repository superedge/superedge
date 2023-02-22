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
	"encoding/base64"
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
	"net/url"
	"os"
	"strings"
)

func HandleServerConn(proxyConn net.Conn, category string, noAccess func(host string) error) {
	defer proxyConn.Close()
	req, raw, err := util.GetRequestFromConn(proxyConn)
	if err != nil {
		klog.V(8).Infof("Failed to get http request, error: %v", err)
		return
	}
	if os.Getenv(util.PROXY_AUTHORIZATION_ENV) == "true" {
		if category == util.HTTP_PROXY {
			proxyAuth := req.Header.Get("Proxy-Authorization")
			if proxyAuth == "" {
				util.WriteMsg(proxyConn, util.Unauthorized)
				return
			}
			auths := strings.Split(proxyAuth, " ")
			if len(auths) != 2 && auths[0] != "Basic" {
				util.WriteMsg(proxyConn, util.Forbidden)
				return
			}
			infos, err := base64.StdEncoding.DecodeString(auths[1])
			if err != nil {
				klog.Error(err)
				util.WriteMsg(proxyConn, util.Forbidden)
				return
			}
			userinfos := strings.Split(string(infos), ":")
			_, err = os.Stat(util.AuthorizationPath + "/" + userinfos[0])
			if err == nil {
				pwd, err := os.ReadFile(util.AuthorizationPath + "/" + userinfos[0])
				if err != nil {
					klog.Error(err)
					util.WriteMsg(proxyConn, util.Forbidden)
					return
				}
				if string(pwd) != userinfos[1] {
					klog.Errorf("Incorrect password, username: %s", userinfos[0])
					util.WriteMsg(proxyConn, util.Forbidden)
					return
				}
			} else {
				klog.Errorf("Username does not exist")
				util.WriteMsg(proxyConn, util.Forbidden)
				return
			}
		}
	}
	if req.Method != http.MethodConnect {
		if category != util.HTTP_PROXY {
			return
		}
		localConn, err := net.Dial("tcp", common.GetRemoteAddr("127.0.0.1", category))
		if err != nil {
			klog.Errorf("Failed to forward http request through localhost [Dial], error:%v", err)
			return
		}

		proxyReq := &http.Request{
			Method: http.MethodConnect,
			URL:    &url.URL{Host: req.Host},
		}

		err = proxyReq.Write(localConn)
		if err != nil {
			klog.Errorf("Failed to forward http request through localhost [Write Request], error:%v", err)
			return
		}
		r := bufio.NewReader(localConn)
		resp, err := http.ReadResponse(r, proxyReq)
		if err != nil {
			klog.Errorf("Failed to forward http request through localhost [Response], error:%v", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			_, err = localConn.Write(raw.Bytes())
			if err != nil {
				klog.Errorf("Failed to forward http request through localhost [Write Origin Request], error:%v", err)
				return
			}
			go func() {
				_, writeErr := io.Copy(localConn, proxyConn)
				if writeErr != nil {
					klog.Errorf("Failed to copy data to remoteConn, error: %v", writeErr)
					return
				}
			}()
			_, err = io.Copy(proxyConn, localConn)
			if err != nil {
				klog.Errorf("Failed to read data from remoteConn, error: %v", err)
				return
			}
		}
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
				err = common.DailDirect(v, port, category, proxyConn)
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
				err = common.DailDirect(host, port, category, proxyConn)
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
				err = common.DailDirect(host, port, category, proxyConn)
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
		err = common.DailDirect(host, port, category, proxyConn)
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
					err = common.DailDirect(host, port, category, proxyConn)
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
					err = common.DailDirect(host, port, category, proxyConn)
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
			err = common.DailDirect(host, port, category, proxyConn)
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
