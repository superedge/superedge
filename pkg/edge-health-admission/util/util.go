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
	"k8s.io/api/core/v1"
)

// TaintExists checks whether or not the given taint exists in the list of taints. Returns true and index if exists and false otherwise.
func TaintExistsPosition(taints []v1.Taint, taintToFind v1.Taint) (int, bool) {
	for index, taint := range taints {
		if taint.MatchTaint(&taintToFind) {
			return index, true
		}
	}
	return -1, false
}

// GetNodeCondition extracts the node condition from the provided node status and specified condition type.
// Returns nil and -1 if the condition is not present, otherwise the index of the located condition.
func GetNodeCondition(status *v1.NodeStatus, conditionType v1.NodeConditionType) (int, v1.NodeCondition) {
	if status == nil {
		return -1, v1.NodeCondition{}
	}
	for i := range status.Conditions {
		if status.Conditions[i].Type == conditionType {
			return i, status.Conditions[i]
		}
	}
	return -1, v1.NodeCondition{}
}
