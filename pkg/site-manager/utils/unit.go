package utils

import (
	"context"
	"fmt"
	sitev1 "github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha1"
	crdClientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

//  GetUnitsByNode
func GetUnitsByNode(crdClient *crdClientset.Clientset, node *corev1.Node) (nodeUnits []sitev1.NodeUnit, err error) {
	allNodeUnit, err := crdClient.SiteV1alpha1().NodeUnits().List(context.TODO(), metav1.ListOptions{})
	if err != nil && !errors.IsConflict(err) {
		klog.Errorf("List nodeUnit error: %#v", err)
		return nil, err
	}

	for _, nodeunit := range allNodeUnit.Items {
		for _, nodeName := range nodeunit.Status.ReadyNodes {
			if nodeName == node.Name {
				nodeUnits = append(nodeUnits, nodeunit)
			}
		}
		for _, nodeName := range nodeunit.Status.NotReadyNodes {
			if nodeName == node.Name {
				nodeUnits = append(nodeUnits, nodeunit)
			}
		}
	}
	return
}

/*
  NodeUit Rate
*/
func AddNodeUitReadyRate(nodeUnit *sitev1.NodeUnit) string {
	unitStatus := nodeUnit.Status
	return fmt.Sprintf("%d/%d", len(unitStatus.ReadyNodes), len(unitStatus.ReadyNodes)+len(unitStatus.NotReadyNodes)+1)
}

func GetNodeUitReadyRate(nodeUnit *sitev1.NodeUnit) string {
	unitStatus := nodeUnit.Status
	return fmt.Sprintf("%d/%d", len(unitStatus.ReadyNodes), len(unitStatus.ReadyNodes)+len(unitStatus.NotReadyNodes))
}
