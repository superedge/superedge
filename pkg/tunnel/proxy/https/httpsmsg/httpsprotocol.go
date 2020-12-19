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

package httpsmsg

import (
	"fmt"
	"k8s.io/klog"
	"superedge/pkg/tunnel/context"
	"superedge/pkg/tunnel/proto"
	"superedge/pkg/tunnel/proxy/https/httpsmng"
)

func ConnectingHandler(msg *proto.StreamMsg) error {
	go httpsmng.Request(msg)
	return nil
}

func ConnectedAndTransmission(msg *proto.StreamMsg) error {
	conn := context.GetContext().GetConn(msg.Topic)
	if conn == nil {
		klog.Errorf("trace_id = %s the stream module failed to distribute the side message module = %s type = %s", msg.Topic, msg.Category, msg.Type)
		return fmt.Errorf("trace_id = %s the stream module failed to distribute the side message module = %s type = %s", msg.Topic, msg.Category, msg.Type)
	}
	conn.Send2Conn(msg)
	return nil
}
