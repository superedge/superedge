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
	"k8s.io/klog/v2"
	"sync"

	"github.com/superedge/superedge/pkg/edge-health/data"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type storageCache struct {
	// hostName is the nodeName of node which application-grid-wrapper deploys on
	hostName                          string
	wrapperInCluster                  bool
	serviceAutonomyEnhancementEnabled bool
	supportEndpointSlice              bool

	// mu lock protect the following map structure
	mu                      sync.RWMutex
	servicesMap             map[types.NamespacedName]*serviceContainer
	endpointsMap            map[types.NamespacedName]*endpointsContainer
	endpointSliceV1Map      map[types.NamespacedName]*endpointSliceV1Container
	endpointSliceV1Beta1Map map[types.NamespacedName]*endpointSliceV1Beta1Container
	nodesMap                map[types.NamespacedName]*nodeContainer
	localNodeInfo           map[string]data.ResultDetail

	// service watch channel
	serviceBroardcaster *watch.Broadcaster
	// endpoints watch channel
	endpointsBroadcaster *watch.Broadcaster
	// endpointSlice watch channel
	endpointSliceV1Broardcaster *watch.Broadcaster

	endpointSliceV1Beta1Boardcaster *watch.Broadcaster
	nodeBroadcaster                 *watch.Broadcaster
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

// endpointSliceContainer stores original kubernetes endpointSlice and relevant modified serviceTopology endpointSlice
type endpointSliceV1Container struct {
	endpointSlice *discoveryv1.EndpointSlice
	modified      *discoveryv1.EndpointSlice
}
type endpointSliceV1Beta1Container struct {
	endpointSlice *discoveryv1beta1.EndpointSlice
	modified      *discoveryv1beta1.EndpointSlice
}

var _ Cache = &storageCache{}

func NewStorageCache(hostName string, wrapperInCluster, serviceAutonomyEnhancementEnabled bool, serviceBroadcaster, endpointSliceV1Broadcaster, endpointSliceV1Beta1Broadcaster, endpointBroadcaster, nodeBroadcaster *watch.Broadcaster, supportEndpointSlice bool) *storageCache {
	msc := &storageCache{
		hostName:                          hostName,
		wrapperInCluster:                  wrapperInCluster,
		serviceAutonomyEnhancementEnabled: serviceAutonomyEnhancementEnabled,
		supportEndpointSlice:              supportEndpointSlice,
		servicesMap:                       make(map[types.NamespacedName]*serviceContainer),
		endpointsMap:                      make(map[types.NamespacedName]*endpointsContainer),
		endpointSliceV1Map:                make(map[types.NamespacedName]*endpointSliceV1Container),
		endpointSliceV1Beta1Map:           make(map[types.NamespacedName]*endpointSliceV1Beta1Container),
		nodesMap:                          make(map[types.NamespacedName]*nodeContainer),
		serviceBroardcaster:               serviceBroadcaster,
		endpointsBroadcaster:              endpointBroadcaster,
		endpointSliceV1Broardcaster:       endpointSliceV1Broadcaster,
		endpointSliceV1Beta1Boardcaster:   endpointSliceV1Beta1Broadcaster,
		nodeBroadcaster:                   nodeBroadcaster,
		localNodeInfo:                     make(map[string]data.ResultDetail),
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

func (sc *storageCache) EndpointSliceV1EventHandler() cache.ResourceEventHandler {
	return &endpointSliceV1Handler{cache: sc}
}

func (sc *storageCache) EndpointSliceV1Beta1EventHandler() cache.ResourceEventHandler {
	return &endpointSliceV1Beta1Handler{cache: sc}
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

func (sc *storageCache) GetEndpointSliceV1() []*discoveryv1.EndpointSlice {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	epsList := make([]*discoveryv1.EndpointSlice, 0, len(sc.endpointSliceV1Map))
	for _, v := range sc.endpointSliceV1Map {
		epsList = append(epsList, v.modified)
	}
	return epsList
}

func (sc *storageCache) GetEndpointSliceV1Beta1() []*discoveryv1beta1.EndpointSlice {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	epsList := make([]*discoveryv1beta1.EndpointSlice, 0, len(sc.endpointSliceV1Beta1Map))
	for _, v := range sc.endpointSliceV1Beta1Map {
		epsList = append(epsList, v.modified)
	}
	return epsList
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

func (sc *storageCache) GetLocalNodeInfo() map[string]data.ResultDetail {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.localNodeInfo
}

func (sc *storageCache) ClearLocalNodeInfo() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.localNodeInfo = make(map[string]data.ResultDetail)
}

func (sc *storageCache) SetLocalNodeInfo(info map[string]data.ResultDetail) {
	klog.V(4).Infof("Set local node info %#v", info)
	sc.mu.Lock()
	sc.localNodeInfo = info
	sc.mu.Unlock()
	sc.resync()
}

func (sc *storageCache) resync() {
	if len(sc.endpointSliceV1Map) > 0 {
		// update endpointSlice
		sc.mu.Lock()
		changedEndpointSliceV1 := sc.rebuildEndpointSliceV1Map()
		sc.mu.Unlock()
		for _, eps := range changedEndpointSliceV1 {
			sc.endpointSliceV1Broardcaster.Action(eps.Type, eps.Object)
		}
	}

	if len(sc.GetEndpointSliceV1Beta1()) > 0 {
		// update endpointSlice
		sc.mu.Lock()
		changedEndpointSliceV1Beta1 := sc.rebuildEndpointSliceV1Beta1Map()
		sc.mu.Unlock()
		for _, eps := range changedEndpointSliceV1Beta1 {
			sc.endpointSliceV1Broardcaster.Action(eps.Type, eps.Object)
		}
	}

	if len(sc.endpointsMap) > 0 {
		// update endpoints
		sc.mu.Lock()
		changedEps := sc.rebuildEndpointsMap()
		sc.mu.Unlock()
		for _, v := range changedEps {
			sc.endpointsBroadcaster.ActionOrDrop(v.Type, v.Object)
		}
	}
}

// rebuildEndpointsMap updates all endpoints stored in storageCache.endpointsMap dynamically and constructs relevant modified events
func (sc *storageCache) rebuildEndpointsMap() []watch.Event {
	evts := make([]watch.Event, 0)
	for name, endpointsContainer := range sc.endpointsMap {
		newEps := pruneEndpoints(sc.hostName, sc.nodesMap, sc.servicesMap, endpointsContainer.endpoints, sc.localNodeInfo, sc.wrapperInCluster, sc.serviceAutonomyEnhancementEnabled)
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

// rebuildEndpointSliceMap updates all endpoints stored in storageCache.endpointSliceMap dynamically and constructs relevant modified events
func (sc *storageCache) rebuildEndpointSliceV1Map() []watch.Event {
	evts := make([]watch.Event, 0)
	for name, endpointSliceContainer := range sc.endpointSliceV1Map {
		newEps := pruneEndpointSliceV1(sc.hostName, sc.nodesMap, sc.servicesMap, endpointSliceContainer.endpointSlice, sc.localNodeInfo, sc.wrapperInCluster, sc.serviceAutonomyEnhancementEnabled)
		if apiequality.Semantic.DeepEqual(newEps, endpointSliceContainer.modified) {
			continue
		}
		sc.endpointSliceV1Map[name].modified = newEps
		evts = append(evts, watch.Event{
			Type:   watch.Modified,
			Object: newEps,
		})
	}
	return evts
}

func (sc *storageCache) rebuildEndpointSliceV1Beta1Map() []watch.Event {
	evts := make([]watch.Event, 0)
	for name, endpointSliceContainer := range sc.endpointSliceV1Beta1Map {
		newEps := pruneEndpointSliceV1Beta1(sc.hostName, sc.nodesMap, sc.servicesMap, endpointSliceContainer.endpointSlice, sc.localNodeInfo, sc.wrapperInCluster, sc.serviceAutonomyEnhancementEnabled)
		if apiequality.Semantic.DeepEqual(newEps, endpointSliceContainer.modified) {
			continue
		}
		sc.endpointSliceV1Beta1Map[name].modified = newEps
		evts = append(evts, watch.Event{
			Type:   watch.Modified,
			Object: newEps,
		})
	}
	return evts
}

func (sc *storageCache) GetNodeName() string {
	return sc.hostName
}
