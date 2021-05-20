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

package proxy

import (
	"encoding/json"
)

// EdgeResponseDataHolder hold all data of a response
type EdgeResponseDataHolder struct {
	Code   int                 `json:"code"`
	Body   []byte              `json:"body"`
	Header map[string][]string `json:"header"`
}

func (holder *EdgeResponseDataHolder) Output() ([]byte, error) {
	res, err := json.Marshal(holder)
	if err != nil {
		return []byte{}, err
	}
	return res, nil
}

func (holder *EdgeResponseDataHolder) Input(in []byte) error {
	err := json.Unmarshal(in, holder)
	if err != nil {
		return err
	}
	return nil
}
