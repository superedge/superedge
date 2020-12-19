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
	"superedge/pkg/application-grid-controller/controller"
	"superedge/pkg/application-grid-controller/controller/common"
	"time"

	"k8s.io/klog"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
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

	crdv1 "superedge/pkg/application-grid-controller/apis/superedge.io/v1"

	crdclientset "superedge/pkg/application-grid-controller/generated/clientset/versioned"
	crdinformers "superedge/pkg/application-grid-controller/generated/informers/externalversions/superedge.io/v1"
	crdv1listers "superedge/pkg/application-grid-controller/generated/listers/superedge.io/v1"
)

var controllerKind = crdv1.SchemeGroupVersion.WithKind("DeploymentGrid")

type DeploymentGridController struct {
	dpControl          controller.DPControlInterface
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
	dgc.dpControl = controller.RealDPControl{
		KubeClient: kubeClient,
	}

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

func (dgc *DeploymentGridController) syncDeploymentGrid(key string) error {
	startTime := time.Now()
	klog.V(4).Infof("Started syncing deployment-grid %q (%v)", key, startTime)
	defer func() {
		klog.V(4).Infof("Finished syncing deployment-grid %q (%v)", key, time.Since(startTime))
	}()

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	grid, err := dgc.dpGridLister.DeploymentGrids(ns).Get(name)
	if errors.IsNotFound(err) {
		klog.V(2).Infof("Deployment-grid %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	g := grid.DeepCopy()
	if g.Spec.GridUniqKey == "" {
		dgc.eventRecorder.Eventf(g, corev1.EventTypeWarning, "Empty", "This deployment-grid has an empty grid key")
		return nil
	}

	/* get deploy list for this grid
	 */
	dpList, err := dgc.getDeploymentForGrid(g)
	if err != nil {
		return err
	}

	/* gridValues: grid labels in all nodes
	 */
	gridValues, err := dgc.getGridValueFromNode(g)
	if err != nil {
		return err
	}

	if g.DeletionTimestamp != nil {
		return dgc.syncStatus(g, dpList, gridValues)
	}

	/*
	 */
	return dgc.reconcile(g, dpList, gridValues)
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

func (dgc *DeploymentGridController) getDeploymentForGrid(g *crdv1.DeploymentGrid) ([]*appsv1.Deployment, error) {
	dpList, err := dgc.dpLister.Deployments(g.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	labelSelector, err := common.GetDefaultSelector(g.Name)
	if err != nil {
		return nil, err
	}
	canAdoptFunc := controller.RecheckDeletionTimestamp(func() (metav1.Object, error) {
		fresh, err := dgc.crdClient.SuperedgeV1().DeploymentGrids(g.Namespace).Get(context.TODO(), g.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		if fresh.UID != g.UID {
			return nil, fmt.Errorf("orignal Deployment-grid %v/%v is gone: got uid %v, wanted %v", g.Namespace,
				g.Name, fresh.UID, g.UID)
		}
		return fresh, nil
	})

	cm := controller.NewDeploymentControllerRefManager(dgc.dpControl, g, labelSelector, controllerKind, canAdoptFunc)
	return cm.ClaimDeployment(dpList)
}

func (dgc *DeploymentGridController) getGridValueFromNode(g *crdv1.DeploymentGrid) ([]string, error) {
	labelSelector := labels.NewSelector()
	gridRequirement, err := labels.NewRequirement(g.Spec.GridUniqKey, selection.Exists, nil)
	if err != nil {
		return nil, err
	}
	labelSelector = labelSelector.Add(*gridRequirement)

	nodes, err := dgc.nodeLister.List(labelSelector)
	if err != nil {
		return nil, err
	}

	values := make([]string, 0)
	for _, n := range nodes {
		gridVal := n.Labels[g.Spec.GridUniqKey]
		if gridVal != "" {
			values = append(values, gridVal)
		}
	}
	return values, nil
}

func (dgc *DeploymentGridController) enqueue(deploymentGrid *crdv1.DeploymentGrid) {
	key, err := controller.KeyFunc(deploymentGrid)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", deploymentGrid, err))
		return
	}

	dgc.queue.Add(key)
}
