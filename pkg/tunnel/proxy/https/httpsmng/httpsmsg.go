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

package httpsmng

import (
	"encoding/json"
	"k8s.io/klog"
	"net/http"
)

type HttpsMsg struct {
	StatusCode  int         `json:"status_code"`
	HttpsStatus string      `json:"https_status"`
	HttpBody    []byte      `json:"http_body"`
	Header      http.Header `json:"header"`
	Method      string      `json:"method"`
}

func (msg *HttpsMsg) Serialization() []byte {
	bmsg, err := json.Marshal(msg)
	if err != nil {
		klog.Errorf("httpsmsg serialization failed err = %v", err)
		return nil
	}
	return bmsg
}
func Deserialization(data []byte) (*HttpsMsg, error) {
	msg := &HttpsMsg{}
	err := json.Unmarshal(data, msg)
	if err != nil {
		klog.Errorf("httpsmsg deserialization failed err = %v", err)
		return nil, err
	}
	return msg, nil
}
