package utils

import (
	"context"
	sitev1 "github.com/superedge/superedge/pkg/site-manager/apis/site/v1"
	siteClientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	"github.com/superedge/superedge/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"strings"
)

func GetUnitsByNodeGroup(siteClient *siteClientset.Clientset, nodeGroup *sitev1.NodeGroup) (nodeUnits []string, err error) {
	// Get units by selector
	var unitList *sitev1.NodeUnitList
	selector := nodeGroup.Spec.Selector
	if selector != nil {
		if len(selector.MatchLabels) > 0 || len(selector.MatchExpressions) > 0 {
			labelSelector := &metav1.LabelSelector{
				MatchLabels:      selector.MatchLabels,
				MatchExpressions: selector.MatchExpressions,
			}
			selector, err := metav1.LabelSelectorAsSelector(labelSelector)
			if err != nil {
				return nodeUnits, err
			}

			listOptions := metav1.ListOptions{LabelSelector: selector.String()}
			unitList, err = siteClient.SiteV1().NodeUnits().List(context.TODO(), listOptions)
			if err != nil {
				klog.Errorf("Get nodes by selector, error: %v", err)
				return nodeUnits, err
			}
		}

		if len(selector.Annotations) > 0 { //todo: add Annotations selector

		}

		for _, unit := range unitList.Items {
			nodeUnits = append(nodeUnits, unit.Name)
		}
	}

	// Get units by nodeName
	unitsNames := nodeGroup.Spec.NodeUnits
	for _, unitName := range unitsNames {
		unit, err := siteClient.SiteV1().NodeUnits().Get(context.TODO(), unitName, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				klog.Warningf("Get node: %s nil", nodeGroup.Name)
				continue
			} else {
				klog.Errorf("Get nodes by node name, error: %v", err)
				return nodeUnits, err
			}
		}
		nodeUnits = append(nodeUnits, unit.Name)
	}

	return util.RemoveDuplicateElement(nodeUnits), nil
}
