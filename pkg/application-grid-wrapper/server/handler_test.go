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

package server

import (
	"fmt"
	discoveryv1 "k8s.io/api/discovery/v1"
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"testing"
)

func TestGetIngressPath(t *testing.T) {
	getIngressPath("/superedge-ingress/zone/api/v1?a=1")
}

func TestConvert(t *testing.T) {
	targetVersion := schema.GroupVersion{
		Group:   discoveryv1beta1.GroupName,
		Version: "__internal",
	}
	epsV1 := &discoveryv1.EndpointSlice{}
	epsV1.Name = "test"
	epsV1.Namespace = "default"
	runtimeScheme := runtime.NewScheme()
	discoveryv1.AddToScheme(runtimeScheme)
	discoveryv1beta1.AddToScheme(runtimeScheme)
	out, err := runtimeScheme.ConvertToVersion(epsV1, targetVersion)
	if err != nil {
		klog.Errorf("failed to convert, error: %v", err)
	}
	fmt.Println(out)
}
