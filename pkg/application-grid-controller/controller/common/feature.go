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

package common

import (
	"k8s.io/component-base/featuregate"
)

const (
	ServiceTopology        featuregate.Feature = "ServiceTopology"
	EndpointSlice          featuregate.Feature = "EndpointSlice"
	EvenPodsSpread         featuregate.Feature = "EvenPodsSpread"
	TopologyAnnotationsKey                     = "topologyKeys"
)

var DefaultKubernetesFeatureGates = map[featuregate.Feature]featuregate.FeatureSpec{
	ServiceTopology: {Default: false, PreRelease: featuregate.Alpha},
	EndpointSlice:   {Default: false, PreRelease: featuregate.Beta},
	EvenPodsSpread:  {Default: false, PreRelease: featuregate.Alpha},
}
