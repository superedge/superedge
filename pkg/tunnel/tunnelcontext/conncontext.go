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
	"github.com/superedge/superedge/pkg/tunnel/proto"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"sync"
)

type connContext struct {
	conns    map[string]*conn
	connLock sync.RWMutex
}

func (entity *connContext) AddConn(uid string) *conn {
	entity.connLock.Lock()
	c := &conn{uid: uid, ch: make(chan *proto.StreamMsg, util.MSG_CHANNEL_CAP)}
	entity.conns[uid] = c
	entity.connLock.Unlock()
	return c
}

func (entity *connContext) RemoveConn(uid string) {
	entity.connLock.Lock()
	delete(entity.conns, uid)
	entity.connLock.Unlock()
}

func (entity *connContext) GetConn(uid string) *conn {
	entity.connLock.RLock()
	defer entity.connLock.RUnlock()
	return entity.conns[uid]
}

func (entity *connContext) GetConns(names []string) []*conn {
	entity.connLock.RLock()
	defer entity.connLock.RUnlock()
	var conns []*conn
	for _, v := range names {
		conn := entity.conns[v]
		if conn != nil {
			conns = append(conns, conn)
		}
	}
	return conns
}

func (entity *connContext) SetConn(uid string, ch chan *proto.StreamMsg) {
	entity.connLock.Lock()
	defer entity.connLock.Unlock()
	entity.conns[uid] = &conn{
		uid,
		ch,
	}
}
