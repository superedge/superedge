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

package controller

import (
	"testing"
)

func TestCheckifcontains(t *testing.T) {
	testcases := []struct {
		nodelabel     map[string]string
		keyslices     []string
		expectinclude bool
		expectresult  string
	}{
		{
			nodelabel: map[string]string{
				"localzone": "eu-west-1",
				"usage":     "edge",
			},
			keyslices:     []string{"localzone"},
			expectinclude: true,
			expectresult:  "eu-west-1",
		},
		{
			nodelabel: map[string]string{
				"localzone":              "eu-west-2",
				"kubernetes.io/arch":     "arm64",
				"kubernetes.io/hostname": "vm-111-centos",
				"kubernetes.io/os":       "linux",
			},
			keyslices:     []string{"kubernetes.io/hostname", "kubernetes.io/os"},
			expectinclude: true,
			expectresult:  "vm-111-centos-linux",
		},
		{
			nodelabel: map[string]string{
				"localzone":          "eu-west-2",
				"kubernetes.io/arch": "arm64",
			},
			keyslices:     []string{"beta.kubernetes.io/os"},
			expectinclude: false,
			expectresult:  "",
		},
	}
	for _, tc := range testcases {
		include, result, _ := checkIfContains(tc.nodelabel, tc.keyslices)
		if include != tc.expectinclude || result != tc.expectresult {
			t.Fatal("not as expected", include, tc.expectinclude, result, tc.expectresult)
		}
	}
}
