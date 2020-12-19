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

package v1

import (
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ServiceGrid struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ServiceGridSpec `json:"spec,omitempty"`
}

type ServiceGridSpec struct {
	GridUniqKey string         `json:"gridUniqKey,omitempty"`
	Template    v1.ServiceSpec `json:"template,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DeploymentGrid struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeploymentGridSpec       `json:"spec,omitempty"`
	Status DeploymentGridStatusList `json:"status,omitempty"`
}

type DeploymentGridStatusList struct {
	States map[string]appv1.DeploymentStatus `json:"states,omitempty"`
}

type DeploymentGridSpec struct {
	GridUniqKey string               `json:"gridUniqKey,omitempty"`
	Template    appv1.DeploymentSpec `json:"template,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceGridList is a list of ServiceGrid resources
type ServiceGridList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ServiceGrid `json:"items"`
}

func NewServiceGrid(namespace, name string, obj ServiceGrid) *ServiceGrid {
	obj.APIVersion, obj.Kind = SchemeGroupVersion.WithKind("ServiceGrid").ToAPIVersionAndKind()
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DeploymentGridList is a list of DeploymentGrid resources
type DeploymentGridList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DeploymentGrid `json:"items"`
}

func NewDeploymentGrid(namespace, name string, obj DeploymentGrid) *DeploymentGrid {
	obj.APIVersion, obj.Kind = SchemeGroupVersion.WithKind("DeploymentGrid").ToAPIVersionAndKind()
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}
