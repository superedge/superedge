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

package storage

import (
	"github.com/superedge/superedge/pkg/edge-health/data"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	"k8s.io/client-go/tools/cache"
)

type Cache interface {
	CacheList
	CacheEventHandler
	LocalNodeInfoHandler
	GetNodeName() string
}

type CacheList interface {
	GetNode(hostName string) *v1.Node
	GetServices() []*v1.Service
	GetEndpoints() []*v1.Endpoints
	GetEndpointSliceV1() []*discoveryv1.EndpointSlice
	GetEndpointSliceV1Beta1() []*discoveryv1beta1.EndpointSlice
}

type CacheEventHandler interface {
	NodeEventHandler() cache.ResourceEventHandler
	ServiceEventHandler() cache.ResourceEventHandler
	EndpointsEventHandler() cache.ResourceEventHandler
	EndpointSliceV1EventHandler() cache.ResourceEventHandler
	EndpointSliceV1Beta1EventHandler() cache.ResourceEventHandler
}

type LocalNodeInfoHandler interface {
	GetLocalNodeInfo() map[string]data.ResultDetail
	SetLocalNodeInfo(map[string]data.ResultDetail)
	ClearLocalNodeInfo()
}
