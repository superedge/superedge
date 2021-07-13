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

package context

import (
	"github.com/superedge/superedge/pkg/tunnel/proto"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"k8s.io/klog/v2"
	"sync"
)

var (
	once    sync.Once
	context *Context
)

func GetContext() *Context {
	once.Do(func() {
		context = &Context{}
		context.conns = &connContext{
			conns:    make(map[string]*conn),
			connLock: sync.RWMutex{},
		}
		context.nodes = &nodeContext{
			nodes:    make(map[string]*node),
			nodeLock: sync.RWMutex{},
		}
		context.protocol = &protocolContext{
			protocols:    make(map[string]map[string]CallBack),
			protocolLock: sync.RWMutex{},
		}
	})
	return context
}

func (ctx *Context) AddModule(module string) {
	klog.V(8).Infof("starting add module: %s", module)
	ctx.protocol.AddModule(module)
	klog.V(8).Infof(" add module: %s success !", module)
}

func (ctx *Context) RemoveModule(module string) {
	klog.V(8).Infof("starting remove module: %s", module)
	ctx.protocol.RemoveModule(module)
	klog.V(8).Infof("remove module: %s success !", module)
}

func (ctx *Context) GetHandler(key, module string) CallBack {
	defer klog.V(8).Infof("get handler success key: %s module: %s", key, module)
	klog.V(8).Infof("starting get handler  key: %s module: %s", key, module)
	return ctx.protocol.GetHandler(key, module)
}

func (ctx *Context) RegisterHandler(key, module string, handler CallBack) {
	klog.V(8).Infof("starting register handler key: %s value: %s", key, module)
	ctx.protocol.RegisterHandler(key, module, handler)
	klog.V(8).Infof(" register handler success key: %s value: %s", key, module)
}

func (ctx *Context) GetNodes() []string {
	defer klog.V(8).Info(" get nodes success !")
	klog.V(8).Info("starting get nodes")
	return ctx.nodes.GetNodes()
}

func (ctx *Context) AddNode(node string) *node {
	defer klog.V(8).Infof(" add node success node: %s", node)
	klog.V(8).Infof("starting add node node: %s", node)
	return ctx.nodes.AddNode(node)
}

func (ctx *Context) RemoveNode(name string) {
	klog.V(8).Infof("starting remove node node: %s", name)
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
	klog.V(8).Infof("remove node success node: %s", name)
}

func (ctx *Context) GetNode(node string) *node {
	defer klog.V(8).Infof("get node success node: %s", node)
	klog.V(8).Infof("starting get node node: %s", node)
	return ctx.nodes.GetNode(node)
}

func (ctx *Context) NodeIsExist(node string) bool {
	defer klog.V(8).Infof(" check node success node: %s", node)
	klog.V(8).Infof("starting check node node: %s", node)
	return ctx.nodes.NodeIsExist(node)
}

func (ctx *Context) AddConn(uid string) *conn {
	defer klog.V(8).Infof(" add conn success uuid: %s", uid)
	klog.V(8).Infof("starting add conn uuid: %s", uid)
	return ctx.conns.AddConn(uid)
}

func (ctx *Context) GetConn(uid string) *conn {
	defer klog.V(8).Infof("get conn success uuid: %s", uid)
	klog.V(8).Infof("starting get conn uuid: %s", uid)
	return ctx.conns.GetConn(uid)
}

func (ctx *Context) RemoveConn(uid string) {
	klog.V(8).Infof("starting remove conn uuid: %s", uid)
	ctx.conns.RemoveConn(uid)
	klog.V(8).Infof("remove conn success uuid: %s", uid)
}

func (ctx *Context) Handler(msg *proto.StreamMsg, key, module string) {
	f := ctx.GetHandler(key, module)
	if f != nil {
		err := f(msg)
		if err != nil {
			klog.Errorf("handler execution error err = %v", err)
		}
	}
}
