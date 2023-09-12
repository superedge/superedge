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
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/superedge/superedge/pkg/tunnel/proto"
	"github.com/superedge/superedge/pkg/tunnel/proxy/common"
	"github.com/superedge/superedge/pkg/tunnel/proxy/common/indexers"
	"github.com/superedge/superedge/pkg/tunnel/proxy/modules/stream/streammng/connect"
	"github.com/superedge/superedge/pkg/tunnel/tunnelcontext"
	"github.com/superedge/superedge/pkg/tunnel/util"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/util/proxy"
	"k8s.io/klog/v2"
)

type forwardInfo struct {
	podIp    string
	port     string
	nodeName string
}

func HandleServerConn(proxyConn net.Conn, category string, noAccess func(host string) error) error {
	defer proxyConn.Close()
	req, reqraw, err := util.GetRequestFromConn(proxyConn)
	if err != nil {
		klog.V(2).ErrorS(err, "failed to get http request", "remoteAddr", proxyConn.RemoteAddr(), "localAddr", proxyConn.LocalAddr())
		return err
	}
	if traceIds, ok := req.Header[util.STREAM_TRACE_ID]; ok && len(traceIds) == 1 {
		req = req.WithContext(context.WithValue(req.Context(), util.STREAM_TRACE_ID, traceIds[0]))
	} else {
		req = req.WithContext(context.WithValue(req.Context(), util.STREAM_TRACE_ID, uuid.NewV4().String()))
	}
	klog.V(2).InfoS("receive request", "method", req.Method, "category", category, "host",
		req.Host, "remoteAddr", proxyConn.RemoteAddr(), "localAddr", proxyConn.LocalAddr(), "req", req,
		util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID).(string))

	if os.Getenv(util.PROXY_AUTHORIZATION_ENV) == "true" {
		if category == util.HTTP_PROXY {
			proxyAuth := req.Header.Get("Proxy-Authorization")
			if proxyAuth == "" {
				klog.Errorf("no username and password provided, %s:%v", util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
				writeErr := util.WriteResponseMsg(proxyConn, "No username and password provided", req.Context().Value(util.STREAM_TRACE_ID).(string), "Unauthorized", http.StatusUnauthorized)
				if writeErr != nil {
					klog.Error(err)
				}
				return fmt.Errorf("no username and password provided, %s:%v", util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
			}
			auths := strings.Split(proxyAuth, " ")
			if len(auths) != 2 && auths[0] != "Basic" {
				klog.Errorf("failed to parse  the Proxy-Authorization field of the request.Header, %s:%v", util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
				writeErr := util.WriteResponseMsg(proxyConn, "failed to parse  the Proxy-Authorization field of the request.Header", req.Context().Value(util.STREAM_TRACE_ID).(string), "Forbidden", http.StatusForbidden)
				if writeErr != nil {
					klog.Error(err)
				}
				return fmt.Errorf("failed to parse  the Proxy-Authorization field of the request.Header")
			}
			infos, err := base64.StdEncoding.DecodeString(auths[1])
			if err != nil {
				klog.ErrorS(err, "failed to decode authInfo", util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
				writeErr := util.WriteResponseMsg(proxyConn, fmt.Sprintf("failed to decode authInfo, error:%v", err), req.Context().Value(util.STREAM_TRACE_ID).(string), "Forbidden", http.StatusForbidden)
				if err != nil {
					klog.Error(writeErr)
				}
				return err
			}
			userinfos := strings.Split(string(infos), ":")
			_, err = os.Stat(util.AuthorizationPath + "/" + userinfos[0])
			if err == nil {
				pwd, err := os.ReadFile(util.AuthorizationPath + "/" + userinfos[0])
				if err != nil {
					klog.ErrorS(err, "failed to get user info", util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
					writeErr := util.WriteResponseMsg(proxyConn, "user does not exist", req.Context().Value(util.STREAM_TRACE_ID).(string), "Forbidden", http.StatusForbidden)
					if writeErr != nil {
						klog.Error(err)
					}
					return err
				}
				if string(pwd) != userinfos[1] {
					klog.Errorf("incorrect password, username:%s, %s:%v", userinfos[0], util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
					writeErr := util.WriteResponseMsg(proxyConn, "incorrect password", req.Context().Value(util.STREAM_TRACE_ID).(string), "Forbidden", http.StatusForbidden)
					if writeErr != nil {
						klog.Error(writeErr)
					}
					return fmt.Errorf("incorrect password, username:%s", userinfos[0])
				}
			} else {
				klog.ErrorS(err, "username does not exist", "username", userinfos[0], util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
				writeErr := util.WriteResponseMsg(proxyConn, fmt.Sprintf("user %s does not exist", userinfos[0]), req.Context().Value(util.STREAM_TRACE_ID).(string), "Forbidden", http.StatusForbidden)
				if writeErr != nil {
					klog.Error(writeErr)
				}
				return err
			}
		}
	}

	if noAccess != nil {
		err = noAccess(req.Host)
		if err != nil {
			klog.ErrorS(err, "host  forbidden to access", "host", req.Host)
			return err
		}
	}

	host, port, err := net.SplitHostPort(req.Host)
	if err != nil {
		klog.ErrorS(err, "failed to get host and port", "category", category, util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
		writeErr := util.InternalServerErrorMsg(proxyConn, err.Error(), req.Context().Value(util.STREAM_TRACE_ID).(string))
		if writeErr != nil {
			klog.Error(writeErr)
		}
		return err
	}

	info, directDial, err := getForwardInfo(host, port)
	if err != nil {
		klog.ErrorS(err, "failed to get forwarding info", util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
		writeErr := util.InternalServerErrorMsg(proxyConn, fmt.Sprintf("failed to get forwarding info, error:%v", err), req.Context().Value(util.STREAM_TRACE_ID).(string))
		if writeErr != nil {
			klog.Error(writeErr)
		}
		return err
	}

	if req.Method == http.MethodConnect {
		if directDial {
			if info != nil {
				host = info.podIp
				port = info.port
			}
			err = common.DirectDial(host, port, category, proxyConn, req.Context())
			if err != nil {
				klog.Error(err)
				return err
			}
			return nil
		}

		err = common.ForwardNode(info.nodeName, info.podIp, info.port, category, proxyConn, req.Context())
		if err != nil {
			klog.ErrorS(err, "failed to  forward request to node", "nodeName", info.nodeName, "addr", net.JoinHostPort(info.podIp, info.port))
			return err
		}
		return nil
	} else {
		if category != util.HTTP_PROXY {
			klog.Errorf("request.Method must be CONNECT")
			writeErr := util.WriteResponseMsg(proxyConn, "request.Method must be CONNECT", req.Context().Value(util.STREAM_TRACE_ID).(string), "Method Not Allowed", http.StatusMethodNotAllowed)
			if writeErr != nil {
				klog.Error(writeErr)
			}
			return fmt.Errorf("request.Method must be CONNECT")
		}
		// direct forwarding
		if directDial {
			remoteAddr := req.Host
			if info != nil {
				remoteAddr = net.JoinHostPort(info.podIp, info.port)
			}
			remoteConn, err := net.Dial("tcp", remoteAddr)
			if err != nil {
				klog.ErrorS(err, "failed to establish a connection between proxyServer and backendServer", util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
				respErr := util.InternalServerErrorMsg(proxyConn, fmt.Sprintf("failed to establish a connection between proxyServer and backendServer, error: %v", err), req.Context().Value(util.STREAM_TRACE_ID).(string))
				if respErr != nil {
					return respErr
				}
				return err
			}
			defer remoteConn.Close()
			err = req.Write(remoteConn)
			if err != nil {
				klog.ErrorS(err, "failed to write request data to remoteConn", util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
				respErr := util.InternalServerErrorMsg(proxyConn, fmt.Sprintf("failed to write request data to remoteConn, error: %v", err), req.Context().Value(util.STREAM_TRACE_ID).(string))
				if respErr != nil {
					return respErr
				}
				return err
			}

			// copy data and close conn
			return util.ConnCopyAndClose(remoteConn, proxyConn, req.Context().Value(util.STREAM_TRACE_ID).(string))
		}

		// Forward to edge node
		edgeNode := tunnelcontext.GetContext().GetNode(info.nodeName)
		if edgeNode != nil {
			conn, err := edgeNode.ConnectNode(category, net.JoinHostPort(info.podIp, info.port), req.Context())
			if err == nil {
				edgeNode.Send2Node(&proto.StreamMsg{
					Node:     info.nodeName,
					Category: category,
					Type:     util.TCP_FORWARD,
					Topic:    conn.GetUid(),
					Data:     reqraw.Bytes(),
				})
				go common.Read(proxyConn, edgeNode, category, util.TCP_FORWARD, conn.GetUid())
				common.Write(proxyConn, conn)
				return nil
			} else {
				klog.ErrorS(err, "failed to connect edge node", "nodeName", info.nodeName, util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
				respErr := util.InternalServerErrorMsg(proxyConn, err.Error(), req.Context().Value(util.STREAM_TRACE_ID).(string))
				if respErr != nil {
					return respErr
				}
				return err
			}
		} else {
			tunnelCloudPodIp, ok := connect.Route.EdgeNode[info.nodeName]
			// Forwarding through tunnel-cloud
			if ok {
				remoteConn, err := net.Dial("tcp", common.GetRemoteAddr(category, tunnelCloudPodIp))
				if err != nil {
					klog.ErrorS(err, "failed to forward http request through localhost [Dial]", util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
					writeErr := util.InternalServerErrorMsg(proxyConn, fmt.Sprintf("failed to forward http request through localhost [Dial], error:%v", err), req.Context().Value(util.STREAM_TRACE_ID).(string))
					if writeErr != nil {
						klog.Error(writeErr)
					}
					return err
				}
				defer remoteConn.Close()

				proxyReq := &http.Request{
					Method: http.MethodConnect,
					URL:    &url.URL{Host: net.JoinHostPort(info.podIp, info.port)},
					Header: map[string][]string{
						util.STREAM_TRACE_ID: {req.Context().Value(util.STREAM_TRACE_ID).(string)},
					},
				}

				proxyCtx, cancle := context.WithTimeout(req.Context(), 5*time.Second)
				defer cancle()
				proxyReq = proxyReq.WithContext(proxyCtx)
				err = proxyReq.Write(remoteConn)
				if err != nil {
					klog.ErrorS(err, "failed to forward http request through localhost [Write Request]", util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
					writeErr := util.InternalServerErrorMsg(proxyConn, fmt.Sprintf("failed to forward http request through localhost [Write Request], error:%v", err), req.Context().Value(util.STREAM_TRACE_ID).(string))
					if writeErr != nil {
						klog.Error(writeErr)
					}
					return err
				}
				resp, respRawData, err := util.GetRespFromConn(remoteConn, nil)
				if err != nil {
					klog.ErrorS(err, "failed to forward http request through localhost [Response]", util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
					writeErr := util.InternalServerErrorMsg(proxyConn, fmt.Sprintf("failed to forward http request through localhost [Write Request], error:%v", err), req.Context().Value(util.STREAM_TRACE_ID).(string))
					if writeErr != nil {
						klog.Error(writeErr)
					}
					return err
				}
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					err = req.Write(remoteConn)
					if err != nil {
						klog.ErrorS(err, "failed to forward http request through localhost [Write Origin Request]", util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
						writeErr := util.InternalServerErrorMsg(proxyConn, fmt.Sprintf("failed to forward http request through localhost [Write Origin Request], error:%v", err), req.Context().Value(util.STREAM_TRACE_ID).(string))
						if writeErr != nil {
							klog.Error(writeErr)
						}
						return err
					}

					// copy data and close conn
					return util.ConnCopyAndClose(remoteConn, proxyConn, req.Context().Value(util.STREAM_TRACE_ID).(string))
				} else {
					klog.Errorf("failed to establish a connection with the target server of the edge node, nodeName:%s, Addr:%s, errorResponse:%s, %s:%s", info.nodeName, net.JoinHostPort(info.podIp, info.port), respRawData.String(), util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
					writeErr := util.InternalServerErrorMsg(proxyConn, fmt.Sprintf("failed to establish a connection with the target server of the edge node, nodeName:%s, Addr:%s, errorResponse:%s", info.nodeName, net.JoinHostPort(info.podIp, info.port), respRawData.String()), "")
					if writeErr != nil {
						klog.Error(writeErr)
					}
					return fmt.Errorf("failed to establish a connection with the target server of the edge node, nodeName:%s, Addr:%s, errorResponse:%s", info.nodeName, net.JoinHostPort(info.podIp, info.port), respRawData.String())
				}
			} else {
				klog.Errorf("the edge node is disconnected from the cloud, nodeName:%s, %s:%v", info.nodeName, util.STREAM_TRACE_ID, req.Context().Value(util.STREAM_TRACE_ID))
				writeErr := util.InternalServerErrorMsg(proxyConn, fmt.Sprintf("the edge node is disconnected from the cloud, nodeName:%s", info.nodeName), req.Context().Value(util.STREAM_TRACE_ID).(string))
				if writeErr != nil {
					klog.Error(writeErr)
				}
				return fmt.Errorf("the edge node is disconnected from the cloud, nodeName:%s", info.nodeName)
			}
		}
	}
}

func getForwardInfoFromEdgeService(svc *v1.Service, port string) (*forwardInfo, error) {

	portInt32, err := strconv.ParseInt(port, 10, 32)
	if err != nil {
		klog.ErrorS(err, "failed to resolve port")
		return nil, err
	}
	podUrl, err := proxy.ResolveEndpoint(indexers.ServiceLister, indexers.EndpointLister, svc.Namespace, svc.Name, int32(portInt32))
	if err != nil {
		klog.ErrorS(err, "failed to get podIp from service")
		return nil, err
	}
	// Only handle access within the cluster
	nodeName, err := indexers.GetNodeByPodIP(podUrl.Hostname())
	if err != nil {
		klog.ErrorS(err, "failed to get the node name where the pod is located")
		return nil, err
	}
	return &forwardInfo{
		podIp:    podUrl.Hostname(),
		port:     podUrl.Port(),
		nodeName: nodeName,
	}, nil
}

func getForwardInfo(host, port string) (*forwardInfo, bool, error) {
	if net.ParseIP(host) == nil {
		/*
		 1. nodeName(kubectl logs/exec; ssh)
		 2. serviceName()
		 3. domain
		*/

		// nodeName
		node, err := indexers.NodeLister.Get(host)
		if err == nil {
			var interIp string
			for _, addr := range node.Status.Addresses {
				if addr.Type == "InternalIP" {
					interIp = addr.Address
				}
			}

			// cloud
			if _, ok := connect.Route.CloudNode[node.Name]; ok {
				return &forwardInfo{
					podIp:    interIp,
					port:     port,
					nodeName: node.Name,
				}, true, nil
			}

			// edge
			if _, ok := connect.Route.EdgeNode[node.Name]; ok {
				return &forwardInfo{
					podIp:    interIp,
					port:     port,
					nodeName: node.Name,
				}, false, nil
			}

			return nil, false, fmt.Errorf("node %s is not registered with routeCache", node.GetName())
		}

		svcs := strings.Split(host, ".")
		if len(svcs) == 2 {
			svc, err := indexers.ServiceLister.Services(svcs[1]).Get(svcs[0])
			if err == nil {
				return resolvForwardInfo(svc, port)
			} else if !apierrors.IsNotFound(err) {
				return nil, false, err
			}
		}
		return nil, true, nil
	} else {
		/*
		  1. clusterIp
		  2. podIp
		*/

		// clusterIp
		svc, err := indexers.GetServiceByClusterIP(host)
		if err != nil && !apierrors.IsNotFound(err) {
			klog.ErrorS(err, "failed to get service by clusterIp", "clusterIp", host)
			return nil, false, err
		}

		if svc != nil {
			return resolvForwardInfo(svc, port)
		}

		// Request pods on edge nodes
		nodeName, err := indexers.GetNodeByPodIP(host)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return nil, true, nil
			}
			return nil, false, err
		}

		return &forwardInfo{
			podIp:    host,
			port:     port,
			nodeName: nodeName,
		}, false, nil

	}
}

func resolvForwardInfo(svc *v1.Service, port string) (*forwardInfo, bool, error) {

	// externalName service
	if svc.Spec.Type == v1.ServiceTypeExternalName {
		return &forwardInfo{
			podIp:    svc.Spec.ExternalName,
			port:     port,
			nodeName: "",
		}, true, nil
	}

	// user services
	if v, ok := connect.Route.UserServicesMap[fmt.Sprintf("%s.%s", svc.Name, svc.Namespace)]; ok {
		if v == util.CLOUD {
			return nil, true, nil
		}
		if v == util.EDGE {
			info, err := getForwardInfoFromEdgeService(svc, port)
			if err != nil {
				klog.ErrorS(err, "failed to get the backend pod instance by the user-defined edge service", "service", svc)
				return nil, false, err
			}
			return info, false, nil
		}
	}

	// service
	if v, ok := connect.Route.ServicesMap[fmt.Sprintf("%s.%s", svc.Name, svc.Namespace)]; ok {
		if v == util.CLOUD {
			return nil, true, nil
		}
		if v == util.EDGE {
			info, err := getForwardInfoFromEdgeService(svc, port)
			if err != nil {
				klog.ErrorS(err, "failed to get the backend pod instance by the edge service", "service", svc)
				return nil, false, err
			}
			return info, false, nil
		}
	}

	return nil, false, fmt.Errorf("failed to forward service request, service.name:%s, service.clusterIp:%s", fmt.Sprintf("%s:%s", svc.Name, svc.Namespace), svc.Spec.ClusterIP)
}
