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

package connect

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/superedge/superedge/pkg/tunnel/proxy/common/indexers"
	tunnelutil "github.com/superedge/superedge/pkg/tunnel/util"
	"github.com/superedge/superedge/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog/v2"
)

var Route = &RouteCache{
	EdgeNode:        map[string]string{},
	CloudNode:       map[string]string{},
	ServicesMap:     map[string]string{},
	UserServicesMap: map[string]string{},
}

type RouteCache struct {
	EdgeNode        map[string]string
	CloudNode       map[string]string
	ServicesMap     map[string]string
	UserServicesMap map[string]string
}

func SyncRoute(userClient kubernetes.Interface) {
	id := os.Getenv(tunnelutil.POD_NAME)
	lock, err := resourcelock.New(resourcelock.EndpointsResourceLock,
		os.Getenv(tunnelutil.USER_NAMESPACE_ENV),
		"tunnel-route-cache",
		userClient.CoreV1(),
		userClient.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity: id,
		})
	if err != nil {
		klog.ErrorS(err, "Cache sync failed")
		return
	}

	leaderelection.RunOrDie(context.Background(), leaderelection.LeaderElectionConfig{
		Lock: lock,
		// IMPORTANT: you MUST ensure that any code you have that
		// is protected by the lease must terminate **before**
		// you call cancel. Otherwise, you could have a background
		// loop still running and another process could
		// get elected before your background loop finished, violating
		// the stated goal of the lease.
		ReleaseOnCancel: true,
		LeaseDuration:   60 * time.Second,
		RenewDeadline:   15 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				// we're notified when we start - this is where you would
				// usually put your code
				klog.Info("start syncCache")
				for {
					err = syncCache()
					if err != nil {
						klog.ErrorS(err, "failed to syncCache")
					}
					time.Sleep(10 * time.Second)
				}

			},
			OnStoppedLeading: func() {
				// we can do cleanup here
				klog.InfoS("leader lost", "id", id)
				os.Exit(0)
			},
			OnNewLeader: func(identity string) {
				// we're notified when new leader elected
				if identity == id {
					// I just got the lock
					return
				}
				klog.InfoS("new leader elected", "id", identity)
			},
		},
	})
}

