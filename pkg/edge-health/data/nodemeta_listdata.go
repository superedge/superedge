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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"sync"
)

var (
	NodeListOnce sync.Once
	NodeListMu   sync.Mutex
	NodeList     NodeListData
)

type NodeListData struct {
	NodeList   []metav1.PartialObjectMetadata
	NodeListMu *sync.Mutex
}

func NewNodeListData() NodeListData {
	NodeListOnce.Do(func() {
		NodeList = NodeListData{
			NodeList:   make([]metav1.PartialObjectMetadata, 5),
			NodeListMu: &NodeListMu,
		}
	})
	return NodeList
}

func (n *NodeListData) SetNodeListDataByNode(node metav1.PartialObjectMetadata) {
	n.NodeListMu.Lock()
	defer n.NodeListMu.Unlock()
	var flag bool
	for k, existNode := range n.NodeList {
		if existNode.Name == node.Name {
			n.NodeList[k] = node
			flag = true
		}
	}
	if !flag {
		n.NodeList = append(n.NodeList, node)
	}
}

func (n *NodeListData) DeleteNodeListDataByNode(node metav1.PartialObjectMetadata) {
	n.NodeListMu.Lock()
	defer n.NodeListMu.Unlock()
	for k, existNode := range n.NodeList {
		if existNode.Name == node.Name {
			n.NodeList = append(n.NodeList[:k], n.NodeList[k+1:]...)
			return
		}
	}
}

func (n *NodeListData) SetNodeListDataByNodeSlice(nodeslice []runtime.Object) {
	n.NodeListMu.Lock()
	defer n.NodeListMu.Unlock()
	var nodelist []metav1.PartialObjectMetadata
	for _, v := range nodeslice {
		nodelist = append(nodelist, *(v.(*metav1.PartialObjectMetadata)))
	}
	n.NodeList = nodelist
}

func (n *NodeListData) GetLenListData() int {
	n.NodeListMu.Lock()
	defer n.NodeListMu.Unlock()
	return len(n.NodeList)
}

func (n NodeListData) CopyNodeListData() []metav1.PartialObjectMetadata {
	temp := []metav1.PartialObjectMetadata{}
	n.NodeListMu.Lock()
	defer n.NodeListMu.Unlock()
	temp = append(temp, n.NodeList...)
	return temp
}
