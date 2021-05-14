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
	"context"
	"time"
)

type NodeTaskContext struct {
	ctx context.Context
}

func NewContext(ctx context.Context) *NodeTaskContext {
	return &NodeTaskContext{
		ctx,
	}
}

func (ctx *NodeTaskContext) Deadline() (deadline time.Time, ok bool) {
	return ctx.ctx.Deadline()
}

func (ctx *NodeTaskContext) Done() <-chan struct{} {
	return ctx.ctx.Done()
}

func (ctx *NodeTaskContext) Err() error {
	return ctx.ctx.Err()
}

func (ctx *NodeTaskContext) Value(key interface{}) interface{} {
	return ctx.ctx.Value(key)
}
