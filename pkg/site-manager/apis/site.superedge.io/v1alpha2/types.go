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

package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
type WorkloadType string

const (
	WorkloadPod          WorkloadType = "Pod"
	WorkloadJob          WorkloadType = "Job"
	WorkloadCronjob      WorkloadType = "CronJob"
	WorkloadDeploy       WorkloadType = "Deployment"
	WorkloadService      WorkloadType = "Service"
	WorkloadReplicaSet   WorkloadType = "ReplicaSet"
	WorkloadDaemonset    WorkloadType = "DaemonSet"
	WorkloadStatuefulset WorkloadType = "StatuefulSet"
)

type Workload struct {
	// workload name
	// +optional
	Name string `json:"name,omitempty"`
	// workload type, Value can be pod, deploy, ds, service, job, st
	// +optional
	Type WorkloadType `json:"type,omitempty"`
	// If specified, Label selector for workload.
	// +optional
	Selector *Selector `json:"selector,omitempty"`
}

// NodeGroupSpec defines the desired state of NodeGroup
type NodeGroupSpec struct {
	// If specified, If nodeUnit exists, join NodeGroup directly
	// +optional
	NodeUnits []string `json:"nodeUnits,omitempty"`

	// If specified, Label selector for nodeUnit.
	// +optional
	Selector *Selector `json:"selector,omitempty"`

	// If specified, create new NodeUnits based on node have same label keys, for different values will create different nodeunites
	// +optional
	AutoFindNodeKeys []string `json:"autoFindNodeKeys,omitempty"`

	// If specified, Nodegroup bound workload
	// +optional
	Workload []Workload `json:"workload,omitempty"`
}

// NodeGroupStatus defines the observed state of NodeGroup
type WorkloadStatus struct {
	// workload Name
	// +optional
	WorkloadName string `json:"workloadName,omitempty"`
	// workload Ready Units
	// +optional
	ReadyUnit []string `json:"readyUnit,omitempty"`
	// workload NotReady Units
	// +optional
	NotReadyUnit []string `json:"notReadyUnit,omitempty"`
}

