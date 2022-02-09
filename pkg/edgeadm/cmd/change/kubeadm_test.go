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

package change

import (
	"context"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestDeployTunnelCloud(t *testing.T) {
	c := newChange()
	c.clientSet = fake.NewSimpleClientset()

	var selector map[string]string
	selector = make(map[string]string)
	selector["appGrid"] = "echo"

	// mock create deployment first
	deploymentobj := test.BuildTestDeployment(constant.ServiceTunnelCloud, constant.NamespaceKubeSystem, 1, selector)
	c.clientSet.AppsV1().Deployments(constant.NamespaceKubeSystem).Create(
		context.TODO(), deploymentobj, metav1.CreateOptions{})
	result, err := c.deployTunnelCloud()

	if err != nil {
		t.Fatal(err)
	}
	_, err = c.clientSet.AppsV1().Deployments(constant.NamespaceKubeSystem).Get(
		context.TODO(), constant.ServiceTunnelCloud, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result)
	if err != nil {
		t.Fatal(err)
	}
}
