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

package httpsmng

import (
	"superedge/pkg/tunnel/context"
	"superedge/pkg/tunnel/proto"
	"superedge/pkg/tunnel/util"
	"io"
	"k8s.io/klog"
	"net"
)

func NetRead(conn net.Conn, uid string, node context.Node, stop, complete chan struct{}) {
	rrunning := true
	for rrunning {
		rbyte := make([]byte, 1024)
		n, err := conn.Read(rbyte)
		msg := &proto.StreamMsg{
			Node:     node.GetName(),
			Category: util.HTTPS,
			Type:     util.TRANSNMISSION,
			Topic:    uid,
			Data:     rbyte[:n],
		}
		if err != nil {
			if err == io.EOF {
				klog.V(4).Infof("traceid = %s  read failed err = %v", uid, err)
			} else {
				klog.Errorf("traceid = %s  read failed err = %v", uid, err)
			}
			rrunning = false
			msg.Type = util.CLOSED
		}
		node.Send2Node(msg)
	}
	conn.Close()
	stop <- struct{}{}
	complete <- struct{}{}
}

func NetWrite(nc net.Conn, node context.Node, conn context.Conn, stop, complete chan struct{}) {
	defer nc.Close()
	wrunning := true
	for wrunning {
		select {
		case msg := <-conn.ConnRecv():
			if msg.Data != nil && len(msg.Data) != 0 {
				_, err := nc.Write(msg.Data)
				if err != nil {
					klog.Errorf("traceid = %s httpserver net write failed err = %v", conn.GetUid(), err)
					wrunning = false
				}
			}
			if msg.Type == util.CLOSED {
				klog.Errorf("traceid = %s The server receives a client shutdown message!", conn.GetUid())
				wrunning = false
			}
		case <-stop:
			klog.Infof("traceid = %s net write stop", conn.GetUid())
			wrunning = false
			break
		}
	}
	node.UnbindNode(conn.GetUid())
	context.GetContext().RemoveConn(conn.GetUid())
	complete <- struct{}{}
}

func IoRead(body io.ReadWriteCloser, uid string, node context.Node, stop, complete chan struct{}) {
	defer body.Close()
	rrunning := true
	for rrunning {
		rbyte := make([]byte, 1024)
		n, err := body.Read(rbyte)
		msg := &proto.StreamMsg{
			Node:     node.GetName(),
			Category: util.HTTPS,
			Type:     util.TRANSNMISSION,
			Topic:    uid,
			Data:     rbyte[:n],
		}
		if err != nil {
			if err == io.EOF {
				klog.V(4).Infof("traceid = %s  io read failed err = %v", uid, err)
			} else {
				klog.Errorf("traceid = %s  io read failed err = %v", uid, err)
			}
			rrunning = false
			msg.Type = util.CLOSED
		}
		node.Send2Node(msg)
	}
	stop <- struct{}{}
	complete <- struct{}{}
}

func IoWrite(body io.ReadWriteCloser, node context.Node, conn context.Conn, stop, complete chan struct{}) {
	defer body.Close()
	wrunning := true
	for wrunning {
		select {
		case msg := <-conn.ConnRecv():
			if msg.Data != nil && len(msg.Data) != 0 {
				_, err := body.Write(msg.Data)
				if err != nil {
					klog.Errorf("traceid = %s  io write failed err = %v", conn.GetUid(), err)
					wrunning = false
				}
			}
			if msg.Type == util.CLOSED {
				klog.Errorf("traceid = %s io write receive close msg", conn.GetUid())
				wrunning = false
			}
		case <-stop:
			klog.Infof("traceid = %s io write stop", conn.GetUid())
			wrunning = false
		}
	}
	node.UnbindNode(conn.GetUid())
	context.GetContext().RemoveConn(conn.GetUid())
	complete <- struct{}{}
}
