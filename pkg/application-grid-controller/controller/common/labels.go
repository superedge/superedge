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
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/validation"
	corelisters "k8s.io/client-go/listers/core/v1"
)

const (
	// GridSelectorName is the selector key name for controller to retrieve deployments
	GridSelectorName = "superedge.io/grid-selector"
	// GridSelectorKey corresponds to gridUniqKey of the custom-defined workload
	GridSelectorUniqKeyName = "superedge.io/grid-uniq-key"
	// TemplateHashKey is a key for storing template's hash value in labels.
	TemplateHashKey = "superedge.io/template-hash-key"
	// Indicate this app is for federation, "yes" or "no"
	FedrationKey = "superedge.io/fed"
	// Indicate this app has been distributed
	FedrationDisKey = "superedge.io/dis"
	// ClusterId of managed cluster id
	FedManagedClustIdKey = "clusters.clusternet.io/cluster-id"
	// Target namespace of app in managed cluster
	FedTargetNameSpace = "superedge.io/target-namespace"
)

func GetDefaultSelector(val string) (labels.Selector, error) {
	labelSelector := labels.NewSelector()
	if errList := validation.IsValidLabelValue(val); errList != nil {
		return labelSelector, fmt.Errorf("Invalid label value %s err %v", val, errList)
	}

	dpRequirement, err := labels.NewRequirement(GridSelectorName, selection.Equals, []string{val})
	if err != nil {
		return labelSelector, err
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

func GetNodesSelector(nodes ...*corev1.Node) (labels.Selector, error) {
	labelSelector := labels.NewSelector()
	valueList := make([]string, 0)
	seen := make(map[string]struct{})
	for _, node := range nodes {
		for key := range node.Labels {
			if _, exist := seen[key]; !exist {
				seen[key] = struct{}{}
				if errList := validation.IsValidLabelValue(key); errList != nil {
					continue
				}
				valueList = append(valueList, key)
			}
		}
	}
	if len(valueList) == 0 {
		return labelSelector, nil
	}
	requirement, err := labels.NewRequirement(GridSelectorUniqKeyName, selection.In, valueList)
	if err != nil {
		return labelSelector, err
	}
	labelSelector = labelSelector.Add(*requirement)
	return labelSelector, nil
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

	var values []string
	for _, n := range nodes {
		if gridVal := n.Labels[gridUniqKey]; gridVal != "" {
			values = append(values, gridVal)
		}
	}
	return values, nil
}
