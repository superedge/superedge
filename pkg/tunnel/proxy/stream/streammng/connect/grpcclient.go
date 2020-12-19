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
	ctx "context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"k8s.io/klog"
	"math"
	"net/http"
	"os"
	"superedge/pkg/tunnel/conf"
	"superedge/pkg/tunnel/proto"
	"superedge/pkg/tunnel/proxy/stream/streammng/stream"
	"superedge/pkg/tunnel/util"
	"time"
)

var kacp = keepalive.ClientParameters{
	Time:                10 * time.Millisecond,
	Timeout:             time.Second,
	PermitWithoutStream: true,
}

var streamConn *grpc.ClientConn

func StartClient() (*grpc.ClientConn, ctx.Context, ctx.CancelFunc, error) {
	creds, err := credentials.NewClientTLSFromFile(conf.TunnelConf.TunnlMode.EDGE.StreamEdge.Client.Cert, conf.TunnelConf.TunnlMode.EDGE.StreamEdge.Client.Dns)
	if err != nil {
		klog.Errorf("failed to load credentials: %v", err)
		return nil, nil, nil, err
	}
	opts := []grpc.DialOption{grpc.WithKeepaliveParams(kacp), grpc.WithStreamInterceptor(ClientStreamInterceptor), grpc.WithTransportCredentials(creds)}
	conn, err := grpc.Dial(conf.TunnelConf.TunnlMode.EDGE.StreamEdge.Client.ServerName, opts...)
	if err != nil {
		klog.Error("edge start client fail !")
		return nil, nil, nil, err
	}
	clictx, cancle := ctx.WithTimeout(ctx.Background(), time.Duration(math.MaxInt64))
	return conn, clictx, cancle, nil
}

func StartSendClient() {
	conn, clictx, cancle, err := StartClient()
	if err != nil {
		klog.Error("edge start client error !")
		klog.Flush()
		os.Exit(1)
	}
	streamConn = conn
	defer func() {
		conn.Close()
		cancle()
	}()

	go func(monitor *grpc.ClientConn) {
		mcount := 0
		for {
			if conn.GetState() == connectivity.Ready {
				mcount = 0
			} else {
				mcount += 1
			}
			klog.V(8).Infof("grpc connection status = %s count = %v", conn.GetState(), mcount)
			if mcount >= util.TIMEOUT_EXIT {
				klog.Error("grpc connection rebuild timed out, container exited !")
				klog.Flush()
				os.Exit(1)
			}
			klog.V(8).Infof("grpc connection status of node = %v", conn.GetState())
			time.Sleep(1 * time.Second)
		}
	}(conn)
	running := true
	count := 0
	for running {
		if conn.GetState() == connectivity.Ready {
			cli := proto.NewStreamClient(conn)
			stream.Send(cli, clictx)
			count = 0
		}
		count += 1
		klog.V(8).Infof("node connection status = %s count = %v", conn.GetState(), count)
		time.Sleep(1 * time.Second)
		if count >= util.TIMEOUT_EXIT {
			klog.Error("the streamClient retrying to establish a connection timed out and the container exited !")
			klog.Flush()
			os.Exit(1)
		}
	}
}

func EdgeHealthCheck(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		if streamConn == nil {
			writer.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(writer, "Edge node connection is not established")
		}
		if streamConn.GetState() == connectivity.Ready {
			writer.WriteHeader(http.StatusOK)
			fmt.Fprintf(writer, "the connection of the node is being established status = %s", streamConn.GetState())
		} else {
			writer.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(writer, "the connection of the node is abnormal status = %s", streamConn.GetState())
		}
	} else {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(writer, "only supports GET method")
	}
}
