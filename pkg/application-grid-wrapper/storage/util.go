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
	"encoding/json"
	"github.com/superedge/superedge/pkg/edge-health/data"
	"net"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

const (
	TopologyAnnotationsKey = "topologyKeys"

	EdgeLocalEndpoint  = "superedge.io/local-endpoint"
	EdgeLocalPort      = "superedge.io/local-port"
	MasterEndpointName = "kubernetes"
)

func getTopologyKeys(objectMeta *metav1.ObjectMeta) []string {
	if !hasTopologyKey(objectMeta) {
		return nil
	}

	var keys []string
	keyData := objectMeta.Annotations[TopologyAnnotationsKey]
	if err := json.Unmarshal([]byte(keyData), &keys); err != nil {
		klog.Errorf("can't parse topology keys %s, %v", keyData, err)
		return nil
	}

	return keys
}

func hasTopologyKey(objectMeta *metav1.ObjectMeta) bool {
	if objectMeta.Annotations == nil {
		return false
	}

	_, ok := objectMeta.Annotations[TopologyAnnotationsKey]
	return ok
}

func genLocalEndpoints(eps *v1.Endpoints) *v1.Endpoints {
	if eps.Namespace != metav1.NamespaceDefault || eps.Name != MasterEndpointName {
		return eps
	}

	klog.V(4).Infof("begin to gen local ep %v", eps)
	ipAddress, e := eps.Annotations[EdgeLocalEndpoint]
	if !e {
		return eps
	}

	portStr, e := eps.Annotations[EdgeLocalPort]
	if !e {
		return eps
	}

	klog.V(4).Infof("get local endpoint %s:%s", ipAddress, portStr)
	port, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		klog.Errorf("parse int %s err %v", portStr, err)
		return eps
	}

	ip := net.ParseIP(ipAddress)
	if ip == nil {
		klog.Warningf("parse ip %s nil", ipAddress)
		return eps
	}

	nep := eps.DeepCopy()
	nep.Subsets = []v1.EndpointSubset{
		{
			Addresses: []v1.EndpointAddress{
				{
					IP: ipAddress,
				},
			},
			Ports: []v1.EndpointPort{
				{
					Protocol: v1.ProtocolTCP,
					Port:     int32(port),
					Name:     "https",
				},
			},
		},
	}

	klog.V(4).Infof("gen new endpoint complete %v", nep)
	return nep
}

func genLocalEndpointSliceV1(eps *discoveryv1.EndpointSlice) *discoveryv1.EndpointSlice {
	if eps.Namespace != metav1.NamespaceDefault || eps.Name != MasterEndpointName {
		return eps
	}

	klog.V(4).Infof("begin to gen local endpointslice %v", eps)
	ipAddress, e := eps.Annotations[EdgeLocalEndpoint]
	if !e {
		return eps
	}

	portStr, e := eps.Annotations[EdgeLocalPort]
	if !e {
		return eps
	}

	klog.V(4).Infof("get local endpoint %s:%s", ipAddress, portStr)
	port, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		klog.Errorf("parse int %s err %v", portStr, err)
		return eps
	}

	ip := net.ParseIP(ipAddress)
	if ip == nil {
		klog.Warningf("parse ip %s nil", ipAddress)
		return eps
	}

	nep := eps.DeepCopy()
	nep.Endpoints = []discoveryv1.Endpoint{
		{
			Addresses: []string{
				ipAddress,
			},
		},
	}
	nameHelper := "https"
	portHelper := int32(port)
	protocolHelper := v1.ProtocolTCP
	nep.Ports = []discoveryv1.EndpointPort{
		{
			Protocol: &protocolHelper,
			Port:     &portHelper,
			Name:     &nameHelper,
		},
	}

	klog.V(4).Infof("gen new endpointslice complete %v", nep)
	return nep
}

