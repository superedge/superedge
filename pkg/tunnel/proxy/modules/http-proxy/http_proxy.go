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

package http_proxy

import (
	"fmt"
	"github.com/superedge/superedge/pkg/tunnel/conf"
	"github.com/superedge/superedge/pkg/tunnel/module"
	"github.com/superedge/superedge/pkg/tunnel/proxy/handlers"
	"github.com/superedge/superedge/pkg/tunnel/proxy/modules/http-proxy/connect"
	"github.com/superedge/superedge/pkg/tunnel/tunnelcontext"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"k8s.io/klog/v2"
	"net"
	"os"
	"strconv"
)

type HttpProxy struct {
	stop chan struct{}
}

func (h HttpProxy) Name() string {
	return util.HTTP_PROXY
}

func (h HttpProxy) Start(mode string) {
	//Handle HTTP_CONNECT requests for tunnel establishment
	tunnelcontext.GetContext().RegisterHandler(util.TCP_FORWARD, util.HTTP_PROXY, handlers.DirectHandler)
	tunnelcontext.GetContext().RegisterHandler(tunnelcontext.CONNECT_SUCCESSED, util.HTTP_PROXY, handlers.DirectHandler)
	tunnelcontext.GetContext().RegisterHandler(tunnelcontext.CONNECT_FAILED, util.HTTP_PROXY, handlers.DirectHandler)
	tunnelcontext.GetContext().RegisterHandler(util.CLOSED, util.HTTP_PROXY, handlers.DirectHandler)
	go func() {
		if mode == util.EDGE {
			tunnelcontext.GetContext().RegisterHandler(tunnelcontext.CONNECT_REQ, util.HTTP_PROXY, handlers.ConnectingHandler)
			listener, err := net.Listen("tcp", conf.TunnelConf.TunnlMode.EDGE.HttpProxy.ProxyIP+":"+conf.TunnelConf.TunnlMode.EDGE.HttpProxy.ProxyPort)
			if err != nil {
				klog.Errorf("Failed to start http_proxy edge server, error: %v", err)
				return
			}
			for {
				conn, err := listener.Accept()
				if err != nil {
					klog.Errorf("http_proxy edge server accept failed, error: %v", err)
					continue
				}
				go connect.HttpProxyEdgeServer(conn)
			}

		} else if mode == util.CLOUD {
			tunnelcontext.GetContext().RegisterHandler(tunnelcontext.CONNECT_REQ, util.HTTP_PROXY, handlers.AccessHandler)
			listener, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(conf.TunnelConf.TunnlMode.Cloud.HttpProxy.ProxyPort))
			if err != nil {
				klog.Errorf("Failed to start http_proxy edge server, error: %v", err)
				return
			}
			for {
				conn, err := listener.Accept()
				if err != nil {
					klog.Errorf("http_proxy edge server accept failed, error: %v", err)
					continue
				}
				//go connect.HttpProxyCloudServer(conn)
				go handlers.HandleServerConn(conn, util.HTTP_PROXY, func(host string) error {
					if os.Getenv(util.CloudProxy) != "" {
						config := util.NewHttpProxyConfig(os.Getenv(util.CloudProxy))
						if !config.UseProxy(host) {
							klog.V(8).Infof("Forbid access to service %s in the cluster", host)
							return fmt.Errorf("forbid access to service %s in the cluster", host)
						}
					}
					return nil
				})
			}
		}
	}()

}

func (h HttpProxy) CleanUp() {
	h.stop <- struct{}{}
	tunnelcontext.GetContext().RemoveModule(util.HTTP_PROXY)
}

func InitHttpProxy() {
	module.Register(&HttpProxy{})
	klog.Infof("init module: %s success !", util.HTTP_PROXY)
}
