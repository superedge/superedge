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
	corev1 "k8s.io/api/core/v1"
	"time"
)

const (
	CheckScoreMax          = 100
	CheckScoreMin          = 0
	TopologyZone           = "superedgehealth/topology-zone"
	TaintZoneConfigMap     = "edge-health-zone-config"
	TaintZoneConfigMapKey  = "TaintZoneAdmission"
	HmacConfig             = "hmac-config"
	HmacKey                = "hmackey"
	MasterLabel            = "node-role.kubernetes.io/master"
	TokenFile              = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	ReListTime             = 2 * time.Minute
	NodeUnhealthAnnotation = "nodeunhealth"
)

var (
	UnreachableNoExecuteTaint = &corev1.Taint{
		Key:    corev1.TaintNodeUnreachable,
		Effect: corev1.TaintEffectNoExecute,
	}
)
