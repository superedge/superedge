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
	sitev1 "github.com/superedge/superedge/pkg/site-manager/apis/site/v1"
	crdClientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	crdinformers "github.com/superedge/superedge/pkg/site-manager/generated/informers/externalversions/site/v1"
	crdv1listers "github.com/superedge/superedge/pkg/site-manager/generated/listers/site/v1"
	"github.com/superedge/superedge/pkg/statefulset-grid-daemon/common"
	"github.com/superedge/superedge/pkg/statefulset-grid-daemon/hosts"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"time"
)

var (
	KeyFunc        = cache.DeletionHandlingMetaNamespaceKeyFunc
	controllerKind = appsv1.SchemeGroupVersion.WithKind("StatefulSet")
)

type SitesManagerDaemonController struct {
	hostName string
	hosts    *hosts.Hosts

	podLister       corelisters.PodLister
	podListerSynced cache.InformerSynced

	nodeLister       corelisters.NodeLister
	nodeListerSynced cache.InformerSynced

	serviceLister       corelisters.ServiceLister
	serviceListerSynced cache.InformerSynced

	nodeUnitLister       crdv1listers.NodeUnitLister
	nodeUnitListerSynced cache.InformerSynced

	nodeGroupLister       crdv1listers.NodeGroupLister
	nodeGroupListerSynced cache.InformerSynced

	eventRecorder record.EventRecorder
	queue         workqueue.RateLimitingInterface
	kubeClient    clientset.Interface
	crdClient     *crdClientset.Clientset

	syncHandler     func(dKey string) error
	enqueueNodeUnit func(set *sitev1.NodeUnit)
}

func NewSitesManagerDaemonController(
	nodeInformer coreinformers.NodeInformer,
	podInformer coreinformers.PodInformer,
	nodeUnitInformer crdinformers.NodeUnitInformer,
	nodeGroupInformer crdinformers.NodeGroupInformer,
	serviceInformer coreinformers.ServiceInformer,
	kubeClient clientset.Interface,
	crdClient *crdClientset.Clientset,
	hostName string, hosts *hosts.Hosts) *SitesManagerDaemonController {

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1.EventSinkImpl{
		Interface: kubeClient.CoreV1().Events(""),
	})

	siteController := &SitesManagerDaemonController{
		hostName:      hostName,
		hosts:         hosts,
		kubeClient:    kubeClient,
		crdClient:     crdClient,
		eventRecorder: eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "site-manager-daemon"}),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "site-manager-daemon"),
	}

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    siteController.addNode,
		UpdateFunc: siteController.updateNode,
		DeleteFunc: siteController.deleteNode,
	})

	//podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
	//	AddFunc:    siteManager.addPod,
	//	UpdateFunc: siteManager.updatePod,
	//	DeleteFunc: siteManager.deletePod,
	//})

	nodeUnitInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    siteController.addNodeUnit,
		UpdateFunc: siteController.updateNodeUnit,
		DeleteFunc: siteController.deleteNodeUnit,
	})

	//serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
	//	AddFunc:    siteManager.addService,
	//	UpdateFunc: siteManager.updateService,
	//	DeleteFunc: siteManager.deleteService,
	//})

	//siteController.syncHandler = siteController.syncDnsHosts
	siteController.enqueueNodeUnit = siteController.enqueue

	siteController.podLister = podInformer.Lister()
	siteController.podListerSynced = podInformer.Informer().HasSynced

	siteController.nodeLister = nodeInformer.Lister()
	siteController.nodeListerSynced = nodeInformer.Informer().HasSynced

	siteController.serviceLister = serviceInformer.Lister()
	siteController.serviceListerSynced = serviceInformer.Informer().HasSynced

	siteController.nodeUnitLister = nodeUnitInformer.Lister()
	siteController.nodeUnitListerSynced = nodeUnitInformer.Informer().HasSynced

	siteController.nodeGroupLister = nodeGroupInformer.Lister()
	siteController.nodeGroupListerSynced = nodeGroupInformer.Informer().HasSynced

	klog.V(4).Infof("Site-manager set handler success")

	return siteController
}

func (siteManager *SitesManagerDaemonController) Run(workers, syncPeriodAsWhole int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer siteManager.queue.ShutDown()

	klog.V(1).Infof("Starting site-manager daemon")
	defer klog.V(1).Infof("Shutting down site-manager daemon")

	if !cache.WaitForNamedCacheSync("site-manager-daemon", stopCh,
		siteManager.nodeListerSynced, siteManager.podListerSynced,
		siteManager.nodeUnitListerSynced, siteManager.nodeGroupListerSynced, siteManager.serviceListerSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(siteManager.worker, time.Second, stopCh)
		klog.V(4).Infof("Site-manager set worker-%d success", i)
	}

	// sync dns hosts as a whole
	//go wait.Until(siteManager.syncDnsHostsAsWhole, time.Duration(syncPeriodAsWhole)*time.Second, stopCh)
	<-stopCh
}

