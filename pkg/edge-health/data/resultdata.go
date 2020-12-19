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
	"superedge/pkg/edge-health/common"
	"sync"
	"time"
)

var (
	Result       ResultData
	ResultDataMu sync.Mutex
	ResultOnce   sync.Once
)

type ResultDetail struct {
	Normal bool
	Hamc   string
	Time   time.Time
}

type ResultData struct {
	Result       map[string]map[string]ResultDetail //string:checker ip string:checked ip bool: whether normal
	ResultDataMu *sync.Mutex
}

func NewResultData() ResultData {
	ResultOnce.Do(func() {
		Result = ResultData{
			Result:       make(map[string]map[string]ResultDetail),
			ResultDataMu: &ResultDataMu,
		}
	})
	return Result
}

func (r *ResultData) SetResult(c *CommunicateData) {
	r.ResultDataMu.Lock()
	defer r.ResultDataMu.Unlock()

	if _, ok := r.Result[c.SourceIP]; !ok {
		r.Result[c.SourceIP] = make(map[string]ResultDetail)
	}
	for k, v := range c.ResultDetail { //change time be local time, different machines have different time
		v.Time = time.Now()
		c.ResultDetail[k] = v
	}
	r.Result[c.SourceIP] = c.ResultDetail
}

func (r *ResultData) SetResultFromCheckInfo(LocalIp, desip string, result ResultDetail) {
	r.ResultDataMu.Lock()
	defer r.ResultDataMu.Unlock()
	if _, ok := r.Result[LocalIp]; !ok {
		r.Result[LocalIp] = make(map[string]ResultDetail)
	}
	r.Result[LocalIp][desip] = result
}

func (r ResultData) CopyLocalResultData(localip string) map[string]ResultDetail {
	r.ResultDataMu.Lock()
	defer r.ResultDataMu.Unlock()
	temp := make(map[string]ResultDetail)
	for ip, result := range r.Result[localip] {
		r := ResultDetail{
			result.Normal,
			result.Hamc,
			result.Time,
		}
		temp[ip] = r
	}
	return temp
}

func (r ResultData) CopyResultDataAll() map[string]map[string]ResultDetail {
	r.ResultDataMu.Lock()
	defer r.ResultDataMu.Unlock()
	temp := make(map[string]map[string]ResultDetail)
	for checkerip, m := range r.Result {
		temp[checkerip] = make(map[string]ResultDetail)
		for checkedip, detail := range m {
			d := ResultDetail{
				detail.Normal,
				detail.Hamc,
				detail.Time,
			}
			temp[checkerip][checkedip] = d
		}
	}
	return temp
}

func (r *ResultData) DeleteResultData(ip string) {
	r.ResultDataMu.Lock()
	defer r.ResultDataMu.Unlock()
	delete(r.Result[common.LocalIp], ip)
	delete(r.Result, ip)
}
