package utils

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha2"
	sitev1 "github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha2"
	"github.com/superedge/superedge/pkg/site-manager/constant"
	crdClientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	crdv1listers "github.com/superedge/superedge/pkg/site-manager/generated/listers/site.superedge.io/v1alpha2"
	"github.com/superedge/superedge/pkg/util"
	utilkube "github.com/superedge/superedge/pkg/util/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"sort"

	"k8s.io/klog/v2"
)

const Finalizers = `
{"metadata":{"finalizers":null}}
`

func CaculateNodeUnitStatus(nodeMap map[string]*corev1.Node, nu *sitev1.NodeUnit) (*sitev1.NodeUnitStatus, error) {
	status := &sitev1.NodeUnitStatus{}
	var readyList, notReadyList []string
	for k, v := range nodeMap {
		if utilkube.IsReadyNode(v) {
			readyList = append(readyList, k)
		} else {
			notReadyList = append(notReadyList, k)
		}
	}
	sort.Strings(readyList)
	sort.Strings(notReadyList)
	status.ReadyNodes = readyList
	status.NotReadyNodes = notReadyList
	status.ReadyRate = fmt.Sprintf("%d/%d", len(readyList), len(readyList)+len(notReadyList))

	return status, nil
}

func CaculateNodeGroupStatus(unitSet sets.String, ng *sitev1.NodeGroup) (*sitev1.NodeGroupStatus, error) {
	status := &sitev1.NodeGroupStatus{}
	status.NodeUnits = unitSet.List()
	status.UnitNumber = unitSet.Len()
	return status, nil
}

func ListNodeFromLister(nodeLister corelisters.NodeLister, selector labels.Selector, appendFn cache.AppendFunc) error {

	nodes, err := nodeLister.List(selector)
	if err != nil {
		return err
	}

	klog.Infof("selector = %v,", selector)
	for _, n := range nodes {
		klog.Infof("selector = %v, labels = %v", selector, n.Labels)
		appendFn(n)
	}
	return nil
}

func ListNodeUnitFromLister(unitLister crdv1listers.NodeUnitLister, selector labels.Selector, appendFn cache.AppendFunc) error {

	units, err := unitLister.List(selector)
	if err != nil {
		return err
	}
	for _, n := range units {
		appendFn(n)
	}
	return nil
}

// GetUnitsByNode
func GetGroupsByUnit(groupLister crdv1listers.NodeGroupLister, nu *sitev1.NodeUnit) (nodeGroups []*sitev1.NodeGroup, groupList []string, err error) {

	for k, v := range nu.Labels {
		if v == constant.NodeGroupSuperedge {
			groupList = append(groupList, k)
		}
	}
	for _, ngName := range groupList {
		ng, err := groupLister.Get(ngName)
		if err != nil {
			return nil, nil, err
		}
		nodeGroups = append(nodeGroups, ng)
	}
	return
}

// GetUnitsByNode
func GetUnitsByNode(unitLister crdv1listers.NodeUnitLister, node *corev1.Node) (nodeUnits []*sitev1.NodeUnit, unitList []string, err error) {

	for k, v := range node.Labels {
		if v == constant.NodeUnitSuperedge {
			unitList = append(unitList, k)
		}
	}
	for _, nuName := range unitList {
		nu, err := unitLister.Get(nuName)
		if err != nil {
			return nil, nil, err
		}
		nodeUnits = append(nodeUnits, nu)
	}
	return
}

