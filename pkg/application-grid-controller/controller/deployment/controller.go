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

package deployment

import (
	"context"
	"fmt"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/deployment/util"
	"time"

	"k8s.io/klog"

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

	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"

	crdclientset "github.com/superedge/superedge/pkg/application-grid-controller/generated/clientset/versioned"
	crdinformers "github.com/superedge/superedge/pkg/application-grid-controller/generated/informers/externalversions/superedge.io/v1"
	crdv1listers "github.com/superedge/superedge/pkg/application-grid-controller/generated/listers/superedge.io/v1"
)

type DeploymentGridController struct {
	dpClient           controller.DeployClientInterface
	dpGridLister       crdv1listers.DeploymentGridLister
	dpLister           appslisters.DeploymentLister
	nodeLister         corelisters.NodeLister
	dpGridListerSynced cache.InformerSynced
	dpListerSynced     cache.InformerSynced
	nodeListerSynced   cache.InformerSynced

	eventRecorder record.EventRecorder
	queue         workqueue.RateLimitingInterface
	kubeClient    clientset.Interface
	crdClient     crdclientset.Interface

	// To allow injection of syncDeploymentGrid for testing.
	syncHandler func(dKey string) error
	// used for unit testing
	enqueueDeploymentGrid func(deploymentGrid *crdv1.DeploymentGrid)
}

func NewDeploymentGridController(dpGridInformer crdinformers.DeploymentGridInformer, dpInformer appsinformers.DeploymentInformer,
	nodeInformer coreinformers.NodeInformer, kubeClient clientset.Interface, crdClient crdclientset.Interface) *DeploymentGridController {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1.EventSinkImpl{
		Interface: kubeClient.CoreV1().Events(""),
	})

	dgc := &DeploymentGridController{
		kubeClient: kubeClient,
		crdClient:  crdClient,
		eventRecorder: eventBroadcaster.NewRecorder(scheme.Scheme,
			corev1.EventSource{Component: "deployment-grid-controller"}),
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(),
			"deployment-grid-controller"),
	}
	dgc.dpClient = controller.NewRealDeployClient(kubeClient)

	dpGridInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    dgc.addDeploymentGrid,
		UpdateFunc: dgc.updateDeploymentGrid,
		DeleteFunc: dgc.deleteDeploymentGrid,
	})

	dpInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    dgc.addDeployment,
		UpdateFunc: dgc.updateDeployment,
		DeleteFunc: dgc.deleteDeployment,
	})

	// TODO: node label changed causing deployment deletion?
	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    dgc.addNode,
		UpdateFunc: dgc.updateNode,
		DeleteFunc: dgc.deleteNode,
	})

	dgc.syncHandler = dgc.syncDeploymentGrid
	dgc.enqueueDeploymentGrid = dgc.enqueue

	dgc.dpLister = dpInformer.Lister()
	dgc.dpListerSynced = dpInformer.Informer().HasSynced

	dgc.dpGridLister = dpGridInformer.Lister()
	dgc.dpGridListerSynced = dpGridInformer.Informer().HasSynced

	dgc.nodeLister = nodeInformer.Lister()
	dgc.nodeListerSynced = nodeInformer.Informer().HasSynced

	return dgc
}

func (dgc *DeploymentGridController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer dgc.queue.ShutDown()

	klog.Infof("Starting deployment grid controller")
	defer klog.Infof("Shutting down deployment grid controller")

	if !cache.WaitForNamedCacheSync("deployment-grid", stopCh,
		dgc.dpGridListerSynced, dgc.dpListerSynced, dgc.nodeListerSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(dgc.worker, time.Second, stopCh)
	}
	<-stopCh
}

func (dgc *DeploymentGridController) worker() {
	for dgc.processNextWorkItem() {
	}
}

func (dgc *DeploymentGridController) processNextWorkItem() bool {
	key, quit := dgc.queue.Get()
	if quit {
		return false
	}
	defer dgc.queue.Done(key)

	err := dgc.syncHandler(key.(string))
	dgc.handleErr(err, key)

	return true
}

