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
	"time"

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

	sitev1 "github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha1"
	"github.com/superedge/superedge/pkg/site-manager/constant"
	crdClientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	crdinformers "github.com/superedge/superedge/pkg/site-manager/generated/informers/externalversions/site.superedge.io/v1alpha1"
	crdv1listers "github.com/superedge/superedge/pkg/site-manager/generated/listers/site.superedge.io/v1alpha1"
	"github.com/superedge/superedge/pkg/site-manager/utils"
)

var (
	KeyFunc        = cache.DeletionHandlingMetaNamespaceKeyFunc
	controllerKind = appsv1.SchemeGroupVersion.WithKind("site-manager-daemon")
	finalizerID    = "site.superedge.io/finalizer"
)

type SitesManagerDaemonController struct {
	nodeLister       corelisters.NodeLister
	nodeListerSynced cache.InformerSynced

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
	nodeUnitInformer crdinformers.NodeUnitInformer,
	nodeGroupInformer crdinformers.NodeGroupInformer,
	kubeClient clientset.Interface,
	crdClient *crdClientset.Clientset,
) *SitesManagerDaemonController {

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1.EventSinkImpl{
		Interface: kubeClient.CoreV1().Events(""),
	})

	siteController := &SitesManagerDaemonController{
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

	nodeUnitInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    siteController.addNodeUnit,
		UpdateFunc: siteController.updateNodeUnit,
		DeleteFunc: siteController.deleteNodeUnit,
	})

	nodeGroupInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    siteController.addNodeGroup,
		UpdateFunc: siteController.updateNodeGroup,
		DeleteFunc: siteController.deleteNodeGroup,
	})

	//siteController.syncHandler = siteController.syncNodeUnit
	//siteController.enqueueNodeUnit = siteController.enqueue

	siteController.nodeLister = nodeInformer.Lister()
	siteController.nodeListerSynced = nodeInformer.Informer().HasSynced

	siteController.nodeUnitLister = nodeUnitInformer.Lister()
	siteController.nodeUnitListerSynced = nodeUnitInformer.Informer().HasSynced

	siteController.nodeGroupLister = nodeGroupInformer.Lister()
	siteController.nodeGroupListerSynced = nodeGroupInformer.Informer().HasSynced

	if err := utils.InitUnitToNode(kubeClient, crdClient); err != nil {
		klog.Errorf("Init unit info to node error: %#v", err)
	}

	klog.V(4).Infof("Site-manager set handler success")

	return siteController
}

func (siteManager *SitesManagerDaemonController) Run(workers, syncPeriodAsWhole int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer siteManager.queue.ShutDown()

	klog.V(1).Infof("Starting site-manager daemon")
	defer klog.V(1).Infof("Shutting down site-manager daemon")

	if !cache.WaitForNamedCacheSync("site-manager-daemon", stopCh,
		siteManager.nodeListerSynced, siteManager.nodeUnitListerSynced, siteManager.nodeGroupListerSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(siteManager.worker, time.Second, stopCh)
		klog.V(4).Infof("Site-manager set worker-%d success", i)
	}

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

	siteManager.handleErr(nil, key)

	return true
}

func (siteManager *SitesManagerDaemonController) handleErr(err error, key interface{}) {
	if err == nil {
		siteManager.queue.Forget(key)
		return
	}

	if siteManager.queue.NumRequeues(key) < constant.MaxRetries {
		klog.V(2).Infof("Error syncing siteManager %v: %v", key, err)
		siteManager.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	klog.V(2).Infof("Dropping siteManager %q out of the queue: %v", key, err)
	siteManager.queue.Forget(key)
}

func (siteManager *SitesManagerDaemonController) enqueue(nodeunit *sitev1.NodeUnit) {
	key, err := KeyFunc(nodeunit)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", nodeunit, err))
		return
	}

	siteManager.queue.Add(key)
}