func SetNodeToNodeUnits(crdClient *crdClientset.Clientset, ng *v1alpha2.NodeGroup, unitMaps map[string]*v1alpha2.NodeUnit) error {
	klog.V(4).InfoS("SetNode to node unit: %s", "nodegroup", ng.Name)

	for _, nu := range unitMaps {
		setNodeReady, ownerLabelReady := true, true
		if nu.Spec.SetNode.Labels != nil && nu.Spec.SetNode.Labels[ng.Name] != nu.Name {
			setNodeReady = false
		}

		if nu.Labels != nil && nu.Labels[ng.Name] != constant.NodeGroupSuperedge {
			ownerLabelReady = false
		}

		if setNodeReady && ownerLabelReady {
			continue
		}
		newNu := nu.DeepCopy()
		// update setnode for add label({nodegroup name}: {nodeunit name}) for node

		if newNu.Spec.SetNode.Labels == nil {
			newNu.Spec.SetNode.Labels = make(map[string]string)
		}
		newNu.Spec.SetNode.Labels[ng.Name] = newNu.Name

		if newNu.Labels == nil {
			newNu.Labels = make(map[string]string)
		}
		// set node group as owner
		newNu.Labels[ng.Name] = constant.NodeGroupSuperedge

		if _, err := crdClient.SiteV1alpha2().NodeUnits().Update(context.TODO(), newNu, metav1.UpdateOptions{}); err != nil {
			return err
		}
	}
	return nil
}

func GetUnitByGroup(unitLister crdv1listers.NodeUnitLister, ng *sitev1.NodeGroup) (sets.String, map[string]*sitev1.NodeUnit, error) {
	groupUnitSet := sets.NewString()
	unitMap := make(map[string]*sitev1.NodeUnit, 3)
	appendFunc := func(n interface{}) {
		nu, ok := n.(*sitev1.NodeUnit)
		if !ok {
			return
		}
		unitMap[nu.Name] = nu
		groupUnitSet.Insert(nu.Name)
	}
	if len(ng.Spec.NodeUnits) > 0 {
		for _, nuName := range ng.Spec.NodeUnits {
			nu, err := unitLister.Get(nuName)
			if err != nil && errors.IsNotFound(err) {
				klog.V(4).InfoS("Nodegroup units not found", "node group", ng.Name, "node unit", nuName)
				continue
			} else if err != nil {
				return nil, nil, err
			}
			unitMap[nu.Name] = nu
			groupUnitSet.Insert(nu.Name)
		}
	}
	if ng.Spec.Selector != nil {
		labelSelector := &metav1.LabelSelector{
			MatchLabels:      ng.Spec.Selector.MatchLabels,
			MatchExpressions: ng.Spec.Selector.MatchExpressions,
		}
		unitSelector, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			return nil, nil, err
		}
		ListNodeUnitFromLister(unitLister, unitSelector, appendFunc)
	}
	if len(ng.Spec.AutoFindNodeKeys) > 0 {
		labelSelector := &metav1.LabelSelector{
			MatchLabels: map[string]string{
				constant.NodeUnitAutoFindLabel: HashAutoFindKeys(ng.Spec.AutoFindNodeKeys),
				ng.Name:                        constant.NodeGroupSuperedge,
			},
		}
		unitSelector, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			return nil, nil, err
		}
		ListNodeUnitFromLister(unitLister, unitSelector, appendFunc)
	}

	return groupUnitSet, unitMap, nil

}

func GetNodesByUnit(nodeLister corelisters.NodeLister, nu *sitev1.NodeUnit) (sets.String, map[string]*corev1.Node, error) {

	unitNodeSet := sets.NewString()
	nodeMap := make(map[string]*corev1.Node, 5)
	appendFunc := func(n interface{}) {
		node, ok := n.(*corev1.Node)
		if !ok {
			return
		}
		nodeMap[node.Name] = node
		unitNodeSet.Insert(node.Name)
	}
	if len(nu.Spec.Nodes) > 0 {
		for _, nodeName := range nu.Spec.Nodes {
			node, err := nodeLister.Get(nodeName)
			if err != nil && errors.IsNotFound(err) {
				klog.V(4).InfoS("NodeUnit node not found", "node unit", nu.Name, "node", nodeName)
				continue
			} else if err != nil {
				return nil, nil, err
			}
			nodeMap[nodeName] = node
			unitNodeSet.Insert(nodeName)
		}
	}
	if nu.Spec.Selector != nil {
		labelSelector := &metav1.LabelSelector{
			MatchLabels:      nu.Spec.Selector.MatchLabels,
			MatchExpressions: nu.Spec.Selector.MatchExpressions,
		}
		nodeSelector, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			return nil, nil, err
		}
		if (nu.Spec.Selector.MatchLabels == nil || len(nu.Spec.Selector.MatchLabels) == 0) &&
			(nu.Spec.Selector.MatchExpressions == nil || len(nu.Spec.Selector.MatchExpressions) == 0) {
			nodeSelector = labels.Nothing()
		}
		ListNodeFromLister(nodeLister, nodeSelector, appendFunc)
	}

	return unitNodeSet, nodeMap, nil
}

