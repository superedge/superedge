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
	"net"
	"strconv"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
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

func pruneEndpoints(hostName string,
	nodes map[types.NamespacedName]*nodeContainer,
	services map[types.NamespacedName]*serviceContainer,
	eps *v1.Endpoints, wrapperInCluster bool) *v1.Endpoints {

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
		return eps
	}

	// topology endpoints
	newEps := eps.DeepCopy()
	for si := range newEps.Subsets {
		subnet := &newEps.Subsets[si]
		subnet.Addresses = filterConcernedAddress(svc.keys, hostName, nodes, subnet.Addresses)
		subnet.NotReadyAddresses = filterConcernedAddress(svc.keys, hostName, nodes, subnet.NotReadyAddresses)
	}
	klog.V(4).Infof("Topology endpoints %s: subnets from %+#v to %+#v", eps.Name, eps.Subsets, newEps.Subsets)

	return newEps
}

func filterConcernedAddress(topologyKeys []string, hostName string, nodes map[types.NamespacedName]*nodeContainer,
	address []v1.EndpointAddress) []v1.EndpointAddress {
	hostNode, found := nodes[types.NamespacedName{Name: hostName}]
	if !found {
		return nil
	}

	filtered := make([]v1.EndpointAddress, 0)
	for i := range address {
		addr := address[i]
		if addr.NodeName != nil {
			nodeName := *addr.NodeName
			epsNode, found := nodes[types.NamespacedName{Name: nodeName}]
			if !found {
				continue
			}
			if hasIntersectionLabel(topologyKeys, hostNode.labels, epsNode.labels) {
				filtered = append(filtered, addr)
			}
		}
	}

	return filtered
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
