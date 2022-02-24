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
	"fmt"
	"strings"

	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ControllerKind = crdv1.SchemeGroupVersion.WithKind(common.ServiceGridKind)

func GetServiceName(sg *crdv1.ServiceGrid) string {
	return strings.Join([]string{sg.Name, "svc"}, "-")
}

func CreateService(sg *crdv1.ServiceGrid) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetServiceName(sg),
			Namespace: sg.Namespace,
			// Append existed ServiceGrid labels to service to be created
			Labels: func() map[string]string {
				if sg.Labels != nil {
					newLabels := sg.Labels
					newLabels[common.GridSelectorName] = sg.Name
					newLabels[common.GridSelectorUniqKeyName] = sg.Spec.GridUniqKey
					return newLabels
				} else {
					return map[string]string{
						common.GridSelectorName:        sg.Name,
						common.GridSelectorUniqKeyName: sg.Spec.GridUniqKey,
					}
				}
			}(),
			Annotations: func() map[string]string {
				keys := make([]string, 1)
				keys[0] = sg.Spec.GridUniqKey
				keyData, _ := json.Marshal(keys)
				if sg.Annotations != nil {
					newAnnotation := sg.Annotations
					newAnnotation[common.TopologyAnnotationsKey] = string(keyData)
					return newAnnotation
				} else {
					return map[string]string{
						common.TopologyAnnotationsKey: string(keyData),
					}
				}
			}(),
		},
		Spec: sg.Spec.Template,
	}

	return svc
}

// constructServicePortIdentify construct the name identify for service port
func constructServicePortIdentify(port *corev1.ServicePort) string {
	return fmt.Sprintf("%s-%s-%d-%s-%d", port.Name, port.Protocol, port.Port, port.TargetPort.String(), port.NodePort)
}

func KeepConsistence(sg *crdv1.ServiceGrid, svc *corev1.Service) *corev1.Service {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	copyObj := svc.DeepCopy()
	// Append existed ServiceGrid labels to service to be checked
	if sg.Labels != nil {
		copyObj.Labels = sg.Labels
		copyObj.Labels[common.GridSelectorName] = sg.Name
		copyObj.Labels[common.GridSelectorUniqKeyName] = sg.Spec.GridUniqKey
	} else {
		copyObj.Labels = map[string]string{
			common.GridSelectorName:        sg.Name,
			common.GridSelectorUniqKeyName: sg.Spec.GridUniqKey,
		}
	}

	if copyObj.Annotations == nil {
		copyObj.Annotations = make(map[string]string)
	}

	// Perform weak check for service annotations since it may update during execution
	keys := make([]string, 1)
	keys[0] = sg.Spec.GridUniqKey
	keyData, _ := json.Marshal(keys)
	copyObj.Annotations[common.TopologyAnnotationsKey] = string(keyData)

	var oldServiceNameNodePort = make(map[string]int32)
	var newServiceNameNodePort = make(map[string]int32)
	if sg.Spec.Template.Type == corev1.ServiceTypeNodePort && copyObj.Spec.Type == corev1.ServiceTypeNodePort {
		for _, port := range copyObj.Spec.Ports {
			oldServiceNameNodePort[constructServicePortIdentify(&port)] = port.NodePort
		}
		for _, port := range sg.Spec.Template.Ports {
			newServiceNameNodePort[constructServicePortIdentify(&port)] = port.NodePort
		}
	}

	// TODO: serviceGrid keepConsistence for more fields
	copyObj.Spec.Selector = sg.Spec.Template.Selector
	copyObj.Spec.Ports = sg.Spec.Template.Ports
	if sg.Spec.Template.Type == corev1.ServiceTypeNodePort && copyObj.Spec.Type == corev1.ServiceTypeNodePort {
		for k, port := range copyObj.Spec.Ports {
			if _, ok := oldServiceNameNodePort[constructServicePortIdentify(&port)]; ok {
				if newServiceNameNodePort[port.Name] == 0 && oldServiceNameNodePort[port.Name] != 0 {
					copyObj.Spec.Ports[k].NodePort = oldServiceNameNodePort[constructServicePortIdentify(&port)]
				}
			}
		}
	}
	return copyObj
}

func CreateServiceGrid(sg *crdv1.ServiceGrid, nameSpace string) *crdv1.ServiceGrid {
	sgcopy := sg.DeepCopy()
	sgcopy.ResourceVersion = ""
	TargetNameSpace := sgcopy.Namespace
	sgcopy.Namespace = nameSpace
	sgcopy.Labels[common.FedTargetNameSpace] = TargetNameSpace
	sgcopy.Labels[common.FedrationDisKey] = "yes"
	sgcopy.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(sg, ControllerKind)}
	return sgcopy
}

func UpdateServiceGrid(sg, fed *crdv1.ServiceGrid) *crdv1.ServiceGrid {
	sgcopy := fed.DeepCopy()
	sgcopy.Spec = sg.Spec
	return sgcopy
}
