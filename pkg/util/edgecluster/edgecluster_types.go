/*


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

package edgecluster

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterPhase defines the phase of cluster constructor.
type ClusterPhase string

const (
	// ClusterRunning is the normal running phase.
	ClusterRunning ClusterPhase = "Running"
	// ClusterInitializing is the initialize phase.
	ClusterInitializing ClusterPhase = "Initializing"
	// ClusterFailed is the failed phase.
	ClusterFailed ClusterPhase = "Failed"
	// ClusterTerminating means the cluster is undergoing graceful termination.
	ClusterTerminating ClusterPhase = "Terminating"
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
	Type string `json:"type" protobuf:"bytes,1,opt,name=type"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// Last time we probed the condition.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

// ResourceList is a set of (resource name, quantity) pairs.
type ResourceList map[string]resource.Quantity

// ResourceRequirements describes the compute resource requirements.
type ResourceRequirements struct {
	Limits   ResourceList `json:"limits,omitempty" protobuf:"bytes,1,rep,name=limits,casttype=ResourceList"`
	Requests ResourceList `json:"requests,omitempty" protobuf:"bytes,2,rep,name=requests,casttype=ResourceList"`
}

// ClusterResource records the current available and maximum resource quota
// information for the cluster.
type ClusterResource struct {
	// Capacity represents the total resources of a cluster.
	// +optional
	Capacity ResourceList `json:"capacity,omitempty" protobuf:"bytes,1,rep,name=capacity,casttype=ResourceList"`
	// Allocatable represents the resources of a cluster that are available for scheduling.
	// Defaults to Capacity.
	// +optional
	Allocatable ResourceList `json:"allocatable,omitempty" protobuf:"bytes,2,rep,name=allocatable,casttype=ResourceList"`
	// +optional
	Allocated ResourceList `json:"allocated,omitempty" protobuf:"bytes,3,rep,name=allocated,casttype=ResourceList"`
}

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
	Type AddressType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=AddressType"`
	// The cluster address.
	Host string `json:"host" protobuf:"bytes,2,opt,name=host"`
	Port int32  `json:"port" protobuf:"varint,3,name=port"`
	Path string `json:"path" protobuf:"bytes,4,opt,name=path"`
}

// FinalizerName is the name identifying a finalizer during cluster lifecycle.
type FinalizerName string

const (
	// ClusterFinalize is an internal finalizer values to Cluster.
	ClusterFinalize FinalizerName = "cluster"

	// MachineFinalize is an internal finalizer values to Machine.
	MachineFinalize FinalizerName = "machine"
)

// NetworkType defines the network type of cluster.
type NetworkType string

const (
	// NetworkPhysics indicates the communication network using the physics network to establish the pod between nodes.
	NetworkPhysics NetworkType = "Physics"
	// NetworkVPC indicates the communication network using the VPC to establish the pod between nodes.
	NetworkVPC NetworkType = "VPC"
	// NetworkFlannel indicates the communication network using the flannel to establish the pod between nodes.
	NetworkFlannel NetworkType = "Flannel"
	// NetworkCalico indicates the communication network using the calico to establish the pod between nodes.
	NetworkCalico NetworkType = "Calico"
	// NetworkIPIP indicates the communication network using the IPIP to establish the pod between nodes.
	NetworkIPIP NetworkType = "IPIP"
)

// ClusterCredential records the credential information needed to access the cluster.
type Credential struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	TenantID    string `json:"tenantID" protobuf:"bytes,2,opt,name=tenantID"`
	ClusterName string `json:"clusterName" protobuf:"bytes,3,opt,name=clusterName"`

	// For TKE in global reuse
	// +optional
	ETCDCACert []byte `json:"etcdCACert,omitempty" protobuf:"bytes,4,opt,name=etcdCACert"`
	// +optional
	ETCDCAKey []byte `json:"etcdCAKey,omitempty" protobuf:"bytes,5,opt,name=etcdCAKey"`
	// +optional
	ETCDAPIClientCert []byte `json:"etcdAPIClientCert,omitempty" protobuf:"bytes,6,opt,name=etcdAPIClientCert"`
	// +optional
	ETCDAPIClientKey []byte `json:"etcdAPIClientKey,omitempty" protobuf:"bytes,7,opt,name=etcdAPIClientKey"`

	// For connect the cluster
	// +optional
	CACert []byte `json:"caCert,omitempty" protobuf:"bytes,8,opt,name=caCert"`
	// +optional
	CAKey []byte `json:"caKey,omitempty" protobuf:"bytes,9,opt,name=caKey"`
	// For kube-apiserver X509 auth
	// +optional
	ClientCert []byte `json:"clientCert,omitempty" protobuf:"bytes,10,opt,name=clientCert"`
	// For kube-apiserver X509 auth
	// +optional
	ClientKey []byte `json:"clientKey,omitempty" protobuf:"bytes,11,opt,name=clientKey"`
	// For kube-apiserver token auth
	// +optional
	Token *string `json:"token,omitempty" protobuf:"bytes,12,opt,name=token"`
	// For kubeadm init or join
	// +optional
	BootstrapToken *string `json:"bootstrapToken,omitempty" protobuf:"bytes,13,opt,name=bootstrapToken"`
	// For kubeadm init or join
	// +optional
	CertificateKey *string `json:"certificateKey,omitempty" protobuf:"bytes,14,opt,name=certificateKey"`
}

// ClusterFeature records the features that are enabled by the cluster.
type ClusterFeature struct {
	// +optional
	IPVS *bool `json:"ipvs,omitempty" protobuf:"varint,1,opt,name=ipvs"`
	// +optional
	PublicLB *bool `json:"publicLB,omitempty" protobuf:"varint,2,opt,name=publicLB"`
	// +optional
	InternalLB *bool `json:"internalLB,omitempty" protobuf:"varint,3,opt,name=internalLB"`
	// +optional
	GPUType *GPUType `json:"gpuType,omitempty" protobuf:"bytes,4,opt,name=gpuType"`
	// +optional
	EnableMasterSchedule bool `json:"enableMasterSchedule,omitempty" protobuf:"bytes,5,opt,name=enableMasterSchedule"`
	// +optional
	HA *HA `json:"ha,omitempty" protobuf:"bytes,6,opt,name=ha"`
}

// GPUType defines the gpu type of cluster.
type GPUType string

const (
	// GPUPhysical indicates the gpu type of cluster is physical.
	GPUPhysical GPUType = "Physical"
	// GPUVirtual indicates the gpu type of cluster is virtual.
	GPUVirtual GPUType = "Virtual"
)

type HA struct {
	TKEHA        *TKEHA        `json:"tke,omitempty" protobuf:"bytes,1,opt,name=tke"`
	ThirdPartyHA *ThirdPartyHA `json:"thirdParty,omitempty" protobuf:"bytes,2,opt,name=thirdParty"`
}

type TKEHA struct {
	VIP string `json:"vip" protobuf:"bytes,1,name=vip"`
}

type ThirdPartyHA struct {
	VIP   string `json:"vip" protobuf:"bytes,1,name=vip"`
	VPort int32  `json:"vport" protobuf:"bytes,2,name=vport"`
}

// ClusterProperty records the attribute information of the cluster.
type ClusterProperty struct {
	// +optional
	MaxClusterServiceNum *int32 `json:"maxClusterServiceNum,omitempty" protobuf:"bytes,1,opt,name=maxClusterServiceNum"`
	// +optional
	MaxNodePodNum *int32 `json:"maxNodePodNum,omitempty" protobuf:"bytes,2,opt,name=maxNodePodNum"`
	// +optional
	OversoldRatio map[string]string `json:"oversoldRatio,omitempty" protobuf:"bytes,3,opt,name=oversoldRatio"`
}

type CIDR struct {
	// +optional
	ClusterCIDR string `json:"clusterCIDR,omitempty" protobuf:"bytes,8,opt,name=clusterCIDR"`
	// ServiceCIDR is used to set a separated CIDR for k8s service, it's exclusive with MaxClusterServiceNum.
	// +optional
	ServiceCIDR string `json:"serviceCIDR,omitempty" protobuf:"bytes,19,opt,name=serviceCIDR"`
	// +optional
	PodCIDR string `json:"podCIDR,omitempty" protobuf:"bytes,19,opt,name=podCIDR"`
}

type NodeHosts struct {
	IP     string `json:"ip"     protobuf:"bytes,8,opt,name=ip"`
	Domain string `json:"domain" protobuf:"bytes,64,opt,name=domain"`
}

// Edge access center ways
type AccessType string

const (
	AccessTypeIps     AccessType = "ip"      // ips
	AccessTypeGateway AccessType = "gateway" // gateway
	AccessTypeIngress AccessType = "ingress" // ingress
)

type EdgeAccessCenter struct {
	// +optional
	AccessType AccessType `json:"accessType" protobuf:"bytes,4,opt,name=accessType"`
	// ServiceCIDR is used to set a separated CIDR for k8s service, it's exclusive with MaxClusterServiceNum.
	// +optional
	AccessAddr []string `json:"accessAddr,omitempty" protobuf:"bytes,8,opt,name=accessAddr"`
	// +optional
	ResolveIPs []string `json:"resolveIPs,omitempty" protobuf:"bytes,8,opt,name=resolveIPs"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EdgeClusterSpec defines the desired state of EdgeCluster
type EdgeClusterSpec struct {
	// +optional
	ClusterId string `json:"clusterId,omitempty" protobuf:"bytes,20,opt,name=clusterId"`

	// +optional
	DisplayName string `json:"displayName" protobuf:"bytes,3,opt,name=displayName"`

	Type string `json:"type" protobuf:"bytes,4,opt,name=type"`

	Version string `json:"version" protobuf:"bytes,5,opt,name=version"`
	// +optional
	NetworkType NetworkType `json:"networkType,omitempty" protobuf:"bytes,6,opt,name=networkType,casttype=NetworkType"`
	// +optional
	NetworkDevice string `json:"networkDevice,omitempty" protobuf:"bytes,7,opt,name=networkDevice"`

	// DNSDomain is the dns domain used by k8s services. Defaults to "cluster.local".
	DNSDomain string `json:"dnsDomain,omitempty" protobuf:"bytes,9,opt,name=dnsDomain"`
	// +optional
	PublicAlternativeNames []string `json:"publicAlternativeNames,omitempty" protobuf:"bytes,10,opt,name=publicAlternativeNames"`
	// +optional
	Features ClusterFeature `json:"features,omitempty" protobuf:"bytes,11,opt,name=features,casttype=ClusterFeature"`
	// +optional
	Properties ClusterProperty `json:"properties,omitempty" protobuf:"bytes,12,opt,name=properties,casttype=ClusterProperty"`
	// +optional
	//Machines []ClusterMachine `json:"machines,omitempty" protobuf:"bytes,13,rep,name=addresses"`

	Cidr CIDR `json:"cidr,omitempty" protobuf:"bytes,24,opt,name=cidr,casttype=CIDR"`
	// +optional
	NodeHosts        []NodeHosts      `json:"nodeHosts,omitempty" protobuf:"bytes,256,name=nodeHosts,casttype=NodeHosts"`
	EdgeAccessCenter EdgeAccessCenter `json:"edgeAccessCenter" protobuf:"bytes,20,opt,name=edgeAccessCenter,casttype=EdgeAccessCenter"`
	// +optional
	Credential Credential `json:"credential,omitempty" protobuf:"bytes,256,opt,name=credential"`

	// +optional
	DockerExtraArgs map[string]string `json:"dockerExtraArgs,omitempty" protobuf:"bytes,14,name=dockerExtraArgs"`
	// +optional
	KubeletExtraArgs map[string]string `json:"kubeletExtraArgs,omitempty" protobuf:"bytes,15,name=kubeletExtraArgs"`
	// +optional
	APIServerExtraArgs map[string]string `json:"apiServerExtraArgs,omitempty" protobuf:"bytes,16,name=apiServerExtraArgs"`
	// +optional
	ControllerManagerExtraArgs map[string]string `json:"controllerManagerExtraArgs,omitempty" protobuf:"bytes,17,name=controllerManagerExtraArgs"`
	// +optional
	SchedulerExtraArgs map[string]string `json:"schedulerExtraArgs,omitempty" protobuf:"bytes,18,name=schedulerExtraArgs"`
}

// EdgeClusterStatus defines the observed state of EdgeCluster
type EdgeClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +optional
	Locked  *bool  `json:"locked,omitempty" protobuf:"varint,1,opt,name=locked"`
	Version string `json:"version" protobuf:"bytes,2,opt,name=version"`
	// +optional
	Phase ClusterPhase `json:"phase,omitempty" protobuf:"bytes,3,opt,name=phase,casttype=ClusterPhase"`
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []ClusterCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,4,rep,name=conditions"`
	// A human readable message indicating details about why the cluster is in this condition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
	// A brief CamelCase message indicating details about why the cluster is in this state.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,6,opt,name=reason"`
	// List of addresses reachable to the cluster.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Addresses []ClusterAddress `json:"addresses,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,7,rep,name=addresses"`
	// +optional
	Resource ClusterResource `json:"resource,omitempty" protobuf:"bytes,9,opt,name=resource,casttype=ClusterResource"`
	//// +optional
	//// +patchMergeKey=type
	//// +patchStrategy=merge
	//Components []ClusterComponent `json:"components,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,10,rep,name=components"`
	//// +optional
	ServiceCIDR string `json:"serviceCIDR,omitempty" protobuf:"bytes,11,opt,name=serviceCIDR"`

	NodeCIDRMaskSize int32 `json:"nodeCIDRMaskSize,omitempty" protobuf:"varint,12,opt,name=nodeCIDRMaskSize"`

	DNSIP string `json:"dnsIP,omitempty" protobuf:"bytes,13,opt,name=dnsIP"`
	// +optional
	RegistryIPs []string `json:"registryIPs,omitempty" protobuf:"bytes,14,opt,name=registryIPs"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=ec,scope=Cluster
// +kubebuilder:printcolumn:name="Namespace",type=string,JSONPath=`.spec.clusterId`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// EdgeCluster is the Schema for the edgeclusters API
type EdgeCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EdgeClusterSpec   `json:"spec,omitempty"`
	Status EdgeClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// EdgeClusterList contains a list of EdgeCluster
type EdgeClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EdgeCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EdgeCluster{}, &EdgeClusterList{})
}
