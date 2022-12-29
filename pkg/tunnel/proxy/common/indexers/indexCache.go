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
	"github.com/superedge/superedge/pkg/tunnel/util"
	"github.com/superedge/superedge/pkg/util/kubeclient"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	informcorev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sync"
	"time"
)

var (
	podIndexer, nodeIndexer, serviceIndexer, endpointIndexer cache.Indexer
	once                                                     sync.Once
	ServiceLister                                            listersv1.ServiceLister
	EndpointLister                                           listersv1.EndpointsLister
	NodeLister                                               listersv1.NodeLister
)

//Index pods by podIp
func PodIPKeyFunc(obj interface{}) ([]string, error) {
	pod, ok := obj.(*v1.Pod)
	if ok {
		return []string{pod.Status.PodIP}, nil
	}
	return []string{}, nil
}

func ServiceIPKeyFunc(obj interface{}) ([]string, error) {
	svc, ok := obj.(*v1.Service)
	if ok {
		return []string{svc.Spec.ClusterIP}, nil
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

func InitCache(path string, stopCh chan struct{}) {
	once.Do(func() {
		clientset, err := kubeclient.GetInclusterClientSet(path)
		if err != nil {
			klog.Errorf("Failed to get kubeclient, error: %v", err)
			return
		}

		informerFactory := informers.NewSharedInformerFactory(clientset, 1*time.Minute)

		//Initialize podIndexer
		podInformer := informerFactory.InformerFor(&v1.Pod{}, func(k kubernetes.Interface, duration time.Duration) cache.SharedIndexInformer {
			return informcorev1.NewPodInformer(k, "", duration, cache.Indexers{util.PODIP_INDEXER: PodIPKeyFunc})
		})
		go podInformer.Run(stopCh)
		// Wait for all involved caches to be synced, before processing items from the queue is started
		if !cache.WaitForCacheSync(stopCh, podInformer.HasSynced) {
			runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
			return
		}
		podIndexer = podInformer.GetIndexer()

		//initialize nodeIndexer
		nodeInformer := informerFactory.InformerFor(&v1.Node{}, func(k kubernetes.Interface, duration time.Duration) cache.SharedIndexInformer {
			return informcorev1.NewNodeInformer(k, duration, cache.Indexers{util.METANAME_INDEXER: MetaNameIndexFunc})
		})
		go nodeInformer.Run(stopCh)
		if !cache.WaitForCacheSync(stopCh, nodeInformer.HasSynced) {
			runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
			return
		}
		nodeIndexer = nodeInformer.GetIndexer()
		NodeLister = listersv1.NewNodeLister(nodeIndexer)

		//initialize serviceIndexer、serviceLister
		serviceInform := informerFactory.InformerFor(&v1.Service{}, func(k kubernetes.Interface, duration time.Duration) cache.SharedIndexInformer {
			return informcorev1.NewServiceInformer(k, "", duration, cache.Indexers{util.SERVICEIP_INDEXER: ServiceIPKeyFunc})
		})
		go serviceInform.Run(stopCh)
		if !cache.WaitForCacheSync(stopCh, serviceInform.HasSynced) {
			runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
			return
		}
		serviceIndexer = serviceInform.GetIndexer()
		ServiceLister = listersv1.NewServiceLister(serviceIndexer)

		//initialize endpointIndexer、endpointLister
		endpointInform := informerFactory.Core().V1().Endpoints().Informer()
		go endpointInform.Run(stopCh)
		if !cache.WaitForCacheSync(stopCh, endpointInform.HasSynced) {
			runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
			return
		}
		endpointIndexer = endpointInform.GetIndexer()
		EndpointLister = listersv1.NewEndpointsLister(endpointIndexer)
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
		return "", fmt.Errorf("Failed to get pods by PodIP %s", podIp)
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
		return "", apierrors.NewNotFound(schema.GroupResource{}, name)
	}
	for _, addr := range nodes[0].(*v1.Node).Status.Addresses {
		if addr.Type == "InternalIP" {
			return addr.Address, nil
		}
	}
	return "", fmt.Errorf("Failed to get internalIp from node.status.Addresses")
}

func GetServiceByClusterIP(clusterIp string) (*v1.Service, error) {
	if serviceIndexer == nil {
		return nil, fmt.Errorf("serviceIndexer is not initialized")
	}
	svcs, err := serviceIndexer.ByIndex(util.SERVICEIP_INDEXER, clusterIp)
	if err != nil {
		return nil, err
	}

	if len(svcs) < 1 {
		return nil, apierrors.NewNotFound(schema.GroupResource{}, clusterIp)
	}
	return svcs[0].(*v1.Service), nil
}
