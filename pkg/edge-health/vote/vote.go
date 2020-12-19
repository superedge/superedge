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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	log "k8s.io/klog"
	"superedge/pkg/edge-health/check"
	"superedge/pkg/edge-health/common"
	"superedge/pkg/edge-health/data"
	"superedge/pkg/edge-health/util"
	"time"
	admissionutil "superedge/pkg/edge-health-admission/util"
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
	for k, v := range tempNodeStatus {                //k is checker ip
		for ip, resultdetail := range v { //ip is checked ip
			if k == common.LocalIp || (k != common.LocalIp && !time.Now().After(resultdetail.Time.Add(time.Duration(vote.GetVoteTimeout())*time.Second))) {
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
				name := util.GetNodeNameByIp(data.NodeList.NodeList.Items, ip)
				if node, err := check.NodeManager.NodeLister.Get(name); err == nil && name != "" {
					if _, ok := node.Annotations["nodeunhealth"]; ok {
						nodenew := node.DeepCopy()
						delete(nodenew.Annotations, "nodeunhealth")
						if _, err := common.ClientSet.CoreV1().Nodes().Update(context.TODO(), nodenew, metav1.UpdateOptions{}); err != nil {
							log.Errorf("update yes vote to master error: %v ", err)
						} else {
							log.V(2).Infof("update yes vote of %s to master", nodenew.Name)
						}
					} else if index, flag := admissionutil.TaintExistsPosition(node.Spec.Taints, UnreachNoExecuteTaint); flag{
						nodenew := node.DeepCopy()
						nodenew.Spec.Taints = append(nodenew.Spec.Taints[:index], nodenew.Spec.Taints[index+1:]...)
						if _, err := common.ClientSet.CoreV1().Nodes().Update(context.TODO(), nodenew, metav1.UpdateOptions{}); err != nil {
							log.Errorf("remove no excute taint for health node error: %v ", err)
						} else {
							log.V(2).Infof("remove no excute taint for health node: %s to master", nodenew.Name)
						}
					}
				}
			}
		}
		if _, ok := v["no"]; ok {
			if float64(v["no"]) >= num {
				log.V(4).Infof("vote: vote no to master begin")
				name := util.GetNodeNameByIp(data.NodeList.NodeList.Items, ip)
				if node, err := check.NodeManager.NodeLister.Get(name); err == nil && name != "" {
					if _, ok := node.Annotations["nodeunhealth"]; !ok {
						nodenew := node.DeepCopy()
						nodenew.Annotations["nodeunhealth"] = "yes"
						if _, err := common.ClientSet.CoreV1().Nodes().Update(context.TODO(), nodenew, metav1.UpdateOptions{}); err != nil {
							log.Errorf("update no vote to master error: %v ", err)
						} else {
							log.V(2).Infof("update no vote of %s to master", nodenew.Name)
						}
					}
				}
			}
		}
	}
}
