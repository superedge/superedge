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

var ControllerKind = crdv1.SchemeGroupVersion.WithKind("DeploymentGrid")

func GetDeploymentName(dg *crdv1.DeploymentGrid, gridValue string) string {
	return fmt.Sprintf("%s-%s", dg.Name, gridValue)
}

func GetGridValueFromName(dg *crdv1.DeploymentGrid, name string) string {
	return strings.TrimPrefix(name, dg.Name+"-")
}

func CreateDeployment(dg *crdv1.DeploymentGrid, gridValue string) *appsv1.Deployment {
	dp := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            GetDeploymentName(dg, gridValue),
			Namespace:       dg.Namespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(dg, ControllerKind)},
			Labels: map[string]string{
				common.GridSelectorName: dg.Name,
			},
		},
		Spec: dg.Spec.Template,
	}

	dp.Spec.Template.Spec.NodeSelector = map[string]string{
		dg.Spec.GridUniqKey: gridValue,
	}

	return dp
}

func KeepConsistence(dg *crdv1.DeploymentGrid, dp *appsv1.Deployment, gridValue string) *appsv1.Deployment {
	copyObj := dp.DeepCopy()
	if copyObj.Labels == nil {
		copyObj.Labels = make(map[string]string)
	}
	copyObj.Labels[common.GridSelectorName] = dg.Name
	copyObj.Spec.Replicas = dg.Spec.Template.Replicas
	copyObj.Spec.Selector = dg.Spec.Template.Selector
	// TODO: this line will cause DeepEqual fails always since actual generated deployment.Spec.Template is definitely different with ones of relevant deploymentGrid
	copyObj.Spec.Template = dg.Spec.Template.Template
	copyObj.Spec.Template.Spec.NodeSelector = map[string]string{
		dg.Spec.GridUniqKey: gridValue,
	}

	return copyObj
}
