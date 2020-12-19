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

package util

import (
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

func ToCustomResourceDefinition(gvk schema.GroupVersionKind) apiext.CustomResourceDefinition {
	plural := strings.ToLower(gvk.Kind + "s")
	name := strings.ToLower(plural + "." + gvk.Group)

	crd := apiext.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: apiext.CustomResourceDefinitionSpec{
			Group:   gvk.Group,
			Version: gvk.Version,
			Versions: []apiext.CustomResourceDefinitionVersion{
				{
					Name:    gvk.Version,
					Storage: true,
					Served:  true,
				},
			},
			Names: apiext.CustomResourceDefinitionNames{
				Plural: plural,
				Kind:   gvk.Kind,
			},
			Subresources: &apiext.CustomResourceSubresources{
				Status: &apiext.CustomResourceSubresourceStatus{},
			},
		},
	}
	crd.Spec.Scope = apiext.NamespaceScoped

	return crd
}
