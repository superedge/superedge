package utils

import (
	"context"
	"encoding/json"
	edgeadmConstant "github.com/superedge/superedge/pkg/edgeadm/constant"
	sitev1 "github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha1"
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

const (
	KubernetesEdgeNodeRoleKey   = "node-role.kubernetes.io/Edge"
	KubernetesCloudNodeRoleKey  = "node-role.kubernetes.io/Cloud"
	KubernetesMasterNodeRoleKey = "node-role.kubernetes.io/Master"
)

func SetNodeRole(kubeClient clientset.Interface, node *corev1.Node) error {
	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}

	if _, ok := node.Labels[edgeadmConstant.EdgeNodeLabelKey]; ok {
		edgeNodeLabel := map[string]string{
			KubernetesEdgeNodeRoleKey: "",
		}
		if err := utilkube.AddNodeLabel(kubeClient, node.Name, edgeNodeLabel); err != nil {
			klog.Errorf("Add edge Node role label error: %v", err)
			return err
		}
		return nil
	}

	if _, ok := node.Labels[edgeadmConstant.CloudNodeLabelKey]; ok {
		cloudNodeLabel := map[string]string{
			KubernetesCloudNodeRoleKey: "",
		}
		if err := utilkube.AddNodeLabel(kubeClient, node.Name, cloudNodeLabel); err != nil {
			klog.Errorf("Add Cloud node label error: %v", err)
			return err
		}
		return nil
	}

	return nil
}

func SetNodeToNodes(kubeClient clientset.Interface, setNode sitev1.SetNode, nodeNames []string) {
	for _, nodeName := range nodeNames {
		node, err := kubeClient.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				klog.Warningf("Get node: %s nil", node.Name)
				continue
			}
		}

		if setNode.Labels != nil {
			if node.Labels == nil {
				node.Labels = make(map[string]string)
				node.Labels = setNode.Labels
			} else {
				for key, val := range setNode.Labels {
					node.Labels[key] = val
				}
			}
		}
		if setNode.Annotations != nil {
			if node.Annotations == nil {
				node.Annotations = make(map[string]string)
				node.Annotations = setNode.Annotations
			} else {
				for key, val := range setNode.Annotations {
					node.Annotations[key] = val
				}
			}
		}
		if setNode.Taints != nil {
			if node.Spec.Taints == nil {
				node.Spec.Taints = []corev1.Taint{}
			}
			node.Spec.Taints = append(node.Spec.Taints, setNode.Taints...)
		}
		if _, err := kubeClient.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{}); err != nil {
			klog.Errorf("Update Node: %s, error: %#v", node.Name, err)
			continue
		}
	}
}

func DeleteNodesFromSetNode(kubeClient clientset.Interface, setNode sitev1.SetNode, nodeNames []string) {
	for _, nodeName := range nodeNames {
		node, err := kubeClient.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				klog.Warningf("Get node: %s nil", node.Name)
				continue
			}
		}

		if setNode.Labels != nil {
			if node.Labels != nil {
				for k, _ := range setNode.Labels {
					delete(node.Labels, k)
				}
			}
		}

		if setNode.Annotations != nil {
			if node.Annotations != nil {
				for k, _ := range setNode.Annotations {
					delete(node.Annotations, k)
				}
			}
		}
		node.Spec.Taints = deleteTaintItems(node.Spec.Taints, setNode.Taints)

		if _, err := kubeClient.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{}); err != nil {
			klog.Errorf("Update Node: %s, error: %#v", node.Name, err)
			continue
		}
	}
}

func SetUpdatedValue(oldValues map[string]string, curValues map[string]string, modifyValues *map[string]string) {

	// delete old values
	for k, _ := range oldValues {
		if _, found := (*modifyValues)[k]; found {
			delete((*modifyValues), k)
		}
	}
	// set new values
	for k, v := range curValues {
		(*modifyValues)[k] = v
	}
}

func UpdtateNodeFromSetNode(kubeClient clientset.Interface, oldSetNode sitev1.SetNode, curSetNode sitev1.SetNode, nodeNames []string) {
	for _, nodeName := range nodeNames {
		node, err := kubeClient.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				klog.Warningf("Get node: %s nil", node.Name)
				continue
			}
		}

		if node.Labels == nil {
			node.Labels = make(map[string]string)
			node.Labels = curSetNode.Labels
		} else {
			SetUpdatedValue(oldSetNode.Labels, curSetNode.Labels, &node.Labels)
		}

		if node.Annotations == nil {
			node.Annotations = make(map[string]string)
			node.Annotations = curSetNode.Annotations
		} else {
			SetUpdatedValue(oldSetNode.Annotations, curSetNode.Annotations, &node.Annotations)
		}

		if node.Spec.Taints == nil {
			node.Spec.Taints = []corev1.Taint{}
			node.Spec.Taints = append(node.Spec.Taints, curSetNode.Taints...)
		} else {
			node.Spec.Taints = updateTaintItems(node.Spec.Taints, oldSetNode.Taints, curSetNode.Taints)
		}

		if _, err := kubeClient.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{}); err != nil {
			klog.Errorf("Update Node: %s, error: %#v", node.Name, err)
			continue
		}
	}
}

func updateTaintItems(currTaintItems []corev1.Taint, oldTaintItems []corev1.Taint, newTaintItems []corev1.Taint) []corev1.Taint {

	deletedResults := []corev1.Taint{}
	Results := []corev1.Taint{}

	newMap := make(map[string]bool)
	for _, s := range newTaintItems {
		newMap[s.Key] = true
	}

	deleteMap := make(map[string]bool)
	for _, s := range oldTaintItems {
		deleteMap[s.Key] = true
	}

	// filter old values
	for _, val := range currTaintItems {
		if _, ok := deleteMap[val.Key]; ok {
			continue
		}
		deletedResults = append(deletedResults, val)
	}
	// filter new values
	for _, val := range deletedResults {
		if _, ok := newMap[val.Key]; ok {
			continue
		}
		Results = append(Results, val)
	}

	Results = append(Results, newTaintItems...)
	return Results
}

func deleteTaintItems(currentTaintItems, deleteTaintItems []corev1.Taint) []corev1.Taint {
	result := []corev1.Taint{}
	deleteMap := make(map[string]bool)
	for _, s := range deleteTaintItems {
		deleteMap[s.Key] = true
	}

	for _, val := range currentTaintItems {
		if _, ok := deleteMap[val.Key]; ok {
			continue
		} else {
			result = append(result, val)
		}
	}
	return result
}
