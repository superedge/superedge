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

package tcpmng

import (
	"superedge/pkg/tunnel/context"
	"superedge/pkg/tunnel/proto"
	"superedge/pkg/tunnel/util"
	"io"
	"k8s.io/klog"
	"net"
	"sync"
)

type TcpConn struct {
	Conn     net.Conn
	uid      string
	stopChan chan struct{}
	Type     string
	C        context.Conn
	n        context.Node
	Addr     string
	once     sync.Once
}

func NewTcpConn(uuid, addr, node string) *TcpConn {
	tcp := &TcpConn{
		uid:      uuid,
		stopChan: make(chan struct{}, 1),
		C:        context.GetContext().AddConn(uuid),
		Addr:     addr,
		n:        context.GetContext().AddNode(node),
	}
	tcp.n.BindNode(uuid)
	return tcp
}

func (tcp *TcpConn) Write() {
	running := true
	for running {
		select {
		case msg := <-tcp.C.ConnRecv():
			if msg.Type == util.TCP_CONTROL {
				tcp.cleanUp()
				break
			}
			_, err := tcp.Conn.Write(msg.Data)
			if err != nil {
				klog.Errorf("write conn fail err = %v", err)
				tcp.cleanUp()
				break
			}
		case <-tcp.stopChan:
			klog.Info("disconnect tcp and stop sending !")
			tcp.Conn.Close()
			tcp.closeOpposite()
			running = false
		}
	}
}

func (tcp *TcpConn) Read() {
	running := true
	for running {
		select {
		case <-tcp.stopChan:
			klog.Info("Disconnect tcp and stop receiving !")
			tcp.Conn.Close()
			running = false
		default:
			size := 32 * 1024
			if l, ok := interface{}(tcp.Conn).(*io.LimitedReader); ok && int64(size) > l.N {
				if l.N < 1 {
					size = 1
				} else {
					size = int(l.N)
				}
			}
			buf := make([]byte, size)
			n, err := tcp.Conn.Read(buf)
			if err != nil {
				klog.Errorf("conn read fail，err = %s ", err)
				tcp.cleanUp()
				break
			}
			tcp.n.Send2Node(&proto.StreamMsg{
				Category: util.TCP,
				Type:     tcp.Type,
				Topic:    tcp.uid,
				Data:     buf[0:n],
				Addr:     tcp.Addr,
				Node:     tcp.n.GetName(),
			})
			if err != nil {
				klog.Errorf("tcp conn failed to send a message to the node err = %v", err)
				running = false
				break
			}
		}
	}
}

func (tcp *TcpConn) cleanUp() {
	tcp.stopChan <- struct{}{}
}
func (tcp *TcpConn) closeOpposite() {
	tcp.once.Do(func() {
		tcp.n.Send2Node(&proto.StreamMsg{
			Category: util.TCP,
			Type:     util.TCP_CONTROL,
			Topic:    tcp.uid,
			Data:     []byte{},
			Addr:     tcp.Addr,
		})
		tcp.n.UnbindNode(tcp.uid)

		context.GetContext().RemoveConn(tcp.uid)
	})
}
