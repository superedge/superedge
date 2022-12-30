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

package check

import (
	"context"
	"fmt"
	"time"

	"github.com/superedge/superedge/pkg/edge-health/common"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/client-go/informers"
	informcorev1 "k8s.io/client-go/informers/core/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

const (
	EdgeHealthPodLabelKey   = "name"
	EdgeHealthPodLabelValue = "edge-health"
	PodIPIndexKey           = "PodIP"
	NodeIPIndexKey          = "NodeIP"
	NodeNameIndexKey        = "NodeName"
)

var PodManager *PodController

func NewPodController(clientset kubernetes.Interface) *PodController {
	labelSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{EdgeHealthPodLabelKey: EdgeHealthPodLabelValue},
	}
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		panic("edge-health pod label selector error")
	}
	// Make informers that filter out objects that want a non-default service proxy.

	informerFactory := informers.NewSharedInformerFactory(clientset, 10*time.Minute)

	//Initialize podIndexer
	podInformer := informerFactory.InformerFor(&v1.Pod{}, func(k kubernetes.Interface, duration time.Duration) cache.SharedIndexInformer {
		return informcorev1.NewFilteredPodInformer(k, common.Namespace, duration,
			cache.Indexers{
				PodIPIndexKey:    PodIPKeyFunc,
				NodeIPIndexKey:   NodeIPKeyFunc,
				NodeNameIndexKey: NodeNameKeyFunc,
			},
			func(options *metav1.ListOptions) {
				// only watch edge-health pod and it's status Running
				options.LabelSelector = selector.String()
				options.FieldSelector = fields.OneTermEqualSelector("status.phase", string(v1.PodRunning)).String()
			},
		)
	})
	n := &PodController{}
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: n.handlePodAddUpdate,
		UpdateFunc: func(old, cur interface{}) {
			n.handlePodAddUpdate(cur)
		},
		DeleteFunc: n.handlePodDelete,
	})
	n.podStore = podInformer.GetIndexer()
	n.clientset = clientset
	n.PodInformer = podInformer
	n.PodListerSynced = podInformer.HasSynced
	PodManager = n
	return n
}

type PodController struct {
	podStore        cache.Indexer
	clientset       kubernetes.Interface
	PodInformer     cache.SharedIndexInformer
	PodListerSynced cache.InformerSynced
}

func (n *PodController) handlePodAddUpdate(obj interface{}) {
	node, ok := obj.(*v1.Pod)
	if !ok {
		return
	}
	klog.V(4).Infof("Add/Update Node %s", node.Name)
}

func (n *PodController) handlePodDelete(obj interface{}) {
	node, ok := obj.(*v1.Pod)
	if !ok {
		return
	}
	klog.V(4).Infof("Delete Node %s", node.Name)
}

func (n *PodController) Run(ctx context.Context) {
	defer runtimeutil.HandleCrash()

	go n.PodInformer.Run(ctx.Done())

	if ok := cache.WaitForCacheSync(
		ctx.Done(),
		n.PodListerSynced,
	); !ok {
		klog.Fatal("failed to wait for caches to sync")
	}

	<-ctx.Done()
}

func (n *PodController) GetPodIPByNodeIP(nodeIP string) (string, error) {
	if n == nil || n.podStore == nil {
		return "", fmt.Errorf("podStore is not initialized")
	}
	pods, err := n.podStore.ByIndex(NodeIPIndexKey, nodeIP)
	if err != nil {
		return "", err
	}

	if len(pods) < 1 {
		return "", fmt.Errorf("failed to get pods by NodeIP %s", nodeIP)
	} else if len(pods) > 1 {
		klog.V(4).InfoS("GetPodIPByNodeIP return multi pods", "pod length", len(pods))
		return "", fmt.Errorf("multi pods return by NodeIP %s", nodeIP)
	}
	return pods[0].(*v1.Pod).Status.PodIP, nil
}

func (n *PodController) GetNodeIPByNodeName(nodeName string) (string, error) {
	if n == nil || n.podStore == nil {
		return "", fmt.Errorf("podStore is not initialized")
	}

	pods, err := n.podStore.ByIndex(NodeNameIndexKey, nodeName)
	if err != nil {
		return "", err
	}

	if len(pods) < 1 {
		return "", fmt.Errorf("failed to get pods by nodeName %s", nodeName)
	} else if len(pods) > 1 {
		klog.V(4).InfoS("GetNodeIPByNodeName return multi pods", "pod length", len(pods))
		return "", fmt.Errorf("multi pods return by nodeName %s", nodeName)
	}
	return pods[0].(*v1.Pod).Status.HostIP, nil
}
func (n *PodController) GetNodeNameByNodeIP(nodeIP string) (string, error) {
	if n == nil || n.podStore == nil {
		return "", fmt.Errorf("podStore is not initialized")
	}

	pods, err := n.podStore.ByIndex(NodeIPIndexKey, nodeIP)
	if err != nil {
		return "", err
	}

	if len(pods) < 1 {
		return "", fmt.Errorf("failed to get pods by nodeIP %s", nodeIP)
	} else if len(pods) > 1 {
		klog.V(4).InfoS("GetNodeNameByNodeIP return multi pods", "pod length", len(pods))
		return "", fmt.Errorf("multi pods return by nodeIP %s", nodeIP)
	}
	return pods[0].(*v1.Pod).Spec.NodeName, nil
}

func PodIPKeyFunc(obj interface{}) ([]string, error) {
	pod, ok := obj.(*v1.Pod)
	if ok {
		return []string{pod.Status.PodIP}, nil
	}
	return []string{}, nil
}

// TODO: node internal IP in superedge will same in two node in a cluster
func NodeIPKeyFunc(obj interface{}) ([]string, error) {
	pod, ok := obj.(*v1.Pod)
	if ok {
		return []string{pod.Status.HostIP}, nil
	}
	return []string{}, nil
}

func NodeNameKeyFunc(obj interface{}) ([]string, error) {
	pod, ok := obj.(*v1.Pod)
	if ok {
		return []string{pod.Spec.NodeName}, nil
	}
	return []string{}, nil
}
