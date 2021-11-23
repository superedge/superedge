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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeUnitType string

const (
	EdgeNodeUnit  NodeUnitType = "edge"
	CloudNodeUnit NodeUnitType = "cloud"
)

// NodeUnitSpec defines the desired state of NodeUnit
type NodeSelector struct {
	// Type of nodeunit， vaule: Cloud、Edge
	// +optional
	Type NodeUnitType `json:"type,omitempty" protobuf:"bytes,2,rep,name=type"`
}

type Selector struct {
	// matchLabels is a map of {key,value} pairs.
	// +optional
	MatchLabels map[string]string `json:"matchLabels,omitempty" protobuf:"bytes,1,rep,name=matchLabels"`
	// matchExpressions is a list of label selector requirements. The requirements are ANDed.
	// +optional
	MatchExpressions []metav1.LabelSelectorRequirement `json:"matchExpressions,omitempty" protobuf:"bytes,2,rep,name=matchExpressions"`
	//If specified, select node to join nodeUnit according to Annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,12,rep,name=annotations"`
}

type SetNode struct {
	//If specified, set labels to all nodes of nodeunit
	// +optional
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,11,rep,name=labels"`

	//If specified, set annotations to all nodes of nodeunit
	// +optional
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,12,rep,name=annotations"`

	// If specified, set taints to all nodes of nodeunit
	// +optional
	Taints []corev1.Taint `json:"taints,omitempty" protobuf:"bytes,5,opt,name=taints"`
}

// NodeUnitSpec defines the desired state of NodeUnit
type NodeUnitSpec struct {
	// Type of nodeunit， vaule: Cloud、Edge
	// +optional
	//+kubebuilder:default=edge
	Type NodeUnitType `json:"type,omitempty" protobuf:"bytes,2,rep,name=type"`

	// Unschedulable controls nodeUnit schedulability of new workwolads. By default, nodeUnit is schedulable.
	// +optional
	//+kubebuilder:default=false
	Unschedulable bool `json:"unschedulable,omitempty" protobuf:"varint,4,opt,name=unschedulable"`

	// If specified, If node exists, join nodeunit directly
	// +optional
	Nodes []string `json:"nodes" protobuf:"bytes,12,rep,name=nodes"`

	// If specified, Label selector for nodes.
	// +optional
	Selector *Selector `json:"selector" protobuf:"bytes,2,opt,name=selector"`

	// If specified, set the relevant properties to the node of nodeunit.
	// +optional
	SetNode SetNode `json:"setnode" protobuf:"bytes,12,opt,name=setnode"`
	// If specified, allow to set taints to nodeunit for the scheduler to choose
	// +optional
	Taints []corev1.Taint `json:"taints,omitempty" protobuf:"bytes,5,opt,name=taints"`
}

// NodeUnitStatus defines the observed state of NodeUnit
type NodeUnitStatus struct {
	// Node that is ready in nodeunit
	//+kubebuilder:default='1/1'
	// +optional
	ReadyNodes string `json:"readynodes" protobuf:"bytes,4,rep,name=readynodes"`
	// Node selected by nodeunit
	// +optional
	Nodes []string `json:"nodes" protobuf:"bytes,12,rep,name=nodes"`
	// Node that is not ready in nodeunit
	// +optional
	NotReadyNodes []string `json:"notreadynodes" protobuf:"bytes,12,rep,name=notreadynodes"`
}

// +genclient
// +genclient:nonNamespaced
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=nu,scope=Cluster,path=nodeunits,categories=all
//+kubebuilder:printcolumn:name="TYPE",type="string",JSONPath=`.spec.type`
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=`.status.readynodes`
//+kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodeUnit is the Schema for the nodeunits API
type NodeUnit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeUnitSpec   `json:"spec,omitempty"`
	Status NodeUnitStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodeUnitList contains a list of NodeUnit
type NodeUnitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeUnit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeUnit{}, &NodeUnitList{})
}
