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

package controller

import (
	"fmt"
	controllercommon "github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	crdinformers "github.com/superedge/superedge/pkg/application-grid-controller/generated/informers/externalversions/superedge.io/v1"
	crdv1listers "github.com/superedge/superedge/pkg/application-grid-controller/generated/listers/superedge.io/v1"
	"github.com/superedge/superedge/pkg/statefulset-grid-daemon/common"
	"github.com/superedge/superedge/pkg/statefulset-grid-daemon/hosts"
	"github.com/superedge/superedge/pkg/statefulset-grid-daemon/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"strings"
	"time"
)

var (
	KeyFunc        = cache.DeletionHandlingMetaNamespaceKeyFunc
	controllerKind = appsv1.SchemeGroupVersion.WithKind("StatefulSet")
)

type StatefulSetGridDaemonController struct {
	hostName string
	hosts    *hosts.Hosts

	setGridLister       crdv1listers.StatefulSetGridLister
	setGridListerSynced cache.InformerSynced

	nodeLister       corelisters.NodeLister
	nodeListerSynced cache.InformerSynced

	podLister       corelisters.PodLister
	podListerSynced cache.InformerSynced

	setLister       appslisters.StatefulSetLister
	setListerSynced cache.InformerSynced

	svcLister       corelisters.ServiceLister
	svcListerSynced cache.InformerSynced

	eventRecorder record.EventRecorder
	queue         workqueue.RateLimitingInterface
	kubeClient    clientset.Interface

	syncHandler        func(dKey string) error
	enqueueStatefulSet func(set *appsv1.StatefulSet)
}

func NewStatefulSetGridDaemonController(nodeInformer coreinformers.NodeInformer, podInformer coreinformers.PodInformer,
	setInformer appsinformers.StatefulSetInformer, setGridInformer crdinformers.StatefulSetGridInformer,
	svcInformer coreinformers.ServiceInformer, kubeClient clientset.Interface,
	hostName string, hosts *hosts.Hosts) *StatefulSetGridDaemonController {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1.EventSinkImpl{
		Interface: kubeClient.CoreV1().Events(""),
	})

	ssgdc := &StatefulSetGridDaemonController{
		hostName:   hostName,
		hosts:      hosts,
		kubeClient: kubeClient,
		eventRecorder: eventBroadcaster.NewRecorder(scheme.Scheme,
			corev1.EventSource{Component: "statefulset-grid-daemon"}),
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(),
			"statefulset-grid-daemon"),
	}

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ssgdc.addNode,
		UpdateFunc: ssgdc.updateNode,
		DeleteFunc: ssgdc.deleteNode,
	})

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ssgdc.addPod,
		UpdateFunc: ssgdc.updatePod,
		DeleteFunc: ssgdc.deletePod,
	})

	setInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ssgdc.addStatefulSet,
		UpdateFunc: ssgdc.updateStatefulSet,
		DeleteFunc: ssgdc.deleteStatefulSet,
	})

	svcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ssgdc.addService,
		UpdateFunc: ssgdc.updateService,
		DeleteFunc: ssgdc.deleteService,
	})

	ssgdc.syncHandler = ssgdc.syncDnsHosts
	ssgdc.enqueueStatefulSet = ssgdc.enqueue

	ssgdc.nodeLister = nodeInformer.Lister()
	ssgdc.nodeListerSynced = nodeInformer.Informer().HasSynced

	ssgdc.podLister = podInformer.Lister()
	ssgdc.podListerSynced = podInformer.Informer().HasSynced

	ssgdc.setLister = setInformer.Lister()
	ssgdc.setListerSynced = setInformer.Informer().HasSynced

	ssgdc.setGridLister = setGridInformer.Lister()
	ssgdc.setGridListerSynced = setGridInformer.Informer().HasSynced

	ssgdc.svcLister = svcInformer.Lister()
	ssgdc.svcListerSynced = svcInformer.Informer().HasSynced

	return ssgdc
}

