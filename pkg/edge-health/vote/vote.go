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

package vote

import (
	"context"
	"fmt"
	"time"

	"github.com/superedge/superedge/pkg/edge-health/check"
	"github.com/superedge/superedge/pkg/edge-health/common"
	"github.com/superedge/superedge/pkg/edge-health/data"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	log "k8s.io/klog/v2"
)

const (
	NodeAnnotationPatchAddUpdateTemplate = `{"metadata":{"annotations":{"%s":"%s"}}}`
	NodeAnnotationPatchDeleteTemplate    = `{"metadata":{"annotations":{"%s":null}}}`
)

var UnreachNoExecuteTaint = &corev1.Taint{
	Key:    corev1.TaintNodeUnreachable,
	Effect: corev1.TaintEffectNoExecute,
}

type Vote interface {
	Vote()
	GetVotePeriod() int
	GetVoteTimeout() int
}

type VoteEdge struct {
	VoteTimeOut int
	VotePeriod  int
}

func NewVoteEdge(voteTimeOut, votePeriod int) Vote {
	return VoteEdge{
		VoteTimeOut: voteTimeOut,
		VotePeriod:  votePeriod,
	}
}

func (vote VoteEdge) GetVoteTimeout() int {
	return vote.VoteTimeOut
}

func (vote VoteEdge) GetVotePeriod() int {
	return vote.VotePeriod
}

func (vote VoteEdge) Vote() {
	voteCountMap := make(map[string]map[string]int) // {"a":{"yes":1,"no":2}}
	healthNodeMap := make(map[string]string)

	tempNodeStatus := data.Result.CopyResultDataAll() //map[string]map[string]ResultDetail string:checker ip string:checked ip bool:noraml
	nowUTC := time.Now().UTC()
	for k, v := range tempNodeStatus { //k is checker ip
		for ip, resultdetail := range v { //ip is checked ip
			rtime := time.Unix(resultdetail.Time, 0).UTC()
			if k == common.NodeIP || (k != common.NodeIP && !nowUTC.After(rtime.Add(time.Duration(vote.GetVoteTimeout())*time.Second))) {
				healthNodeMap[k] = "" //node is a health node if it has at least one valid check
				if _, ok := voteCountMap[ip]; !ok {
					voteCountMap[ip] = make(map[string]int)
				}
				if resultdetail.Normal {
					if _, ok := voteCountMap[ip]["yes"]; !ok {
						voteCountMap[ip]["yes"] = 0
					}
					voteCountMap[ip]["yes"] += 1
				} else {
					if _, ok := voteCountMap[ip]["no"]; !ok {
						voteCountMap[ip]["no"] = 0
					}
					voteCountMap[ip]["no"] += 1
				}
			}
		}
	}
	log.V(4).Infof("Vote: healthNodeMap is %v , voteCountMap is %v", healthNodeMap, voteCountMap)

	//num := (float64(len(healthNodeMap)) + 1) / 2
	num := (float64(data.CheckInfoResult.GetLenCheckInfo()) + 1) / 2

	if len(healthNodeMap) == 1 {
		return
	}
	for ip, v := range voteCountMap {
		if _, ok := v["yes"]; ok {
			if float64(v["yes"]) >= num {
				log.V(4).Infof("vote: vote yes to master begin")
				name, err := check.PodManager.GetNodeNameByNodeIP(ip)
				if err != nil {
					log.ErrorS(err, "GetNodeNameByNodeIP error")
					continue
				}
				if nodeObject, err := check.NodeMetaManager.NodeMetaILister.Get(name); err == nil && name != "" {
					node := nodeObject.(*metav1.PartialObjectMetadata)
					if _, ok := node.Annotations["nodeunhealth"]; ok {
						// delete nodeunhealth annotation
						patchBytes := []byte(fmt.Sprintf(NodeAnnotationPatchDeleteTemplate, "nodeunhealth"))
						if _, err := common.ClientSet.CoreV1().Nodes().Patch(context.TODO(), node.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{}); err != nil {
							log.Errorf("patch yes vote to master error: %v ", err)
						} else {
							log.V(2).Infof("patch yes vote of %s to master", node.Name)
						}
					}
				}
			}
		}
		if _, ok := v["no"]; ok {
			if float64(v["no"]) >= num {
				log.V(4).Infof("vote: vote no to master begin")
				name, err := check.PodManager.GetNodeNameByNodeIP(ip)
				if err != nil {
					log.ErrorS(err, "GetNodeNameByNodeIP error")
					continue
				}
				if nodeObject, err := check.NodeMetaManager.NodeMetaILister.Get(name); err == nil && name != "" {
					node := nodeObject.(*metav1.PartialObjectMetadata)
					if _, ok := node.Annotations["nodeunhealth"]; !ok {

						patchBytes := []byte(fmt.Sprintf(NodeAnnotationPatchAddUpdateTemplate, "nodeunhealth", "yes"))
						if _, err := common.ClientSet.CoreV1().Nodes().Patch(context.TODO(), node.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{}); err != nil {
							log.Errorf("patch no vote to master error: %v ", err)
						} else {
							log.V(2).Infof("patch no vote of %s to master", node.Name)
						}
					}
				}
			}
		}
	}
}
