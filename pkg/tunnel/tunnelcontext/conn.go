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

import "github.com/superedge/superedge/pkg/tunnel/proto"

type conn struct {
	uid string
	ch  chan *proto.StreamMsg
}

func (c *conn) Send2Conn(msg *proto.StreamMsg) {
	c.ch <- msg
}

func (c *conn) ConnRecv() <-chan *proto.StreamMsg {
	return c.ch
}
func (c *conn) GetUid() string {
	return c.uid
}
