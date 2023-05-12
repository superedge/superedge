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
	"fmt"
	"github.com/superedge/superedge/pkg/edge-health-admission/config"
	"github.com/superedge/superedge/pkg/edge-health-admission/util"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"net/http"
)

func EndPoint(w http.ResponseWriter, r *http.Request) {
	serve(w, r, endpoint)
}

func endpoint(ar v1.AdmissionReview) *v1.AdmissionResponse {
	logger := klog.NewKlogr().
		WithValues("resource", ar.Request.Resource).
		WithValues("name", ar.Request.Name).
		WithValues("namespace", ar.Request.Namespace)

	logger.V(2).Info("admitting endpoints")
	endpointResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "endpoints"}
	if ar.Request.Resource != endpointResource {
		logger.Error(fmt.Errorf("bad resource"), fmt.Sprintf("expect resource to be %s", endpointResource))
		return Allow()
	}
	if ar.Request.Operation != v1.Update {
		return Allow()
	}

	curr, err := decodeRawEndpoint(ar, "new")
	if err != nil {
		logger.Error(err, "decode endpoint failed")
		return DenyWithMessage(err.Error())
	}
	prev, err := decodeRawEndpoint(ar, "old")
	if err != nil {
		logger.Error(err, "decode endpoint failed")
		return DenyWithMessage(err.Error())
	}

	patches := patchEndpoints(curr, prev)
	logger.Info("Patch", patches)
	return AllowWithJsonPatch(patches)
}

func decodeRawEndpoint(ar v1.AdmissionReview, version string) (*corev1.Endpoints, error) {
	var raw []byte
	if version == "new" {
		raw = ar.Request.Object.Raw
	} else if version == "old" {
		raw = ar.Request.OldObject.Raw
	}

	endpoint := &corev1.Endpoints{}
	deserializer := Codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, endpoint); err != nil {
		klog.Error(err)
		return nil, err
	}
	return endpoint, nil
}

func patchEndpoints(newEndpoints, oldEndpoints *corev1.Endpoints) []Patch {
	var (
		removeAllPatched  []Patch
		removePatches     []Patch
		addPatches        []Patch
		oldReadyEndpoints = map[string]struct{}{}
	)

	for _, subset := range oldEndpoints.Subsets {
		for _, endpoint := range subset.Addresses {
			oldReadyEndpoints[endpoint.IP] = struct{}{}
		}
	}

	for i, subset := range newEndpoints.Subsets {
		addressesCreated := false
		removeCount := 0
		for j, address := range subset.NotReadyAddresses {
			if address.NodeName == nil || !nodeNotReady(*address.NodeName) {
				continue
			}
			if _, ok := oldReadyEndpoints[address.IP]; !ok {
				continue
			}
			removeCount++
			removePatches = append(removePatches, Patch{
				OP:   "remove",
				Path: fmt.Sprintf("/subsets/%d/notReadyAddresses/%d", i, j),
			})

			if len(subset.Addresses) == 0 && !addressesCreated {
				addressesCreated = true
				addPatches = append(addPatches, Patch{
					OP:    "add",
					Path:  fmt.Sprintf("/subsets/%d/addresses", i),
					Value: []string{},
				})
			}

			addPatches = append(addPatches, Patch{
				OP:    "add",
				Path:  fmt.Sprintf("/subsets/%d/addresses/-", i),
				Value: marshal(&address),
			})
		}
		if removeCount == len(subset.NotReadyAddresses) && removeCount > 0 {
			removeAllPatched = append(removeAllPatched, Patch{
				OP:   "remove",
				Path: fmt.Sprintf("/subsets/%d/notReadyAddresses", i),
			})
		}
	}
	return append(append(addPatches, reverse(removePatches)...), removeAllPatched...)
}

func reverse(patches []Patch) []Patch {
	i, j := 0, len(patches)-1
	for i < j {
		patches[i], patches[j] = patches[j], patches[i]
		i++
		j--
	}
	return patches
}

func marshal(address *corev1.EndpointAddress) map[string]interface{} {
	return map[string]interface{}{
		"ip":       address.IP,
		"hostname": address.Hostname,
		"nodeName": address.NodeName,
		"targetRef": map[string]interface{}{
			"kind":            address.TargetRef.Kind,
			"namespace":       address.TargetRef.Namespace,
			"name":            address.TargetRef.Name,
			"uid":             address.TargetRef.UID,
			"apiVersion":      address.TargetRef.APIVersion,
			"resourceVersion": address.TargetRef.ResourceVersion,
			"fieldPath":       address.TargetRef.FieldPath,
		},
	}
}

func nodeNotReady(name string) bool {
	node, err := config.Kubeclient.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return false
	}

	_, condition := util.GetNodeCondition(&node.Status, corev1.NodeReady)
	_, ok := node.Annotations["nodeunhealth"]

	return !ok && condition.Status == corev1.ConditionUnknown
}