func SetNodeToNodes(kubeClient clientset.Interface, nu *sitev1.NodeUnit, nodeMaps map[string]*corev1.Node) error {
	klog.V(4).InfoS("SetNodeToNodes SetNode", "nu.Spec.SetNode", util.ToJson(nu.Spec.SetNode))
	for _, node := range nodeMaps {
		// first check if need update node
		labelSet, annoSet, taintSet := true, true, true
		if nu.Spec.SetNode.Labels != nil {
			for k, v := range nu.Spec.SetNode.Labels {
				if node.Labels == nil || node.Labels[k] != v {
					labelSet = false
					break
				}
			}
		}
		if nu.Spec.SetNode.Annotations != nil {
			for k, v := range nu.Spec.SetNode.Annotations {
				if node.Annotations == nil || node.Labels[k] != v {
					annoSet = false
					break
				}
			}
		}

		if nu.Spec.SetNode.Taints != nil {
			// if node.Spec.Taints ==nil || reflect.DeepEqual()
			if node.Spec.Taints == nil {
				taintSet = false
			} else {
				for _, setTaint := range nu.Spec.SetNode.Taints {
					if !TaintInSlices(node.Spec.Taints, setTaint) {
						taintSet = false
						break
					}
				}
			}
		}
		if labelSet && annoSet && taintSet {
			continue
		}
		// deep copy for don't update informer cache
		node := node.DeepCopy()
		// set labels
		if nu.Spec.SetNode.Labels != nil {
			if node.Labels == nil {
				node.Labels = make(map[string]string)
				node.Labels = nu.Spec.SetNode.Labels
			} else {
				for key, val := range nu.Spec.SetNode.Labels {
					node.Labels[key] = val
				}
			}
		}
		klog.V(4).InfoS("SetNodeToNodes node labels", "node.Labels", util.ToJson(node.Labels))

		// setNode annotations
		if nu.Spec.SetNode.Annotations != nil {
			if node.Annotations == nil {
				node.Annotations = make(map[string]string)
				node.Annotations = nu.Spec.SetNode.Annotations
			} else {
				for key, val := range nu.Spec.SetNode.Annotations {
					node.Annotations[key] = val
				}
			}
		}

		// setNode taints
		if nu.Spec.SetNode.Taints != nil {
			if node.Spec.Taints == nil {
				node.Spec.Taints = nu.Spec.SetNode.Taints
			} else {
				tmp := []corev1.Taint{}
				for _, setTaint := range nu.Spec.SetNode.Taints {
					if !TaintInSlices(node.Spec.Taints, setTaint) {
						tmp = append(tmp, setTaint)
					}
					node.Spec.Taints = append(node.Spec.Taints, tmp...)
				}
			}

		}

		if _, err := kubeClient.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{}); err != nil {
			klog.Errorf("SetNode to node update Node: %s, error: %#v", node.Name, err)
			return err
		}
	}
	return nil
}

