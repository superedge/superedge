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
	"fmt"
	"sync"
	"time"

	"github.com/superedge/superedge/pkg/tunnel/proto"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"k8s.io/klog/v2"
)

const (
	CONNECT_REQ       = "connecting"
	CONNECT_FAILED    = "connect-failed"
	CONNECT_SUCCESSED = "connected"
)

type node struct {
	name      string
	ch        chan *proto.StreamMsg
	conns     []string
	connsLock sync.RWMutex
	pairnodes map[string]string
	nodesLock sync.RWMutex
}

func (edge *node) BindNode(uuid string) {
	edge.connsLock.Lock()
	if edge.conns == nil {
		edge.conns = []string{uuid}
	} else {
		edge.conns = append(edge.conns, uuid)
	}
	edge.connsLock.Unlock()
}

func (edge *node) UnbindNode(uuid string) {
	// 删除连接绑定
	edge.connsLock.Lock()
	for k, v := range edge.conns {
		if v == uuid {
			edge.conns = append(edge.conns[:k], edge.conns[k+1:len(edge.conns)]...)
		}
	}
	edge.connsLock.Unlock()
	// 删除节点绑定
	edge.nodesLock.Lock()
	delete(edge.pairnodes, uuid)
	edge.nodesLock.Unlock()
}

func (edge *node) ConnectNode(category, addr string, ctx context.Context) (*conn, error) {
	uid := ctx.Value(util.STREAM_TRACE_ID).(string)
	subCtx, cancle := context.WithTimeout(ctx, 5*time.Second)
	defer cancle()
	var err error
	conn := tunnelContext.AddConn(uid)
	edge.BindNode(uid)
	defer func() {
		if err != nil {
			edge.UnbindNode(uid)
			tunnelContext.RemoveConn(uid)
		}
	}()
	edge.Send2Node(&proto.StreamMsg{
		Node:     edge.name,
		Category: category,
		Type:     CONNECT_REQ,
		Topic:    uid,
		Data:     []byte(fmt.Sprintf("CONNECT %s HTTP/1.1\r\n\r\n\r\n", addr)),
		Addr:     addr,
	})
	select {
	case resp := <-conn.ch:
		if resp.Type == CONNECT_SUCCESSED {
			return conn, err
		}
		if resp.Type == CONNECT_FAILED {
			err = fmt.Errorf("failed to connect target server %s, error:%s, %s:%s", addr, string(resp.Data), util.STREAM_TRACE_ID, uid)
			return conn, err
		}
	case <-subCtx.Done():
		err = fmt.Errorf("the connection to the target server %s of the edge node %s timed out, %s:%s", addr, edge.name, util.STREAM_TRACE_ID, uid)
		return conn, err
	}
	return conn, err
}

func (edge *node) Send2Node(msg *proto.StreamMsg) {
	klog.V(3).InfoS("node send msg", "nodeName", edge.name, "category", msg.GetCategory(),
		"type", msg.GetType(), util.STREAM_TRACE_ID, msg.GetTopic())
	edge.ch <- msg
}

func (edge *node) NodeRecv() <-chan *proto.StreamMsg {
	return edge.ch
}

func (edge *node) GetName() string {
	return edge.name
}

func (edge *node) GetBindConns() []string {
	edge.connsLock.RLock()
	defer edge.connsLock.RUnlock()
	if edge.conns == nil {
		return nil
	}
	return edge.conns
}
func (edge *node) GetChan() chan *proto.StreamMsg {
	return edge.ch
}

func (edge *node) AddPairNode(uid, nodeName string) {
	edge.nodesLock.Lock()
	defer edge.nodesLock.Unlock()
	edge.pairnodes[uid] = nodeName
}
func (edge *node) RemovePairNode(uid string) {
	edge.nodesLock.Lock()
	defer edge.nodesLock.Unlock()
	delete(edge.pairnodes, uid)
}

func (edge *node) GetPairNode(uid string) string {
	edge.nodesLock.RLock()
	defer edge.nodesLock.RUnlock()
	return edge.pairnodes[uid]
}
