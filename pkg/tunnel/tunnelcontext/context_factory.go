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
	"sync"

	"github.com/superedge/superedge/pkg/tunnel/proto"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"k8s.io/klog/v2"
)

var (
	once          sync.Once
	tunnelContext *TunnelContext
)

func GetContext() *TunnelContext {
	once.Do(func() {
		tunnelContext = &TunnelContext{}
		tunnelContext.conns = &connContext{
			conns:    make(map[string]*conn),
			connLock: sync.RWMutex{},
		}
		tunnelContext.nodes = &nodeContext{
			nodes:    make(map[string]*node),
			nodeLock: sync.RWMutex{},
		}
		tunnelContext.protocol = &protocolContext{
			protocols:    make(map[string]map[string]CallBack),
			protocolLock: sync.RWMutex{},
		}
	})
	return tunnelContext
}

func (ctx *TunnelContext) AddModule(module string) {
	klog.V(3).InfoS("starting add module", "module", module)
	ctx.protocol.AddModule(module)
	klog.V(3).InfoS(" add module successfully", "module", module)
}

func (ctx *TunnelContext) RemoveModule(module string) {
	klog.V(3).InfoS("starting remove module", "module", module)
	ctx.protocol.RemoveModule(module)
	klog.V(3).InfoS("remove module successfully", "module", module)
}

func (ctx *TunnelContext) GetHandler(typekey, category string) CallBack {
	defer klog.V(3).InfoS("get handler successfully", "category", category, "type", typekey)
	klog.V(3).InfoS("starting get handler", "category", category, "type", typekey)
	return ctx.protocol.GetHandler(typekey, category)
}

func (ctx *TunnelContext) RegisterHandler(typekey, category string, handler CallBack) {
	klog.V(3).InfoS("starting register handler", "category", category, "type", typekey)
	ctx.protocol.RegisterHandler(typekey, category, handler)
	klog.V(3).InfoS(" register handler successfully", "category", category, "type", typekey)
}

func (ctx *TunnelContext) GetNodes() []string {
	defer klog.V(3).Info(" get nodes success !")
	klog.V(3).Info("starting get nodes")
	return ctx.nodes.GetNodes()
}

func (ctx *TunnelContext) AddNode(node string) *node {
	defer klog.V(3).InfoS(" add node successfully", "node", node)
	klog.V(3).InfoS("starting add node", "node", node)
	return ctx.nodes.AddNode(node)
}

func (ctx *TunnelContext) RemoveNode(name string) {
	klog.V(3).InfoS("starting remove node", "node", name)
	node := ctx.nodes.GetNode(name)
	if node != nil {
		names := node.GetBindConns()
		if names != nil {
			conns := ctx.conns.GetConns(names)
			stopMsg := &proto.StreamMsg{
				Type: util.CLOSED,
			}
			for _, v := range conns {
				v.Send2Conn(stopMsg)
			}
		}
		ctx.nodes.RemoveNode(name)
	}
	klog.V(3).InfoS("remove node successfully", "node", name)
}

func (ctx *TunnelContext) GetNode(node string) *node {
	defer klog.V(3).InfoS("get node successfully", "node", node)
	klog.V(3).InfoS("starting get node", "node", node)
	return ctx.nodes.GetNode(node)
}

func (ctx *TunnelContext) NodeIsExist(node string) bool {
	defer klog.V(3).Infof(" check node success node: %s", node)
	klog.V(3).Infof("starting check node node: %s", node)
	return ctx.nodes.NodeIsExist(node)
}

func (ctx *TunnelContext) AddConn(uid string) *conn {
	defer klog.V(3).InfoS("add conn successfully", util.STREAM_TRACE_ID, uid)
	klog.V(3).InfoS("starting add conn", util.STREAM_TRACE_ID, uid)
	return ctx.conns.AddConn(uid)
}

func (ctx *TunnelContext) GetConn(uid string) *conn {
	defer klog.V(3).InfoS("get conn successfully", util.STREAM_TRACE_ID, uid)
	klog.V(3).InfoS("starting get conn", util.STREAM_TRACE_ID, uid)
	return ctx.conns.GetConn(uid)
}

func (ctx *TunnelContext) RemoveConn(uid string) {
	klog.V(3).InfoS("starting remove conn", util.STREAM_TRACE_ID, uid)
	ctx.conns.RemoveConn(uid)
	klog.V(3).InfoS("remove conn successfully", util.STREAM_TRACE_ID, uid)
}

func (ctx *TunnelContext) SetConn(uid string, ch chan *proto.StreamMsg) {
	klog.V(3).InfoS("starting set conn", util.STREAM_TRACE_ID, uid)
	ctx.conns.SetConn(uid, ch)
	klog.V(3).Infof("remove conn successfully", util.STREAM_TRACE_ID, uid)
}

func (ctx *TunnelContext) Handler(msg *proto.StreamMsg, key, module string) {
	f := ctx.GetHandler(key, module)
	if f != nil {
		err := f(msg)
		if err != nil {
			klog.ErrorS(err, "handler execution error", "category", msg.Category,
				"type", msg.Type, util.STREAM_TRACE_ID, msg.Topic)
		}
	}
	klog.V(3).InfoS("get handler successfully", "category", msg.Category,
		"type", msg.Type, util.STREAM_TRACE_ID, msg.Topic)
}
