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
	"io/ioutil"

	"github.com/superedge/superedge/pkg/edge-health-admission/config"
	"github.com/superedge/superedge/pkg/edge-health-admission/util"
	edgeutil "github.com/superedge/superedge/pkg/util"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"net/http"
)

type admitFunc func(admissionv1.AdmissionReview) *admissionv1.AdmissionResponse

type Patch struct {
	OP    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

func NodeTaint(w http.ResponseWriter, r *http.Request) {
	serve(w, r, nodeTaint)
}

func nodeTaint(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	var nodeNew, nodeOld corev1.Node

	UnreachNoExecuteTaint := &corev1.Taint{
		Key:    corev1.TaintNodeUnreachable,
		Effect: corev1.TaintEffectNoExecute,
	}

	klog.V(7).Info("admitting nodes")
	nodeResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"}
	reviewResponse := admissionv1.AdmissionResponse{}
	if ar.Request.Resource != nodeResource {
		//klog.V(4).Infof("Request is not nodes, ignore, is %s", ar.Request.Resource.String())
		reviewResponse = admissionv1.AdmissionResponse{Allowed: true}
		return &reviewResponse
	}

	//klog.V(4).Infof("Request is nodes, is %s", ar.Request)

	reviewResponseNode, nodeNew, err := decodeRawNode(ar, "new")
	if err != nil {
		return reviewResponseNode
	}
	klog.V(4).Infof("nodeNew is %s", edgeutil.ToJson(nodeNew))

	reviewResponseNode, nodeOld, err = decodeRawNode(ar, "old")
	if err != nil {
		return reviewResponseNode
	}
	klog.V(4).Infof("nodeOld is %s", edgeutil.ToJson(nodeOld))

	_, condition := util.GetNodeCondition(&nodeNew.Status, corev1.NodeReady)
	patches := []*Patch{}
	if condition.Status == corev1.ConditionUnknown {
		if _, ok := nodeNew.Annotations["nodeunhealth"]; !ok || config.NodeAlwaysReachable {
			taintsToAdd, _ := util.TaintSetDiff(nodeNew.Spec.Taints, nodeOld.Spec.Taints)
			if _, flag := util.TaintExistsPosition(taintsToAdd, UnreachNoExecuteTaint); flag {
				index, _ := util.TaintExistsPosition(nodeNew.Spec.Taints, UnreachNoExecuteTaint)
				patches = append(patches, &Patch{
					OP:   "remove",
					Path: fmt.Sprintf("/spec/taints/%d", index),
				})
				klog.V(7).Infof("UnreachNoExecuteTaint: remove %d taints : %s", index, nodeNew.Spec.Taints[index])
			} else if _, resflag := util.TaintExistsPosition(nodeNew.Spec.Taints, UnreachNoExecuteTaint); resflag {
				index, _ := util.TaintExistsPosition(nodeNew.Spec.Taints, UnreachNoExecuteTaint)
				patches = append(patches, &Patch{
					OP:   "remove",
					Path: fmt.Sprintf("/spec/taints/%d", index),
				})
				klog.V(7).Infof("UnreachNoExecuteTaint still existed: remove %d taints : %s", index, nodeNew.Spec.Taints[index])
			}

			if len(patches) != 0 {
				bytes, _ := json.Marshal(patches)
				reviewResponse.Patch = bytes
				pt := admissionv1.PatchTypeJSONPatch
				reviewResponse.PatchType = &pt
			}
		}
	}

	reviewResponse.Allowed = true

	return &reviewResponse
}

func decodeRawNode(ar admissionv1.AdmissionReview, version string) (*admissionv1.AdmissionResponse, corev1.Node, error) {
	var raw []byte
	if version == "new" {
		raw = ar.Request.Object.Raw
	} else if version == "old" {
		raw = ar.Request.OldObject.Raw
	}

	node := corev1.Node{}
	deserializer := Codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &node); err != nil {
		klog.Error(err)
		return toAdmissionResponse(err), node, err
	}
	return nil, node, nil
}

func serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	klog.V(7).Info(fmt.Sprintf("handling request: %s", body))

	// The AdmissionReview that was sent to the webhook
	admissionReview := admissionv1.AdmissionReview{}

	deserializer := Codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &admissionReview); err != nil {
		klog.Error(err)
		admissionReview.Response = toAdmissionResponse(err)
	} else {
		// pass to admitFunc
		admissionReview.Response = admit(admissionReview)
	}

	admissionReview.Response.UID = admissionReview.Request.UID

	klog.V(7).Info(fmt.Sprintf("sending response: %v", admissionReview))

	respBytes, err := json.Marshal(admissionReview)
	if err != nil {
		klog.Error(err)
	}
	if _, err := w.Write(respBytes); err != nil {
		klog.Error(err)
	}
}

func toAdmissionResponse(err error) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}
