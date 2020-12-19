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
	"superedge/pkg/application-grid-controller/controller/common"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	crdv1 "superedge/pkg/application-grid-controller/apis/superedge.io/v1"
)

func GetDeploymentName(g *crdv1.DeploymentGrid, gridValue string) string {
	return fmt.Sprintf("%s-%s", g.Name, gridValue)
}

func GetGridValueFromName(g *crdv1.DeploymentGrid, name string) string {
	return strings.TrimPrefix(name, g.Name+"-")
}

func CreateDeployment(g *crdv1.DeploymentGrid, gridValue string) *appsv1.Deployment {
	dp := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetDeploymentName(g, gridValue),
			Namespace: g.Namespace,
			Labels: map[string]string{
				common.GridSelectorName: g.Name,
			},
		},
		Spec: g.Spec.Template,
	}

	dp.Spec.Template.Spec.NodeSelector = map[string]string{
		g.Spec.GridUniqKey: gridValue,
	}

	return dp
}

func KeepConsistence(g *crdv1.DeploymentGrid, dp *appsv1.Deployment, gridValue string) *appsv1.Deployment {
	copyObj := dp.DeepCopy()
	if copyObj.Labels == nil {
		copyObj.Labels = make(map[string]string)
	}
	copyObj.Labels[common.GridSelectorName] = g.Name
	copyObj.Spec.Replicas = g.Spec.Template.Replicas
	copyObj.Spec.Selector = g.Spec.Template.Selector
	copyObj.Spec.Template = g.Spec.Template.Template
	copyObj.Spec.Template.Spec.NodeSelector = map[string]string{
		g.Spec.GridUniqKey: gridValue,
	}

	return copyObj
}
