/*
Copyright 2021 The SuperEdge authors.

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
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EDeploymentSpec defines the desired state of EDeployment
type EDeploymentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of EDeployment. Edit edeployment_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// EDeploymentStatus defines the observed state of EDeployment
type EDeploymentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=edeploy
//+kubebuilder:printcolumn:name="READY",type="integer",JSONPath=`.status.readyReplicas`
//+kubebuilder:printcolumn:name="UP-TO-DATE",type="integer",JSONPath=`.status.updatedReplicas`
//+kubebuilder:printcolumn:name="AVAILABLE",type="integer",JSONPath=`.status.availableReplicas`
//+kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EDeployment is the Schema for the edeployments API
type EDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   appsv1.DeploymentSpec   `json:"spec,omitempty"`
	Status appsv1.DeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EDeploymentList contains a list of EDeployment
type EDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EDeployment{}, &EDeploymentList{})
}
