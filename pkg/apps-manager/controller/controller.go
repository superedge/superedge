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
	sitev1 "github.com/superedge/superedge/pkg/apps-manager/apis/apps/v1"
	crdClientset "github.com/superedge/superedge/pkg/apps-manager/generated/clientset/versioned"
	crdinformers "github.com/superedge/superedge/pkg/apps-manager/generated/informers/externalversions/apps/v1"
	crdv1listers "github.com/superedge/superedge/pkg/apps-manager/generated/listers/apps/v1"
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
	controllerKind = appsv1.SchemeGroupVersion.WithKind("apps-manager")
)

type SitesManagerController struct {
	hostName string
	hosts    *hosts.Hosts

	podLister       corelisters.PodLister
	podListerSynced cache.InformerSynced

	nodeLister       corelisters.NodeLister
	nodeListerSynced cache.InformerSynced

	serviceLister       corelisters.ServiceLister
	serviceListerSynced cache.InformerSynced

	eDeploymentLister    crdv1listers.EDeploymentLister
	nodeUnitListerSynced cache.InformerSynced

	eventRecorder record.EventRecorder
	queue         workqueue.RateLimitingInterface
	kubeClient    clientset.Interface
	crdClient     *crdClientset.Clientset

	syncHandler     func(dKey string) error
	enqueueNodeUnit func(set *sitev1.EDeployment)
}

func NewAppsManagerDaemonController(
	nodeInformer coreinformers.NodeInformer,
	podInformer coreinformers.PodInformer,
	edeployUnitInformer crdinformers.EDeploymentInformer,
	serviceInformer coreinformers.ServiceInformer,
	kubeClient clientset.Interface,
	crdClient *crdClientset.Clientset,
	hostName string, hosts *hosts.Hosts) *SitesManagerController {

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1.EventSinkImpl{
		Interface: kubeClient.CoreV1().Events(""),
	})

	appsManager := &SitesManagerController{
		hostName:      hostName,
		hosts:         hosts,
		kubeClient:    kubeClient,
		crdClient:     crdClient,
		eventRecorder: eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "apps-manager-daemon"}),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "apps-manager-daemon"),
	}

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		//AddFunc:    appsController.addNode,
		//UpdateFunc: appsController.updateNode,
		//DeleteFunc: appsController.deleteNode,
	})

	//podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
	//	AddFunc:   appsManager.addPod,
	//	UpdateFunc:appsManager.updatePod,
	//	DeleteFunc:appsManager.deletePod,
	//})

	edeployUnitInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    appsManager.addEDeploy,
		UpdateFunc: appsManager.updateEDeploy,
		DeleteFunc: appsManager.deleteEDeploy,
	})

	//serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
	//	AddFunc:   appsManager.addService,
	//	UpdateFunc:appsManager.updateService,
	//	DeleteFunc:appsManager.deleteService,
	//})

	//siteController.syncHandler = siteController.syncDnsHosts
	//siteController.enqueueNodeUnit = siteController.enqueue

	//siteController.podLister = podInformer.Lister()
	//siteController.podListerSynced = podInformer.Informer().HasSynced

	appsManager.nodeLister = nodeInformer.Lister()
	appsManager.nodeListerSynced = nodeInformer.Informer().HasSynced

	//siteController.serviceLister = serviceInformer.Lister()
	//siteController.serviceListerSynced = serviceInformer.Informer().HasSynced

	appsManager.eDeploymentLister = edeployUnitInformer.Lister()
	appsManager.nodeUnitListerSynced = edeployUnitInformer.Informer().HasSynced

	klog.V(4).Infof("Site-manager set handler success")

	return appsManager
}

func (appsManager *SitesManagerController) Run(workers, syncPeriodAsWhole int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer appsManager.queue.ShutDown()

	klog.V(1).Infof("Starting apps-manager daemon")
	defer klog.V(1).Infof("Shutting down apps-manager daemon")

	if !cache.WaitForNamedCacheSync("apps-manager-daemon", stopCh,
		appsManager.nodeListerSynced, appsManager.nodeUnitListerSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(appsManager.worker, time.Second, stopCh)
		klog.V(4).Infof("Site-manager set worker-%d success", i)
	}

	//sync dns hosts as a whole
	//go wait.Until(siteManager.syncDnsHostsAsWhole, time.Duration(syncPeriodAsWhole)*time.Second, stopCh)
	<-stopCh
}

func (appsManager *SitesManagerController) worker() {
	for appsManager.processNextWorkItem() {
	}
}

func (appsManager *SitesManagerController) processNextWorkItem() bool {
	key, quit := appsManager.queue.Get()
	if quit {
		return false
	}
	defer appsManager.queue.Done(key)
	klog.V(4).Infof("GetappsManager queue key: %s", key)

	//err :=appsManager.syncHandler(key.(string))
	//klog.V(4).Infof("GetappsManager syncHandler error: %#v", err)

	appsManager.handleErr(nil, key)

	return true
}

func (appsManager *SitesManagerController) handleErr(err error, key interface{}) {
	if err == nil {
		appsManager.queue.Forget(key)
		return
	}

	if appsManager.queue.NumRequeues(key) < common.MaxRetries {
		klog.V(2).Infof("Error syncing statefulset %v: %v", key, err)
		appsManager.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	klog.V(2).Infof("Dropping statefulset %q out of the queue: %v", key, err)
	appsManager.queue.Forget(key)
}

func (appsManager *SitesManagerController) enqueue(nodeunit *sitev1.EDeployment) {
	key, err := KeyFunc(nodeunit)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", nodeunit, err))
		return
	}

	appsManager.queue.Add(key)
}
