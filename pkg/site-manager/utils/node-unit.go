package utils

import (
	"context"
	crdClientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

func InitUnitToNode(kubeclient clientset.Interface, crdClient *crdClientset.Clientset) error {
	nodes, err := kubeclient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Get nodes by node name, error: %v", err)
		return err
	}

	for _, node := range nodes.Items {
		nodeUnits, err := GetUnitsByNode(crdClient, &node)
		if err != nil {
			klog.Errorf("Get nodeUnit by node, error： %#v", err)
			return err
		}

		var nodeUnitsName []string
		for _, unit := range nodeUnits {
			nodeUnitsName = append(nodeUnitsName, unit.Name)
		}

		// Processing stock node annotations
		if err := ResetNodeUnitAnnotations(kubeclient, &node, nodeUnitsName); err != nil {
			klog.Errorf("Node: %s add annotations error: %#v", node.Name, err)
			return err
		}
	}

	return nil
}
