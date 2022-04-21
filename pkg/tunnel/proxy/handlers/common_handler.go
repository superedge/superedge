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

package handlers

import (
	"fmt"
	"github.com/superedge/superedge/pkg/tunnel/context"
	"github.com/superedge/superedge/pkg/tunnel/proto"
	"k8s.io/klog/v2"
)

func DirectHandler(msg *proto.StreamMsg) error {
	context.GetContext().GetNode(msg.Node)
	chConn := context.GetContext().GetConn(msg.Topic)
	if chConn != nil {
		chConn.Send2Conn(msg)
	} else {
		sendFlag := false
		if localNode := context.GetContext().GetNode(msg.Node); localNode != nil {
			if remoteNodeName := localNode.GetPairNode(msg.GetTopic()); remoteNodeName != "" {
				if remoteNode := context.GetContext().GetNode(remoteNodeName); remoteNode != nil {
					remoteNode.Send2Node(msg)
					sendFlag = true
				}
			}
		}
		if !sendFlag {
			klog.Errorf("%s connection has been disconnected", msg.Category)
			return fmt.Errorf("Failed to get %s connection", msg.Category)
		}

	}
	return nil
}
