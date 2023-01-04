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
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
)

const (
	CmdName                 = "edge-health"
	TopologyZone            = "superedgehealth/topology-zone"
	HealthCheckUnitsKey     = "check-units"
	HealthCheckUnitEnable   = "unit-internal-check"
	EdgeHealthConfigMapName = "edge-health-config"
	HmacConfig              = "hmac-config"
	HmacKey                 = "hmackey"
	// if without hmac-config configmap and without hmackey, it will use default hmac key
	DefaultHmacKey   = "hSZbJsKAVmWTxRPi"
	MasterLabel      = "node-role.kubernetes.io/master"
	TokenFile        = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	ReListTime       = 2 * time.Minute
	DefaultNamespace = "kube-system"
)

var (
	Namespace string
	PodName   string
	PodIP     string

	NodeName  string
	NodeIP    string
	ClientSet *kubernetes.Clientset
	// watch node only use metadata client, it will reduce a lot of apiserver bandwidth
	MetadataClientSet metadata.Interface
)
