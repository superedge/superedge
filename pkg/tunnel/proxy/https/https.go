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

package https

import (
	"superedge/pkg/tunnel/context"
	"superedge/pkg/tunnel/model"
	"superedge/pkg/tunnel/proxy/https/httpsmng"
	"superedge/pkg/tunnel/proxy/https/httpsmsg"
	"superedge/pkg/tunnel/util"
	"k8s.io/klog"
)

type Https struct {
}

func (https *Https) Name() string {
	return util.HTTPS
}

func (https *Https) Start(mode string) {
	context.GetContext().RegisterHandler(util.CONNECTING, util.HTTPS, httpsmsg.ConnectingHandler)
	context.GetContext().RegisterHandler(util.CONNECTED, util.HTTPS, httpsmsg.ConnectedAndTransmission)
	context.GetContext().RegisterHandler(util.CLOSED, util.HTTPS, httpsmsg.ConnectedAndTransmission)
	context.GetContext().RegisterHandler(util.TRANSNMISSION, util.HTTPS, httpsmsg.ConnectedAndTransmission)
	if mode == util.CLOUD {
		go httpsmng.StartServer()
	}
}

func (https *Https) CleanUp() {
	context.GetContext().RemoveModule(util.HTTPS)
}

func InitHttps() {
	model.Register(&Https{})
	klog.Infof("init module: %s success !", util.HTTPS)
}
