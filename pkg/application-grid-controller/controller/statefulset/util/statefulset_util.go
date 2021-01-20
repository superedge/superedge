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
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	"strings"

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

func CreateStatefulSet(ssg *crdv1.StatefulSetGrid, gridValue string) *appsv1.StatefulSet {
	set := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            GetStatefulSetName(ssg, gridValue),
			Namespace:       ssg.Namespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(ssg, ControllerKind)},
			Labels: map[string]string{
				common.GridSelectorName: ssg.Name,
			},
		},
		Spec: ssg.Spec.Template,
	}

	set.Spec.Template.Spec.NodeSelector = map[string]string{
		ssg.Spec.GridUniqKey: gridValue,
	}

	return set
}

func KeepConsistence(ssg *crdv1.StatefulSetGrid, set *appsv1.StatefulSet, gridValue string) *appsv1.StatefulSet {
	copyObj := set.DeepCopy()
	if copyObj.Labels == nil {
		copyObj.Labels = make(map[string]string)
	}
	copyObj.Labels[common.GridSelectorName] = ssg.Name
	copyObj.Spec.Replicas = ssg.Spec.Template.Replicas
	copyObj.Spec.Selector = ssg.Spec.Template.Selector
	// TODO: this line will cause DeepEqual fails always since actual generated statefulset.Spec.Template is definitely different with ones of relevant statefulsetGrid
	copyObj.Spec.Template = ssg.Spec.Template.Template
	copyObj.Spec.Template.Spec.NodeSelector = map[string]string{
		ssg.Spec.GridUniqKey: gridValue,
	}

	return copyObj
}
