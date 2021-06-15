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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +genclient:skipVerbs=update
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type NodeTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeTaskSpec   `json:"spec,omitempty"`
	Status NodeTaskStatus `json:"status,omitempty"`
}
type NodeTaskSpec struct {
	NodeNamePrefix    string            `json:"nodeNamePrefix,omitempty"`
	TargetMachines    []string          `json:"targetMachines,omitempty"`
	NodeNamesOverride map[string]string `json:"nodeNamesOverride,omitempty"`
	ProxyNode         string            `json:"proxyNode,omitempty"`
	SSHCredential     string            `json:"sshCredential,omitempty"`
	SSHPort           int               `json:"sshPort,omitempty"`
}

type NodeTaskStatus struct {
	NodeStatus     map[string]string `json:"nodeStatus,omitempty"`
	NodeTaskStatus string            `json:"nodetaskStatus,omitempty"`
}

const (
	NodeTaskStatusCreating = "creating"
	NodeTaskStatusReady    = "ready"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NodeTaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []NodeTask `json:"items"`
}
