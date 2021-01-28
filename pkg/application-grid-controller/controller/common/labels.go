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

package common

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	corelisters "k8s.io/client-go/listers/core/v1"
)

const (
	// GridSelectorName is the selector key name for controller to retrieve deployments
	GridSelectorName = "superedge.io/grid-selector"
	// GridSelectorKey corresponds to gridUniqKey of the custom-defined workload
	GridSelectorUniqKeyName = "superedge.io/grid-uniq-key"
)

func GetDefaultSelector(val string) (labels.Selector, error) {
	labelSelector := labels.NewSelector()
	dpRequirement, err := labels.NewRequirement(GridSelectorName, selection.Equals, []string{val})
	if err != nil {
		return nil, err
	}
	return labelSelector.Add(*dpRequirement), nil
}

func IsConcernedObject(objMeta metav1.ObjectMeta) bool {
	if objMeta.Labels == nil {
		return false
	}

	_, found := objMeta.Labels[GridSelectorName]
	return found
}

func GetGridValuesFromNode(nodeLister corelisters.NodeLister, gridUniqKey string) ([]string, error) {
	labelSelector := labels.NewSelector()
	gridRequirement, err := labels.NewRequirement(gridUniqKey, selection.Exists, nil)
	if err != nil {
		return nil, err
	}
	labelSelector = labelSelector.Add(*gridRequirement)

	nodes, err := nodeLister.List(labelSelector)
	if err != nil {
		return nil, err
	}

	values := make([]string, 0)
	for _, n := range nodes {
		if gridVal := n.Labels[gridUniqKey]; gridVal != "" {
			values = append(values, gridVal)
		}
	}
	return values, nil
}
