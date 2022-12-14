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

package data

// import (
// 	"k8s.io/api/core/v1"
// 	"sync"
// )

// var (
// 	NodeListOnce sync.Once
// 	NodeListMu   sync.Mutex
// 	NodeList     NodeListData
// )

// type NodeListData struct {
// 	NodeList   v1.NodeList
// 	NodeListMu *sync.Mutex
// }

// func NewNodeListData() NodeListData {
// 	NodeListOnce.Do(func() {
// 		NodeList = NodeListData{
// 			NodeList:   v1.NodeList{},
// 			NodeListMu: &NodeListMu,
// 		}
// 	})
// 	return NodeList
// }

// func (n *NodeListData) SetNodeListDataByNode(node v1.Node) {
// 	n.NodeListMu.Lock()
// 	defer n.NodeListMu.Unlock()
// 	var flag bool
// 	for k, existNode := range n.NodeList.Items {
// 		if existNode.Name == node.Name {
// 			n.NodeList.Items[k] = node
// 			flag = true
// 		}
// 	}
// 	if !flag {
// 		n.NodeList.Items = append(n.NodeList.Items, node)
// 	}
// }

// func (n *NodeListData) DeleteNodeListDataByNode(node v1.Node) {
// 	n.NodeListMu.Lock()
// 	defer n.NodeListMu.Unlock()
// 	for k, existNode := range n.NodeList.Items {
// 		if existNode.Name == node.Name {
// 			n.NodeList.Items = append(n.NodeList.Items[:k], n.NodeList.Items[k+1:]...)
// 			return
// 		}
// 	}
// }

// func (n *NodeListData) SetNodeListDataByNodeSlice(nodeslice []*v1.Node) {
// 	n.NodeListMu.Lock()
// 	defer n.NodeListMu.Unlock()
// 	var nodelist []v1.Node
// 	for _, v := range nodeslice {
// 		nodelist = append(nodelist, *v)
// 	}
// 	n.NodeList.Items = nodelist
// }

// func (n *NodeListData) GetLenListData() int {
// 	n.NodeListMu.Lock()
// 	defer n.NodeListMu.Unlock()
// 	return len(n.NodeList.Items)
// }

// func (n NodeListData) CopyNodeListData() []v1.Node {
// 	temp := []v1.Node{}
// 	n.NodeListMu.Lock()
// 	defer n.NodeListMu.Unlock()
// 	temp = append(temp, n.NodeList.Items...)
// 	return temp
// }
