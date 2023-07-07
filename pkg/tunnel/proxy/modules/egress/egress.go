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
	"strconv"

	"github.com/superedge/superedge/pkg/tunnel/conf"
	"github.com/superedge/superedge/pkg/tunnel/module"
	"github.com/superedge/superedge/pkg/tunnel/proxy/handlers"
	"github.com/superedge/superedge/pkg/tunnel/tunnelcontext"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"k8s.io/klog/v2"
)

type EgressSelector struct {
}

func (e EgressSelector) Name() string {
	return util.EGRESS
}

func (e EgressSelector) Start(mode string) {
	tunnelcontext.GetContext().RegisterHandler(util.TCP_FORWARD, util.EGRESS, handlers.DirectHandler)
	tunnelcontext.GetContext().RegisterHandler(util.CLOSED, util.EGRESS, handlers.DirectHandler)
	tunnelcontext.GetContext().RegisterHandler(tunnelcontext.CONNECT_REQ, util.EGRESS, handlers.ConnectingHandler)
	tunnelcontext.GetContext().RegisterHandler(tunnelcontext.CONNECT_SUCCESSED, util.EGRESS, handlers.DirectHandler)
	tunnelcontext.GetContext().RegisterHandler(tunnelcontext.CONNECT_FAILED, util.EGRESS, handlers.DirectHandler)
	if mode == util.CLOUD {
		if conf.TunnelConf.TunnlMode.Cloud.Egress == nil {
			return
		}
		config, err := util.LoadTLSConfig(util.EgressCertPath, util.EgressKeyPath,
			conf.TunnelConf.TunnlMode.Cloud.TLS.CipherSuites, conf.TunnelConf.TunnlMode.Cloud.TLS.MinTLSVersion, true)
		if err != nil {
			return
		}

		listener, err := tls.Listen("tcp", "0.0.0.0:"+strconv.Itoa(conf.TunnelConf.TunnlMode.Cloud.Egress.EgressPort), config)
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
				go func() {
					handlerErr := handlers.HandleServerConn(conn, util.EGRESS, nil)
					if handlerErr != nil {
						klog.Error(handlerErr)
					}
				}()
			}
		}()
	}
}

func (e EgressSelector) CleanUp() {
	tunnelcontext.GetContext().RemoveModule(e.Name())
}

func InitEgress() {
	module.Register(&EgressSelector{})
	klog.Infof("init module: %s success !", util.EGRESS)
}
