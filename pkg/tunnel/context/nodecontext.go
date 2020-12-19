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
	"superedge/pkg/tunnel/proto"
	"superedge/pkg/tunnel/util"
	"sync"
)

type nodeContext struct {
	nodes    map[string]*node
	nodeLock sync.RWMutex
}

func (entity *nodeContext) AddNode(name string) *node {
	entity.nodeLock.Lock()
	defer entity.nodeLock.Unlock()
	edge := &node{
		ch:        make(chan *proto.StreamMsg, util.MSG_CHANNEL_CAP),
		connsLock: sync.RWMutex{},
		name:      name,
	}
	entity.nodes[name] = edge
	return edge
}

func (entity *nodeContext) GetNode(name string) *node {
	entity.nodeLock.Lock()
	defer entity.nodeLock.Unlock()
	return entity.nodes[name]
}

func (entity *nodeContext) RemoveNode(name string) {
	entity.nodeLock.Lock()
	defer entity.nodeLock.Unlock()
	delete(entity.nodes, name)
}

func (entity *nodeContext) GetNodes() []string {
	entity.nodeLock.RLock()
	defer entity.nodeLock.RUnlock()
	var nodes []string
	for k := range entity.nodes {
		nodes = append(nodes, k)
	}
	return nodes
}
func (entity *nodeContext) NodeIsExist(node string) bool {
	entity.nodeLock.RLock()
	defer entity.nodeLock.RUnlock()
	_, ok := entity.nodes[node]
	return ok
}
