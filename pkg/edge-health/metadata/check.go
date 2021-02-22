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

package metadata

import (
	"sync"
	"time"
)

// TODO: more efficient data structures
type CheckDetail struct {
	Normal bool
	Time   time.Time
}

type CommunInfo struct {
	SourceIP    string                 // ClientIP，Checker ip
	CheckDetail map[string]CheckDetail // Checked ip:Check detail
	Hmac        string
}

type CheckMetadata struct {
	CheckInfo            map[string]map[string]CheckDetail // Checker ip:{Checked ip:Check detail}
	CheckPluginScoreInfo map[string]map[string]float64     // Checked ip:{Plugin name:Check score}
	sync.RWMutex
}

// CheckPluginScoreInfo relevant functions
func (cm *CheckMetadata) SetByPluginScore(checkedIp, pluginName string, weight float64, score int) {
	cm.Lock()
	defer cm.Unlock()

	if _, existed := cm.CheckPluginScoreInfo[checkedIp]; !existed {
		cm.CheckPluginScoreInfo[checkedIp] = make(map[string]float64)
	}
	cm.CheckPluginScoreInfo[checkedIp][pluginName] = float64(score) * weight
}

func (cm *CheckMetadata) InitCheckPluginScore(checkedIp string) {
	cm.Lock()
	defer cm.Unlock()
	if _, existed := cm.CheckPluginScoreInfo[checkedIp]; !existed {
		cm.CheckPluginScoreInfo[checkedIp] = make(map[string]float64)
	}
}

func (cm *CheckMetadata) CopyCheckedIp() []string {
	cm.RLock()
	defer cm.RUnlock()
	var checkedIps []string
	for checkedIp := range cm.CheckPluginScoreInfo {
		checkedIps = append(checkedIps, checkedIp)
	}
	return checkedIps
}

func (cm *CheckMetadata) CopyCheckPluginScore() map[string]map[string]float64 {
	cm.RLock()
	defer cm.RUnlock()
	copyScores := make(map[string]map[string]float64)
	for checkedIp, pluginScores := range cm.CheckPluginScoreInfo {
		copyScores[checkedIp] = make(map[string]float64)
		for plugin, score := range pluginScores {
			copyScores[checkedIp][plugin] = score
		}
	}
	return copyScores
}

func (cm *CheckMetadata) DeleteCheckPluginScore(checkedIp string) {
	cm.Lock()
	defer cm.Unlock()
	delete(cm.CheckPluginScoreInfo, checkedIp)
}

// CheckInfo relevant functions
func (cm *CheckMetadata) SetByCommunInfo(c CommunInfo) {
	cm.Lock()
	defer cm.Unlock()

	if _, existed := cm.CheckInfo[c.SourceIP]; !existed {
		cm.CheckInfo[c.SourceIP] = make(map[string]CheckDetail)
	}
	for k, detail := range c.CheckDetail {
		// Update time to local timestamp since different machines have different ones
		detail.Time = time.Now()
		c.CheckDetail[k] = detail
	}
	cm.CheckInfo[c.SourceIP] = c.CheckDetail
}

func (cm *CheckMetadata) SetByCheckDetail(LocalIp, dstIp string, checkDetail CheckDetail) {
	cm.Lock()
	defer cm.Unlock()
	if _, existed := cm.CheckInfo[LocalIp]; !existed {
		cm.CheckInfo[LocalIp] = make(map[string]CheckDetail)
	}
	checkDetail.Time = time.Now()
	cm.CheckInfo[LocalIp][dstIp] = checkDetail
}

func (cm *CheckMetadata) CopyLocal(localIp string) map[string]CheckDetail {
	cm.RLock()
	defer cm.RUnlock()
	localCheckMetadata := make(map[string]CheckDetail)
	for ip, detail := range cm.CheckInfo[localIp] {
		localCheckMetadata[ip] = CheckDetail{
			detail.Normal,
			detail.Time,
		}
	}
	return localCheckMetadata
}

func (cm *CheckMetadata) CopyAll() map[string]map[string]CheckDetail {
	cm.RLock()
	defer cm.RUnlock()
	copyCheckInfo := make(map[string]map[string]CheckDetail)
	for checkerIp, checkedDetails := range cm.CheckInfo {
		copyCheckInfo[checkerIp] = make(map[string]CheckDetail)
		for checkedIp, detail := range checkedDetails {
			copyCheckInfo[checkedIp][checkedIp] = CheckDetail{
				detail.Normal,
				detail.Time,
			}
		}
	}
	return copyCheckInfo
}

func (cm *CheckMetadata) DeleteByIp(localIp, ip string) {
	cm.Lock()
	defer cm.Unlock()
	delete(cm.CheckInfo[localIp], ip)
	delete(cm.CheckInfo, ip)
}
