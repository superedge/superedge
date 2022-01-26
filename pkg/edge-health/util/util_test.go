/*
Copyright 2022 The SuperEdge Authors.

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

package util

import (
	"github.com/superedge/superedge/test"
	"k8s.io/api/core/v1"
	"testing"
)

func TestGetHmacCode(t *testing.T) {
	testcases := []struct {
		parastring string
		parakey    string
		expect     string
		error      error
	}{
		{
			parastring: "abcde",
			parakey:    "12345",
			expect:     "f8f78e4c506669fdd116876d08adf809ed7c33585078dcfd4e6fe863b7ea966a",
			error:      nil,
		},
		{
			parastring: "helloworld",
			parakey:    "passkey",
			expect:     "b33350ac6988d60bf85739f71f73df3b2252a20ee650d271e34ca5799b8a5334",
			error:      nil,
		},
	}

	for _, tc := range testcases {
		result, err := GetHmacCode(tc.parastring, tc.parakey)

		if err != tc.error || result != tc.expect {
			t.Fatal("unexpected error", err)
		}
	}

}

func TestGetNodeNameByIp(t *testing.T) {
	node1 := test.BuildTestNode("node1", 1000, 2000, 9, nil)
	node2 := test.BuildTestNode("node2", 1000, 2000, 9, nil)
	nodelist := []v1.Node{*node1, *node2}

	result := GetNodeNameByIp(nodelist, "10.0.0.1")
	if result != "" {
		t.Fatal("unexpected err")
	}
}
