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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkloadType string

const (
	WorkloadPod          WorkloadType = "pod"
	WorkloadJob          WorkloadType = "job"
	WorkloadDeploy       WorkloadType = "deploy"
	WorkloadService      WorkloadType = "service"
	WorkloadDaemonset    WorkloadType = "daemonset"
	WorkloadStatuefulset WorkloadType = "statuefulset"
)

type Workload struct {
	// workload name
	// +optional
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`
	// workload type, Value can be pod, deploy, ds, service, job, st
	// +optional
	Type WorkloadType `json:"type" protobuf:"bytes,2,opt,name=type"`
	// If specified, Label selector for workload.
	// +optional
	Selector *Selector `json:"selector" protobuf:"bytes,2,opt,name=selector"`
}

// NodeGroupSpec defines the desired state of NodeGroup
type NodeGroupSpec struct {
	// If specified, If nodeUnit exists, join NodeGroup directly
	// +optional
	NodeUnits []string `json:"nodeunits" protobuf:"bytes,12,rep,name=nodeunits"`

	// If specified, Label selector for nodeUnit.
	// +optional
	Selector *Selector `json:"selector" protobuf:"bytes,2,opt,name=selector"`

	// If specified, Nodegroup bound workload
	// +optional
	Workload []Workload `json:"workload" protobuf:"bytes,12,rep,name=workload"`
}

// NodeGroupStatus defines the observed state of NodeGroup
type WorkloadStatus struct {
	// workload Name
	// +optional
	WorkloadName string `json:"workloadname" protobuf:"bytes,12,rep,name=workloadname"`
	// workload Ready Units
	// +optional
	ReadyUnit []string `json:"readyunit" protobuf:"bytes,12,rep,name=readyunit"`
	// workload NotReady Units
	// +optional
	NotReadyUnit []string `json:"notreadyunit" protobuf:"bytes,12,rep,name=notreadyunit"`
}

// NodeGroupStatus defines the observed state of NodeGroup
type NodeGroupStatus struct {
	// Nodeunit contained in nodegroup
	// +optional
	NodeUnits []string `json:"nodeunits" protobuf:"bytes,12,rep,name=nodeunits"`

	// The status of the workload in the nodegroup in each nodeunit
	// +optional
	WorkloadStatus []WorkloadStatus `json:"workloadstatus" protobuf:"bytes,12,rep,name=workloadstatus"`
}

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NodeGroup is the Schema for the nodegroups API
type NodeGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeGroupSpec   `json:"spec,omitempty"`
	Status NodeGroupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodeGroupList contains a list of NodeGroup
type NodeGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeGroup{}, &NodeGroupList{})
}
