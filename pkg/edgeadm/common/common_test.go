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

package common

import (
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestReadYaml(t *testing.T) {
	tests := []struct {
		inputPath string
		defaults  string
		expected  string
	}{
		{
			inputPath: "/tmp/notexistpath/notexistpath",
			defaults:  "test",
			expected:  "test",
		},
		{
			inputPath: "../../../test/testdata/tunnel-coredns.yaml",
			defaults:  "test",
			expected:  "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: tunnel-coredns\ndata:\n  Corefile: |\n    .:53 {\n        errors\n        health {\n           lameduck 5s\n        }\n        hosts /etc/edge/hosts {\n            reload 300ms\n            fallthrough\n        }\n        ready\n        prometheus :9153\n        forward . /etc/resolv.conf\n        cache 30\n        reload 2s\n        loadbalance\n    }",
		},
	}

	for _, tc := range tests {
		got := ReadYaml(tc.inputPath, tc.defaults)
		if got != tc.expected {
			t.Fatal("expect is, actual is", tc.expected, got)
		}
	}
}

func TestEnsureEdgexNamespace(t *testing.T) {
	kubeClient := fake.NewSimpleClientset()
	err := EnsureEdgexNamespace(kubeClient)
	if err != nil {
		t.Fatal("unexpected err", err)
	}
}
