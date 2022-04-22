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

package egress

import (
	"crypto/tls"
	"github.com/superedge/superedge/pkg/tunnel/conf"
	"github.com/superedge/superedge/pkg/tunnel/context"
	"github.com/superedge/superedge/pkg/tunnel/module"
	"github.com/superedge/superedge/pkg/tunnel/proxy/handlers"
	"github.com/superedge/superedge/pkg/tunnel/proxy/modules/egress/connect"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"k8s.io/klog/v2"
)

type EgressSelector struct {
}

func (e EgressSelector) Name() string {
	return util.EGRESS
}

func (e EgressSelector) Start(mode string) {
	context.GetContext().RegisterHandler(util.TCP_FRONTEND, util.EGRESS, handlers.FrontendHandler)
	context.GetContext().RegisterHandler(util.TCP_BACKEND, util.EGRESS, handlers.DirectHandler)
	context.GetContext().RegisterHandler(util.CLOSED, util.EGRESS, handlers.DirectHandler)
	if mode == util.CLOUD {
		if conf.TunnelConf.TunnlMode.Cloud.Egress == nil {
			return
		}
		cert, err := tls.LoadX509KeyPair(conf.TunnelConf.TunnlMode.Cloud.Egress.ServerCert, conf.TunnelConf.TunnlMode.Cloud.Egress.ServerKey)
		if err != nil {
			klog.Errorf("client load cert fail, error: %v", err)
			return
		}
		config := &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		}
		listener, err := tls.Listen("tcp", "0.0.0.0:"+conf.TunnelConf.TunnlMode.Cloud.Egress.EgressPort, config)
		if err != nil {
			klog.Errorf("Failed to start SSH Server, error:%v", err)
			return
		}
		go func() {
			for {
				conn, err := listener.Accept()
				if err != nil {
					klog.Errorf("SSH Server accept failed, error: %v", err)
					continue
				}
				go connect.HandleEgressConn(conn)
			}
		}()
	}
}

func (e EgressSelector) CleanUp() {
	context.GetContext().RemoveModule(e.Name())
}

func InitEgress() {
	module.Register(&EgressSelector{})
	klog.Infof("init module: %s success !", util.EGRESS)
}
