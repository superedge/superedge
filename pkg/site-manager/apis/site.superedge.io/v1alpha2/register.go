/*
Copyright 2021 The SuperEdge Authors.

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

// +k8s:deepcopy-gen=package
// +groupName=superedge.io
package v1alpha2

import (
	"github.com/superedge/superedge/pkg/site-manager/constant"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: "site.superedge.io", Version: "v1alpha2"}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder      = runtime.NewSchemeBuilder(addKnownTypes, registerDefaults)
	localSchemeBuilder = &SchemeBuilder
	AddToScheme        = SchemeBuilder.AddToScheme
)

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&NodeGroup{},
		&NodeGroupList{},
		&NodeUnit{},
		&NodeUnitList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

func registerDefaults(scheme *runtime.Scheme) error {
	scheme.AddTypeDefaultingFunc(&NodeUnit{}, func(obj interface{}) { SetObjectDefaults_NodeUnit(obj.(*NodeUnit)) })
	scheme.AddTypeDefaultingFunc(&NodeUnitList{}, func(obj interface{}) { SetObjectDefaults_NodeUnitList(obj.(*NodeUnitList)) })
	scheme.AddTypeDefaultingFunc(&NodeGroup{}, func(obj interface{}) { SetObjectDefaults_NodeGroup(obj.(*NodeGroup)) })
	scheme.AddTypeDefaultingFunc(&NodeGroupList{}, func(obj interface{}) { SetObjectDefaults_NodeGroupList(obj.(*NodeGroupList)) })
	return nil
}

func SetObjectDefaults_NodeUnit(in *NodeUnit) {
	SetDefaults_NodeUnitSpec(in)
	SetDefaults_NodeUnitTypeMeta(in)
}

func SetDefaults_NodeUnitTypeMeta(in *NodeUnit) {
	in.APIVersion = SchemeGroupVersion.String()
	in.Kind = "NodeUnit"
}

func SetDefaults_NodeUnitSpec(in *NodeUnit) {
	if in.Spec.SetNode.Labels == nil {
		in.Spec.SetNode.Labels = map[string]string{
			in.Name: constant.NodeUnitSuperedge,
		}
	} else if _, ok := in.Spec.SetNode.Labels[in.Name]; !ok {
		in.Spec.SetNode.Labels[in.Name] = constant.NodeUnitSuperedge
	}
	if in.Spec.AutonomyLevel == "" {
		in.Spec.AutonomyLevel = "L3"
	}

}

func SetObjectDefaults_NodeUnitList(in *NodeUnitList) {
	SetDefaults_NodeUnitListTypeMeta(in)
}

func SetDefaults_NodeUnitListTypeMeta(in *NodeUnitList) {
	in.APIVersion = SchemeGroupVersion.String()
	in.Kind = "NodeUnitList"
}

func SetObjectDefaults_NodeGroup(in *NodeGroup) {
	SetDefaults_NodeGroupTypeMeta(in)
}

func SetDefaults_NodeGroupTypeMeta(in *NodeGroup) {
	in.APIVersion = SchemeGroupVersion.String()
	in.Kind = "NodeGroup"
}

func SetObjectDefaults_NodeGroupList(in *NodeGroupList) {
	SetDefaults_NodeGroupListTypeMeta(in)
}

func SetDefaults_NodeGroupListTypeMeta(in *NodeGroupList) {
	in.APIVersion = SchemeGroupVersion.String()
	in.Kind = "NodeGroupList"
}
