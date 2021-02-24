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
	"io/ioutil"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"net/http"
	"time"
)

type EdgeHealthAdmission struct {
	cfg        *config.EdgeHealthAdmissionConfig
	nodeLister corelisters.NodeLister
}

func NewEdgeHealthAdmission(c *config.EdgeHealthAdmissionConfig) *EdgeHealthAdmission {
	return &EdgeHealthAdmission{
		cfg:        c,
		nodeLister: c.NodeInformer.Lister(),
	}
}

// toAdmissionResponse is a helper function to create an AdmissionResponse
// with an embedded error
func toAdmissionResponse(err error) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

// admitFunc is the type we use for all of our validators and mutators
type admitFunc func(admissionv1.AdmissionReview) *admissionv1.AdmissionResponse

// serve handles the http portion of a request prior to handing to an admit function
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

	klog.V(4).Info(fmt.Sprintf("handling request: %s", body))

	// The AdmissionReview that was sent to the webhook
	requestedAdmissionReview := admissionv1.AdmissionReview{}

	// The AdmissionReview that will be returned
	responseAdmissionReview := admissionv1.AdmissionReview{}

	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &requestedAdmissionReview); err != nil {
		klog.Error(err)
		responseAdmissionReview.Response = toAdmissionResponse(err)
	} else {
		// pass to admitFunc
		responseAdmissionReview.Response = admit(requestedAdmissionReview)
	}

	// Return the same UID
	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID

	klog.V(4).Info(fmt.Sprintf("sending response: %v", responseAdmissionReview.Response))

	respBytes, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		klog.Error(err)
	}
	if _, err := w.Write(respBytes); err != nil {
		klog.Error(err)
	}
}

func (eha *EdgeHealthAdmission) serveNodeTaint(w http.ResponseWriter, r *http.Request) {
	serve(w, r, eha.mutateNodeTaint)
}

func (eha *EdgeHealthAdmission) serveEndpoint(w http.ResponseWriter, r *http.Request) {
	serve(w, r, eha.mutateEndpoint)
}

func (eha *EdgeHealthAdmission) Run(stopCh <-chan struct{}) {
	if !cache.WaitForNamedCacheSync("edge-health-admission", stopCh, eha.cfg.NodeInformer.Informer().HasSynced) {
		return
	}

	http.HandleFunc("/node-taint", eha.serveNodeTaint)
	http.HandleFunc("/endpoint", eha.serveEndpoint)
	server := &http.Server{
		Addr: eha.cfg.Addr,
	}

	go func() {
		if err := server.ListenAndServeTLS(eha.cfg.CertFile, eha.cfg.KeyFile); err != http.ErrServerClosed {
			klog.Fatalf("ListenAndServeTLS err %v", err)
		}
	}()

	for {
		select {
		case <-stopCh:
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				klog.Errorf("Server: program exit, server exit error %v", err)
			}
			return
		default:
		}
	}
}
