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

// func TestNodeListData_SetNodeListDataByNode(t *testing.T) {
// 	nodeListData := NewNodeListData()
// 	node1 := test.BuildTestNode("node1", 1000, 2000, 9, nil)

// 	testcases := []struct {
// 		description string
// 		nodes       []v1.Node
// 		result      int
// 	}{
// 		{
// 			description: "add one node into node list",
// 			nodes:       []v1.Node{*node1},
// 			result:      1,
// 		},
// 	}

// 	for _, tc := range testcases {
// 		for _, node := range tc.nodes {
// 			t.Log("now prepare to run", tc.description)
// 			nodeListData.SetNodeListDataByNode(node)

// 			t.Log(len(nodeListData.NodeList.Items))
// 			if len(nodeListData.NodeList.Items) != tc.result {
// 				t.Fatal("unexpected result")
// 			}
// 		}
// 	}
// }

// func TestNodeListData_DeleteNodeListDataByNode(t *testing.T) {
// 	nodeListData := NewNodeListData()
// 	node1 := test.BuildTestNode("node1", 1000, 2000, 9, nil)
// 	nodeListData.SetNodeListDataByNode(*node1)
// 	if len(nodeListData.NodeList.Items) != 1 {
// 		t.Fatal("insert node into nodeList fail")
// 	}

// 	nodeListData.DeleteNodeListDataByNode(*node1)
// 	if len(nodeListData.NodeList.Items) != 0 {
// 		t.Fatal("delete node into nodeList fail")
// 	}
// }

// func TestNodeListData_CopyNodeListData(t *testing.T) {
// 	nodeListData := NewNodeListData()
// 	node1 := test.BuildTestNode("node1", 1000, 2000, 9, nil)
// 	node2 := test.BuildTestNode("node2", 1000, 2000, 9, nil)
// 	nodeListData.SetNodeListDataByNode(*node1)
// 	nodeListData.SetNodeListDataByNode(*node2)

// 	if nodeListData.GetLenListData() != 2 {
// 		t.Fatal("unexpected result, wrong insert data")
// 	}

// 	nodelist := nodeListData.CopyNodeListData()
// 	if len(nodelist) != 2 {
// 		t.Fatal("unexpected copy, wrong nodes len")
// 	}
// }
