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

package common

import (
	"io"
	"net"

	"github.com/superedge/superedge/pkg/tunnel/proto"
	"github.com/superedge/superedge/pkg/tunnel/tunnelcontext"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"k8s.io/klog/v2"
)

const (
	BuferSize = 1 * 1024
)

func Read(conn net.Conn, node tunnelcontext.Node, category, handleType, uuid string) {
	defer func() {
		node.UnbindNode(uuid)
		tunnelcontext.GetContext().RemoveConn(uuid)
		conn.Close()
	}()
	for {
		rb := make([]byte, BuferSize)
		n, err := conn.Read(rb)
		if err != nil {
			closeMsg := &proto.StreamMsg{
				Node:     node.GetName(),
				Category: category,
				Type:     util.CLOSED,
				Topic:    uuid,
				Data:     []byte(err.Error()),
			}
			node.Send2Node(closeMsg)
			klog.V(2).InfoS("send closeMsg", "closeMsg", closeMsg, util.STREAM_TRACE_ID, uuid)
			if err != io.EOF {
				klog.ErrorS(err, "failed to read data", util.STREAM_TRACE_ID, uuid)
			}
			return
		}
		node.Send2Node(&proto.StreamMsg{
			Node:     node.GetName(),
			Category: category,
			Type:     handleType,
			Topic:    uuid,
			Data:     rb[:n],
		})
	}
}

func Write(conn net.Conn, ch tunnelcontext.Conn) {
	defer func() {
		conn.Close()
	}()

	for {
		select {
		case msg := <-ch.ConnRecv():
			if msg.Type == util.CLOSED {
				klog.V(2).InfoS("receive a closeMsg", "closeMsg", msg)
				return
			}
			_, err := conn.Write(msg.Data)
			if err != nil {
				klog.ErrorS(err, "failed to write data", util.STREAM_TRACE_ID, msg.Topic)
				return
			}
		}
	}
}
