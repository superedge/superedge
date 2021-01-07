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

package options

import (
	"encoding/json"
	"gotest.tools/assert"
	"testing"

	"github.com/superedge/superedge/pkg/lite-apiserver/config"
)

func TestJson(t *testing.T) {
	s := `
[{"key":"key1","cert":"cert1"},{"key":"key2","cert":"cert2"}]
`
	n := []config.TLSKeyPair{}
	err := json.Unmarshal([]byte(s), &n)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, n[1].KeyPath, "key2")

}
