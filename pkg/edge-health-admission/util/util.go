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

package util

import (
	"io/ioutil"
	"k8s.io/api/core/v1"
)

func GetCABundle(certFile string) ([]byte, error) {
	certByte, err := ioutil.ReadFile(certFile)
	if err != nil {
		return []byte{}, err
	}
	return certByte, nil
}

func TaintSetDiff(t1, t2 []v1.Taint) (taintsToAdd []v1.Taint, taintsToRemove []*v1.Taint) {
	for _, taint := range t1 {
		if !TaintExists(t2, &taint) {
			t := taint
			taintsToAdd = append(taintsToAdd, t)
		}
	}

	for _, taint := range t2 {
		if !TaintExists(t1, &taint) {
			t := taint
			taintsToRemove = append(taintsToRemove, &t)
		}
	}

	return
}

// TaintExists checks if the given taint exists in list of taints. Returns true if exists false otherwise.
func TaintExists(taints []v1.Taint, taintToFind *v1.Taint) bool {
	for _, taint := range taints {
		if taint.MatchTaint(taintToFind) {
			return true
		}
	}
	return false
}

func TaintExistsPosition(taints []v1.Taint, taintToFind *v1.Taint) (int, bool) {
	for index, taint := range taints {
		if taint.MatchTaint(taintToFind) {
			return index, true
		}
	}
	return -1, false
}

// GetNodeCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func GetNodeCondition(status *v1.NodeStatus, conditionType v1.NodeConditionType) (int, *v1.NodeCondition) {
	if status == nil {
		return -1, nil
	}
	for i := range status.Conditions {
		if status.Conditions[i].Type == conditionType {
			return i, &status.Conditions[i]
		}
	}
	return -1, nil
}