func (dgc *DeploymentGridController) handleErr(err error, key interface{}) {
	if err == nil {
		dgc.queue.Forget(key)
		return
	}

	if dgc.queue.NumRequeues(key) < common.MaxRetries {
		klog.V(2).Infof("Error syncing deployment grid %v: %v", key, err)
		dgc.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	klog.V(2).Infof("Dropping deployment grid %q out of the queue: %v", key, err)
	dgc.queue.Forget(key)
}

func (dgc *DeploymentGridController) syncDeploymentGrid(key string) error {
	startTime := time.Now()
	klog.V(4).Infof("Started syncing deployment grid %q (%v)", key, startTime)
	defer func() {
		klog.V(4).Infof("Finished syncing deployment grid %q (%v)", key, time.Since(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	dg, err := dgc.dpGridLister.DeploymentGrids(namespace).Get(name)
	if errors.IsNotFound(err) {
		klog.V(2).Infof("deployment grid %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	if dg.Spec.GridUniqKey == "" {
		dgc.eventRecorder.Eventf(dg, corev1.EventTypeWarning, "Empty", "This deployment-grid has an empty grid key")
		return nil
	}

	// get deployment workload list of this grid
	dpList, err := dgc.getDeploymentForGrid(dg)
	if err != nil {
		return err
	}

	// get all grid labels in all nodes
	gridValues, err := common.GetGridValuesFromNode(dgc.nodeLister, dg.Spec.GridUniqKey)
	if err != nil {
		return err
	}

	// sync deployment grid workload status
	if dg.DeletionTimestamp != nil {
		return dgc.syncStatus(dg, dpList, gridValues)
	}

	// sync deployment grid status and its relevant deployments workload
	return dgc.reconcile(dg, dpList, gridValues)
}

func (dgc *DeploymentGridController) getDeploymentForGrid(dg *crdv1.DeploymentGrid) ([]*appsv1.Deployment, error) {
	dpList, err := dgc.dpLister.Deployments(dg.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	labelSelector, err := common.GetDefaultSelector(dg.Name)
	if err != nil {
		return nil, err
	}
	canAdoptFunc := controller.RecheckDeletionTimestamp(func() (metav1.Object, error) {
		fresh, err := dgc.crdClient.SuperedgeV1().DeploymentGrids(dg.Namespace).Get(context.TODO(), dg.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		if fresh.UID != dg.UID {
			return nil, fmt.Errorf("orignal Deployment-grid %v/%v is gone: got uid %v, wanted %v", dg.Namespace,
				dg.Name, fresh.UID, dg.UID)
		}
		return fresh, nil
	})

	cm := controller.NewDeploymentControllerRefManager(dgc.dpClient, dg, labelSelector, util.ControllerKind, canAdoptFunc)
	return cm.ClaimDeployment(dpList)
}

func (dgc *DeploymentGridController) addDeploymentGrid(obj interface{}) {
	dg := obj.(*crdv1.DeploymentGrid)
	klog.V(4).Infof("Adding deployment grid %s", dg.Name)
	dgc.enqueueDeploymentGrid(dg)
}

func (dgc *DeploymentGridController) updateDeploymentGrid(oldObj, newObj interface{}) {
	oldDg := oldObj.(*crdv1.DeploymentGrid)
	curDg := newObj.(*crdv1.DeploymentGrid)
	klog.V(4).Infof("Updating deployment grid %s", oldDg.Name)
	if curDg.ResourceVersion == oldDg.ResourceVersion {
		// Periodic resync will send update events for all known DeploymentGrids.
		// Two different versions of the same DeploymentGrid will always have different RVs.
		return
	}
	dgc.enqueueDeploymentGrid(curDg)
}

func (dgc *DeploymentGridController) deleteDeploymentGrid(obj interface{}) {
	dg, ok := obj.(*crdv1.DeploymentGrid)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		dg, ok = tombstone.Obj.(*crdv1.DeploymentGrid)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object that is not a deployment grid %#v", obj))
			return
		}
	}
	klog.V(4).Infof("Deleting deployment grid %s", dg.Name)
	dgc.enqueueDeploymentGrid(dg)
}

func (dgc *DeploymentGridController) enqueue(deploymentGrid *crdv1.DeploymentGrid) {
	key, err := controller.KeyFunc(deploymentGrid)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", deploymentGrid, err))
		return
	}

	dgc.queue.Add(key)
}
