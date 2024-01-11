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
	"errors"
	"fmt"
	"github.com/superedge/superedge/pkg/penetrator/apis/nodetask.apps.superedge.io/v1beta1"
	clientset "github.com/superedge/superedge/pkg/penetrator/client/clientset/versioned"
	"github.com/superedge/superedge/pkg/penetrator/constants"
	"github.com/superedge/superedge/pkg/penetrator/operator/context"
	"io/ioutil"
	admissionv1 "k8s.io/api/admission/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"math/rand"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"
)

var (
	scheme  = runtime.NewScheme()
	codecs  = serializer.NewCodecFactory(scheme)
	letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
)

type Patch struct {
	Op    string      `json:"op,inline"`
	Path  string      `json:"path,inline"`
	Value interface{} `json:"value"`
}

func Handler(kubeclient kubernetes.Interface, nodetaskclient clientset.Interface, ctx *context.NodeTaskContext) http.Handler {

	handler := mutatingHandler{
		kubeclient:        kubeclient,
		nodetaskClientset: nodetaskclient,
		ctx:               ctx,
	}

	return http.HandlerFunc(handler.Handle)
}

type mutatingHandler struct {
	kubeclient        kubernetes.Interface
	nodetaskClientset clientset.Interface
	ctx               *context.NodeTaskContext
}

func (handler *mutatingHandler) Handle(w http.ResponseWriter, r *http.Request) {
	klog.Infof("request = %v\n", r)
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return
	}

	if r.Body == nil {
		return
	}

	admissionReview := &admissionv1.AdmissionReview{}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		klog.Errorf("failed to read webhook request data, error:%v", err)
		admissionReview.Response = &admissionv1.AdmissionResponse{
			Allowed: false,
			Result:  &metav1.Status{Message: err.Error()},
		}
		writeResponse(admissionReview, w)
		return
	}

	// Decode to get admissionReview
	deserializer := codecs.UniversalDeserializer()
	_, _, err = deserializer.Decode(data, nil, admissionReview)
	if err != nil {
		klog.Errorf("failed to decode webhook admissionReview, error:%v", err)
		admissionReview.Response = &admissionv1.AdmissionResponse{
			Allowed: false,
			Result:  &metav1.Status{Message: err.Error()},
		}
		writeResponse(admissionReview, w)
		return
	}

	nt := &v1beta1.NodeTask{}

	// Decode to get NodeTask
	_, _, err = deserializer.Decode(admissionReview.Request.Object.Raw, nil, nt)
	if err != nil {
		klog.Errorf("failed to decode webhook nodetask, error:%v", err)
		admissionReview.Response = &admissionv1.AdmissionResponse{
			Allowed: false,
			UID:     admissionReview.Request.UID,
			Result:  &metav1.Status{Message: err.Error()},
		}
		writeResponse(admissionReview, w)
		return
	}

	switch admissionReview.Request.Operation {
	case admissionv1.Create:

		// Verify nodetask
		if errs := handler.validate(nt); len(errs) > 0 {
			klog.Errorf("webhook check failed, error: %v", errs)
			errmsgs := make([]string, 0)
			for _, err := range errs {
				errmsgs = append(errmsgs, err.Error())
			}
			admissionReview.Response = &admissionv1.AdmissionResponse{
				Allowed: false,
				UID:     admissionReview.Request.UID,
				Result:  &metav1.Status{Message: strings.Join(errmsgs, ",")},
			}
			writeResponse(admissionReview, w)
			return
		}

		klog.Infof("nodetask = %v\n", nt)
		patches, err := handler.getCreatePatch(nt)
		if err != nil {
			klog.Errorf("failed to create webhook patches, error: %v", err)
			admissionReview.Response = &admissionv1.AdmissionResponse{
				Allowed: false,
				UID:     admissionReview.Request.UID,
				Result:  &metav1.Status{Message: err.Error()},
			}
			writeResponse(admissionReview, w)
			return
		}

		writeResponseWithPatch(patches, admissionReview, w)

	case admissionv1.Update:
		//Does not support updating the spec and status of nodetask
		old, err := handler.nodetaskClientset.NodestaskV1beta1().NodeTasks().Get(handler.ctx, nt.Name, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("failed to get nodetask, error: %v", err)
			admissionReview.Response = &admissionv1.AdmissionResponse{
				Allowed: false,
				UID:     admissionReview.Request.UID,
				Result:  &metav1.Status{Message: fmt.Sprintf("webhook failed to get nodetask, error: %v", err)},
			}
			writeResponse(admissionReview, w)
			return
		}

		if !reflect.DeepEqual(old.Spec, nt.Spec) {
			admissionReview.Response = &admissionv1.AdmissionResponse{
				Allowed: false,
				UID:     admissionReview.Request.UID,
				Result:  &metav1.Status{Message: "nodetask does not support update"},
			}
			writeResponse(admissionReview, w)
			return
		}

		admissionReview.Response = &admissionv1.AdmissionResponse{
			Allowed: true,
			UID:     admissionReview.Request.UID,
		}
		writeResponse(admissionReview, w)

	default:
		admissionReview.Response = &admissionv1.AdmissionResponse{
			Allowed: true,
			UID:     admissionReview.Request.UID,
		}
		writeResponse(admissionReview, w)

	}
}

func writeResponse(response *admissionv1.AdmissionReview, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		klog.Errorf("failed to write admissionReview, error: %v", err)
	}
}

func writeResponseWithPatch(patches []Patch, response *admissionv1.AdmissionReview, w http.ResponseWriter) {
	klog.Infof("write patches %+v", patches)
	patchData, err := json.Marshal(patches)
	if err != nil {
		klog.Error(err)
		return
	}

	v1JSONPatch := admissionv1.PatchTypeJSONPatch
	response.Response = &admissionv1.AdmissionResponse{
		UID:       response.Request.UID,
		Patch:     patchData,
		Allowed:   true,
		PatchType: &v1JSONPatch,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		klog.Errorf("failed to write admissionReview, error: %v", err)
	}
}