func syncCache() error {
	err := loadCacheFromLocalFile()
	if err != nil {
		return err
	}
	edgeNodeFile, err := os.Open(tunnelutil.EdgeNodesFilePath)
	if err != nil {
		return err
	}
	defer edgeNodeFile.Close()

	updateFlag := false
	// check edge node
	edgeNodes := hosts2Array(edgeNodeFile)
	if len(edgeNodes) != len(Route.EdgeNode) {
		updateFlag = true
	} else {
		for _, v := range edgeNodes {
			if value, ok := Route.EdgeNode[string(v[1])]; ok {
				if value != string(v[0]) {
					updateFlag = true
				}
			} else {
				updateFlag = true
			}
		}
	}

	// check cloud node

	r, err := labels.NewRequirement(util.CloudNodeLabelKey, selection.Equals, []string{"enable"})
	if err != nil {
		return err
	}
	cloudNodes, err := indexers.NodeLister.List(labels.NewSelector().Add(*r))
	if err != nil {
		return err
	}

	// by default, add masters to the cloud node
	ls, err := labels.NewRequirement(util.KubernetesDefaultRoleLabel, selection.Exists, []string{})
	if err != nil {
		return err
	}
	masters, err := indexers.NodeLister.List(labels.NewSelector().Add(*ls))
	if err != nil {
		return err
	}
	if len(masters) != 0 {
		cloudNodes = append(cloudNodes, masters...)
	}

	nodesMap := make(map[string]int)
	for i, n := range cloudNodes {
		nodesMap[n.Name] = i
		var interIp string
		for _, addr := range n.Status.Addresses {
			if addr.Type == "InternalIP" {
				interIp = addr.Address
			}
		}
		if v, ok := Route.CloudNode[n.Name]; ok {
			if v != interIp {
				updateFlag = true
				Route.CloudNode[n.Name] = interIp
			}
		} else {
			updateFlag = true
			Route.CloudNode[n.Name] = interIp
		}
	}

	// 筛选 cache 中 已经写入的node 已经被删除的情况，这时也需要进行更新；例如 cloud 节点被 delete
	for name := range Route.CloudNode {
		if _, ok := nodesMap[name]; ok {
			// 实际 node 存在，什么也不做
		} else {
			// 实际 node 已经删除，需要更新Route.CloudNode
			updateFlag = true
			delete(Route.CloudNode, name)
		}
	}

	// check service
	svcs, err := indexers.ServiceLister.List(labels.Everything())
	if err != nil {
		return err
	}
	svcMaps := make(map[string]int)
	for i, svc := range svcs {
		svcMaps[fmt.Sprintf("%s.%s", svc.Name, svc.Namespace)] = i
		eps, err := indexers.EndpointLister.Endpoints(svc.Namespace).Get(svc.Name)
		if err != nil {
			// klog.Errorf("Failed to get endpoints %s, error:%v", fmt.Sprintf("%s.%s", svc.Name, svc.Namespace), err)
			continue
		}
		if len(eps.Subsets) == 0 {
			continue
		}
		epnodes := []string{}
		for _, ep := range eps.Subsets[0].Addresses {
			// endpoints kubernetes.default subsets has no field NodeName
			if eps.Name == "kubernetes" {
				epnodes = append(epnodes, ep.IP)
				continue
			}
			if ep.NodeName == nil {
				continue
			}
			epnodes = append(epnodes, *ep.NodeName)
		}
		if len(epnodes) == 0 {
			continue
		}
		if v, ok := Route.ServicesMap[fmt.Sprintf("%s.%s", svc.Name, svc.Namespace)]; ok {
			if v == tunnelutil.EDGE {
				edgeFlag := true
				for _, v := range epnodes {
					if _, ok := Route.EdgeNode[v]; ok {
						continue
					}
					edgeFlag = false
					break
				}
				if edgeFlag {
					continue
				}
				delete(Route.ServicesMap, fmt.Sprintf("%s.%s", svc.Name, svc.Namespace))
				updateFlag = true
			} else if v == tunnelutil.CLOUD {
				cloudFlag := true
				for _, v := range epnodes {
					if _, ok := Route.CloudNode[v]; ok {
						continue
					}
					cloudFlag = false
					break
				}
				if cloudFlag {
					continue
				}
				delete(Route.ServicesMap, fmt.Sprintf("%s.%s", svc.Name, svc.Namespace))
				updateFlag = true
			}
		} else {
			edgeFlag := true
			for _, v := range epnodes {
				if _, ok := Route.EdgeNode[v]; ok {
					continue
				}
				edgeFlag = false
				break
			}

			if edgeFlag {
				updateFlag = true
				Route.ServicesMap[fmt.Sprintf("%s.%s", svc.Name, svc.Namespace)] = tunnelutil.EDGE
			} else {
				cloudFlag := true
				for _, v := range epnodes {
					if _, ok := Route.CloudNode[v]; ok {
						continue
					}
					cloudFlag = false
					break
				}
				if cloudFlag {
					updateFlag = true
					Route.ServicesMap[fmt.Sprintf("%s.%s", svc.Name, svc.Namespace)] = tunnelutil.CLOUD
				}
			}
		}
	}

	for svc := range Route.ServicesMap {
		if _, ok := svcMaps[svc]; !ok {
			// configmap中的 services 存在，实际 svc 已经被删除，需要删除 Route.ServicesMap
			updateFlag = true
			delete(Route.ServicesMap, svc)
		}
	}

	if updateFlag {
		cfg, err := register.ClientSet.CoreV1().ConfigMaps(os.Getenv(tunnelutil.POD_NAMESPACE_ENV)).Get(context.Background(), tunnelutil.CacheConfig, metav1.GetOptions{})
		if err != nil {
			return err
		}
		edgeNodeBuffer := &bytes.Buffer{}
		cloudNodeBuffer := &bytes.Buffer{}
		serviceBuffer := &bytes.Buffer{}
		for k, v := range Route.EdgeNode {
			edgeNodeBuffer.WriteString(v)
			edgeNodeBuffer.WriteString("    ")
			edgeNodeBuffer.WriteString(k)
			edgeNodeBuffer.WriteString("\n")
		}
		for k, v := range Route.CloudNode {
			cloudNodeBuffer.WriteString(v)
			cloudNodeBuffer.WriteString("    ")
			cloudNodeBuffer.WriteString(k)
			cloudNodeBuffer.WriteString("\n")
		}
		for k, v := range Route.ServicesMap {
			serviceBuffer.WriteString(k)
			serviceBuffer.WriteString("    ")
			serviceBuffer.WriteString(v)
			serviceBuffer.WriteString("\n")
		}
		cfg.Data[tunnelutil.EdgeNodesFile] = edgeNodeBuffer.String()
		cfg.Data[tunnelutil.CloudNodesFile] = cloudNodeBuffer.String()
		cfg.Data[tunnelutil.ServicesFile] = serviceBuffer.String()
		_, err = register.ClientSet.CoreV1().ConfigMaps(os.Getenv(tunnelutil.POD_NAMESPACE_ENV)).Update(context.Background(), cfg, metav1.UpdateOptions{})
		if err != nil {
			klog.ErrorS(err, "failed to update services of route cache")
			return err
		}

	}
	return nil
}

func loadCacheFromLocalFile() error {
	hosts, err := os.Open(tunnelutil.HostsPath)
	if err != nil {
		return err
	}
	defer hosts.Close()

	cloudNodeFile, err := os.Open(tunnelutil.CloudNodesFilePath)
	if err != nil {
		return err
	}
	defer cloudNodeFile.Close()

	servicesFile, err := os.Open(tunnelutil.ServicesFilePath)
	if err != nil {
		return err
	}
	defer servicesFile.Close()

	userServiceFile, err := os.Open(tunnelutil.UserServiceFilepath)
	if err != nil {
		return err
	}
	defer userServiceFile.Close()

	for _, v := range hosts2Array(hosts) {
		Route.EdgeNode[string(v[1])] = string(v[0])
	}

	for _, v := range hosts2Array(cloudNodeFile) {
		Route.CloudNode[string(v[1])] = string(v[0])
	}

	for _, v := range service2Array(servicesFile) {
		Route.ServicesMap[string(v[0])] = string(v[1])
	}

	for _, v := range service2Array(userServiceFile) {
		Route.UserServicesMap[string(v[0])] = string(v[1])
	}
	return nil
}

func service2Array(fileread io.Reader) [][][]byte {
	scanner := bufio.NewScanner(fileread)
	hostsArray := [][][]byte{}
	for scanner.Scan() {
		// copy byte slice before append to hostsArray
		f := bytes.Fields([]byte(scanner.Text()))
		if len(f) == 2 {
			hostsArray = append(hostsArray, f)
		}
	}
	return hostsArray
}
