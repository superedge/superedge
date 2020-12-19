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
	"encoding/json"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crdv1 "superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	"superedge/pkg/application-grid-controller/controller/common"
)

//var (
//	ServiceTopologyEnabled = utilfeature.DefaultMutableFeatureGate.Enabled(common.ServiceTopology) &&
//		utilfeature.DefaultMutableFeatureGate.Enabled(common.EndpointSlice)
//)

func GetServiceName(g *crdv1.ServiceGrid) string {
	return strings.Join([]string{g.Name, "svc"}, "-")
}

func CreateService(g *crdv1.ServiceGrid) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetServiceName(g),
			Namespace: g.Namespace,
			Labels: map[string]string{
				common.GridSelectorName: g.Name,
			},
			Annotations: make(map[string]string),
		},
		Spec: g.Spec.Template,
	}

	//if ServiceTopologyEnabled {
	//	svc.Spec.TopologyKeys = append(svc.Spec.TopologyKeys, g.Spec.GridUniqKey)
	//} else {
	keys := make([]string, 1)
	keys[0] = g.Spec.GridUniqKey
	keyData, _ := json.Marshal(keys)
	svc.Annotations[common.TopologyAnnotationsKey] = string(keyData)
	//}

	return svc
}

func KeepConsistence(g *crdv1.ServiceGrid, svc *corev1.Service) *corev1.Service {
	copyObj := svc.DeepCopy()
	if copyObj.Labels == nil {
		copyObj.Labels = make(map[string]string)
	}
	copyObj.Labels[common.GridSelectorName] = g.Name

	if copyObj.Annotations == nil {
		copyObj.Annotations = make(map[string]string)
	}

	//if ServiceTopologyEnabled {
	//	copyObj.Spec.TopologyKeys = []string{g.Spec.GridUniqKey}
	//} else {
	keys := make([]string, 1)
	keys[0] = g.Spec.GridUniqKey
	keyData, _ := json.Marshal(keys)
	copyObj.Annotations[common.TopologyAnnotationsKey] = string(keyData)
	//}
	var oldServiceNameNodePort = make(map[string]int32)
	var newServiceNameNodePort = make(map[string]int32)
	if g.Spec.Template.Type == corev1.ServiceTypeNodePort && copyObj.Spec.Type == corev1.ServiceTypeNodePort {
		for _, port := range copyObj.Spec.Ports {
			oldServiceNameNodePort[port.Name] = port.NodePort
		}
		for _, port := range g.Spec.Template.Ports {
			newServiceNameNodePort[port.Name] = port.NodePort
		}
	}

	copyObj.Spec.Ports = g.Spec.Template.Ports
	if g.Spec.Template.Type == corev1.ServiceTypeNodePort && copyObj.Spec.Type == corev1.ServiceTypeNodePort {
		for k, port := range copyObj.Spec.Ports {
			if _, ok := oldServiceNameNodePort[port.Name]; ok{
				if newServiceNameNodePort[port.Name] == 0 && oldServiceNameNodePort[port.Name] != 0 {
					copyObj.Spec.Ports[k].NodePort = oldServiceNameNodePort[port.Name]
				}
			}
		}
	}
	return copyObj
}
