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

package streammsg

import (
	"fmt"
	"github.com/superedge/superedge/pkg/tunnel/proto"
	"github.com/superedge/superedge/pkg/tunnel/tunnelcontext"
	"k8s.io/klog/v2"
)

func HeartbeatHandler(msg *proto.StreamMsg) error {
	node := tunnelcontext.GetContext().GetNode(msg.Node)
	if node == nil {
		klog.Errorf("failed to send heartbeat to edge node node: %s", msg.Node)
		return fmt.Errorf("failed to send heartbeat to edge node node: %s", msg.Node)
	}
	node.Send2Node(msg)
	return nil
}
