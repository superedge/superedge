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

package metadata

import (
	"k8s.io/api/core/v1"
	"sync"
)

// TODO: more efficient data structures
type NodeMetadata struct {
	NodeList []v1.Node
	sync.RWMutex
}

func (nm *NodeMetadata) AddOrUpdateByNode(node *v1.Node) {
	nm.Lock()
	defer nm.Unlock()
	var flag bool
	for k, existedNode := range nm.NodeList {
		if existedNode.Name == node.Name {
			nm.NodeList[k] = *node
			flag = true
			break
		}
	}
	// Append when it does not exist
	if !flag {
		nm.NodeList = append(nm.NodeList, *node)
	}
	return
}

func (nm *NodeMetadata) DeleteByNode(node *v1.Node) {
	nm.Lock()
	defer nm.Unlock()
	for k, existedNode := range nm.NodeList {
		if existedNode.Name == node.Name {
			nm.NodeList = append(nm.NodeList[:k], nm.NodeList[k+1:]...)
			break
		}
	}
	return
}

func (nm *NodeMetadata) GetNodeByName(hostname string) *v1.Node {
	nm.RLock()
	defer nm.RUnlock()
	for _, existedNode := range nm.NodeList {
		if existedNode.Name == hostname {
			tmpNode := existedNode
			return &tmpNode
		}
	}
	return nil
}

func (nm *NodeMetadata) GetNodeByAddr(ip string) *v1.Node {
	nm.RLock()
	defer nm.RUnlock()
	for _, existedNode := range nm.NodeList {
		for _, addr := range existedNode.Status.Addresses {
			if addr.Type == v1.NodeInternalIP && addr.Address == ip {
				tmpNode := existedNode
				return &tmpNode
			}
		}
	}
	return nil
}

func (nm *NodeMetadata) SetByNodeList(nodeList []*v1.Node) {
	nm.Lock()
	defer nm.Unlock()
	var nodes []v1.Node
	for _, node := range nodeList {
		nodes = append(nodes, *node)
	}
	nm.NodeList = nodes
}

func (nm *NodeMetadata) GetLen() int {
	nm.RLock()
	defer nm.RUnlock()
	return len(nm.NodeList)
}

func (nm *NodeMetadata) Copy() []v1.Node {
	nm.RLock()
	defer nm.RUnlock()
	var nodes []v1.Node
	nodes = append(nodes, nm.NodeList...)
	return nodes
}
