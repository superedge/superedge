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
		result, res := checkifcontains(node.Labels, nodeGroup.Spec.AutoFindNodeKeys)
		if result {
			matchNodes = append(matchNodes, node)
			newNodeUnit(crdClient, nodeGroup, res, nodeGroup.Namespace, node.Name)
		}
	}
}

func newNodeUnit(crdClient *sitecrdClientset.Clientset, nodeGroup *v1alpha1.NodeGroup, name string, namespace string, Nodes string) error {
	newNodeUnit := &v1alpha1.NodeUnit{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{{
				Kind:       nodeGroup.Kind,
				APIVersion: nodeGroup.APIVersion,
				Name:       nodeGroup.Name,
				UID:        nodeGroup.UID,
			},
			},
		},
		Spec: v1alpha1.NodeUnitSpec{
			Type:  EdgeNodeUnit,
			Nodes: []string{Nodes},
		},
	}
	get, err := crdClient.SiteV1alpha1().NodeUnits().Get(context.TODO(), name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		klog.Warning("obj not found, will create nodeunit now")
		_, err = crdClient.SiteV1alpha1().NodeUnits().Create(context.TODO(), newNodeUnit, metav1.CreateOptions{})
		if err != nil {
			klog.Warningf("error to create nodeunites", err)
		}
	} else {
		return err
	}
	get.Spec.Nodes = append(get.Spec.Nodes)

	get.OwnerReferences = append(get.OwnerReferences, metav1.OwnerReference{
		Kind:       nodeGroup.Kind,
		APIVersion: nodeGroup.APIVersion,
		Name:       nodeGroup.Name,
		UID:        nodeGroup.UID,
	})
	crdClient.SiteV1alpha1().NodeUnits().Update(context.TODO(), get, metav1.UpdateOptions{})
	return nil

}

func checkifcontains(nodelabel map[string]string, keyslices []string) (bool, string) {
	//tmplabel := map[string]string{}
	var res string
	sort.Strings(keyslices)
	for _, value := range keyslices {
		if _, ok := nodelabel[value]; ok {
			//tmplabel[value] = nodelabel[value]
			res = res + value + "-" + nodelabel[value]
			continue
		} else {
			return false, ""
		}
	}

	return true, res
}

func GetUnitsByNodeGroup(siteClient *siteClientset.Clientset, nodeGroup *v1alpha1.NodeGroup) (nodeUnits []string, err error) {
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

func UnitMatchNodeGroups(siteClient *siteClientset.Clientset, unitName string) (nodeGroups []*v1alpha1.NodeGroup, err error) {
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
		units, err := GetUnitsByNodeGroup(siteClient, &nodeGroup)
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
