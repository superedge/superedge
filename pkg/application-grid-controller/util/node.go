package util

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	corelisters "k8s.io/client-go/listers/core/v1"
)

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
