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
	"github.com/superedge/superedge/pkg/tunnel/proto"
	"k8s.io/klog/v2"
)

type Server struct{}

func (s *Server) TunnelStreaming(stream proto.Stream_TunnelStreamingServer) error {
	errChan := make(chan error, 2)

	go func(sendStream proto.Stream_TunnelStreamingServer, sendChan chan error) {
		sendErr := sendStream.SendMsg(nil)
		if sendErr != nil {
			klog.Errorf("streamServer failed to send message err = %v", sendErr)
		}
		sendChan <- sendErr
	}(stream, errChan)

	go func(recvStream proto.Stream_TunnelStreamingServer, recvChan chan error) {
		recvErr := stream.RecvMsg(nil)
		if recvErr != nil {
			klog.Errorf("streamServer failed to receive message err = %v", recvErr)
		}
		recvChan <- recvErr
	}(stream, errChan)

	e := <-errChan
	klog.Errorf("the stream of streamServer is disconnected err = %v", e)
	return e
}
