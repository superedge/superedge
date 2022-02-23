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

package e2e

import (
	"context"
	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"testing"
)

func TestApplicationgrid(t *testing.T) {
	ctx := context.Background()
	clientSet, crdclient, _ := initClient(t)
	var selector map[string]string
	var err error
	selector = make(map[string]string)
	selector["appGrid"] = "echo"

	sharedInformerFactory := informers.NewSharedInformerFactory(clientSet, 0)
	nodeInformer := sharedInformerFactory.Core().V1().Nodes()

	nodelist, err := ReadyNodes(ctx, clientSet, nodeInformer, "zone1=nodeunit1")
	t.Log(len(nodelist))
	if err != nil {
		t.Fatal("check node list fail", err)
	}

	deploymentGridObj := &crdv1.DeploymentGrid{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deploymentgrid-e2e",
			Namespace: "default",
		},
		Spec: crdv1.DeploymentGridSpec{
			GridUniqKey: "zone1",
			Template: appsv1.DeploymentSpec{
				Replicas: func() *int32 { i := int32(1); return &i }(),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "deploymentgrid-e2e"},
				},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"app": "deploymentgrid-e2e"},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{{
							Name:            "pause",
							ImagePullPolicy: "Always",
							//Image:           "kubernetes/pause",
							Image: "gcr.io/kubernetes-e2e-test-images/echoserver:2.2",
							Ports: []v1.ContainerPort{{ContainerPort: 8080, Protocol: "TCP"}},
						}},
					},
				},
			},
		},
	}

	_, err = crdclient.SuperedgeV1().DeploymentGrids("default").Create(ctx, deploymentGridObj, metav1.CreateOptions{})
	if err != nil {
		t.Fatal("create DeploymentGrids fail", err)
	}
	defer crdclient.SuperedgeV1().DeploymentGrids("default").Delete(ctx, "deploymentgrid-e2e", metav1.DeleteOptions{})

	// check is it have deployment deploymentgrid-e2e-nodeunitname
	result, err := waitForDeployment(t, ctx, "default", "deploymentgrid-e2e-nodeunit1", clientSet)
	if !result {
		t.Fatal("deployment deploymentgrid-e2e-nodeunit1 not found", err)
	}

	// new ServiceGrid "servicegrid-e2e" and apply to cluster
	svcGridObj := &crdv1.ServiceGrid{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "servicegrid-e2e",
			Namespace: "default",
		},
		Spec: crdv1.ServiceGridSpec{
			GridUniqKey: "zone1",
			Template: v1.ServiceSpec{
				Selector: selector,
				Ports:    []v1.ServicePort{{Port: 8080, Protocol: "TCP"}},
			},
		},
	}

	// new ServiceGrids "servicegrid-e2e" and apply to cluster
	_, err = crdclient.SuperedgeV1().ServiceGrids("default").Create(ctx, svcGridObj, metav1.CreateOptions{})
	if err != nil {
		t.Fatal("create svcGridObj fail", err)
	}

	defer crdclient.SuperedgeV1().ServiceGrids("default").Delete(ctx, "servicegrid-e2e", metav1.DeleteOptions{})
	// make sure the service create success
	result, err = waitForSVC(t, ctx, "default", "servicegrid-e2e-svc", clientSet)
	if !result {
		t.Fatal("deployment deploymentgrid-e2e-nodeunit1 not found", err)
	}
}
