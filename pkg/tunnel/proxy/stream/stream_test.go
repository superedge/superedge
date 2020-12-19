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
	ctx "context"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/channelz/grpc_channelz_v1"
	"k8s.io/klog"
	"os"
	"superedge/pkg/tunnel/conf"
	"superedge/pkg/tunnel/context"
	"superedge/pkg/tunnel/model"
	"superedge/pkg/tunnel/proto"
	"superedge/pkg/tunnel/proxy/stream/streammng/connect"
	"superedge/pkg/tunnel/util"
	"testing"
	"time"
)

func Test_StreamServer(t *testing.T) {
	err := conf.InitConf(util.CLOUD, "../../../../conf/cloud_mode.toml")
	if err != nil {
		t.Errorf("failed to initialize stream server configuration file err = %v", err)
		return
	}
	model.InitModules(util.CLOUD)
	InitStream(util.CLOUD)
	model.LoadModules(util.CLOUD)
	context.GetContext().RegisterHandler(util.MODULE_DEBUG, util.STREAM, StreamDebugHandler)
	model.ShutDown()

}

func StreamDebugHandler(msg *proto.StreamMsg) error {
	klog.Infof("received the msg node = %s uuid = %s data = %s", msg.Node, msg.Topic, string(msg.Data))
	node := context.GetContext().GetNode(msg.Node)
	if node == nil {
		klog.Errorf("failed to send debug to edge node node: %s", msg.Node)
		return fmt.Errorf("failed to send debug to edge node node: %s", msg.Node)
	}
	if len(msg.Data) == 1 && msg.Data[0] == 's' {
		msg.Data[0] = 'c'
	} else {
		msg.Data[0] = 's'
		node.Send2Node(msg)
	}
	return nil
}

func Test_StreamClient(t *testing.T) {
	os.Setenv(util.NODE_NAME_ENV, "node1")
	err := conf.InitConf(util.EDGE, "../../../../conf/edge_mode.toml")
	if err != nil {
		t.Errorf("failed to initialize stream client configuration file err = %v", err)
		return
	}
	model.InitModules(util.EDGE)
	InitStream(util.EDGE)
	model.LoadModules(util.EDGE)
	context.GetContext().RegisterHandler(util.MODULE_DEBUG, util.STREAM, StreamDebugHandler)
	go func() {
		running := true
		for running {
			node := context.GetContext().GetNode(os.Getenv(util.NODE_NAME_ENV))
			if node != nil {
				node.Send2Node(&proto.StreamMsg{
					Node:     os.Getenv(util.NODE_NAME_ENV),
					Category: util.STREAM,
					Type:     util.MODULE_DEBUG,
					Topic:    uuid.NewV4().String(),
					Data:     []byte{'c'},
				})
			}
			time.Sleep(10 * time.Second)
		}
	}()
	model.ShutDown()

}

func Test_ChannelzSever(t *testing.T) {
	os.Setenv(util.NODE_NAME_ENV, "node1")
	err := conf.InitConf(util.EDGE, "../../../../conf/edge_mode.toml")
	if err != nil {
		t.Errorf("failed to initialize stream client configuration file err = %v", err)
		return
	}
	conn, clictx, cancle, err :=connect.StartClient()
	if err != nil {
		t.Errorf("failed to grpc client err: %v", err)
		return
	}
	cli := grpc_channelz_v1.NewChannelzClient(conn)

	respServers, err := cli.GetServers(clictx, &grpc_channelz_v1.GetServersRequest{})
	if err != nil {
		t.Errorf("failed to get channelz servers response err = %v", err)
		return
	}
	fmt.Println(respServers)

	respServerSockets, err := cli.GetServerSockets(clictx, &grpc_channelz_v1.GetServerSocketsRequest{})
	if err != nil {
		t.Errorf("failed to get channnelz server sockets err = %v ", err)
		return
	}
	fmt.Println(respServerSockets)

	respSocket, err := cli.GetSocket(clictx, &grpc_channelz_v1.GetSocketRequest{
		SocketId: respServers.Server[0].GetListenSocket()[0].GetSocketId(),
	})
	if err != nil {
		t.Errorf("failed to get channel socket err = %v", err)
		return
	}
	fmt.Println(respSocket)
	conn.Close()
	cancle()
}

func Test_ChannelzClient(t *testing.T) {
	cliContext := ctx.Background()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.DialContext(cliContext, "127.0.0.1:5000", opts...)
	if err != nil {
		t.Errorf("failed to establish grpc connection err = %v ", err)
		return
	}
	cli := grpc_channelz_v1.NewChannelzClient(conn)

	respTopChannel, err := cli.GetTopChannels(cliContext, &grpc_channelz_v1.GetTopChannelsRequest{
		StartChannelId: 0,
	})
	if err != nil {
		t.Errorf("failed to get channelz topchannel err = %v", err)
		return
	}
	//fmt.Println(respTopChannel)

	respChannel, err := cli.GetChannel(cliContext, &grpc_channelz_v1.GetChannelRequest{
		ChannelId: respTopChannel.Channel[0].GetRef().ChannelId,
	})
	if err != nil {
		t.Errorf("failed to channelz channel resp err = %v", err)
		return
	}
	for _, v := range respChannel.Channel.Data.GetTrace().GetEvents() {
		fmt.Println(v)
	}
	//fmt.Println(respChannel)

	conn.Close()

}
