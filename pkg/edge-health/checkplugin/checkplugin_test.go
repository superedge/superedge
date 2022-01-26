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
	"net/http"
	"testing"
	"time"
)

func TestBasePlugin_SetWeight(t *testing.T) {
	testcases := []struct {
		weight         float64
		expectedweight float64
	}{
		{
			weight:         1.5,
			expectedweight: 1.5,
		},
		{
			weight:         3.0,
			expectedweight: 3.0,
		},
	}
	for _, tc := range testcases {
		baseplugin := NewBasePlugin(5, 5, 80, tc.weight, "test")
		setvalue := tc.weight + float64(3)
		baseplugin.SetWeight(setvalue)
		if baseplugin.GetWeight() != tc.expectedweight {
			t.Fatal("unexpected error, actual weight is ", baseplugin.GetWeight())
		}
	}
}

func TestPingDo(t *testing.T) {
	testcases := []struct {
		description string
		url         string
		result      bool
	}{
		{
			description: "request github",
			url:         "https://github.com/",
			result:      true,
		},
		{
			description: "request url which not exist",
			url:         "http://notexistrealdomainip/notrealpath",
			result:      false,
		},
	}
	client := http.Client{Timeout: time.Duration(10) * time.Second}

	for _, tc := range testcases {
		t.Log("now prepare to run", tc.description)
		req, err := http.NewRequest("HEAD", tc.url, nil)
		if err != nil {
			t.Fatal("unexpected reuqest")
		}

		result, err := PingDo(client, req)
		if result != tc.result {
			t.Fatal("unexpected error, actual result and err is", tc.result, err)
		}
	}
}
