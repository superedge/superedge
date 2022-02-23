/*
Copyright 2022 The SuperEdge Authors.

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

package service

import (
	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newServiceGrid(name string, gridUniqKey string, selector map[string]string, port int) *crdv1.ServiceGrid {
	serviceGrid := crdv1.ServiceGrid{
		TypeMeta: metav1.TypeMeta{APIVersion: "superedge.io/v1", Kind: "ServiceGrid"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   metav1.NamespaceDefault,
			Annotations: make(map[string]string),
		},
		Spec: crdv1.ServiceGridSpec{},
	}
	return &serviceGrid
}