// NodeGroupStatus defines the observed state of NodeGroup
type NodeGroupStatus struct {
	// NodeUnit that is number in nodegroup
	//+kubebuilder:default=0
	// +optional
	UnitNumber int `json:"unitNumber,omitempty"`
	// Nodeunit contained in nodegroup
	// +optional
	NodeUnits []string `json:"nodeUnits,omitempty"`
	// The status of the workload in the nodegroup in each nodeunit
	// +optional
	WorkloadStatus []WorkloadStatus `json:"workloadStatus,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:shortName=ng,scope=Cluster,path=nodegroups,categories=all
// +kubebuilder:printcolumn:name="UNITS",type="integer",JSONPath=`.status.unitNumber`
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

type NodeGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeGroupSpec   `json:"spec,omitempty"`
	Status NodeGroupStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// NodeGroupList contains a list of NodeGroup
type NodeGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeGroup `json:"items"`
}

//func init() {
//	SchemeBuilder.Register(&NodeGroup{}, &NodeGroupList{})
//}

type NodeUnitType string
type AutonomyLevelType string

const (
	EdgeNodeUnit   NodeUnitType = "edge"
	CloudNodeUnit  NodeUnitType = "cloud"
	MasterNodeUnit NodeUnitType = "master"
	OtherNodeUnit  NodeUnitType = "other"

	AutonomyLevelL5 AutonomyLevelType = "L5"
	AutonomyLevelL4 AutonomyLevelType = "L4"
	AutonomyLevelL3 AutonomyLevelType = "L3"

	UnitClusterStorageTypeSqlite = "sqlite"
	UnitClusterStorageTypeETCD   = "etcd"
)

// ConditionStatus defines the status of Condition.
type ConditionStatus string

// These are valid condition statuses.
// "ConditionTrue" means a resource is in the condition.
// "ConditionFalse" means a resource is not in the condition.
// "ConditionUnknown" means server can't decide if a resource is in the condition
// or not.
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// ClusterCondition contains details for the current condition of this cluster.
type ClusterCondition struct {
	// Type is the type of the condition.
	Type string `json:"type,omitempty"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status ConditionStatus `json:"status,omitempty"`
	// Last time we probed the condition.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// ClusterPhase defines the phase of cluster constructor.
type ClusterPhase string

const (
	// ClusterInitializing is the initialize phase.
	ClusterInitializing ClusterPhase = "Initializing"
	// ClusterRunning is the normal running phase.
	ClusterRunning ClusterPhase = "Running"
	// ClusterFailed is the failed phase.
	ClusterFailed ClusterPhase = "Failed"
	// ClusterConfined is the Confined phase.
	ClusterConfined ClusterPhase = "Confined"
	// ClusterIdling is the Idling phase.
	ClusterIdling ClusterPhase = "Idling"
	// ClusterUpgrading means that the cluster is in upgrading process.
	ClusterUpgrading ClusterPhase = "Upgrading"
	// ClusterTerminating means the cluster is undergoing graceful termination.
	ClusterTerminating ClusterPhase = "Terminating"
	// ClusterUpscaling means the cluster is undergoing graceful up scaling.
	ClusterUpscaling ClusterPhase = "Upscaling"
	// ClusterDownscaling means the cluster is undergoing graceful down scaling.
	ClusterDownscaling ClusterPhase = "Downscaling"
)

// AddressType indicates the type of cluster apiserver access address.
type AddressType string

// These are valid address type of cluster.
const (
	// AddressPublic indicates the address of the apiserver accessed from the external network.(such as public lb)
	AddressPublic AddressType = "Public"
	// AddressAdvertise indicates the address of the apiserver accessed from the worker node.(such as internal lb)
	AddressAdvertise AddressType = "Advertise"
	// AddressReal indicates the real address of one apiserver
	AddressReal AddressType = "Real"
	// AddressInternal indicates the address of the apiserver accessed from TKE control plane.
	AddressInternal AddressType = "Internal"
	// AddressSupport used for vpc lb which bind to JNS gateway as known AddressInternal
	AddressSupport AddressType = "Support"
)

// ClusterAddress contains information for the cluster's address.
type ClusterAddress struct {
	// Cluster address type, one of Public, ExternalIP or InternalIP.
	Type AddressType `json:"type,omitempty"`
	// The cluster address.
	Host string `json:"host,omitempty"`
	Port int32  `json:"port,omitempty"`
	Path string `json:"path,omitempty"`
}

// ResourceList is a set of (resource name, quantity) pairs.
type ResourceList map[string]resource.Quantity

// ClusterResource records the current available and maximum resource quota
// information for the cluster.
type ClusterResource struct {
	// Capacity represents the total resources of a cluster.
	// +optional
	Capacity ResourceList `json:"capacity,omitempty"`
	// Allocatable represents the resources of a cluster that are available for scheduling.
	// Defaults to Capacity.
	// +optional
	Allocatable ResourceList `json:"allocatable,omitempty"`
	// +optional
	Allocated ResourceList `json:"allocated,omitempty"`
}
type Selector struct {
	// matchLabels is a map of {key,value} pairs.
	// +optional
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
	// matchExpressions is a list of label selector requirements. The requirements are ANDed.
	// +optional
	MatchExpressions []metav1.LabelSelectorRequirement `json:"matchExpressions,omitempty"`
	//If specified, select node to join nodeUnit according to Annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

type SetNode struct {
	//If specified, set labels to all nodes of nodeunit
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	//If specified, set annotations to all nodes of nodeunit
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// If specified, set taints to all nodes of nodeunit
	// +optional
	Taints []corev1.Taint `json:"taints,omitempty"`
}

// UnitClusterInfoSpec defines the information when unit cluster creating
type UnitClusterInfoSpec struct {
	// StorageType support sqlite(one master node) and built-in etcd(three master node)
	// default is etcd
	// +optional
	StorageType string `json:"storageType,omitempty"`
	// Parameters holds the parameters for the unit cluster create information
	Parameters map[string]string `json:"parameters,omitempty"`
}

// NodeUnitSpec defines the desired state of NodeUnit
type NodeUnitSpec struct {
	// Type of nodeunit, vaule: cloud, edge, master, other
	// +optional
	//+kubebuilder:default=edge
	Type NodeUnitType `json:"type"`

	// Unschedulable controls nodeUnit schedulability of new workwolads. By default, nodeUnit is schedulable.
	// +optional
	//+kubebuilder:default=false
	Unschedulable bool `json:"unschedulable,omitempty"`

	// If specified, If node exists, join nodeunit directly
	// +optional
	Nodes []string `json:"nodes,omitempty"`

	// If specified, Label selector for nodes.
	// +optional
	Selector *Selector `json:"selector,omitempty"`

	// If specified, set the relevant properties to the node of nodeunit.
	// +optional
	SetNode SetNode `json:"setNode,omitempty"`

	// AutonomyLevel represent the current node unit autonomous capability, L3(default)'s autonomous area is node,
	// L4's autonomous area is unit. If AutonomyLevel larger than L3, it will create a independent control plane in unit.
	// +optional
	//+kubebuilder:default=L3
	//+k8s:conversion-gen=false
	AutonomyLevel AutonomyLevelType `json:"autonomyLevel,omitempty"`
	// UnitCredentialConfigMapRef for isolate sensitive NodeUnit credential.
	// site-manager will create one after controller-plane ready
	// +optional
	//+k8s:conversion-gen=false
	UnitCredentialConfigMapRef *corev1.ObjectReference `json:"unitCredentialConfigMapRef,omitempty"`

	// UnitClusterInfo holds configuration for unit cluster information.
	// +optional
	//+k8s:conversion-gen=false
	UnitClusterInfo *UnitClusterInfoSpec `json:"unitClusterInfo,omitempty"`
}

type UnitClusterStatus struct {
	// If AutonomyLevel larger than L3, it will create a independent control plane in unit,
	// +optional
	Version string `json:"version,omitempty"`
	// +optional
	Phase ClusterPhase `json:"phase,omitempty"`
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []ClusterCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	// +optional
	Addresses []ClusterAddress `json:"addresses,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	// +optional
	ClusterResource ClusterResource `json:"resource,omitempty"`
	// +optional
	ServiceCIDR string `json:"serviceCIDR,omitempty"`
}

// NodeUnitStatus defines the observed state of NodeUnit
type NodeUnitStatus struct {
	// Node that is ready in nodeunit
	//+kubebuilder:default='1/1'
	// +optional
	ReadyRate string `json:"readyRate,omitempty"`
	// Node selected by nodeunit
	// +optional
	ReadyNodes []string `json:"readyNodes,omitempty"`
	// Node that is not ready in nodeunit
	// +optional
	NotReadyNodes []string `json:"notReadyNodes,omitempty"`

	// UnitClusterStatus is not nil, when AutonomyLevel is larger than L3
	// +optional
	//+k8s:conversion-gen=false
	UnitCluster UnitClusterStatus `json:"unitClusterStatus,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=nu,scope=Cluster
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="TYPE",type="string",JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=`.status.readyRate`
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="DELETING",type="date",JSONPath=".metadata.deletionTimestamp"

// NodeUnit is the Schema for the nodeunits API
type NodeUnit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeUnitSpec   `json:"spec,omitempty"`
	Status NodeUnitStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// NodeUnitList contains a list of NodeUnit
type NodeUnitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeUnit `json:"items"`
}
