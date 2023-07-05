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

package handlers

import (
	"fmt"
	"net"

	"github.com/superedge/superedge/pkg/tunnel/proto"
	"github.com/superedge/superedge/pkg/tunnel/proxy/common"
	"github.com/superedge/superedge/pkg/tunnel/tunnelcontext"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"k8s.io/klog/v2"
)

func DirectHandler(msg *proto.StreamMsg) error {
	chConn := tunnelcontext.GetContext().GetConn(msg.Topic)
	if chConn != nil {
		chConn.Send2Conn(msg)
	} else {
		sendFlag := false
		if localNode := tunnelcontext.GetContext().GetNode(msg.Node); localNode != nil {
			if remoteNodeName := localNode.GetPairNode(msg.GetTopic()); remoteNodeName != "" {
				if remoteNode := tunnelcontext.GetContext().GetNode(remoteNodeName); remoteNode != nil {
					remoteNode.Send2Node(msg)
					sendFlag = true
				}
			}
		}
		if !sendFlag {
			// skip duplicate closed messages
			if msg.Type == util.CLOSED {
				return nil
			}
			if localNode := tunnelcontext.GetContext().GetNode(msg.Node); localNode != nil {
				localNode.Send2Node(&proto.StreamMsg{
					Node:     msg.Node,
					Category: msg.Category,
					Type:     util.CLOSED,
					Topic:    msg.Topic,
					Data:     []byte("msg cannot be forwarded due to disconnection"),
				})
			}
			klog.InfoS("msg cannot be forwarded due to disconnection", "category", msg.Category,
				"type", msg.Type, util.STREAM_TRACE_ID, msg.Topic)
			return fmt.Errorf("failed to get msgChannel, %s:%s", util.STREAM_TRACE_ID, msg.Topic)
		}

	}
	return nil
}

func ConnectingHandler(msg *proto.StreamMsg) error {
	chConn := tunnelcontext.GetContext().GetConn(msg.Topic)
	if chConn != nil {
		chConn.Send2Conn(msg)
		return nil
	}
	node := tunnelcontext.GetContext().GetNode(msg.Node)
	conn, err := net.Dial(util.TCP, msg.Addr)
	if err != nil {
		if node != nil {
			node.Send2Node(&proto.StreamMsg{
				Node:     node.GetName(),
				Category: msg.Category,
				Type:     tunnelcontext.CONNECT_FAILED,
				Topic:    msg.Topic,
				Data:     []byte(err.Error()),
			})
		}
		return err
	}
	if node != nil {
		node.Send2Node(&proto.StreamMsg{
			Node:     node.GetName(),
			Category: msg.Category,
			Type:     tunnelcontext.CONNECT_SUCCESSED,
			Topic:    msg.Topic,
		})
	}
	ch := tunnelcontext.GetContext().AddConn(msg.Topic)
	node.BindNode(msg.Topic)
	go common.Read(conn, node, msg.Category, util.TCP_FORWARD, msg.Topic)
	go common.Write(conn, ch)
	return nil
}