func genLocalEndpointSliceV1Beta1(eps *discoveryv1beta1.EndpointSlice) *discoveryv1beta1.EndpointSlice {
	if eps.Namespace != metav1.NamespaceDefault || eps.Name != MasterEndpointName {
		return eps
	}

	klog.V(4).Infof("begin to gen local endpointslice %v", eps)
	ipAddress, e := eps.Annotations[EdgeLocalEndpoint]
	if !e {
		return eps
	}

	portStr, e := eps.Annotations[EdgeLocalPort]
	if !e {
		return eps
	}

	klog.V(4).Infof("get local endpoint %s:%s", ipAddress, portStr)
	port, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		klog.Errorf("parse int %s err %v", portStr, err)
		return eps
	}

	ip := net.ParseIP(ipAddress)
	if ip == nil {
		klog.Warningf("parse ip %s nil", ipAddress)
		return eps
	}

	nep := eps.DeepCopy()
	nep.Endpoints = []discoveryv1beta1.Endpoint{
		{
			Addresses: []string{
				ipAddress,
			},
		},
	}
	nameHelper := "https"
	portHelper := int32(port)
	protocolHelper := v1.ProtocolTCP
	nep.Ports = []discoveryv1beta1.EndpointPort{
		{
			Protocol: &protocolHelper,
			Port:     &portHelper,
			Name:     &nameHelper,
		},
	}

	klog.V(4).Infof("gen new endpointslice complete %v", nep)
	return nep
}

// pruneEndpoints filters endpoints using serviceTopology rules combined by services topologyKeys and node labels
func pruneEndpoints(hostName string,
	nodes map[types.NamespacedName]*nodeContainer,
	services map[types.NamespacedName]*serviceContainer,
	eps *v1.Endpoints, localNodeInfo map[string]data.ResultDetail, wrapperInCluster, serviceAutonomyEnhancementEnabled bool) *v1.Endpoints {

	epsKey := types.NamespacedName{Namespace: eps.Namespace, Name: eps.Name}

	if wrapperInCluster {
		eps = genLocalEndpoints(eps)
	}

	// dangling endpoints
	svc, ok := services[epsKey]
	if !ok {
		klog.V(4).Infof("Dangling endpoints %s, %+#v", eps.Name, eps.Subsets)
		return eps
	}

	// normal service
	if len(svc.keys) == 0 {
		klog.V(4).Infof("Normal endpoints %s, %+#v", eps.Name, eps.Subsets)
		if eps.Namespace == metav1.NamespaceDefault && eps.Name == MasterEndpointName {
			return eps
		}
		if serviceAutonomyEnhancementEnabled {
			newEps := eps.DeepCopy()
			for si := range newEps.Subsets {
				subnet := &newEps.Subsets[si]
				subnet.Addresses = filterLocalNodeInfoConcernedAddresses(nodes, subnet.Addresses, localNodeInfo)
				subnet.NotReadyAddresses = filterLocalNodeInfoConcernedAddresses(nodes, subnet.NotReadyAddresses, localNodeInfo)
			}
			klog.V(4).Infof("Normal endpoints after LocalNodeInfo filter %s: subnets from %+#v to %+#v", eps.Name, eps.Subsets, newEps.Subsets)
			return newEps
		}
		return eps
	}

	// topology endpoints
	newEps := eps.DeepCopy()
	for si := range newEps.Subsets {
		subnet := &newEps.Subsets[si]
		subnet.Addresses = filterConcernedAddresses(svc.keys, hostName, nodes, subnet.Addresses, localNodeInfo, serviceAutonomyEnhancementEnabled)
		subnet.NotReadyAddresses = filterConcernedAddresses(svc.keys, hostName, nodes, subnet.NotReadyAddresses, localNodeInfo, serviceAutonomyEnhancementEnabled)
	}
	klog.V(4).Infof("Topology endpoints %s: subnets from %+#v to %+#v", eps.Name, eps.Subsets, newEps.Subsets)

	return newEps
}

