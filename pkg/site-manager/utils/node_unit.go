package utils

import (
	"context"
	"fmt"

	"github.com/superedge/superedge/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	sitev1 "github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha1"
	crdClientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
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

func RemoveUnitSetNode(crdClient *crdClientset.Clientset, units, keys []string) error {
	if len(units) == 0 {
		return nil
	}
	for _, unit := range units {
		nodeUnit, err := crdClient.SiteV1alpha1().NodeUnits().Get(context.TODO(), unit, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Remove unit setNode get nodeUnit error: %#v", err)
			continue
		}
		setNode := &nodeUnit.Spec.SetNode
		for _, key := range keys {
			delete(setNode.Labels, key)
		}
		_, err = crdClient.SiteV1alpha1().NodeUnits().Update(context.TODO(), nodeUnit, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("Remove unit setNode update nodeUnit error: %#v", err)
			continue
		}
	}
	return nil
}

func RemoveSetNode(kubeClient clientset.Interface, nodeUnit *sitev1.NodeUnit, nodes []string) error {
	klog.V(4).Infof("Remove setNode nodeUnit: %s will remove nodes: %s setNode: %s", nodeUnit.Name, nodes, util.ToJson(nodeUnit.Spec.SetNode))
	for _, nodeName := range nodes {
		node, err := kubeClient.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Get node error: %v", err)
			continue
		}

		setNode := &nodeUnit.Spec.SetNode
		if setNode.Labels != nil && node.Labels != nil {
			for labelKey, _ := range setNode.Labels {
				if _, ok := node.Labels[labelKey]; ok {
					delete(node.Labels, labelKey)
				}
			}
		}
		if setNode.Annotations != nil && node.Annotations != nil {
			for annotationKey, _ := range setNode.Annotations {
				if _, ok := node.Annotations[annotationKey]; ok {
					delete(node.Annotations, annotationKey)
				}
			}
		}
		if setNode.Taints != nil && node.Spec.Taints != nil {
			taints := make(map[string]bool, len(setNode.Taints))
			for _, taint := range setNode.Taints {
				taints[taint.Key] = true
			}
			var taintSlice []corev1.Taint
			for _, taint := range node.Spec.Taints {
				if _, ok := taints[taint.Key]; !ok {
					taintSlice = append(taintSlice, taint)
				}
			}
		}

		if _, err := kubeClient.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{}); err != nil {
			klog.Errorf("Remove setNode update node: %s, error: %#v", node.Name, err)
			continue
		}
	}
	return nil
}
