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

package data

import (
	"github.com/superedge/superedge/test"
	v1 "k8s.io/api/core/v1"
	"testing"
)

func TestConfigMapListData_SetConfigListData(t *testing.T) {
	configMapListData := NewConfigMapListData()

	cm1 := test.BuildTestConfigmap("cm1", "default", map[string]string{"username": "admin"})
	testcases := []struct {
		description string
		obj         v1.ConfigMap
		expecteNum  int
	}{
		{
			description: "add one configmap",
			obj:         *cm1,
			expecteNum:  1,
		},
	}
	for _, tc := range testcases {
		configMapListData.SetConfigListData(tc.obj)
		if len(configMapListData.ConfigMapList.Items) != tc.expecteNum {
			t.Fatal("unexpected result , should set configmap to ConfigMapList")
		}
	}

}

func TestConfigMapListData_DeleteConfigListData(t *testing.T) {
	configMapListData := NewConfigMapListData()
	cm1 := test.BuildTestConfigmap("cm1", "default", map[string]string{"username": "admin"})
	cm2 := test.BuildTestConfigmap("cm2", "default", map[string]string{"password": "passwd"})
	configMapListData.SetConfigListData(*cm1)
	configMapListData.SetConfigListData(*cm2)
	if len(configMapListData.ConfigMapList.Items) != 2 {
		t.Fatal("unexpected result , should set configmap to ConfigMapList")
	}
	configMapListData.DeleteConfigListData(*cm2)
	if len(configMapListData.ConfigMapList.Items) != 1 {
		t.Fatal("unexpected result , should delete configmap success")
	}

}
