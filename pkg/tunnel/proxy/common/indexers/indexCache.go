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

package indexers

import (
	"fmt"
	"github.com/superedge/superedge/pkg/tunnel/conf"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"github.com/superedge/superedge/pkg/util/kubeclient"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sync"
	"time"
)

var (
	podIndexer, nodeIndexer cache.Indexer
	once                    sync.Once
)

//Index pods by podIp
func PodIPKeyFunc(obj interface{}) ([]string, error) {
	pod, ok := obj.(*v1.Pod)
	if ok {
		return []string{pod.Status.PodIP}, nil
	}
	return []string{}, nil
}

func MetaNameIndexFunc(obj interface{}) ([]string, error) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		return []string{}, err
	}
	return []string{key}, nil
}

func InitCache(stopCh chan struct{}) {
	once.Do(func() {
		clientset, err := kubeclient.GetInclusterClientSet(conf.TunnelConf.TunnlMode.Cloud.Egress.KubeConfig)
		if err != nil {
			klog.Errorf("Failed to get kubeclient, error: %v", err)
			return
		}

		//Initialize podIndexer
		podListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", v1.NamespaceAll, fields.Everything())
		indexerPod, informerPod := cache.NewIndexerInformer(podListWatcher, &v1.Pod{}, 1*time.Minute, cache.ResourceEventHandlerFuncs{}, cache.Indexers{util.PODIP_INDEXER: PodIPKeyFunc})
		go informerPod.Run(stopCh)
		// Wait for all involved caches to be synced, before processing items from the queue is started
		if !cache.WaitForCacheSync(stopCh, informerPod.HasSynced) {
			runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
			return
		}

		podIndexer = indexerPod

		//initialize nodeIndexer
		nodeListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "nodes", v1.NamespaceAll, fields.Everything())
		indexerNode, informerNode := cache.NewIndexerInformer(nodeListWatcher, &v1.Node{}, 1*time.Minute, cache.ResourceEventHandlerFuncs{}, cache.Indexers{util.METANAME_INDEXER: MetaNameIndexFunc})
		go informerNode.Run(stopCh)
		// Wait for all involved caches to be synced, before processing items from the queue is started
		if !cache.WaitForCacheSync(stopCh, informerNode.HasSynced) {
			runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
			return
		}

		nodeIndexer = indexerNode
	})

}

//Get the name of the node where the pod is located based on podIp
func GetNodeByPodIP(podIp string) (string, error) {
	if podIndexer == nil {
		return "", fmt.Errorf("podIndexer is not initialized")
	}
	pods, err := podIndexer.ByIndex(util.PODIP_INDEXER, podIp)
	if err != nil {
		return "", err
	}

	if len(pods) < 1 {
		return "", fmt.Errorf("Failed to get pods by PodIP, PodIP: %s", podIp)
	}
	return pods[0].(*v1.Pod).Spec.NodeName, nil
}

//Get the internalIp of the node based on the node name
func GetNodeIPByName(name string) (string, error) {
	if nodeIndexer == nil {
		return "", fmt.Errorf("nodeIndexer is not initialized")
	}
	nodes, err := nodeIndexer.ByIndex(util.METANAME_INDEXER, name)
	if err != nil {
		return "", err
	}
	if len(nodes) < 1 {
		return "", fmt.Errorf("Failed to get nodes by NodeName, nodeName: %s", name)
	}
	for _, addr := range nodes[0].(*v1.Node).Status.Addresses {
		if addr.Type == "InternalIP" {
			return addr.Address, nil
		}
	}
	return "", fmt.Errorf("Failed to get internalIp from node.status.Addresses")
}
