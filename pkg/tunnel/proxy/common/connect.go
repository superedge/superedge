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
	"github.com/superedge/superedge/pkg/tunnel/context"
	"github.com/superedge/superedge/pkg/tunnel/proto"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"k8s.io/klog/v2"
	"net"
)

const (
	BuferSize = 1 * 1024
)

func Read(conn net.Conn, node context.Node, category, handleType, uuid, addr string) {
	defer func() {
		node.UnbindNode(uuid)
		context.GetContext().RemoveConn(uuid)
		conn.Close()
	}()
	for {
		rb := make([]byte, BuferSize)
		n, err := conn.Read(rb)
		if err != nil {
			ch := context.GetContext().GetConn(uuid)
			if ch != nil {
				ch.Send2Conn(&proto.StreamMsg{
					Node:     node.GetName(),
					Category: category,
					Type:     util.CLOSED,
					Topic:    uuid,
				})
			}
			klog.Errorf("Failed to read data, error: %v", err)
			return
		}
		node.Send2Node(&proto.StreamMsg{
			Node:     node.GetName(),
			Category: category,
			Type:     handleType,
			Topic:    uuid,
			Data:     rb[:n],
			Addr:     addr,
		})
	}
}

func Write(conn net.Conn, ch context.Conn) {
	defer func() {
		conn.Close()
	}()

	for {
		select {
		case msg := <-ch.ConnRecv():
			if msg.Type == util.CLOSED {
				klog.V(4).Infof("Receive a close message")
				return
			}
			_, err := conn.Write(msg.Data)
			if err != nil {
				klog.Errorf("Failed to write data, error: %v", err)
				return
			}
		}
	}
}
