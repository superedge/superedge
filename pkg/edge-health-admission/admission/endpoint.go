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
	"context"
	"encoding/json"
	"fmt"
	"github.com/superedge/superedge/pkg/edge-health-admission/config"
	"github.com/superedge/superedge/pkg/edge-health-admission/util"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"net/http"
)

func EndPoint(w http.ResponseWriter, r *http.Request) {
	serve(w, r, endPoint)
}

func endPoint(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	var endpointNew corev1.Endpoints

	klog.V(7).Info("admitting endpoints")
	endpointResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "endpoints"}
	reviewResponse := admissionv1.AdmissionResponse{}
	if ar.Request.Resource != endpointResource {
		//klog.V(4).Infof("Request is not nodes, ignore, is %s", ar.Request.Resource.String())
		reviewResponse = admissionv1.AdmissionResponse{Allowed: true}
		return &reviewResponse
	}

	reviewResponseEndPoint, endpointNew, err := decodeRawEndPoint(ar, "new")
	if err != nil {
		return reviewResponseEndPoint
	}

	patches := []*Patch{}

	for i1, EndpointSubset := range endpointNew.Subsets {
		if len(EndpointSubset.NotReadyAddresses) != 0 {
			for i2, EndpointAddress := range EndpointSubset.NotReadyAddresses {
				if node, err := config.Kubeclient.CoreV1().Nodes().Get(context.TODO(), *EndpointAddress.NodeName, metav1.GetOptions{}); err != nil {
					klog.Errorf("can't get pod's node err: %v", err)
				} else {
					_, condition := util.GetNodeCondition(&node.Status, v1.NodeReady)
					if _, ok := node.Annotations["nodeunhealth"]; !ok && condition.Status == v1.ConditionUnknown {

						patches = append(patches, &Patch{
							OP:   "remove",
							Path: fmt.Sprintf("/subsets/%d/notReadyAddresses/%d", i1, i2),
						})

						TargetRef := map[string]interface{}{}
						TargetRef["kind"] = EndpointAddress.TargetRef.Kind
						TargetRef["namespace"] = EndpointAddress.TargetRef.Namespace
						TargetRef["name"] = EndpointAddress.TargetRef.Name
						TargetRef["uid"] = EndpointAddress.TargetRef.UID
						TargetRef["apiVersion"] = EndpointAddress.TargetRef.APIVersion
						TargetRef["resourceVersion"] = EndpointAddress.TargetRef.ResourceVersion
						TargetRef["fieldPath"] = EndpointAddress.TargetRef.FieldPath

						patches = append(patches, &Patch{
							OP:   "add",
							Path: fmt.Sprintf("/subsets/%d/addresses/%d", i1, i2),
							Value: map[string]interface{}{
								"ip":        EndpointAddress.IP,
								"hostname":  EndpointAddress.Hostname,
								"nodeName":  EndpointAddress.NodeName,
								"targetRef": TargetRef,
							},
						})

						if len(patches) != 0 {
							bytes, _ := json.Marshal(patches)
							reviewResponse.Patch = bytes
							pt := admissionv1.PatchTypeJSONPatch
							reviewResponse.PatchType = &pt
						}
					}
				}
			}
		}
	}
	reviewResponse.Allowed = true
	return &reviewResponse
}

func decodeRawEndPoint(ar admissionv1.AdmissionReview, version string) (*admissionv1.AdmissionResponse, corev1.Endpoints, error) {
	var raw []byte
	if version == "new" {
		raw = ar.Request.Object.Raw
	} else if version == "old" {
		raw = ar.Request.OldObject.Raw
	}

	endpoint := corev1.Endpoints{}
	deserializer := Codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &endpoint); err != nil {
		klog.Error(err)
		return toAdmissionResponse(err), endpoint, err
	}
	return nil, endpoint, nil
}
