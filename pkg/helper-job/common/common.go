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

package common

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"superedge/pkg/helper-job/constant"
)

func GetMasterHosts(clientSet *kubernetes.Clientset) (map[string]string, error) {
	nodes, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	masterIPs := make(map[string]string)
	for _, node := range nodes.Items {
		if _, ok := node.Labels[constant.KUBERNETES_DEFAULT_ROLE_LABEL]; ok {
			for _, address := range node.Status.Addresses {
				if address.Type == v1.NodeInternalIP {
					masterIPs[node.Name] = address.Address
				}
			}
		}
	}

	return masterIPs, nil
}
