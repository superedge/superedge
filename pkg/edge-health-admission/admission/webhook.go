package admission

import (
	"encoding/json"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Patch struct {
	OP    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

func Allow() *v1.AdmissionResponse {
	reviewResponse := v1.AdmissionResponse{}
	reviewResponse.Allowed = true
	reviewResponse.Result = &metav1.Status{Message: "this webhook allows all requests"}
	return &reviewResponse
}

func Deny() *v1.AdmissionResponse {
	reviewResponse := &v1.AdmissionResponse{}
	reviewResponse.Allowed = false
	reviewResponse.Result = &metav1.Status{Message: "this webhook denies all requests"}
	return reviewResponse
}

func DenyWithMessage(message string) *v1.AdmissionResponse {
	reviewResponse := &v1.AdmissionResponse{}
	reviewResponse.Allowed = false
	reviewResponse.Result = &metav1.Status{Message: message}
	return reviewResponse
}

func AllowWithJsonPatch(patches []Patch) *v1.AdmissionResponse {
	reviewResponse := &v1.AdmissionResponse{
		Allowed: true,
	}
	if len(patches) != 0 {
		bytes, _ := json.Marshal(patches)
		reviewResponse.Patch = bytes
		pt := v1.PatchTypeJSONPatch
		reviewResponse.PatchType = &pt
	}
	return reviewResponse
}
