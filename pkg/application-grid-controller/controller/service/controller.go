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

package service

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
	"k8s.io/klog"

	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/service/util"
	crdclientset "github.com/superedge/superedge/pkg/application-grid-controller/generated/clientset/versioned"
	crdinformers "github.com/superedge/superedge/pkg/application-grid-controller/generated/informers/externalversions/superedge.io/v1"
	crdv1listers "github.com/superedge/superedge/pkg/application-grid-controller/generated/listers/superedge.io/v1"
)

type ServiceGridController struct {
	svcControl          controller.SVCControlInterface
	svcGridLister       crdv1listers.ServiceGridLister
	svcLister           corelisters.ServiceLister
	svcGridListerSynced cache.InformerSynced
	svcListerSynced     cache.InformerSynced

	eventRecorder record.EventRecorder
	queue         workqueue.RateLimitingInterface
	kubeClient    clientset.Interface
	crdClient     crdclientset.Interface

	// To allow injection of syncServiceGrid for testing.
	syncHandler func(dKey string) error
	// used for unit testing
	enqueueServiceGrid func(service *crdv1.ServiceGrid)
}

func NewServiceGridController(svcGridInformer crdinformers.ServiceGridInformer, svcInformer coreinformers.ServiceInformer,
	client clientset.Interface, crdClient crdclientset.Interface) *ServiceGridController {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1.EventSinkImpl{
		Interface: client.CoreV1().Events(""),
	})

	sgc := &ServiceGridController{
		kubeClient: client,
		crdClient:  crdClient,
		eventRecorder: eventBroadcaster.NewRecorder(scheme.Scheme,
			corev1.EventSource{Component: "service-grid-controller"}),
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service-grid-controller"),
	}
	sgc.svcControl = controller.RealSVCControl{
		KubeClient: client,
	}

	svcGridInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sgc.addServiceGrid,
		UpdateFunc: sgc.updateServiceGrid,
		DeleteFunc: sgc.deleteServiceGrid,
	})

	svcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sgc.addService,
		UpdateFunc: sgc.updateService,
		DeleteFunc: sgc.deleteService,
	})

	sgc.syncHandler = sgc.syncServiceGrid
	sgc.enqueueServiceGrid = sgc.enqueue

	sgc.svcLister = svcInformer.Lister()
	sgc.svcListerSynced = svcInformer.Informer().HasSynced

	sgc.svcGridLister = svcGridInformer.Lister()
	sgc.svcGridListerSynced = svcGridInformer.Informer().HasSynced
	return sgc
}

func (sgc *ServiceGridController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer sgc.queue.ShutDown()

	klog.Infof("Starting service grid controller")
	defer klog.Infof("Shutting down service grid controller")

	if !cache.WaitForNamedCacheSync("service-grid", stopCh, sgc.svcGridListerSynced, sgc.svcListerSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(sgc.worker, time.Second, stopCh)
	}
	<-stopCh
}

func (sgc *ServiceGridController) worker() {
	for sgc.processNextWorkItem() {
	}
}

func (sgc *ServiceGridController) processNextWorkItem() bool {
	key, quit := sgc.queue.Get()
	if quit {
		return false
	}
	defer sgc.queue.Done(key)

	err := sgc.syncHandler(key.(string))
	sgc.handleErr(err, key)

	return true
}

func (sgc *ServiceGridController) handleErr(err error, key interface{}) {
	if err == nil {
		sgc.queue.Forget(key)
		return
	}

	if sgc.queue.NumRequeues(key) < common.MaxRetries {
		klog.V(2).Infof("Error syncing service grid %v: %v", key, err)
		sgc.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	klog.V(2).Infof("Dropping service grid %q out of the queue: %v", key, err)
	sgc.queue.Forget(key)
}

/*
 */
func (sgc *ServiceGridController) syncServiceGrid(key string) error {
	startTime := time.Now()
	klog.V(4).Infof("Started syncing service-grid %q (%v)", key, startTime)
	defer func() {
		klog.V(4).Infof("Finished syncing service-grid %q (%v)", key, time.Since(startTime))
	}()

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	grid, err := sgc.svcGridLister.ServiceGrids(ns).Get(name)
	if errors.IsNotFound(err) {
		klog.V(2).Infof("Service-grid %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	sg := grid.DeepCopy()
	if sg.Spec.GridUniqKey == "" {
		sgc.eventRecorder.Eventf(sg, corev1.EventTypeWarning, "Empty", "This service-grid has an empty grid key")
		return nil
	}

	svcList, err := sgc.getServiceForGrid(sg)
	if err != nil {
		return err
	}

	if sg.DeletionTimestamp != nil {
		return nil
	}
	return sgc.reconcile(sg, svcList)
}

func (sgc *ServiceGridController) enqueue(serviceGrid *crdv1.ServiceGrid) {
	key, err := controller.KeyFunc(serviceGrid)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", serviceGrid, err))
		return
	}

	sgc.queue.Add(key)
}

func (sgc *ServiceGridController) getServiceForGrid(sg *crdv1.ServiceGrid) ([]*corev1.Service, error) {
	svcList, err := sgc.svcLister.Services(sg.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	labelSelector, err := common.GetDefaultSelector(sg.Name)
	if err != nil {
		return nil, err
	}
	canAdoptFunc := controller.RecheckDeletionTimestamp(func() (metav1.Object, error) {
		fresh, err := sgc.crdClient.SuperedgeV1().ServiceGrids(sg.Namespace).Get(context.TODO(), sg.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		if fresh.UID != sg.UID {
			return nil, fmt.Errorf("orignal Service-grid %v/%v is gone: got uid %v, wanted %v", sg.Namespace,
				sg.Name, fresh.UID, sg.UID)
		}
		return fresh, nil
	})

	cm := controller.NewServiceControllerRefManager(sgc.svcControl, sg, labelSelector, util.ControllerKind, canAdoptFunc)
	return cm.ClaimService(svcList)
}
