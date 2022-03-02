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
	"bytes"
	"encoding/json"
	"fmt"
	_ "strings"

	corev1 "k8s.io/api/core/v1"
	k8syaml "sigs.k8s.io/yaml"
)

func OutPutMessage(msg string) {
	fmt.Printf("\033[1;31;40m%s\033[0m\n", msg)
}

func OutSuccessMessage(msg string) {
	fmt.Printf("\033[1;32;40m%s\033[0m\n", msg)
}

func ToJson(v interface{}) string {
	json, _ := json.Marshal(v)
	return string(json)
}

func ToJsonForm(v interface{}) string {
	json, _ := json.MarshalIndent(v, "", "   ")
	return string(json)
}

func DisableEscapeJson(data interface{}) (string, error) {
	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	if err := jsonEncoder.Encode(data); err != nil {
		return "", err
	}
	return bf.String(), nil
}

func PodToYaml(pod *corev1.Pod) ([]byte, error) {
	podJson, err := json.Marshal(pod)
	if err != nil {
		return nil, err
	}
	data, err := k8syaml.JSONToYAML(podJson)
	if err != nil {
		return nil, err
	}
	return data, nil
}