// pruneEndpointSlice filters endpointslice using serviceTopology rules combined by services topologyKeys and node labels
func pruneEndpointSliceV1(hostName string,
	nodes map[types.NamespacedName]*nodeContainer,
	services map[types.NamespacedName]*serviceContainer,
	eps *discoveryv1.EndpointSlice, localNodeInfo map[string]data.ResultDetail, wrapperInCluster, serviceAutonomyEnhancementEnabled bool) *discoveryv1.EndpointSlice {

	//drop suffix
	strArr := strings.Split(eps.Name, "-")
	epsName := strArr[0]
	for i := 1; i < len(strArr)-1; i++ {
		epsName = epsName + "-" + strArr[i]
	}
	epsKey := types.NamespacedName{Namespace: eps.Namespace, Name: epsName}

	if wrapperInCluster {
		eps = genLocalEndpointSliceV1(eps)
	}

	// dangling endpoints
	svc, ok := services[epsKey]
	if !ok {
		klog.V(4).Infof("Dangling endpointSlice %s, %+#v", eps.Name, eps.Endpoints)
		return eps
	}

	// normal service
	if len(svc.keys) == 0 {
		klog.V(4).Infof("Normal endpointSlice %s, %+#v", eps.Name, eps.Endpoints)
		if eps.Namespace == metav1.NamespaceDefault && eps.Name == MasterEndpointName {
			return eps
		}
		if serviceAutonomyEnhancementEnabled {
			newEps := eps.DeepCopy()
			newEps.Endpoints = filterLocalNodeInfoConcernedAddressesForEndpointSliceV1(nodes, newEps.Endpoints, localNodeInfo)
			klog.V(4).Infof("Normal endpointSlice after LocalNodeInfo filter %s: subnets from %+#v to %+#v", eps.Name, eps.Endpoints, newEps.Endpoints)
			return newEps
		}
		return eps
	}

	// topology endpoints
	newEps := eps.DeepCopy()
	newEps.Endpoints = filterConcernedAddressesForEndpointSliceV1(svc.keys, hostName, nodes, newEps.Endpoints, localNodeInfo, serviceAutonomyEnhancementEnabled)
	klog.V(4).Infof("Topology endpointSlice %s: subnets from %+#v to %+#v", eps.Name, eps.Endpoints, newEps.Endpoints)

	return newEps
}

func pruneEndpointSliceV1Beta1(hostName string,
	nodes map[types.NamespacedName]*nodeContainer,
	services map[types.NamespacedName]*serviceContainer,
	eps *discoveryv1beta1.EndpointSlice, localNodeInfo map[string]data.ResultDetail, wrapperInCluster, serviceAutonomyEnhancementEnabled bool) *discoveryv1beta1.EndpointSlice {

	//drop suffix
	strArr := strings.Split(eps.Name, "-")
	epsName := strArr[0]
	for i := 1; i < len(strArr)-1; i++ {
		epsName = epsName + "-" + strArr[i]
	}
	epsKey := types.NamespacedName{Namespace: eps.Namespace, Name: epsName}

	if wrapperInCluster {
		eps = genLocalEndpointSliceV1Beta1(eps)
	}

	// dangling endpoints
	svc, ok := services[epsKey]
	if !ok {
		klog.V(4).Infof("Dangling endpointSlice %s, %+#v", eps.Name, eps.Endpoints)
		return eps
	}

	// normal service
	if len(svc.keys) == 0 {
		klog.V(4).Infof("Normal endpointSlice %s, %+#v", eps.Name, eps.Endpoints)
		if eps.Namespace == metav1.NamespaceDefault && eps.Name == MasterEndpointName {
			return eps
		}
		if serviceAutonomyEnhancementEnabled {
			newEps := eps.DeepCopy()
			newEps.Endpoints = filterLocalNodeInfoConcernedAddressesForEndpointSliceV1Beta1(nodes, newEps.Endpoints, localNodeInfo)
			klog.V(4).Infof("Normal endpointSlice after LocalNodeInfo filter %s: subnets from %+#v to %+#v", eps.Name, eps.Endpoints, newEps.Endpoints)
			return newEps
		}
		return eps
	}

	// topology endpoints
	newEps := eps.DeepCopy()
	newEps.Endpoints = filterConcernedAddressesForEndpointSliceV1Beta1(svc.keys, hostName, nodes, newEps.Endpoints, localNodeInfo, serviceAutonomyEnhancementEnabled)
	klog.V(4).Infof("Topology endpointSlice %s: subnets from %+#v to %+#v", eps.Name, eps.Endpoints, newEps.Endpoints)

	return newEps
}