func DeleteNodesFromSetNode(kubeClient clientset.Interface, nu *sitev1.NodeUnit, nodeMaps map[string]*corev1.Node) error {
	if len(nu.Spec.Nodes) == 0 {
		return nil
	}
	for _, node := range nodeMaps {
		newNode := node.DeepCopy()
		if nu.Spec.SetNode.Labels != nil {
			if newNode.Labels != nil {
				for k := range nu.Spec.SetNode.Labels {
					delete(newNode.Labels, k)
				}
				if !FoundNode(nu.Spec.Nodes, node.Name) {
					delete(newNode.Labels, KinsRoleLabelKey)
				}

			}
		}

		if nu.Spec.SetNode.Annotations != nil {
			if newNode.Annotations != nil {
				for k := range nu.Spec.SetNode.Annotations {
					delete(newNode.Annotations, k)
				}
			}
		}
		newNode.Spec.Taints = deleteTaintItems(newNode.Spec.Taints, nu.Spec.SetNode.Taints)

		if _, err := kubeClient.CoreV1().Nodes().Update(context.TODO(), newNode, metav1.UpdateOptions{}); err != nil {
			klog.Errorf("Update Node: %s, error: %#v", node.Name, err)
			return err
		}
		if !FoundNode(nu.Spec.Nodes, node.Name) {
			if _, ok := node.Labels[KinsRoleLabelKey]; ok {
				//delete pv
				pv, err := kubeClient.CoreV1().PersistentVolumes().Get(context.TODO(), fmt.Sprintf("%s-local-pv-%s", nu.Name, node.Name), metav1.GetOptions{})
				if err != nil {
					if !errors.IsNotFound(err) {
						return err
					}
				} else {
					_, err = kubeClient.CoreV1().PersistentVolumes().Patch(context.TODO(), pv.Name, types.StrategicMergePatchType, []byte(Finalizers), metav1.PatchOptions{})
					if err != nil {
						klog.V(4).ErrorS(err, "Patch kins pv error", "node unit", nu.Name)
						return err
					}
					if err := kubeClient.CoreV1().PersistentVolumes().Delete(context.TODO(), pv.Name, metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
						klog.V(4).ErrorS(err, "Delete kins pv error", "node unit", nu.Name)
						return err
					}
					//delete pvc
					if pv.Spec.ClaimRef != nil {
						_, err = kubeClient.CoreV1().PersistentVolumeClaims(pv.Spec.ClaimRef.Namespace).Patch(context.TODO(), pv.Spec.ClaimRef.Name, types.StrategicMergePatchType, []byte(Finalizers), metav1.PatchOptions{})
						if err != nil {
							klog.V(4).ErrorS(err, "Patch kins pvc error", "node unit", nu.Name)
							return err
						}
						if err := kubeClient.CoreV1().PersistentVolumeClaims(pv.Spec.ClaimRef.Namespace).Delete(context.TODO(), pv.Spec.ClaimRef.Name, metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
							klog.V(4).ErrorS(err, "Delete kins pvc error", "node unit", nu.Name)
							return err
						}
					}
				}
				//delete pods
				if err = kubeClient.CoreV1().Pods("kins-system").DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{
					FieldSelector: fields.SelectorFromSet(map[string]string{"spec.nodeName": node.Name}).String(),
					LabelSelector: labels.SelectorFromSet(labels.Set(map[string]string{KinsRoleLabelKey: "server"})).String(),
				}); err != nil && !errors.IsNotFound(err) {
					klog.V(4).ErrorS(err, "Delete node pods error", "node unit", nu.Name, "node name", node.Name)
					return err
				}
			}
		}
	}
	return nil
}

func DeleteNodeUnitFromSetNode(crdClient *crdClientset.Clientset, ng *sitev1.NodeGroup, unitMaps map[string]*sitev1.NodeUnit) error {
	for _, nu := range unitMaps {
		newNu := nu.DeepCopy()

		// delete node group owner label
		if newNu.Labels != nil {
			delete(newNu.Labels, ng.Name)
		}

		// delete setNode which will trigger node unit controller clean node label
		if newNu.Spec.SetNode.Labels != nil {
			delete(newNu.Spec.SetNode.Labels, ng.Name)
		}

		_, err := crdClient.SiteV1alpha2().NodeUnits().Update(context.TODO(), newNu, metav1.UpdateOptions{})
		if err != nil {
			return err
		}

	}
	return nil
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

func HashAutoFindKeys(keyslices []string) string {
	if len(keyslices) == 0 {
		return ""
	}
	sort.Strings(keyslices)

	h := sha1.New()
	for _, key := range keyslices {
		h.Write([]byte(key))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func TaintInSlices(taintSlice []corev1.Taint, target corev1.Taint) bool {
	for _, t := range taintSlice {
		if t.Key == target.Key && t.Effect == target.Effect && t.Value == target.Value {
			return true
		}
	}
	return false
}

func FoundNode(nodes []string, node string) bool {
	for _, v := range nodes {
		if v == node {
			return true
		}
	}
	return false

}
