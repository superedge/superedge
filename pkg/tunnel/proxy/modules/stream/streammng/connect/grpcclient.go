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

package connect

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/superedge/superedge/pkg/tunnel/conf"
	"github.com/superedge/superedge/pkg/tunnel/proto"
	stream2 "github.com/superedge/superedge/pkg/tunnel/proxy/modules/stream/streammng/stream"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"k8s.io/klog/v2"
)

var kacp = keepalive.ClientParameters{
	Time:                10 * time.Millisecond,
	Timeout:             time.Second,
	PermitWithoutStream: true,
}

var streamConn *grpc.ClientConn

func StartClient() (*grpc.ClientConn, error) {
	creds, err := credentials.NewClientTLSFromFile(util.TunnelEdgeCAPath, conf.TunnelConf.TunnlMode.EDGE.StreamEdge.Client.Dns)
	if err != nil {
		klog.ErrorS(err, "failed to load credentials")
		return nil, err
	}
	opts := []grpc.DialOption{grpc.WithKeepaliveParams(kacp), grpc.WithStreamInterceptor(ClientStreamInterceptor), grpc.WithTransportCredentials(creds), grpc.WithConnectParams(grpc.ConnectParams{
		MinConnectTimeout: 60 * time.Second,
	}), grpc.WithBlock()}
	conn, err := grpc.Dial(conf.TunnelConf.TunnlMode.EDGE.StreamEdge.Client.ServerName, opts...)
	if err != nil {
		klog.Error("edge start client fail !")
		return nil, err
	}
	return conn, nil
}

func StartSendClient() {
	conn, err := StartClient()
	if err != nil {
		klog.ErrorS(err, "edge start client error !")
		os.Exit(1)
	}
	defer conn.Close()
	streamConn = conn
	for {
		if conn.GetState() == connectivity.Ready {
			cli := proto.NewStreamClient(conn)
			stream2.Send(cli, context.Background())
		}
		time.Sleep(1 * time.Second)
	}
}

func EdgeHealthCheck(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		if streamConn == nil {
			writer.WriteHeader(http.StatusInternalServerError)
			klog.Error("Edge node connection is not established")
			return
		}
		if streamConn.GetState() == connectivity.Ready {
			writer.WriteHeader(http.StatusOK)
			klog.Infof("the connection of the node is being established status = %s", streamConn.GetState())
			return
		} else {
			writer.WriteHeader(http.StatusInternalServerError)
			klog.Errorf("the connection of the node is abnormal status = %s", streamConn.GetState())
			return
		}
	} else {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		klog.Error("only supports GET method")
		return
	}
}