// filterConcernedAddresses aims to filter out endpoints addresses within the same node unit
func filterConcernedAddresses(topologyKeys []string, hostName string, nodes map[types.NamespacedName]*nodeContainer,
	addresses []v1.EndpointAddress, localNodeInfo map[string]data.ResultDetail, getLocalNodeInfo bool) []v1.EndpointAddress {
	hostNode, found := nodes[types.NamespacedName{Name: hostName}]
	if !found {
		return nil
	}

	filteredEndpointAddresses := make([]v1.EndpointAddress, 0)
	for i := range addresses {
		addr := addresses[i]
		if nodeName := addr.NodeName; nodeName != nil {
			epsNode, found := nodes[types.NamespacedName{Name: *nodeName}]
			if !found {
				continue
			}
			_, found = localNodeInfo[epsNode.node.Name]
			if hasIntersectionLabel(topologyKeys, hostNode.labels, epsNode.labels) {
				/*
					1.getLocalNodeInfo is enabled, we can find the node from neighbor's nodes status and node is health
					2.getLocalNodeInfo is enabled, but can't find the node from neighbor's nodes status. Failing get status from neighbor will cause this
					3.getLocalNodeInfo is not enabled
				*/
				if (getLocalNodeInfo && found && localNodeInfo[epsNode.node.Name].Normal) ||
					(getLocalNodeInfo && !found) || !getLocalNodeInfo {
					filteredEndpointAddresses = append(filteredEndpointAddresses, addr)
				}
			}
		}
	}

	return filteredEndpointAddresses
}

// filterConcernedAddressesForEndpointSlice aims to filter out endpointSlice addresses within the same node unit
func filterConcernedAddressesForEndpointSliceV1(topologyKeys []string, hostName string, nodes map[types.NamespacedName]*nodeContainer,
	endpoints []discoveryv1.Endpoint, localNodeInfo map[string]data.ResultDetail, getLocalNodeInfo bool) []discoveryv1.Endpoint {
	hostNode, found := nodes[types.NamespacedName{Name: hostName}]
	if !found {
		return nil
	}

	filteredEndpointAddresses := make([]discoveryv1.Endpoint, 0)
	for i := range endpoints {
		endpoint := endpoints[i]
		if endpoint.NodeName != nil {
			epsNode, found := nodes[types.NamespacedName{Name: *endpoint.NodeName}]
			if !found {
				continue
			}
			_, found = localNodeInfo[epsNode.node.Name]
			if hasIntersectionLabel(topologyKeys, hostNode.labels, epsNode.labels) {
				/*
					1.getLocalNodeInfo is enabled, we can find the node from neighbor's nodes status and node is health
					2.getLocalNodeInfo is enabled, but can't find the node from neighbor's nodes status. Failing get status from neighbor will cause this
					3.getLocalNodeInfo is not enabled
				*/
				if (getLocalNodeInfo && found && localNodeInfo[epsNode.node.Name].Normal) ||
					(getLocalNodeInfo && !found) || !getLocalNodeInfo {
					filteredEndpointAddresses = append(filteredEndpointAddresses, endpoint)
				}
			}
		}
	}

	return filteredEndpointAddresses
}

// filterLocalNodeInfoConcernedAddresses aims to filter out endpoints addresses according to LocalNodeInfo
func filterLocalNodeInfoConcernedAddresses(nodes map[types.NamespacedName]*nodeContainer,
	addresses []v1.EndpointAddress, localNodeInfo map[string]data.ResultDetail) []v1.EndpointAddress {

	filteredEndpointAddresses := make([]v1.EndpointAddress, 0)
	for i := range addresses {
		addr := addresses[i]
		if nodeName := addr.NodeName; nodeName != nil {
			epsNode, found := nodes[types.NamespacedName{Name: *nodeName}]
			if !found {
				continue
			}
			_, found = localNodeInfo[epsNode.node.Name]
			if !found || (found && localNodeInfo[epsNode.node.Name].Normal) {
				filteredEndpointAddresses = append(filteredEndpointAddresses, addr)
			}
		}
	}

	return filteredEndpointAddresses
}

