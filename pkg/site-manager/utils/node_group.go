/*
Copyright 2022 The SuperEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha1"
	siteClientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	sitecrdClientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	"github.com/superedge/superedge/pkg/util"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"sort"
	"strings"
)

func AutoFindNodeKeysbyNodeGroup(kubeclient clientset.Interface, crdClient *sitecrdClientset.Clientset, nodeGroup *v1alpha1.NodeGroup) {
	// find nodes by keys
	allnodes, err := kubeclient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf(err.Error())
	}
	matchNodes := []v1.Node{}

	for _, node := range allnodes.Items {
		if len(node.Labels) == 0 {
			continue
		}
		result, res, sel := checkifcontains(node.Labels, nodeGroup.Spec.AutoFindNodeKeys)
		if result && len(sel) > 0 {
			matchNodes = append(matchNodes, node)
			newNodeUnit(crdClient, nodeGroup, res, sel)
		}
	}
}

func filterString(name string) string {
	if withCheckContains(name) || withCheckSize(name) {
		return hashString(name)
	}
	return name
}

func hashString(name string) string {
	h := sha1.New()
	h.Write([]byte(name))
	sha1_hash := hex.EncodeToString(h.Sum(nil))
	return sha1_hash
}

func withCheckContains(name string) bool {
	if strings.Contains(name, "/") {
		return true
	}
	return false
}

// check size is it more than 64
func withCheckSize(name string) bool {
	if len(name) >= 64 {
		return true
	}
	return false
}

func newNodeUnit(crdClient *sitecrdClientset.Clientset, nodeGroup *v1alpha1.NodeGroup, name string, sel map[string]string) error {
	newname := filterString(name)
	klog.Info("prepare to ceate nodeunite ", newname)
	klog.Info("selector is ", sel)

	ng, err := crdClient.SiteV1alpha1().NodeGroups().Get(context.TODO(), nodeGroup.Name, metav1.GetOptions{})
	if err != nil {
		klog.Error("get nodegroup fail", err)
	}
	klog.Info("kind, apiversion, name, uid is ", ng.Kind, ng.APIVersion, nodeGroup.Name, ng.UID)

	var nuLabel = make(map[string]string)
	nuLabel[nodeGroup.Name] = "autofindnodekeys"

	newNodeUnit := &v1alpha1.NodeUnit{
		ObjectMeta: metav1.ObjectMeta{
			Name:        newname,
			Annotations: sel,
			Labels:      nuLabel,
			OwnerReferences: []metav1.OwnerReference{{
				Kind:       "NodeGroup",
				APIVersion: "site.superedge.io/v1alpha1",
				Name:       nodeGroup.Name,
				UID:        ng.UID,
			},
			},
		},
		Spec: v1alpha1.NodeUnitSpec{
			Type: EdgeNodeUnit,
			Selector: &v1alpha1.Selector{
				MatchLabels: sel,
			},
		},
	}
	klog.Info("create nodeunit json is ", util.ToJson(newNodeUnit))
	get, err := crdClient.SiteV1alpha1().NodeUnits().Get(context.TODO(), newname, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		klog.Warning("obj not found, will create nodeunit now")
		_, err = crdClient.SiteV1alpha1().NodeUnits().Create(context.TODO(), newNodeUnit, metav1.CreateOptions{})
		if err != nil {
			klog.Warningf("error to create nodeunites", err)
			return err
		}

	} else if err == nil {
		tmpSel := &v1alpha1.Selector{
			MatchLabels: sel,
		}
		if get.Spec.Selector != tmpSel {
			get.Spec.Selector = tmpSel
		}
		get.Labels = nuLabel
		tmpOwner := metav1.OwnerReference{
			Kind:       "NodeGroup",
			APIVersion: "site.superedge.io/v1alpha1",
			Name:       nodeGroup.Name,
			UID:        ng.UID,
		}
		if !checkOwnerReferenceContains(tmpOwner, get.OwnerReferences) {
			get.OwnerReferences = append(get.OwnerReferences, metav1.OwnerReference{
				Kind:       "NodeGroup",
				APIVersion: "site.superedge.io/v1alpha1",
				Name:       nodeGroup.Name,
				UID:        ng.UID,
			})
		}

		crdClient.SiteV1alpha1().NodeUnits().Update(context.TODO(), get, metav1.UpdateOptions{})
	} else {
		return err
	}

	return nil
}

func checkOwnerReferenceContains(owner metav1.OwnerReference, tmpSlice []metav1.OwnerReference) bool {
	for _, value := range tmpSlice {
		if value == owner {
			return true
		}
	}
	return false
}

func checkifcontains(nodelabel map[string]string, keyslices []string) (bool, string, map[string]string) {
	var res string
	var sel = make(map[string]string)
	sort.Strings(keyslices)
	for _, value := range keyslices {
		if _, ok := nodelabel[value]; ok {
			sel[value] = nodelabel[value]
			if res == "" {
				res = nodelabel[value]
			} else {
				res = res + "-" + nodelabel[value]
			}

			continue
		} else {
			return false, "", sel
		}
	}

	return true, res, sel
}

func GetUnitsByNodeGroup(kubeClient clientset.Interface, siteClient *siteClientset.Clientset, nodeGroup *v1alpha1.NodeGroup) (nodeUnits []string, err error) {
	// Get units by selector
	var unitList *v1alpha1.NodeUnitList
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
			unitList, err = siteClient.SiteV1alpha1().NodeUnits().List(context.TODO(), listOptions)
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
		unit, err := siteClient.SiteV1alpha1().NodeUnits().Get(context.TODO(), unitName, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				klog.Warningf("Get nodeGroup: %s unit nil", nodeGroup.Name)
				continue
			} else {
				klog.Errorf("Get unit by nodeGroup, error: %v", err)
				return nodeUnits, err
			}
		}
		nodeUnits = append(nodeUnits, unit.Name)
	}

	klog.Info("node units is ", nodeUnits)

	copyNodeUnits := make([]string, len(nodeUnits))
	copy(copyNodeUnits, nodeUnits)

	klog.Info("copy node unit is ", copyNodeUnits)
	UpdateNodeLabels(kubeClient, siteClient, copyNodeUnits, nodeGroup.Name)

	if len(nodeGroup.Spec.AutoFindNodeKeys) > 0 {
		nulist, err := siteClient.SiteV1alpha1().NodeUnits().List(context.TODO(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf(nodeGroup.Name + "=autofindnodekeys"),
		})
		if err != nil {
			klog.Errorf("Get unit by nodeGroup, error: %v", err)
		}

		klog.Info("Get unit by nodeGroup number is ", len(nulist.Items))
		for _, nu := range nulist.Items {
			nodeUnits = append(nodeUnits, nu.Name)
		}
	}

	return util.RemoveDuplicateElement(nodeUnits), nil
}

func GetNodeGroupsByUnit(siteClient *siteClientset.Clientset, unitName string) (nodeGroups []*v1alpha1.NodeGroup, err error) {
	allNodeGroups, err := siteClient.SiteV1alpha1().NodeGroups().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			klog.Warningf("Unit:%s does not belong to any nodeGroup", unitName)
			return
		} else {
			klog.Errorf("Get nodeGroup by unit, error: %v", err)
			return nil, err
		}
	}

	for _, nodeGroup := range allNodeGroups.Items {
		for _, unit := range nodeGroup.Status.NodeUnits {
			if unit == unitName {
				nodeGroups = append(nodeGroups, &nodeGroup)
			}
		}
	}
	return nodeGroups, nil
}

func UnitMatchNodeGroups(kubeClient clientset.Interface, siteClient *siteClientset.Clientset, unitName string) (nodeGroups []*v1alpha1.NodeGroup, err error) {
	allNodeGroups, err := siteClient.SiteV1alpha1().NodeGroups().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			klog.Warningf("Unit:%s does not belong to any nodeGroup", unitName)
			return
		} else {
			klog.Errorf("Get nodeGroup by unit, error: %v", err)
			return nil, err
		}
	}

	for _, nodeGroup := range allNodeGroups.Items {
		units, err := GetUnitsByNodeGroup(kubeClient, siteClient, &nodeGroup)
		if err != nil {
			klog.Errorf("Get NodeGroup unit error: %v", err)
			continue
		}

		for _, unit := range units {
			if unit == unitName {
				nodeGroups = append(nodeGroups, &nodeGroup)
			}
		}
	}

	return nodeGroups, nil
}

func UpdateNodeLabels(kubeClient clientset.Interface, crdClient *siteClientset.Clientset, units []string, nodeGroup string) {
	klog.Info("prepare to update nodeunits, ", units)
	for _, nuName := range units {
		obj, err := crdClient.SiteV1alpha1().NodeUnits().Get(context.TODO(), nuName, metav1.GetOptions{})
		if err != nil {
			klog.Error("Get nodeunit fail, ", err)
		}

		var tmpLabel = make(map[string]string)
		tmpLabel[nodeGroup] = nuName
		if obj.Spec.SetNode.Labels == nil {
			obj.Spec.SetNode.Labels = tmpLabel
		} else {
			obj.Spec.SetNode.Labels[nodeGroup] = nuName
		}

		_, err = crdClient.SiteV1alpha1().NodeUnits().Update(context.TODO(), obj, metav1.UpdateOptions{})
		if err != nil {
			klog.Error("Update NodeUnit fail ", err)
		}
	}
}
