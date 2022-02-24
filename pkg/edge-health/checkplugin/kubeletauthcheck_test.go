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

package checkplugin

import (
	"testing"
)

func TestKubeletCheckPlugin_Set(t *testing.T) {
	k := KubeletAuthCheckPlugin{}

	testcases := []struct {
		description               string
		parameter                 string
		result                    int
		expectedtimeout           int
		expecteretrytimeretrytime int
		error                     error
	}{
		{
			description: "wrong parameter format",
			parameter:   " ",
			result:      1,
			error:       nil,
		},
		{
			description: "wrong timeout",
			parameter:   "time=5",
			result:      1,
			error:       nil,
		},
		{
			description:               "pass timeout, retrytime and weight",
			parameter:                 "timeout=5, retrytime=10, weight=1.0",
			result:                    1,
			expectedtimeout:           5,
			expecteretrytimeretrytime: 10,
			error:                     nil,
		},
	}
	for _, tc := range testcases {
		t.Log("prepare to run", tc.description)

		err := k.Set(tc.parameter)
		if err != tc.error || k.HealthCheckoutTimeOut != tc.expectedtimeout || k.HealthCheckRetryTime != tc.expecteretrytimeretrytime {
			t.Fatal("unexpected err", err)
		}
	}
}
