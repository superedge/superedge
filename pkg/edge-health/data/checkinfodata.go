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
	"sync"
)

var (
	CheckOnce       sync.Once
	CheckInfoDataMu sync.Mutex
	CheckInfoResult CheckInfoData
)

type CheckInfoData struct {
	CheckInfo       map[string]map[string]float64 //string:checked ip string:Plugin name int:check score
	CheckInfoDataMu *sync.Mutex
}

func NewCheckInfoData() CheckInfoData {
	CheckOnce.Do(func() {
		CheckInfoResult = CheckInfoData{
			CheckInfo:       make(map[string]map[string]float64),
			CheckInfoDataMu: &CheckInfoDataMu,
		}
	})
	return CheckInfoResult
}

func (c *CheckInfoData) SetCheckInfo(checkedIp, pluginName string, weight float64, score int) {
	c.CheckInfoDataMu.Lock()
	defer c.CheckInfoDataMu.Unlock()

	if _, ok := c.CheckInfo[checkedIp]; !ok {
		c.CheckInfo[checkedIp] = make(map[string]float64)
	}
	c.CheckInfo[checkedIp][pluginName] = float64(score) * weight
}

func (c *CheckInfoData) SetCheckedIpCheckInfo(checkedIp string) {
	c.CheckInfoDataMu.Lock()
	defer c.CheckInfoDataMu.Unlock()
	if _, ok := c.CheckInfo[checkedIp]; !ok {
		c.CheckInfo[checkedIp] = make(map[string]float64)
	}
}

func (c CheckInfoData) TraverseCheckedIpCheckInfo() []string {
	res := []string{}
	c.CheckInfoDataMu.Lock()
	defer c.CheckInfoDataMu.Unlock()
	for v := range c.CheckInfo {
		res = append(res, v)
	}
	return res
}

func (c *CheckInfoData) DeleteCheckedIpCheckInfo(ip string) {
	c.CheckInfoDataMu.Lock()
	defer c.CheckInfoDataMu.Unlock()
	delete(c.CheckInfo, ip)
}

func (c *CheckInfoData) GetLenCheckInfo() int {
	c.CheckInfoDataMu.Lock()
	defer c.CheckInfoDataMu.Unlock()
	return len(c.CheckInfo)
}

func (c CheckInfoData) CopyCheckInfo() map[string]map[string]float64 {
	c.CheckInfoDataMu.Lock()
	defer c.CheckInfoDataMu.Unlock()
	temp := make(map[string]map[string]float64)
	for ip, m := range c.CheckInfo {
		temp[ip] = make(map[string]float64)
		for plugin, score := range m {
			temp[ip][plugin] = score
		}
	}
	return temp
}
