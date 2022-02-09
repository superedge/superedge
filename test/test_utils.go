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

package test

import (
	"fmt"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

// BuildTestNode creates a node with specified capacity.
func BuildTestNode(name string, millicpu int64, mem int64, pods int64, apply func(*v1.Node)) *v1.Node {
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:     name,
			SelfLink: fmt.Sprintf("/api/v1/nodes/%s", name),
			Labels:   map[string]string{},
		},
		Status: v1.NodeStatus{
			Capacity: v1.ResourceList{
				v1.ResourcePods:   *resource.NewQuantity(pods, resource.DecimalSI),
				v1.ResourceCPU:    *resource.NewMilliQuantity(millicpu, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(mem, resource.DecimalSI),
			},
			Allocatable: v1.ResourceList{
				v1.ResourcePods:   *resource.NewQuantity(pods, resource.DecimalSI),
				v1.ResourceCPU:    *resource.NewMilliQuantity(millicpu, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(mem, resource.DecimalSI),
			},
			Phase: v1.NodeRunning,
			Conditions: []v1.NodeCondition{
				{Type: v1.NodeReady, Status: v1.ConditionTrue},
			},
		},
	}
	if apply != nil {
		apply(node)
	}
	return node
}

func BuildTestPod(name string, cpu int64, memory int64, nodeName string, apply func(*v1.Pod)) *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      name,
			SelfLink:  fmt.Sprintf("/api/v1/namespaces/default/pods/%s", name),
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{},
						Limits:   v1.ResourceList{},
					},
				},
			},
			NodeName: nodeName,
		},
	}
	if cpu >= 0 {
		pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU] = *resource.NewMilliQuantity(cpu, resource.DecimalSI)
	}
	if memory >= 0 {
		pod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory] = *resource.NewQuantity(memory, resource.DecimalSI)
	}
	if apply != nil {
		apply(pod)
	}
	return pod
}

func BuildTestConfigmap(name string, namespace string, data map[string]string) *v1.ConfigMap {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Data: data,
	}
	return cm
}

func BuildTestDeployment(name string, namespace string, replicas int, selector map[string]string) *apps.Deployment {
	return &apps.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			UID:         uuid.NewUUID(),
			Name:        name,
			Namespace:   namespace,
			Annotations: make(map[string]string),
		},
		Spec: apps.DeploymentSpec{
			Replicas: func() *int32 { i := int32(replicas); return &i }(),
			Selector: &metav1.LabelSelector{MatchLabels: selector},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: selector,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image: "foo/bar",
						},
					},
				},
			},
		},
	}
}
