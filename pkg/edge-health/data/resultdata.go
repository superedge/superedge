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
	"sync"
	"time"

	"github.com/superedge/superedge/pkg/edge-health/common"
)

var (
	Result       ResultData
	ResultDataMu sync.Mutex
	ResultOnce   sync.Once
)

type ResultDetail struct {
	Normal bool   `json:"normal,omitempty"`
	Hmac   string `json:"hmac,omitempty"`
	Time   int64  `json:"time,omitempty"`
}

type ResultData struct {
	result       map[string]map[string]ResultDetail //string:checker ip string:checked ip bool: whether normal
	resultDataMu *sync.Mutex
}

func NewResultData() ResultData {
	ResultOnce.Do(func() {
		Result = ResultData{
			result:       make(map[string]map[string]ResultDetail),
			resultDataMu: &ResultDataMu,
		}
	})
	return Result
}

func (r *ResultData) SetResult(c *CommunicateData) {
	r.resultDataMu.Lock()
	defer r.resultDataMu.Unlock()

	if _, ok := r.result[c.SourceIP]; !ok {
		r.result[c.SourceIP] = make(map[string]ResultDetail)
	}
	timestamp := time.Now().UTC().Unix()
	for k, v := range c.ResultDetail { //change time be local time, different machines have different time
		v.Time = timestamp
		c.ResultDetail[k] = v
	}
	r.result[c.SourceIP] = c.ResultDetail
}

func (r *ResultData) SetResultFromCheckInfo(LocalIp, desip string, result ResultDetail) {
	r.resultDataMu.Lock()
	defer r.resultDataMu.Unlock()
	if _, ok := r.result[LocalIp]; !ok {
		r.result[LocalIp] = make(map[string]ResultDetail)
	}
	r.result[LocalIp][desip] = result
}

func (r ResultData) CopyLocalResultData(localip string) map[string]ResultDetail {
	r.resultDataMu.Lock()
	defer r.resultDataMu.Unlock()
	temp := make(map[string]ResultDetail)
	for ip, result := range r.result[localip] {
		r := ResultDetail{
			result.Normal,
			result.Hmac,
			result.Time,
		}
		temp[ip] = r
	}
	return temp
}

func (r ResultData) GetLocalResultData(localip string) map[string]ResultDetail {
	r.resultDataMu.Lock()
	defer r.resultDataMu.Unlock()
	return r.result[localip]
}

func (r ResultData) GetResultDataAll() map[string]map[string]ResultDetail {
	r.resultDataMu.Lock()
	defer r.resultDataMu.Unlock()
	return r.result
}

func (r ResultData) CopyResultDataAll() map[string]map[string]ResultDetail {
	r.resultDataMu.Lock()
	defer r.resultDataMu.Unlock()
	temp := make(map[string]map[string]ResultDetail)
	for checkerip, m := range r.result {
		temp[checkerip] = make(map[string]ResultDetail)
		for checkedip, detail := range m {
			d := ResultDetail{
				detail.Normal,
				detail.Hmac,
				detail.Time,
			}
			temp[checkerip][checkedip] = d
		}
	}
	return temp
}

func (r *ResultData) DeleteResultData(ip string) {
	r.resultDataMu.Lock()
	defer r.resultDataMu.Unlock()
	delete(r.result[common.NodeIP], ip)
	delete(r.result, ip)
}
