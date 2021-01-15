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
	"sync"

	v1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type storageCache struct {
	// hostName is the nodeName of node which application-grid-wrapper deploys on
	hostName         string
	wrapperInCluster bool

	// mu lock protect the following map structure
	mu           sync.RWMutex
	servicesMap  map[types.NamespacedName]*serviceContainer
	endpointsMap map[types.NamespacedName]*endpointsContainer
	nodesMap     map[types.NamespacedName]*nodeContainer

	// service watch channel
	serviceChan chan<- watch.Event
	// endpoints watch channel
	endpointsChan chan<- watch.Event
}

// serviceContainer stores kubernetes service and its topologyKeys
type serviceContainer struct {
	svc  *v1.Service
	keys []string
}

// nodeContainer stores kubernetes node and its labels
type nodeContainer struct {
	node   *v1.Node
	labels map[string]string
}

// endpointsContainer stores original kubernetes endpoints and relevant modified serviceTopology endpoints
type endpointsContainer struct {
	endpoints *v1.Endpoints
	modified  *v1.Endpoints
}

var _ Cache = &storageCache{}

func NewStorageCache(hostName string, wrapperInCluster bool, serviceNotifier, endpointsNotifier chan watch.Event) *storageCache {
	msc := &storageCache{
		hostName:         hostName,
		wrapperInCluster: wrapperInCluster,
		servicesMap:      make(map[types.NamespacedName]*serviceContainer),
		endpointsMap:     make(map[types.NamespacedName]*endpointsContainer),
		nodesMap:         make(map[types.NamespacedName]*nodeContainer),
		serviceChan:      serviceNotifier,
		endpointsChan:    endpointsNotifier,
	}

	return msc
}

func (sc *storageCache) NodeEventHandler() cache.ResourceEventHandler {
	return &nodeHandler{cache: sc}
}

func (sc *storageCache) ServiceEventHandler() cache.ResourceEventHandler {
	return &serviceHandler{cache: sc}
}

func (sc *storageCache) EndpointsEventHandler() cache.ResourceEventHandler {
	return &endpointsHandler{cache: sc}
}

func (sc *storageCache) GetServices() []*v1.Service {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	svcList := make([]*v1.Service, 0, len(sc.servicesMap))
	for _, v := range sc.servicesMap {
		svcList = append(svcList, v.svc)
	}
	return svcList
}

func (sc *storageCache) GetEndpoints() []*v1.Endpoints {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	epList := make([]*v1.Endpoints, 0, len(sc.endpointsMap))
	for _, v := range sc.endpointsMap {
		epList = append(epList, v.modified)
	}
	return epList
}

func (sc *storageCache) GetNodes() []*v1.Node {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	nodeList := make([]*v1.Node, 0, len(sc.nodesMap))
	for _, v := range sc.nodesMap {
		nodeList = append(nodeList, v.node)
	}
	return nodeList
}

func (sc *storageCache) GetNode(hostName string) *v1.Node {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	nodeKey := types.NamespacedName{Name: hostName}
	nodeContainer, found := sc.nodesMap[nodeKey]
	if found {
		return nodeContainer.node
	}

	return nil
}

// rebuildEndpointsMap updates all endpoints stored in storageCache.endpointsMap dynamically and constructs relevant modified events
func (sc *storageCache) rebuildEndpointsMap() []watch.Event {
	evts := make([]watch.Event, 0)
	for name, endpointsContainer := range sc.endpointsMap {
		newEps := pruneEndpoints(sc.hostName, sc.nodesMap, sc.servicesMap, endpointsContainer.endpoints, sc.wrapperInCluster)
		if apiequality.Semantic.DeepEqual(newEps, endpointsContainer.modified) {
			continue
		}
		sc.endpointsMap[name].modified = newEps
		evts = append(evts, watch.Event{
			Type:   watch.Modified,
			Object: newEps,
		})
	}
	return evts
}
