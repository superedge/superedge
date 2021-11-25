package utils

import (
	"context"
	"encoding/json"
	sitev1 "github.com/superedge/superedge/pkg/site-manager/apis/site/v1"
	"github.com/superedge/superedge/pkg/site-manager/constant"
	"github.com/superedge/superedge/pkg/util"
	utilkube "github.com/superedge/superedge/pkg/util/kubeclient"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"strings"
)

func GetNodesByUnit(kubeclient clientset.Interface, nodeUnit *sitev1.NodeUnit) (readyNodes, notReadyNodes []string, err error) {
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
			if strings.Contains(err.Error(), "not found") {
				klog.Warningf("Get node: %s nil", nodeUnit.Name)
				continue
			} else {
				klog.Errorf("Get nodes by node name, error: %v", err)
				return readyNodes, notReadyNodes, err
			}
		}
		nodes = append(nodes, *node)
	}

	readyNodes, notReadyNodes = utilkube.GetNodeListStatus(nodes) // get all readynode and notReadyNodes
	return util.RemoveDuplicateElement(readyNodes), util.RemoveDuplicateElement(notReadyNodes), nil
}

/*
  Nodes Annotations
*/

func AddNodesAnnotations(kubeClient clientset.Interface, nodeNames []string, annotations []string) error {
	for _, nodeName := range nodeNames {
		node, err := kubeClient.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Get Node: %s, error: %#v", nodeName, err)
			continue
		}

		if err := AddNodeAnnotations(kubeClient, node, annotations); err != nil {
			klog.Errorf("Update Node: %s, error: %#v", node.Name, err)
			return err
		}
	}

	return nil
}

func RemoveNodesAnnotations(kubeClient clientset.Interface, nodeNames []string, annotations []string) {
	for _, nodeName := range nodeNames {
		node, err := kubeClient.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Get Node: %s, error: %#v", nodeName, err)
			continue
		}

		if err := RemoveNodeAnnotations(kubeClient, node, annotations); err != nil {
			klog.Errorf("Remove node: %s annotations nodeunit: %s flags error: %#v", nodeName, annotations, err)
			continue
		}
	}
}

/*
  Node Annotations
*/

func AddNodeAnnotations(kubeclient clientset.Interface, node *corev1.Node, annotations []string) error {
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}

	var nodeUnits []string
	value, ok := node.Annotations[constant.NodeUnitSuperedge]
	if ok && value != "\"\"" && value != "null" { // nil annotations Unmarshal can be failed
		if err := json.Unmarshal(json.RawMessage(value), &nodeUnits); err != nil {
			klog.Errorf("Unmarshal node: %s annotations: %s, error: %#v", node.Name, util.ToJson(value), err)
			return err
		}
	}

	nodeUnits = append(nodeUnits, annotations...)
	nodeUnits = util.RemoveDuplicateElement(nodeUnits)
	nodeUnits = util.DeleteSliceElement(nodeUnits, "")
	node.Annotations[constant.NodeUnitSuperedge] = util.ToJson(nodeUnits)
	if _, err := kubeclient.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{}); err != nil {
		klog.Errorf("Update Node: %s, error: %#v", node.Name, err)
		return err
	}

	return nil
}

func RemoveNodeAnnotations(kubeclient clientset.Interface, node *corev1.Node, annotations []string) error {
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}

	var nodeUnits []string
	value, ok := node.Annotations[constant.NodeUnitSuperedge]
	if ok && value != "\"\"" && value != "null" { // nil annotations Unmarshal can be failed
		if err := json.Unmarshal(json.RawMessage(value), &nodeUnits); err != nil {
			klog.Errorf("Unmarshal node: %s annotations: %s, error: %#v", node.Name, util.ToJson(value), err)
			return err
		}
	}

	for _, annotation := range annotations {
		nodeUnits = util.DeleteSliceElement(nodeUnits, annotation)
	}
	nodeUnits = util.DeleteSliceElement(nodeUnits, "")
	node.Annotations[constant.NodeUnitSuperedge] = util.ToJson(nodeUnits)
	if _, err := kubeclient.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{}); err != nil {
		klog.Errorf("Update Node: %s, error: %#v", node.Name, err)
		return err
	}

	return nil
}

func ResetNodeUnitAnnotations(kubeclient clientset.Interface, node *corev1.Node, annotations []string) error {
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}
	annotations = util.DeleteSliceElement(annotations, "")
	node.Annotations[constant.NodeUnitSuperedge] = util.ToJson(annotations)
	if _, err := kubeclient.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{}); err != nil {
		klog.Errorf("Update Node: %s, error: %#v", node.Name, err)
		return err
	}

	return nil
}