func (siteManager *SitesManagerDaemonController) worker() {
	for siteManager.processNextWorkItem() {
	}
}

func (siteManager *SitesManagerDaemonController) processNextWorkItem() bool {
	key, quit := siteManager.queue.Get()
	if quit {
		return false
	}
	defer siteManager.queue.Done(key)
	klog.V(4).Infof("Get siteManager queue key: %s", key)

	//err := siteManager.syncHandler(key.(string))
	//klog.V(4).Infof("Get siteManager syncHandler error: %#v", err)

	siteManager.handleErr(nil, key)

	return true
}

func (siteManager *SitesManagerDaemonController) handleErr(err error, key interface{}) {
	if err == nil {
		siteManager.queue.Forget(key)
		return
	}

	if siteManager.queue.NumRequeues(key) < common.MaxRetries {
		klog.V(2).Infof("Error syncing statefulset %v: %v", key, err)
		siteManager.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	klog.V(2).Infof("Dropping statefulset %q out of the queue: %v", key, err)
	siteManager.queue.Forget(key)
}

//func (siteManager *SitesManagerDaemonController) needClearStatefulSetDomains(set *appsv1.StatefulSet) (bool, error) {
//	// Check existence of statefulset relevant service
//	svc, err := siteManager.serviceLister.Services(set.Namespace).Get(set.Spec.ServiceName)
//	if errors.IsNotFound(err) {
//		klog.V(2).Infof("StatefulSet %v relevant service %s not found", set.Name, set.Spec.ServiceName)
//		return true, nil
//	}
//	if err != nil {
//		return false, err
//	}
//	// Check GridSelectorUniqKeyName label value equation between service and statefulset
//	gridUniqKey, _ := set.Labels[controllercommon.GridSelectorUniqKeyName]
//	svcGridUniqKey, found := svc.Labels[controllercommon.GridSelectorUniqKeyName]
//	if !found {
//		return true, nil
//	} else if gridUniqKey != svcGridUniqKey {
//		return true, nil
//	}
//	return false, nil
//}

//func (siteManager *SitesManagerDaemonController) syncDnsHostsAsWhole() {
//	startTime := time.Now()
//	klog.V(4).Infof("Started syncing dns hosts as a whole (%v)", startTime)
//	defer func() {
//		klog.V(4).Infof("Finished syncing dns hosts as a whole (%v)", time.Since(startTime))
//	}()
//
//	// Get node relevant GridSelectorUniqKeyName labels
//	node, err := siteManager.nodeLister.Get(siteManager.hostName)
//	if err != nil {
//		klog.Errorf("Get host node %s err %v", siteManager.hostName, err)
//		return
//	}
//	gridUniqKeyLabels, err := controllercommon.GetNodesSelector(node)
//	if err != nil {
//		klog.Errorf("Get node %s GridSelectorUniqKeyName selector err %v", node.Name, err)
//		return
//	}
//
//	// List all statefulsets by node labels
//	setList, err := siteManager.setLister.List(gridUniqKeyLabels)
//	if err != nil {
//		klog.Errorf("List statefulsets by labels %v err %v", gridUniqKeyLabels, err)
//		return
//	}
//	hostsMap := make(map[string]string)
//
//	// Filter concerned statefulsets and construct dns hosts
//	for _, set := range setList {
//		if rel, err := siteManager.IsConcernedStatefulSet(set); err != nil || !rel {
//			continue
//		}
//		if needClear, err := siteManager.needClearStatefulSetDomains(set); err != nil || needClear {
//			continue
//		}
//
//		// Get pod list of this statefulset
//		podList, err := siteManager.podLister.Pods(set.Namespace).List(labels.Everything())
//		if err != nil {
//			klog.Errorf("Get podList err %v", err)
//			return
//		}
//
//		ControllerRef := metav1.GetControllerOf(set)
//		gridValue := set.Name[len(ControllerRef.Name)+1:]
//		for _, pod := range podList {
//			if util.IsMemberOf(set, pod) {
//				index := strings.Index(pod.Name, gridValue)
//				if index == -1 {
//					klog.Errorf("Invalid pod name %s(statefulset %s)", pod.Name, set.Name)
//					continue
//				}
//				podDomainsToHosts := pod.Name[0:index] + pod.Name[index+len(gridValue)+1:] + "." + set.Spec.ServiceName
//				if pod.Status.PodIP == "" {
//					klog.V(2).Infof("There is currently no ip for pod %s(statefulset %s)", pod.Name, set.Name)
//					continue
//				}
//				hostsMap[hosts.AppendDomainSuffix(podDomainsToHosts, pod.Namespace)] = pod.Status.PodIP
//			}
//		}
//	}
//	// Set dns hosts as a whole
//	if err := siteManager.hosts.SetHostsByMap(hostsMap); err != nil {
//		klog.Errorf("SetHostsByMap err %v", err)
//	}
//	return
//}
//
//func (siteManager *SitesManagerDaemonController) syncDnsHosts(key string) error {
//	startTime := time.Now()
//	klog.V(4).Infof("Started syncing dns hosts of statefulset %q (%v)", key, startTime)
//	defer func() {
//		klog.V(4).Infof("Finished syncing dns hosts of statefulset %q (%v)", key, time.Since(startTime))
//	}()
//
//	namespace, name, err := cache.SplitMetaNamespaceKey(key)
//	if err != nil {
//		return err
//	}
//
//	set, err := siteManager.setLister.StatefulSets(namespace).Get(name)
//	if errors.IsNotFound(err) {
//		klog.V(2).Infof("StatefulSet %v has been deleted", key)
//		return nil
//	}
//	if err != nil {
//		return err
//	}
//
//	var PodDomainInfoToHosts = make(map[string]string)
//	ControllerRef := metav1.GetControllerOf(set)
//	// Check existence of statefulset relevant service and execute delete operations if necessary
//	if needClear, err := siteManager.needClearStatefulSetDomains(set); err != nil {
//		return err
//	} else if needClear {
//		if err := siteManager.hosts.CheckOrUpdateHosts(PodDomainInfoToHosts, set.Namespace, ControllerRef.Name, set.Spec.ServiceName); err != nil {
//			klog.Errorf("Clear statefulset %v dns hosts err %v", key, err)
//			return err
//		}
//		klog.V(4).Infof("Clear statefulset %v dns hosts successfully", key)
//		return nil
//	}
//
//	// Get pod list of this statefulset
//	podList, err := siteManager.podLister.Pods(set.Namespace).List(labels.Everything())
//	if err != nil {
//		klog.Errorf("Get podList err %v", err)
//		return err
//	}
//
//	podToHosts := []*corev1.Pod{}
//	for _, pod := range podList {
//		if util.IsMemberOf(set, pod) {
//			podToHosts = append(podToHosts, pod)
//		}
//	}
//	// Sync dns hosts partly
//	// Attention: this sync can not guarantee the absolute correctness of statefulset grid dns hosts records,
//	// and should be used combined with syncDnsHostsAsWhole to ensure the eventual consistency
//	// Actual statefulset pod FQDN: <controllerRef>-<gridValue>-<ordinal>.<svc>.<ns>.svc.cluster.local
//	// (eg: statefulsetgrid-demo-nodeunit1-0.servicegrid-demo-svc.default.svc.cluster.local)
//	// Converted statefulset pod FQDN: <controllerRef>-<ordinal>.<svc>.<ns>.svc.cluster.local
//	// (eg: statefulsetgrid-demo-0.servicegrid-demo-svc.default.svc.cluster.local)
//	if ControllerRef != nil {
//		gridValue := set.Name[len(ControllerRef.Name)+1:]
//		for _, pod := range podToHosts {
//			index := strings.Index(pod.Name, gridValue)
//			if index == -1 {
//				klog.Errorf("Invalid pod name %s(statefulset %s)", pod.Name, set.Name)
//				continue
//			}
//			podDomainsToHosts := pod.Name[0:index] + pod.Name[index+len(gridValue)+1:] + "." + set.Spec.ServiceName
//			if pod.Status.PodIP == "" {
//				klog.V(2).Infof("There is currently no ip for pod %s(statefulset %s)", pod.Name, set.Name)
//				continue
//			}
//			PodDomainInfoToHosts[hosts.AppendDomainSuffix(podDomainsToHosts, pod.Namespace)] = pod.Status.PodIP
//		}
//		if err := siteManager.hosts.CheckOrUpdateHosts(PodDomainInfoToHosts, set.Namespace, ControllerRef.Name, set.Spec.ServiceName); err != nil {
//			klog.Errorf("update dns hosts err %v", err)
//			return err
//		}
//	}
//	return nil
//}

func (siteManager *SitesManagerDaemonController) enqueue(nodeunit *sitev1.NodeUnit) {
	key, err := KeyFunc(nodeunit)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", nodeunit, err))
		return
	}

	siteManager.queue.Add(key)
}
