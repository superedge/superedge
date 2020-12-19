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

package stream

import (
	"superedge/pkg/tunnel/conf"
	"superedge/pkg/tunnel/context"
	"superedge/pkg/tunnel/model"
	"superedge/pkg/tunnel/proxy/stream/streammng/connect"
	"superedge/pkg/tunnel/proxy/stream/streammsg"
	"superedge/pkg/tunnel/token"
	"superedge/pkg/tunnel/util"
	"k8s.io/klog"
	"os"
)

type Stream struct {
}

func (stream *Stream) Name() string {
	return util.STREAM
}

func (stream *Stream) Start(mode string) {
	context.GetContext().RegisterHandler(util.STREAM_HEART_BEAT, util.STREAM, streammsg.HeartbeatHandler)
	var channelzAddr string
	if mode == util.CLOUD {
		go connect.StartServer()
		if !conf.TunnelConf.TunnlMode.Cloud.Stream.Dns.Debug {
			go connect.SynCorefile()
		}
		channelzAddr = conf.TunnelConf.TunnlMode.Cloud.Stream.Server.ChannelzAddr
	} else {
		go connect.StartSendClient()
		channelzAddr = conf.TunnelConf.TunnlMode.EDGE.StreamEdge.Client.ChannelzAddr
	}

	go connect.StartLogServer(mode)

	go connect.StartChannelzServer(channelzAddr)
}

func (stream *Stream) CleanUp() {
	context.GetContext().RemoveModule(stream.Name())
}

func InitStream(mode string) {
	if mode == util.CLOUD {
		if !conf.TunnelConf.TunnlMode.Cloud.Stream.Dns.Debug {
			err := connect.InitDNS()
			if err != nil {
				klog.Errorf("init client-go fail err = %v", err)
				return
			}
		}
		err := token.InitTokenCache(conf.TunnelConf.TunnlMode.Cloud.Stream.Server.TokenFile)
		if err != nil {
			klog.Error("Error loading token file ！")
		}
	} else {
		err := connect.InitToken(os.Getenv(util.NODE_NAME_ENV), conf.TunnelConf.TunnlMode.EDGE.StreamEdge.Client.Token)
		if err != nil {
			klog.Errorf("initialize the edge node token err = %v", err)
			return
		}
	}
	model.Register(&Stream{})
	klog.Infof("init module: %s success !", util.STREAM)
}