// filterLocalNodeInfoConcernedAddressesForEndpointSlice aims to filter out endpointSlice addresses according to LocalNodeInfo
func filterLocalNodeInfoConcernedAddressesForEndpointSliceV1(nodes map[types.NamespacedName]*nodeContainer,
	endpoints []discoveryv1.Endpoint, localNodeInfo map[string]data.ResultDetail) []discoveryv1.Endpoint {

	filteredEndpointAddresses := make([]discoveryv1.Endpoint, 0)
	for i := range endpoints {
		endpoint := endpoints[i]
		if endpoint.NodeName != nil {
			epsNode, found := nodes[types.NamespacedName{Name: *endpoint.NodeName}]
			if !found {
				continue
			}
			_, found = localNodeInfo[epsNode.node.Name]
			if !found || (found && localNodeInfo[epsNode.node.Name].Normal) {
				filteredEndpointAddresses = append(filteredEndpointAddresses, endpoint)
			}
		}
	}

	return filteredEndpointAddresses
}

func filterConcernedAddressesForEndpointSliceV1Beta1(topologyKeys []string, hostName string, nodes map[types.NamespacedName]*nodeContainer,
	endpoints []discoveryv1beta1.Endpoint, localNodeInfo map[string]data.ResultDetail, getLocalNodeInfo bool) []discoveryv1beta1.Endpoint {
	hostNode, found := nodes[types.NamespacedName{Name: hostName}]
	if !found {
		return nil
	}

	filteredEndpointAddresses := make([]discoveryv1beta1.Endpoint, 0)
	for i := range endpoints {
		endpoint := endpoints[i]
		topology := endpoint.Topology
		if nodeName := topology["kubernetes.io/hostname"]; &nodeName != nil {
			epsNode, found := nodes[types.NamespacedName{Name: nodeName}]
			if !found {
				continue
			}
			_, found = localNodeInfo[epsNode.node.Name]
			if hasIntersectionLabel(topologyKeys, hostNode.labels, epsNode.labels) {
				/*
					1.getLocalNodeInfo is enabled, we can find the node from neighbor's nodes status and node is health
					2.getLocalNodeInfo is enabled, but can't find the node from neighbor's nodes status. Failing get status from neighbor will cause this
					3.getLocalNodeInfo is not enabled
				*/
				if (getLocalNodeInfo && found && localNodeInfo[epsNode.node.Name].Normal) ||
					(getLocalNodeInfo && !found) || !getLocalNodeInfo {
					filteredEndpointAddresses = append(filteredEndpointAddresses, endpoint)
				}
			}
		}
	}

	return filteredEndpointAddresses
}

func filterLocalNodeInfoConcernedAddressesForEndpointSliceV1Beta1(nodes map[types.NamespacedName]*nodeContainer,
	endpoints []discoveryv1beta1.Endpoint, localNodeInfo map[string]data.ResultDetail) []discoveryv1beta1.Endpoint {

	filteredEndpointAddresses := make([]discoveryv1beta1.Endpoint, 0)
	for i := range endpoints {
		endpoint := endpoints[i]
		topology := endpoint.Topology
		if nodeName := topology["kubernetes.io/hostname"]; &nodeName != nil {
			epsNode, found := nodes[types.NamespacedName{Name: nodeName}]
			if !found {
				continue
			}
			_, found = localNodeInfo[epsNode.node.Name]
			if !found || (found && localNodeInfo[epsNode.node.Name].Normal) {
				filteredEndpointAddresses = append(filteredEndpointAddresses, endpoint)
			}
		}
	}

	return filteredEndpointAddresses
}

func hasIntersectionLabel(keys []string, n1, n2 map[string]string) bool {
	if n1 == nil || n2 == nil {
		return false
	}

	for _, key := range keys {
		val1, v1found := n1[key]
		val2, v2found := n2[key]

		if v1found && v2found && val1 == val2 {
			return true
		}
	}

	return false
}
