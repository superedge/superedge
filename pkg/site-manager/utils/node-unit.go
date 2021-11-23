package utils

import (
	"context"
	"encoding/json"
	"fmt"
	sitev1 "github.com/superedge/superedge/pkg/site-manager/apis/site/v1"
	"github.com/superedge/superedge/pkg/site-manager/constant"
	crdClientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	"github.com/superedge/superedge/pkg/util"
	utilkube "github.com/superedge/superedge/pkg/util/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

func AddNodeUnitAnnotations(kubeclient clientset.Interface, node *corev1.Node, nodeUnit []string) error {
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}

	var nodeUnits []string
	value, ok := node.Annotations[constant.NodeUnitSuperedge]
	if ok && value != "\"\"" && value != "null" { // nil annotations Unmarshal can be failed
		if err := json.Unmarshal([]byte(value), &nodeUnits); err != nil {
			klog.Errorf("Unmarshal node: %s annotations: %s, error: %#v", node.Name, util.ToJson(value), err)
			return err
		}
	}

	nodeUnits = append(nodeUnits, nodeUnit...)
	nodeUnits = util.RemoveDuplicateElement(nodeUnits)
	node.Annotations[constant.NodeUnitSuperedge] = util.ToJson(nodeUnits)
	if _, err := kubeclient.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{}); err != nil {
		klog.Errorf("Update Node: %s, error: %#v", node.Name, err)
		return err
	}

	return nil
}

func RemoveNodeUnitAnnotations(kubeclient clientset.Interface, node *corev1.Node, nodeUnit []string) error {
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}

	var nodeUnits []string
	value, ok := node.Annotations[constant.NodeUnitSuperedge]
	if ok && value != "\"\"" && value != "null" { // nil annotations Unmarshal can be failed
		if err := json.Unmarshal([]byte(value), &nodeUnits); err != nil {
			klog.Errorf("Unmarshal node: %s annotations: %s, error: %#v", node.Name, util.ToJson(value), err)
			return err
		}
	}

	for _, nodeUnit := range nodeUnit {
		nodeUnits = util.DeleteSliceElement(nodeUnits, nodeUnit)
	}

	node.Annotations[constant.NodeUnitSuperedge] = util.ToJson(nodeUnits)
	if _, err := kubeclient.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{}); err != nil {
		klog.Errorf("Update Node: %s, error: %#v", node.Name, err)
		return err
	}

	return nil
}

func GetNodeUnitByNode(crdClient *crdClientset.Clientset, node *corev1.Node) (nodeUnits []sitev1.NodeUnit, err error) {
	allNodeUnit, err := crdClient.SiteV1().NodeUnits().List(context.TODO(), metav1.ListOptions{})
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

func GetNodeUnitNodes(kubeclient clientset.Interface, nodeUnit *sitev1.NodeUnit) (readyNodes, notReadyNodes []string, err error) {
	selector := nodeUnit.Spec.Selector
	var nodes []corev1.Node

	// Get Nodes by selector
	if selector != nil {
		if len(selector.MatchLabels) > 0 || len(selector.MatchExpressions) > 0 {
			labelSelector := &metav1.LabelSelector{
				MatchLabels:      selector.MatchLabels,
				MatchExpressions: selector.MatchExpressions,
			}
			selector, err := metav1.LabelSelectorAsSelector(labelSelector)
			if err != nil {
				return readyNodes, notReadyNodes, err
			}
			listOptions := metav1.ListOptions{LabelSelector: selector.String()}
			nodeList, err := kubeclient.CoreV1().Nodes().List(context.TODO(), listOptions)
			if err != nil {
				klog.Errorf("Get nodes by selector, error: %v", err)
				return readyNodes, notReadyNodes, err
			}
			nodes = append(nodes, nodeList.Items...)
		}

		if len(selector.Annotations) > 0 { //todo: add Annotations selector

		}
	}

	// Get Nodes by nodeName
	nodeNames := nodeUnit.Spec.Nodes
	for _, nodeName := range nodeNames {
		node, err := kubeclient.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Get nodes by node name, error: %v", err)
			return readyNodes, notReadyNodes, err
		}
		nodes = append(nodes, *node)
	}

	readyNodes, notReadyNodes = utilkube.GetNodeListStatus(nodes) // get all readynode and notReadyNodes
	return util.RemoveDuplicateElement(readyNodes), util.RemoveDuplicateElement(notReadyNodes), nil
}

func InitUnitToNode(kubeclient clientset.Interface, crdClient *crdClientset.Clientset) error {
	nodes, err := kubeclient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Get nodes by node name, error: %v", err)
		return err
	}

	for _, node := range nodes.Items {
		nodeUnits, err := GetNodeUnitByNode(crdClient, &node)
		if err != nil {
			klog.Errorf("Get nodeUnit by node, error： %#v", err)
			return err
		}

		var nodeUnitsName []string
		for _, unit := range nodeUnits {
			nodeUnitsName = append(nodeUnitsName, unit.Name)
		}

		if err := AddNodeUnitAnnotations(kubeclient, &node, nodeUnitsName); err != nil {
			klog.Errorf("Node: %s add annotations error: %#v", node.Name, err)
			return err
		}
	}

	return nil
}

func NodeUitReadyRateAdd(nodeUnit *sitev1.NodeUnit) string {
	unitStatus := nodeUnit.Status
	return fmt.Sprintf("%d/%d", len(unitStatus.ReadyNodes), len(unitStatus.ReadyNodes)+len(unitStatus.NotReadyNodes)+1)
}

func GetNodeUitReadyRate(nodeUnit *sitev1.NodeUnit) string {
	unitStatus := nodeUnit.Status
	return fmt.Sprintf("%d/%d", len(unitStatus.ReadyNodes), len(unitStatus.ReadyNodes)+len(unitStatus.NotReadyNodes))
}