func (ssgdc *StatefulSetGridDaemonController) Run(workers, syncPeriodAsWhole int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer ssgdc.queue.ShutDown()

	klog.Infof("Starting statefulset grid daemon")
	defer klog.Infof("Shutting down statefulset grid daemon")

	if !cache.WaitForNamedCacheSync("statefulset-grid-daemon", stopCh,
		ssgdc.nodeListerSynced, ssgdc.podListerSynced, ssgdc.setListerSynced, ssgdc.setGridListerSynced, ssgdc.svcListerSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(ssgdc.worker, time.Second, stopCh)
	}

	// sync dns hosts as a whole
	go wait.Until(ssgdc.syncDnsHostsAsWhole, time.Duration(syncPeriodAsWhole)*time.Second, stopCh)
	<-stopCh
}

func (ssgdc *StatefulSetGridDaemonController) worker() {
	for ssgdc.processNextWorkItem() {
	}
}

func (ssgdc *StatefulSetGridDaemonController) processNextWorkItem() bool {
	key, quit := ssgdc.queue.Get()
	if quit {
		return false
	}
	defer ssgdc.queue.Done(key)

	err := ssgdc.syncHandler(key.(string))
	ssgdc.handleErr(err, key)

	return true
}

func (ssgdc *StatefulSetGridDaemonController) handleErr(err error, key interface{}) {
	if err == nil {
		ssgdc.queue.Forget(key)
		return
	}

	if ssgdc.queue.NumRequeues(key) < common.MaxRetries {
		klog.V(2).Infof("Error syncing statefulset %v: %v", key, err)
		ssgdc.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	klog.V(2).Infof("Dropping statefulset %q out of the queue: %v", key, err)
	ssgdc.queue.Forget(key)
}

func (ssgdc *StatefulSetGridDaemonController) needClearStatefulSetDomains(set *appsv1.StatefulSet) (bool, error) {
	// Check existence of statefulset relevant service
	svc, err := ssgdc.svcLister.Services(set.Namespace).Get(set.Spec.ServiceName)
	if errors.IsNotFound(err) {
		klog.V(2).Infof("StatefulSet %v relevant service %s not found", set.Name, set.Spec.ServiceName)
		return true, nil
	}
	if err != nil {
		return false, err
	}
	// Check GridSelectorUniqKeyName label value equation between service and statefulset
	gridUniqKey, _ := set.Labels[controllercommon.GridSelectorUniqKeyName]
	svcGridUniqKey, found := svc.Labels[controllercommon.GridSelectorUniqKeyName]
	if !found {
		return true, nil
	} else if gridUniqKey != svcGridUniqKey {
		return true, nil
	}
	return false, nil
}

func (ssgdc *StatefulSetGridDaemonController) syncDnsHostsAsWhole() {
	startTime := time.Now()
	klog.V(4).Infof("Started syncing dns hosts as a whole (%v)", startTime)
	defer func() {
		klog.V(4).Infof("Finished syncing dns hosts as a whole (%v)", time.Since(startTime))
	}()
	// Get node relevant GridSelectorUniqKeyName labels
	node, err := ssgdc.nodeLister.Get(ssgdc.hostName)
	if err != nil {
		klog.Errorf("Get host node %s err %v", ssgdc.hostName, err)
		return
	}
	gridUniqKeyLabels, err := controllercommon.GetNodesSelector(node)
	if err != nil {
		klog.Errorf("Get node %s GridSelectorUniqKeyName selector err %v", node.Name, err)
		return
	}
	// List all statefulsets by node labels
	setList, err := ssgdc.setLister.List(gridUniqKeyLabels)
	if err != nil {
		klog.Errorf("List statefulsets by labels %v err %v", gridUniqKeyLabels, err)
		return
	}
	hostsMap := make(map[string]string)
	// Filter concerned statefulsets and construct dns hosts
	for _, set := range setList {
		if rel, err := ssgdc.IsConcernedStatefulSet(set); err != nil || !rel {
			continue
		}
		if needClear, err := ssgdc.needClearStatefulSetDomains(set); err != nil || needClear {
			continue
		}
		// Get pod list of this statefulset
		podList, err := ssgdc.podLister.Pods(set.Namespace).List(labels.Everything())
		if err != nil {
			klog.Errorf("Get podList err %v", err)
			return
		}
		ControllerRef := metav1.GetControllerOf(set)
		gridValue := set.Name[len(ControllerRef.Name)+1:]
		for _, pod := range podList {
			if util.IsMemberOf(set, pod) {
				index := strings.Index(pod.Name, gridValue)
				if index == -1 {
					klog.Errorf("Invalid pod name %s(statefulset %s)", pod.Name, set.Name)
					continue
				}
				podDomainsToHosts := pod.Name[0:index] + pod.Name[index+len(gridValue)+1:] + "." + set.Spec.ServiceName
				if pod.Status.PodIP == "" {
					klog.V(2).Infof("There is currently no ip for pod %s(statefulset %s)", pod.Name, set.Name)
					continue
				}
				hostsMap[hosts.AppendDomainSuffix(podDomainsToHosts, pod.Namespace)] = pod.Status.PodIP
			}
		}
	}
	// Set dns hosts as a whole
	if err := ssgdc.hosts.SetHostsByMap(hostsMap); err != nil {
		klog.Errorf("SetHostsByMap err %v", err)
	}
	return
}

func (ssgdc *StatefulSetGridDaemonController) syncDnsHosts(key string) error {
	startTime := time.Now()
	klog.V(4).Infof("Started syncing dns hosts of statefulset %q (%v)", key, startTime)
	defer func() {
		klog.V(4).Infof("Finished syncing dns hosts of statefulset %q (%v)", key, time.Since(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	set, err := ssgdc.setLister.StatefulSets(namespace).Get(name)
	if errors.IsNotFound(err) {
		klog.V(2).Infof("StatefulSet %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	var PodDomainInfoToHosts = make(map[string]string)
	ControllerRef := metav1.GetControllerOf(set)
	// Check existence of statefulset relevant service and execute delete operations if necessary
	if needClear, err := ssgdc.needClearStatefulSetDomains(set); err != nil {
		return err
	} else if needClear {
		if err := ssgdc.hosts.CheckOrUpdateHosts(PodDomainInfoToHosts, set.Namespace, ControllerRef.Name, set.Spec.ServiceName); err != nil {
			klog.Errorf("Clear statefulset %v dns hosts err %v", key, err)
			return err
		}
		klog.V(4).Infof("Clear statefulset %v dns hosts successfully", key)
		return nil
	}

	// Get pod list of this statefulset
	podList, err := ssgdc.podLister.Pods(set.Namespace).List(labels.Everything())
	if err != nil {
		klog.Errorf("Get podList err %v", err)
		return err
	}

	podToHosts := []*corev1.Pod{}
	for _, pod := range podList {
		if util.IsMemberOf(set, pod) {
			podToHosts = append(podToHosts, pod)
		}
	}
	// Sync dns hosts partly
	// Attention: this sync can not guarantee the absolute correctness of statefulset grid dns hosts records,
	// and should be used combined with syncDnsHostsAsWhole to ensure the eventual consistency
	// Actual statefulset pod FQDN: <controllerRef>-<gridValue>-<ordinal>.<svc>.<ns>.svc.cluster.local
	// (eg: statefulsetgrid-demo-nodeunit1-0.servicegrid-demo-svc.default.svc.cluster.local)
	// Converted statefulset pod FQDN: <controllerRef>-<ordinal>.<svc>.<ns>.svc.cluster.local
	// (eg: statefulsetgrid-demo-0.servicegrid-demo-svc.default.svc.cluster.local)
	if ControllerRef != nil {
		gridValue := set.Name[len(ControllerRef.Name)+1:]
		for _, pod := range podToHosts {
			index := strings.Index(pod.Name, gridValue)
			if index == -1 {
				klog.Errorf("Invalid pod name %s(statefulset %s)", pod.Name, set.Name)
				continue
			}
			podDomainsToHosts := pod.Name[0:index] + pod.Name[index+len(gridValue)+1:] + "." + set.Spec.ServiceName
			if pod.Status.PodIP == "" {
				klog.V(2).Infof("There is currently no ip for pod %s(statefulset %s)", pod.Name, set.Name)
				continue
			}
			PodDomainInfoToHosts[hosts.AppendDomainSuffix(podDomainsToHosts, pod.Namespace)] = pod.Status.PodIP
		}
		if err := ssgdc.hosts.CheckOrUpdateHosts(PodDomainInfoToHosts, set.Namespace, ControllerRef.Name, set.Spec.ServiceName); err != nil {
			klog.Errorf("update dns hosts err %v", err)
			return err
		}
	}
	return nil
}

func (ssgdc *StatefulSetGridDaemonController) enqueue(set *appsv1.StatefulSet) {
	key, err := KeyFunc(set)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", set, err))
		return
	}

	ssgdc.queue.Add(key)
}
