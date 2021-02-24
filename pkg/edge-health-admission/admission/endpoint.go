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

package admission

import (
	"encoding/json"
	"fmt"
	"github.com/superedge/superedge/pkg/edge-health-admission/util"
	"github.com/superedge/superedge/pkg/edge-health/common"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

func (eha *EdgeHealthAdmission) mutateEndpoint(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	klog.V(4).Info("mutating endpoint")
	endpointResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "endpoints"}
	if ar.Request.Resource != endpointResource {
		klog.Errorf("expect resource to be %s", endpointResource)
		return nil
	}

	var endpoint corev1.Endpoints
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(ar.Request.Object.Raw, nil, &endpoint); err != nil {
		klog.Error(err)
		return toAdmissionResponse(err)
	}

	reviewResponse := admissionv1.AdmissionResponse{}
	reviewResponse.Allowed = true

	for epSubsetIndex, epSubset := range endpoint.Subsets {
		for notReadyAddrIndex, EndpointAddress := range epSubset.NotReadyAddresses {
			if node, err := eha.nodeLister.Get(*EndpointAddress.NodeName); err == nil {
				if index, condition := util.GetNodeCondition(&node.Status, v1.NodeReady); index != -1 && condition.Status == v1.ConditionUnknown {
					if node.Annotations != nil {
						var patches []*patch
						if healthy, existed := node.Annotations[common.NodeHealthAnnotation]; existed && healthy == common.NodeHealthAnnotationPros {
							// TODO: handle readiness probes failure
							// Remove address on node from endpoint notReadyAddresses
							patches = append(patches, &patch{
								OP:   "remove",
								Path: fmt.Sprintf("/subsets/%d/notReadyAddresses/%d", epSubsetIndex, notReadyAddrIndex),
							})

							// Add address on node to endpoint readyAddresses
							TargetRef := map[string]interface{}{}
							TargetRef["kind"] = EndpointAddress.TargetRef.Kind
							TargetRef["namespace"] = EndpointAddress.TargetRef.Namespace
							TargetRef["name"] = EndpointAddress.TargetRef.Name
							TargetRef["uid"] = EndpointAddress.TargetRef.UID
							TargetRef["apiVersion"] = EndpointAddress.TargetRef.APIVersion
							TargetRef["resourceVersion"] = EndpointAddress.TargetRef.ResourceVersion
							TargetRef["fieldPath"] = EndpointAddress.TargetRef.FieldPath

							patches = append(patches, &patch{
								OP:   "add",
								Path: fmt.Sprintf("/subsets/%d/addresses/0", epSubsetIndex),
								Value: map[string]interface{}{
									"ip":        EndpointAddress.IP,
									"hostname":  EndpointAddress.Hostname,
									"nodeName":  EndpointAddress.NodeName,
									"targetRef": TargetRef,
								},
							})

							if len(patches) != 0 {
								patchBytes, _ := json.Marshal(patches)
								reviewResponse.Patch = patchBytes
								pt := admissionv1.PatchTypeJSONPatch
								reviewResponse.PatchType = &pt
							}
						}
					}
				}
			} else {
				klog.Errorf("Get pod's node err %v", err)
			}
		}

	}

	return &reviewResponse
}