func (handler *mutatingHandler) getCreatePatch(nt *v1beta1.NodeTask) ([]Patch, error) {
	patches := make([]Patch, 0)
	randomToken := nt.Name + "-" + randString(6)

	// nameIps
	if nt.Spec.NodeNamesOverride == nil {
		if nt.Spec.NodeNamePrefix == "" || nt.Spec.TargetMachines == nil {
			return patches, fmt.Errorf("targetMachines or NodeNamePrefix are empty")
		} else {
			// ip de-duplication
			ipsMap := make(map[string]interface{})
			for k, v := range nt.Spec.TargetMachines {
				ipsMap[v] = k
			}

			if len(ipsMap) != len(nt.Spec.TargetMachines) {
				ips := []string{}
				for k := range ipsMap {
					ips = append(ips, k)
				}
				nt.Spec.TargetMachines = ips
			}
			nodes, err := handler.kubeclient.CoreV1().Nodes().List(handler.ctx, metav1.ListOptions{})
			if err != nil {
				if !apierrors.IsNotFound(err) {
					klog.Errorf("failed to get nodes, error: %v", err)
					return patches, err
				}
			}

			nodesMap := make(map[string]interface{})
			for _, node := range nodes.Items {
				nodesMap[node.Name] = struct{}{}
			}

			nameIps := make(map[string]string)
			var nodeName string
			for _, v := range nt.Spec.TargetMachines {
				for true {
					nodeName = nt.Spec.NodeNamePrefix + "-" + randString(6)
					if _, ok := nodesMap[nodeName]; !ok {
						break
					}
				}
				nameIps[nodeName] = v
			}

			patches = append(patches,
				Patch{
					Op:    "add",
					Path:  "/spec/nodeNamesOverride",
					Value: nameIps,
				},
			)
		}
	}

	//annotation
	patches = append(patches,
		Patch{
			Op:   "add",
			Path: "/metadata/annotations",
			Value: map[string]string{
				constants.AnnotationAddNodeJobName:       randomToken,
				constants.AnnotationAddNodeConfigmapName: randomToken,
			},
		},
	)

	//label
	patches = append(patches,
		Patch{
			Op:   "add",
			Path: "/metadata/labels",
			Value: map[string]string{
				constants.NodeLabel: randomToken,
			},
		},
	)

	return patches, nil
}

func (handler *mutatingHandler) validate(nt *v1beta1.NodeTask) []error {
	errs := make([]error, 0)
	fld := field.NewPath("nodetask")
	if len(nt.Spec.NodeNamesOverride) == 0 && len(nt.Spec.TargetMachines) == 0 {
		errs = append(errs, errors.New(fmt.Sprintf("%s or %s must be specified", fld.Key("spec.ips").String(), fld.Key("spec.nameIps").String())))
	}

	if len(nt.Spec.NodeNamesOverride) != 0 && len(nt.Spec.TargetMachines) != 0 {
		errs = append(errs, errors.New(fmt.Sprintf("only one of %s and %s can be selected", fld.Key("spec.ips").String(), fld.Key("spec.nameIps").String())))
	}

	// Verify the format of ip
	if len(nt.Spec.TargetMachines) != 0 {
		for _, ip := range nt.Spec.TargetMachines {
			addr := net.ParseIP(ip)
			if addr == nil {
				errs = append(errs, errors.New(fmt.Sprintf("The ip format of the %s array is err ", fld.Key("spec.Ips").String())))
				break
			}
		}
	}

	if len(nt.Spec.NodeNamesOverride) != 0 {
		for _, ip := range nt.Spec.NodeNamesOverride {
			addr := net.ParseIP(ip)
			if addr == nil {
				errs = append(errs, errors.New(fmt.Sprintf("The ip format of the %s array is err", fld.Key("spec.nameIps").String())))
				break
			}
		}
	}

	loginSecret, err := handler.kubeclient.CoreV1().Secrets(constants.NameSpaceEdge).Get(handler.ctx, nt.Spec.SSHCredential, metav1.GetOptions{})
	if err != nil {
		errs = append(errs, errors.New(fmt.Sprintf("Failed to get secret, error: %s", err.Error())))
	} else {
		_, pwdOk := loginSecret.Data[constants.PassWd]
		_, sshOk := loginSecret.Data[constants.SshKey]
		if !(pwdOk || sshOk) {
			errs = append(errs, errors.New("Failed to obtain login password (passwd) or private key (sshkey) from secret"))
		}
	}

	nl, err := handler.nodetaskClientset.NodestaskV1beta1().NodeTasks().List(handler.ctx, metav1.ListOptions{})
	if err != nil {
		errs = append(errs, errors.New(fmt.Sprintf("Failed to list nodestask, error: %s", err.Error())))
	} else {
		if len(nl.Items) != 0 {
			if nl.Items[0].Status.NodeTaskStatus == v1beta1.NodeTaskStatusCreating {
				errs = append(errs, errors.New(fmt.Sprintf("The task '%s' is executing, please delete the task 'kubectl -n %s delete task %s' that is currently running, and then create it ", nl.Items[0].Name, nl.Items[0].Namespace, nl.Items[0].Name)))
			} else {
				errs = append(errs, errors.New(fmt.Sprintf("The task '%s' has been completed, please delete the task 'kubectl -n %s delete task %s' that is currently running, and then create it ", nl.Items[0].Name, nl.Items[0].Namespace, nl.Items[0].Name)))
			}
		}
	}
	return errs
}

func randString(n int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}
