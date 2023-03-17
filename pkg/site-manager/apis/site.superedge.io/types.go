package site

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkloadType string

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NodeUnit struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   NodeUnitSpec
	Status NodeUnitStatus
}

// NodeUnitSpec defines the desired state of NodeUnit
type NodeUnitSpec struct {
	// Type of nodeunit， vaule: Cloud、Edge
	// +optional
	//+kubebuilder:default=edge
	Type NodeUnitType

	// Unschedulable controls nodeUnit schedulability of new workwolads. By default, nodeUnit is schedulable.
	// +optional
	//+kubebuilder:default=false
	Unschedulable bool

	// If specified, If node exists, join nodeunit directly
	// +optional
	Nodes []string

	// If specified, Label selector for nodes.
	// +optional
	Selector *Selector

	// If specified, set the relevant properties to the node of nodeunit.
	// +optional
	SetNode SetNode
}

// NodeUnitStatus defines the observed state of NodeUnit
type NodeUnitStatus struct {
	// Node that is ready in nodeunit
	//+kubebuilder:default='1/1'
	// +optional
	ReadyRate string
	// Node selected by nodeunit
	// +optional
	ReadyNodes []string
	// Node that is not ready in nodeunit
	// +optional
	NotReadyNodes []string
}

type NodeUnitType string

type Selector struct {
	// matchLabels is a map of {key,value} pairs.
	// +optional
	MatchLabels map[string]string
	// matchExpressions is a list of label selector requirements. The requirements are ANDed.
	// +optional
	MatchExpressions []metav1.LabelSelectorRequirement
	//If specified, select node to join nodeUnit according to Annotations
	// +optional
	Annotations map[string]string
}

type SetNode struct {
	//If specified, set labels to all nodes of nodeunit
	// +optional
	Labels map[string]string

	//If specified, set annotations to all nodes of nodeunit
	// +optional
	Annotations map[string]string

	// If specified, set taints to all nodes of nodeunit
	// +optional
	Taints []corev1.Taint
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodeUnitList contains a list of NodeUnit
type NodeUnitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeUnit `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NodeGroup struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   NodeGroupSpec
	Status NodeGroupStatus
}

// NodeGroupSpec defines the desired state of NodeGroup
type NodeGroupSpec struct {
	// If specified, If nodeUnit exists, join NodeGroup directly
	// +optional
	NodeUnits []string

	// If specified, Label selector for nodeUnit.
	// +optional
	Selector *Selector

	// If specified, create new NodeUnits based on node have same label keys, for different values will create different nodeunites
	// +optional
	AutoFindNodeKeys []string

	// If specified, Nodegroup bound workload
	// +optional
	Workload []Workload
}

type Workload struct {
	// workload name
	// +optional
	Name string
	// workload type, Value can be pod, deploy, ds, service, job, st
	// +optional
	Type WorkloadType
	// If specified, Label selector for workload.
	// +optional
	Selector *Selector
}

// NodeGroupStatus defines the observed state of NodeGroup
type NodeGroupStatus struct {
	// NodeUnit that is number in nodegroup
	//+kubebuilder:default=0
	// +optional
	UnitNumber int
	// Nodeunit contained in nodegroup
	// +optional
	NodeUnits []string
	// The status of the workload in the nodegroup in each nodeunit
	// +optional
	WorkloadStatus []WorkloadStatus
}

// NodeGroupStatus defines the observed state of NodeGroup
type WorkloadStatus struct {
	// workload Name
	// +optional
	WorkloadName string
	// workload Ready Units
	// +optional
	ReadyUnit []string
	// workload NotReady Units
	// +optional
	NotReadyUnit []string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NodeGroupList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []NodeGroup
}
