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
	"github.com/superedge/superedge/pkg/tunnel/context"
	"github.com/superedge/superedge/pkg/tunnel/proto"
	"github.com/superedge/superedge/pkg/tunnel/proxy/common"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"k8s.io/klog/v2"
	"net"
)

func FrontendHandler(msg *proto.StreamMsg) error {
	chConn := context.GetContext().GetConn(msg.Topic)
	if chConn != nil {
		chConn.Send2Conn(msg)
		return nil
	}
	conn, err := net.Dial("tcp", msg.Addr)
	if err != nil {
		klog.Errorf("The Edge TCP client failed to establish a TCP connection with the Edge Server, error: %v", err)
		return err
	}
	node := context.GetContext().GetNode(msg.Node)
	ch := context.GetContext().AddConn(msg.Topic)
	node.BindNode(msg.Topic)
	ch.Send2Conn(msg)
	go common.Read(conn, node, msg.Category, util.TCP_BACKEND, msg.Topic, msg.Addr)
	go common.Write(conn, ch)
	return nil
}
