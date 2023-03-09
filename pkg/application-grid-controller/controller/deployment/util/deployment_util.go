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

var ControllerKind = crdv1.SchemeGroupVersion.WithKind(common.DeploymentGridKind)

func GetDeploymentName(dg *crdv1.DeploymentGrid, gridValue string) string {
	return fmt.Sprintf("%s-%s", dg.Name, gridValue)
}

func GetGridValueFromName(dg *crdv1.DeploymentGrid, name string) string {
	return strings.TrimPrefix(name, dg.Name+"-")
}

func CreateDeployment(dg *crdv1.DeploymentGrid, gridValue string, dth DeploymentTemplateHash) (*appsv1.Deployment, error) {
	template, err := dth.getDeployTemplate(&dg.Spec, gridValue)
	if err != nil {
		klog.Errorf("Failed to get template deploymentgrid %s of grid value %s, err: %v", dg.Name, gridValue, err)
		return nil, err
	}

	dp := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            GetDeploymentName(dg, gridValue),
			Namespace:       dg.Namespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(dg, ControllerKind)},
			// Append existed DeploymentGrid labels to deployment to be created
			Labels: func() map[string]string {
				if dg.Labels != nil {
					newLabels := dg.Labels
					newLabels[common.GridSelectorName] = dg.Name
					newLabels[common.GridSelectorUniqKeyName] = dg.Spec.GridUniqKey
					return newLabels
				} else {
					return map[string]string{
						common.GridSelectorName:        dg.Name,
						common.GridSelectorUniqKeyName: dg.Spec.GridUniqKey,
					}
				}
			}(),
		},
		Spec: *template,
	}

	// Append existed DeploymentGrid NodeSelector to deployment to be created
	if template.Template.Spec.NodeSelector != nil {
		dp.Spec.Template.Spec.NodeSelector = template.Template.Spec.NodeSelector
		dp.Spec.Template.Spec.NodeSelector[dg.Spec.GridUniqKey] = gridValue
	} else {
		dp.Spec.Template.Spec.NodeSelector = map[string]string{
			dg.Spec.GridUniqKey: gridValue,
		}
	}

	return dp, nil
}

func CreateDeploymentGrid(dg *crdv1.DeploymentGrid, nameSpace string) *crdv1.DeploymentGrid {
	dgcopy := dg.DeepCopy()
	dgcopy.ResourceVersion = ""
	TargetNameSpace := dgcopy.Namespace
	dgcopy.Namespace = nameSpace
	dgcopy.Labels[common.FedTargetNameSpace] = TargetNameSpace
	dgcopy.Labels[common.FedrationDisKey] = "yes"
	dgcopy.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(dg, ControllerKind)}
	return dgcopy
}

func UpdateDeploymentGrid(dg, fed *crdv1.DeploymentGrid) *crdv1.DeploymentGrid {
	dgcopy := fed.DeepCopy()
	dgcopy.Spec = dg.Spec
	return dgcopy
}

func KeepConsistence(dg *crdv1.DeploymentGrid, dp *appsv1.Deployment, gridValue string) *appsv1.Deployment {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	copyObj := dp.DeepCopy()
	// Append existed DeploymentGrid labels to deployment to be checked
	if dg.Labels != nil {
		copyObj.Labels = dg.Labels
		copyObj.Labels[common.GridSelectorName] = dg.Name
		copyObj.Labels[common.GridSelectorUniqKeyName] = dg.Spec.GridUniqKey
	} else {
		copyObj.Labels = map[string]string{
			common.GridSelectorName:        dg.Name,
			common.GridSelectorUniqKeyName: dg.Spec.GridUniqKey,
		}
	}
	copyObj.Spec.Replicas = dg.Spec.Template.Replicas
	// Spec.selector field is immutable
	// copyObj.Spec.Selector = dg.Spec.Template.Selector
	// TODO: this line will cause DeepEqual fails always since actual generated deployment.Spec.Template is definitely different with ones of relevant deploymentGrid
	copyObj.Spec.Template = dg.Spec.Template.Template
	// Append existed DeploymentGrid NodeSelector to deployment to be checked
	if dg.Spec.Template.Template.Spec.NodeSelector != nil {
		copyObj.Spec.Template.Spec.NodeSelector[dg.Spec.GridUniqKey] = gridValue
	} else {
		copyObj.Spec.Template.Spec.NodeSelector = map[string]string{
			dg.Spec.GridUniqKey: gridValue,
		}
	}

	return copyObj
}
