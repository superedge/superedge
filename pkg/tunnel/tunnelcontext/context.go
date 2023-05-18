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

package tunnelcontext

import (
	"context"
	"github.com/superedge/superedge/pkg/tunnel/proto"
)

type CallBack func(msg *proto.StreamMsg) error

type Protocol interface {
	GetHandler(key, module string) CallBack
	RegisterHandler(key, module string, handler CallBack)
}

type Conn interface {
	Send2Conn(msg *proto.StreamMsg)
	ConnRecv() <-chan *proto.StreamMsg
	GetUid() string
}

type ConnMng interface {
	AddConn(uid string) *conn
	GetConn(uid string) *conn
	RemoveConn(uid string)
	GetConns(names []string) []*conn
	SetConn(uid string, ch chan *proto.StreamMsg)
}

type Node interface {
	ConnectNode(category, addr string, ctx context.Context) (*conn, error)
	Send2Node(msg *proto.StreamMsg)
	BindNode(uid string)
	UnbindNode(uid string)
	NodeRecv() <-chan *proto.StreamMsg
	GetName() string
	GetBindConns() []string
	GetChan() chan *proto.StreamMsg
	AddPairNode(uid, nodeName string)
	RemovePairNode(uid string)
	GetPairNode(uid string) string
}

type NodeMng interface {
	AddNode(node string) *node
	GetNode(node string) *node
	RemoveNode(node string)
	GetNodes() []string
	NodeIsExist(node string) bool
}

type ModuleMng interface {
	AddModule(module string)
	RemoveModule(module string)
}

type ProtocolContext interface {
	Protocol
	ModuleMng
}

type TunnelContext struct {
	nodes    NodeMng
	conns    ConnMng
	protocol ProtocolContext
}
