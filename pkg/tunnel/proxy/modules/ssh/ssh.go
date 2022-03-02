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

package ssh

import (
	"github.com/superedge/superedge/pkg/tunnel/context"
	"github.com/superedge/superedge/pkg/tunnel/module"
	"github.com/superedge/superedge/pkg/tunnel/proxy/handlers"
	"github.com/superedge/superedge/pkg/tunnel/proxy/modules/ssh/connect"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"k8s.io/klog/v2"
	"net"
)

type SSH struct {
}

func (s SSH) Name() string {
	return util.SSH
}

func (s SSH) Start(mode string) {
	context.GetContext().RegisterHandler(util.TCP_FRONTEND, util.SSH, handlers.FrontendHandler)
	context.GetContext().RegisterHandler(util.TCP_BACKEND, util.SSH, handlers.DirectHandler)
	context.GetContext().RegisterHandler(util.CLOSED, util.SSH, handlers.DirectHandler)
	if mode == util.CLOUD {
		listener, err := net.Listen(util.TCP, "0.0.0.0:22")
		if err != nil {
			klog.Errorf("Failed to start SSH Server, error:%v", err)
			return
		}
		for {
			conn, err := listener.Accept()
			if err != nil {
				klog.Errorf("SSH Server accept failed, error: %v", err)
				continue
			}
			go connect.HandleServerConn(conn)
		}
	}
}

func (s SSH) CleanUp() {
	context.GetContext().RemoveModule(s.Name())
}

func InitSSH() {
	module.Register(&SSH{})
	klog.Infof("init module: %s success !", util.SSH)
}
