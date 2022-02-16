package utils

import (
	"context"
	sitev1 "github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha1"
	crdClientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	AllNodeUnit    = "unit-node-all"
	EdgeNodeUnit   = "unit-node-edge"
	CloudNodeUnit  = "unit-node-cloud"
	MasterNodeUnit = "unit-node-master"
)

func CreateDefaultUnit(crdClient *crdClientset.Clientset) error {
	// All Node Unit
	allNodeUnitSelector := &sitev1.Selector{
		MatchLabels: map[string]string{
			"kubernetes.io/os": "linux",
		},
	}
	allNnodeUnit := &sitev1.NodeUnit{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "site.superedge.io/v1alpha1",
			Kind:       "NodeUnit",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: AllNodeUnit,
		},
		Spec: sitev1.NodeUnitSpec{
			Type:     sitev1.OtherNodeUnit,
			Selector: allNodeUnitSelector,
		},
	}
	if _, err := crdClient.SiteV1alpha1().NodeUnits().Create(context.TODO(), allNnodeUnit, metav1.CreateOptions{}); err != nil {
		klog.Warningf("Create default %s unit error : %#v", AllNodeUnit, err)
	}

	// Master Node Unit
	masterUnitSelector := &sitev1.Selector{
		MatchLabels: map[string]string{
			KubernetesMasterNodeRoleKey: "",
		},
	}
	masterUnit := &sitev1.NodeUnit{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "site.superedge.io/v1alpha1",
			Kind:       "NodeUnit",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: MasterNodeUnit,
		},
		Spec: sitev1.NodeUnitSpec{
			Type:     sitev1.MasterNodeUnit,
			Selector: masterUnitSelector,
		},
	}
	if _, err := crdClient.SiteV1alpha1().NodeUnits().Create(context.TODO(), masterUnit, metav1.CreateOptions{}); err != nil {
		klog.Warningf("Create default %s unit error : %#v", MasterNodeUnit, err)
	}

	// Cloud Node Unit
	cloudNodeUnitSelector := &sitev1.Selector{
		MatchLabels: map[string]string{
			KubernetesCloudNodeRoleKey: "",
		},
	}
	cloudNnodeUnit := &sitev1.NodeUnit{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "site.superedge.io/v1alpha1",
			Kind:       "NodeUnit",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: CloudNodeUnit,
		},
		Spec: sitev1.NodeUnitSpec{
			Type:     sitev1.CloudNodeUnit,
			Selector: cloudNodeUnitSelector,
		},
	}
	if _, err := crdClient.SiteV1alpha1().NodeUnits().Create(context.TODO(), cloudNnodeUnit, metav1.CreateOptions{}); err != nil {
		klog.Warningf("Create default %s unit error : %#v", MasterNodeUnit, err)
	}

	// Edge Node Unit
	edgedNodeUnitSelector := &sitev1.Selector{
		MatchLabels: map[string]string{
			KubernetesEdgeNodeRoleKey: "",
		},
	}
	edgeNnodeUnit := &sitev1.NodeUnit{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "site.superedge.io/v1alpha1",
			Kind:       "NodeUnit",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: EdgeNodeUnit,
		},
		Spec: sitev1.NodeUnitSpec{
			Type:     sitev1.EdgeNodeUnit,
			Selector: edgedNodeUnitSelector,
		},
	}
	if _, err := crdClient.SiteV1alpha1().NodeUnits().Create(context.TODO(), edgeNnodeUnit, metav1.CreateOptions{}); err != nil {
		klog.Warningf("Create default %s unit error : %#v", EdgeNodeUnit, err)
	}

	return nil
}

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
