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

package server

import (
	"github.com/munnerz/goautoneg"
	"k8s.io/apimachinery/pkg/runtime"
)

func (s *interceptorServer) parseAccept(header string, accepted []runtime.SerializerInfo) (runtime.SerializerInfo, bool) {
	if len(header) == 0 && len(accepted) > 0 {
		return accepted[0], true
	}

	clauses := goautoneg.ParseAccept(header)
	for i := range clauses {
		clause := &clauses[i]
		for i := range accepted {
			accepts := &accepted[i]
			switch {
			case clause.Type == accepts.MediaTypeType && clause.SubType == accepts.MediaTypeSubType,
				clause.Type == accepts.MediaTypeType && clause.SubType == "*",
				clause.Type == "*" && clause.SubType == "*":
				return *accepts, true
			}
		}
	}

	return runtime.SerializerInfo{}, false
}
