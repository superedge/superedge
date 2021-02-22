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

package cache

import (
	"bytes"
	"gotest.tools/assert"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/watch"
	"testing"
)

func TestWatchDecode(t *testing.T) {
	str := `
{
	"type": "MODIFIED",
	"object": {
		"kind": "Node",
		"apiVersion": "v1",
		"metadata": {
			"name": "ecm-ga2sxve4",
			"selfLink": "/api/v1/nodes/ecm-ga2sxve4",
			"uid": "92565cfd-bae3-4986-b57f-827ed297af4b",
			"resourceVersion": "605041631",
			"creationTimestamp": "2021-02-04T05:28:19Z",
			"labels": {
				"beta.kubernetes.io/arch": "amd64",
				"beta.kubernetes.io/os": "linux",
				"kubernetes.io/arch": "amd64",
				"kubernetes.io/hostname": "ecm-ga2sxve4",
				"kubernetes.io/os": "linux"
			},
			"annotations": {
				"node.alpha.kubernetes.io/ttl": "0",
				"volumes.kubernetes.io/controller-managed-attach-detach": "true"
			}
		},
		"spec": {
			"podCIDR": "172.19.1.0/24",
			"podCIDRs": ["172.19.1.0/24"]
		}
	}
}`
	buff := bytes.NewBufferString(str)

	decoder := getWatchDecoder(ioutil.NopCloser(buff))
	eventType, obj, err := decoder.Decode()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, eventType, watch.Modified)

	accessor := meta.NewAccessor()
	kind, err := accessor.Kind(obj)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, kind, "Node")
}
