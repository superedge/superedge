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

package data

import (
	v1 "k8s.io/api/core/v1"
	"sync"
)

var (
	ConfigMapListOnce sync.Once
	ConfigMapListMu   sync.Mutex
	ConfigMapList     ConfigMapListData
)

type ConfigMapListData struct {
	ConfigMapList   v1.ConfigMapList
	ConfigMapListMu *sync.Mutex
}

func NewConfigMapListData() ConfigMapListData {
	ConfigMapListOnce.Do(func() {
		ConfigMapList = ConfigMapListData{
			ConfigMapList:   v1.ConfigMapList{},
			ConfigMapListMu: &ConfigMapListMu,
		}
	})
	return ConfigMapList
}

func (n *ConfigMapListData) SetConfigListData(cm v1.ConfigMap) {
	n.ConfigMapListMu.Lock()
	defer n.ConfigMapListMu.Unlock()
	var flag bool
	for k, existConfigMap := range n.ConfigMapList.Items {
		if existConfigMap.Name == cm.Name {
			n.ConfigMapList.Items[k] = cm
			flag = true
		}
	}
	if !flag {
		n.ConfigMapList.Items = append(n.ConfigMapList.Items, cm)
	}
}

func (n *ConfigMapListData) DeleteConfigListData(cm v1.ConfigMap) {
	n.ConfigMapListMu.Lock()
	defer n.ConfigMapListMu.Unlock()
	for k, existConfigMap := range n.ConfigMapList.Items {
		if existConfigMap.Name == cm.Name {
			n.ConfigMapList.Items = append(n.ConfigMapList.Items[:k], n.ConfigMapList.Items[k+1:]...)
			return
		}
	}
}
