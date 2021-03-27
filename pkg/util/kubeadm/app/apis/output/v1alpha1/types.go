/*
Copyright 2019 The Kubernetes Authors.

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

package v1alpha1

import (
	kubeadmapiv1beta2 "github.com/superedge/superedge/pkg/util/kubeadm/app/apis/kubeadm/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BootstrapToken represents information for the bootstrap token output produced by kubeadm
// This is a copy of BoostrapToken struct from ../kubeadm/types.go with 2 additions:
// metav1.TypeMeta and metav1.ObjectMeta
type BootstrapToken struct {
	metav1.TypeMeta `json:",inline"`

	kubeadmapiv1beta2.BootstrapToken
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Images represents information for the output produced by 'kubeadm config images list'
type Images struct {
	metav1.TypeMeta `json:",inline"`

	Images []string `json:"images"`
}

// ComponentUpgradePlan represents information about upgrade plan for one component
type ComponentUpgradePlan struct {
	Name           string `json:"name"`
	CurrentVersion string `json:"currentVersion"`
	NewVersion     string `json:"newVersion"`
}

// ComponentConfigVersionState describes the current and desired version of a component config
type ComponentConfigVersionState struct {
	// Group points to the Kubernetes API group that covers the config
	Group string `json:"group"`

	// CurrentVersion is the currently active component config version
	// NOTE: This can be empty in case the config was not found on the cluster or it was unsupported
	// kubeadm generated version
	CurrentVersion string `json:"currentVersion"`

	// PreferredVersion is the component config version that is currently preferred by kubeadm for use.
	// NOTE: As of today, this is the only version supported by kubeadm.
	PreferredVersion string `json:"preferredVersion"`

	// ManualUpgradeRequired indicates if users need to manually upgrade their component config versions. This happens if
	// the CurrentVersion of the config is user supplied (or modified) and no longer supported. Users should upgrade
	// their component configs to PreferredVersion or any other supported component config version.
	ManualUpgradeRequired bool `json:"manualUpgradeRequired"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// UpgradePlan represents information about upgrade plan for the output
// produced by 'kubeadm upgrade plan'
type UpgradePlan struct {
	metav1.TypeMeta

	Components []ComponentUpgradePlan `json:"components"`

	ConfigVersions []ComponentConfigVersionState `json:"configVersions"`
}
