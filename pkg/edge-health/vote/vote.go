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
	admissionutil "github.com/superedge/superedge/pkg/edge-health-admission/util"
	"github.com/superedge/superedge/pkg/edge-health/common"
	"github.com/superedge/superedge/pkg/edge-health/config"
	"github.com/superedge/superedge/pkg/edge-health/metadata"
	"github.com/superedge/superedge/pkg/edge-health/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	log "k8s.io/klog"
	"time"
)

type Vote interface {
	Vote(*metadata.EdgeHealthMetadata, clientset.Interface, string, <-chan struct{})
}

type votePair struct {
	pros int
	cons int
}

type VoteEdge struct {
	*config.EdgeHealthVote
}

func NewVoteEdge(cfg *config.EdgeHealthVote) *VoteEdge {
	return &VoteEdge{cfg}
}

func (v *VoteEdge) Vote(edgeHealthMetadata *metadata.EdgeHealthMetadata, kubeclient clientset.Interface,
	localIp string, stopCh <-chan struct{}) {
	go wait.Until(func() {
		v.vote(edgeHealthMetadata, kubeclient, localIp, stopCh)
	}, time.Duration(v.VotePeriod)*time.Second, stopCh)
}

func (v *VoteEdge) vote(edgeHealthMetadata *metadata.EdgeHealthMetadata, kubeclient clientset.Interface, localIp string, stopCh <-chan struct{}) {
	var (
		prosVoteIpList, consVoteIpList []string
		vp                             votePair
	)
	voteCountMap := make(map[string]votePair) // {"127.0.0.1":{"pros":1,"cons":2}}
	copyCheckInfo := edgeHealthMetadata.CopyAll()
	voteThreshold := (len(copyCheckInfo) + 1) / 2
	for _, checkedDetails := range copyCheckInfo {
		for checkedIp, checkedDetail := range checkedDetails {
			if !time.Now().After(checkedDetail.Time.Add(time.Duration(v.VoteTimeout) * time.Second)) {
				if _, existed := voteCountMap[checkedIp]; !existed {
					voteCountMap[checkedIp] = votePair{0, 0}
				}
				vp = voteCountMap[checkedIp]
				if checkedDetail.Normal {
					vp.pros++
					if vp.pros >= voteThreshold {
						prosVoteIpList = append(prosVoteIpList, checkedIp)
					}
				} else {
					vp.cons++
					if vp.cons >= voteThreshold {
						consVoteIpList = append(consVoteIpList, checkedIp)
					}
				}
				voteCountMap[checkedIp] = vp
			}
		}
	}
	log.V(4).Infof("Vote: voteCountMap is %v", voteCountMap)

	// Handle prosVoteIpList
	util.ParallelizeUntil(context.TODO(), 16, len(prosVoteIpList), func(index int) {
		if node := edgeHealthMetadata.GetNodeByAddr(prosVoteIpList[index]); node != nil {
			log.V(4).Infof("Vote: vote pros to edge node %s begin ...", node.Name)
			nodeCopy := node.DeepCopy()
			needUpdated := false
			if _, existed := nodeCopy.Annotations[common.NodeUnhealthAnnotation]; existed {
				delete(nodeCopy.Annotations, common.NodeUnhealthAnnotation)
				needUpdated = true
			}
			if index, existed := admissionutil.TaintExistsPosition(node.Spec.Taints, common.UnreachableNoExecuteTaint); existed {
				nodeCopy.Spec.Taints = append(nodeCopy.Spec.Taints[:index], nodeCopy.Spec.Taints[index+1:]...)
				needUpdated = true
			}
			if needUpdated {
				if _, err := kubeclient.CoreV1().Nodes().Update(context.TODO(), nodeCopy, metav1.UpdateOptions{}); err != nil {
					log.Errorf("Vote: update pros vote to edge node %s error %v ", nodeCopy.Name, err)
				} else {
					log.V(2).Infof("Vote: update pros vote to edge node %s successfully", nodeCopy.Name)
				}
			}
		} else {
			log.Warningf("Vote: edge node addr %s not found", prosVoteIpList[index])
		}
	})

	// Handle consVoteIpList
	util.ParallelizeUntil(context.TODO(), 16, len(consVoteIpList), func(index int) {
		if node := edgeHealthMetadata.GetNodeByAddr(consVoteIpList[index]); node != nil {
			log.V(4).Infof("Vote: vote cons to edge node %s begin ...", node.Name)
			if _, existed := node.Annotations[common.NodeUnhealthAnnotation]; !existed {
				nodeCopy := node.DeepCopy()
				nodeCopy.Annotations[common.NodeUnhealthAnnotation] = "yes"
				if _, err := kubeclient.CoreV1().Nodes().Update(context.TODO(), nodeCopy, metav1.UpdateOptions{}); err != nil {
					log.Errorf("Vote: update cons vote to edge node %s error %v ", nodeCopy.Name, err)
				} else {
					log.V(2).Infof("Vote: update cons vote to edge node %s successfully", nodeCopy.Name)
				}
			}
		} else {
			log.Warningf("Vote: edge node addr %s not found", consVoteIpList[index])
		}
	})
}
