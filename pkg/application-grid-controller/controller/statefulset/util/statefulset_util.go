/*
Copyright 2020 The SuperEdge Authors.

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

package util

import (
	"fmt"
	"strings"

	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	"k8s.io/klog/v2"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"
)

var ControllerKind = crdv1.SchemeGroupVersion.WithKind(common.StatefulSetGridKind)

func GetStatefulSetName(ssg *crdv1.StatefulSetGrid, gridValue string) string {
	return fmt.Sprintf("%s-%s", ssg.Name, gridValue)
}

func GetGridValueFromName(ssg *crdv1.StatefulSetGrid, name string) string {
	return strings.TrimPrefix(name, ssg.Name+"-")
}

func CreateStatefulSet(ssg *crdv1.StatefulSetGrid, gridValue string, sth StatefulsetTemplateHash) (*appsv1.StatefulSet, error) {
	template, err := sth.getStatefulsetTemplate(&ssg.Spec, gridValue)
	if err != nil {
		klog.Errorf("Failed to get template StatefulsetGrid %s of grid value %s, err: %v", ssg.Name, gridValue, err)
		return nil, err
	}
	set := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            GetStatefulSetName(ssg, gridValue),
			Namespace:       ssg.Namespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(ssg, ControllerKind)},
			// Append existed StatefulSetGrid labels to statefulset to be created
			Labels: func() map[string]string {
				if ssg.Labels != nil {
					newLabels := ssg.Labels
					newLabels[common.GridSelectorName] = ssg.Name
					newLabels[common.GridSelectorUniqKeyName] = ssg.Spec.GridUniqKey
					return newLabels
				} else {
					return map[string]string{
						common.GridSelectorName:        ssg.Name,
						common.GridSelectorUniqKeyName: ssg.Spec.GridUniqKey,
					}
				}
			}(),
		},
		Spec: *template,
	}

	// Append existed StatefulSetGrid NodeSelector to statefulset to be created
	if template.Template.Spec.NodeSelector != nil {
		set.Spec.Template.Spec.NodeSelector = template.Template.Spec.NodeSelector
		set.Spec.Template.Spec.NodeSelector[ssg.Spec.GridUniqKey] = gridValue
	} else {
		set.Spec.Template.Spec.NodeSelector = map[string]string{
			ssg.Spec.GridUniqKey: gridValue,
		}
	}

	return set, nil
}

func KeepConsistence(ssg *crdv1.StatefulSetGrid, set *appsv1.StatefulSet, gridValue string) *appsv1.StatefulSet {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	copyObj := set.DeepCopy()
	// Append existed StatefulSetGrid labels to statefulset to be checked
	if ssg.Labels != nil {
		copyObj.Labels = ssg.Labels
		copyObj.Labels[common.GridSelectorName] = ssg.Name
		copyObj.Labels[common.GridSelectorUniqKeyName] = ssg.Spec.GridUniqKey
	} else {
		copyObj.Labels = map[string]string{
			common.GridSelectorName:        ssg.Name,
			common.GridSelectorUniqKeyName: ssg.Spec.GridUniqKey,
		}
	}
	copyObj.Spec.Replicas = ssg.Spec.Template.Replicas
	// Updates to statefulset spec for fields other than 'replicas', 'template', and 'updateStrategy' are forbidden
	// copyObj.Spec.Selector = ssg.Spec.Template.Selector
	// copyObj.Spec.VolumeClaimTemplates = ssg.Spec.Template.VolumeClaimTemplates
	// copyObj.Spec.ServiceName = ssg.Spec.Template.ServiceName
	// TODO: this line will cause DeepEqual fails always since actual generated statefulset.Spec.Template is definitely different with ones of relevant statefulsetGrid
	copyObj.Spec.Template = ssg.Spec.Template.Template
	// Append existed StatefulSetGrid NodeSelector to statefulset to be checked
	if ssg.Spec.Template.Template.Spec.NodeSelector != nil {
		copyObj.Spec.Template.Spec.NodeSelector[ssg.Spec.GridUniqKey] = gridValue
	} else {
		copyObj.Spec.Template.Spec.NodeSelector = map[string]string{
			ssg.Spec.GridUniqKey: gridValue,
		}
	}

	return copyObj
}
