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

package tcpmsg

import (
	"fmt"
	"k8s.io/klog"
	"net"
	"superedge/pkg/tunnel/context"
	"superedge/pkg/tunnel/proto"
	"superedge/pkg/tunnel/proxy/tcp/tcpmng"
	"superedge/pkg/tunnel/util"
)

func BackendHandler(msg *proto.StreamMsg) error {
	conn := context.GetContext().GetConn(msg.Topic)
	if conn == nil {
		klog.Errorf("trace_id = %s the stream module failed to distribute the side message module = %s type = %s", msg.Topic, msg.Category, msg.Type)
		return fmt.Errorf("trace_id = %s the stream module failed to distribute the side message module = %s type = %s ", msg.Topic, msg.Category, msg.Type)
	}
	conn.Send2Conn(msg)
	return nil
}

func FrontendHandler(msg *proto.StreamMsg) error {
	c := context.GetContext().GetConn(msg.Topic)
	if c != nil {
		c.Send2Conn(msg)
		return nil
	}
	tp := tcpmng.NewTcpConn(msg.Topic, msg.Addr, msg.Node)
	tp.Type = util.TCP_BACKEND
	tp.C.Send2Conn(msg)
	tcpAddr, err := net.ResolveTCPAddr("tcp", tp.Addr)
	if err != nil {
		klog.Error("edeg proxy resolve addr fail !")
		return err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		klog.Error("edge proxy connect fail!")
		return err
	}
	tp.Conn = conn
	go tp.Read()
	go tp.Write()
	return nil
}

func ControlHandler(msg *proto.StreamMsg) error {
	conn := context.GetContext().GetConn(msg.Topic)
	if conn == nil {
		klog.Errorf("trace_id = %s the stream module failed to distribute the side message module = %s type = %s", msg.Topic, msg.Category, msg.Type)
		return fmt.Errorf("trace_id = %s the stream module failed to distribute the side message module = %s type = %s ", msg.Topic, msg.Category, msg.Type)
	}
	conn.Send2Conn(msg)
	return nil
}
