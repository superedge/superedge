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

type patch struct {
	OP    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

func (eha *EdgeHealthAdmission) mutateNodeTaint(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	klog.V(4).Info("mutating node taint")
	nodeResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"}
	if ar.Request.Resource != nodeResource {
		klog.Errorf("expect resource to be %s", nodeResource)
		return nil
	}

	var node corev1.Node
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(ar.Request.Object.Raw, nil, &node); err != nil {
		klog.Error(err)
		return toAdmissionResponse(err)
	}

	reviewResponse := admissionv1.AdmissionResponse{}
	reviewResponse.Allowed = true

	if index, condition := util.GetNodeCondition(&node.Status, v1.NodeReady); index != -1 && condition.Status == v1.ConditionUnknown {
		if node.Annotations != nil {
			var patches []*patch
			if healthy, existed := node.Annotations[common.NodeHealthAnnotation]; existed && healthy == common.NodeHealthAnnotationPros {
				if index, existed := util.TaintExistsPosition(node.Spec.Taints, common.UnreachableNoExecuteTaint); existed {
					patches = append(patches, &patch{
						OP:   "remove",
						Path: fmt.Sprintf("/spec/taints/%d", index),
					})
					klog.V(4).Infof("UnreachableNoExecuteTaint: remove %d taints %s", index, node.Spec.Taints[index])
				}
			}
			if len(patches) > 0 {
				patchBytes, _ := json.Marshal(patches)
				reviewResponse.Patch = patchBytes
				pt := admissionv1.PatchTypeJSONPatch
				reviewResponse.PatchType = &pt
			}
		}
	}

	return &reviewResponse
}
