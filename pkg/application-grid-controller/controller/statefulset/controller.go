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

package statefulset

import (
	"context"
	"fmt"
	"time"

	"github.com/superedge/superedge/pkg/application-grid-controller/controller"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	appsv1 "k8s.io/api/apps/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/klog/v2"

	"github.com/superedge/superedge/pkg/application-grid-controller/generated/clientset/versioned/scheme"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"

	"github.com/superedge/superedge/pkg/application-grid-controller/controller/statefulset/util"
	crdclientset "github.com/superedge/superedge/pkg/application-grid-controller/generated/clientset/versioned"
	crdinformers "github.com/superedge/superedge/pkg/application-grid-controller/generated/informers/externalversions/superedge.io/v1"
	crdv1listers "github.com/superedge/superedge/pkg/application-grid-controller/generated/listers/superedge.io/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type StatefulSetGridController struct {
	setClient           controller.SetClientInterface
	setGridLister       crdv1listers.StatefulSetGridLister
	setLister           appslisters.StatefulSetLister
	nodeLister          corelisters.NodeLister
	setGridListerSynced cache.InformerSynced
	setListerSynced     cache.InformerSynced
	nodeListerSynced    cache.InformerSynced

	eventRecorder record.EventRecorder
	queue         workqueue.RateLimitingInterface
	kubeClient    clientset.Interface
	crdClient     crdclientset.Interface

	// To allow injection of syncStatefulSetGrid for testing.
	syncHandler func(dKey string) error
	// used for unit testing
	enqueueStatefulSetGrid func(sg *crdv1.StatefulSetGrid)

	templateHasher util.StatefulsetTemplateHash
}

func NewStatefulSetGridController(setGridInformer crdinformers.StatefulSetGridInformer, setInformer appsinformers.StatefulSetInformer,
	nodeInformer coreinformers.NodeInformer, kubeClient clientset.Interface, crdClient crdclientset.Interface) *StatefulSetGridController {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1.EventSinkImpl{
		Interface: kubeClient.CoreV1().Events(""),
	})

	ssgc := &StatefulSetGridController{
		kubeClient: kubeClient,
		crdClient:  crdClient,
		eventRecorder: eventBroadcaster.NewRecorder(scheme.Scheme,
			corev1.EventSource{Component: "statefulset-grid-controller"}),
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "statefulset-grid-controller"),
	}
	ssgc.setClient = controller.NewRealSetClient(kubeClient)

	setGridInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ssgc.addStatefulSetGrid,
		UpdateFunc: ssgc.updateStatefulSetGrid,
		DeleteFunc: ssgc.deleteStatefulSetGrid,
	})

	setInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ssgc.addStatefulSet,
		UpdateFunc: ssgc.updateStatefulSet,
		DeleteFunc: ssgc.deleteStatefulSet,
	})

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ssgc.addNode,
		UpdateFunc: ssgc.updateNode,
		DeleteFunc: ssgc.deleteNode,
	})

	ssgc.syncHandler = ssgc.syncStatefulSetGrid
	ssgc.enqueueStatefulSetGrid = ssgc.enqueue

	ssgc.setGridLister = setGridInformer.Lister()
	ssgc.setGridListerSynced = setGridInformer.Informer().HasSynced

	ssgc.setLister = setInformer.Lister()
	ssgc.setListerSynced = setInformer.Informer().HasSynced

	ssgc.nodeLister = nodeInformer.Lister()
	ssgc.nodeListerSynced = nodeInformer.Informer().HasSynced

	ssgc.templateHasher = util.NewStatefulsetTemplateHash()

	return ssgc
}

func (ssgc *StatefulSetGridController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer ssgc.queue.ShutDown()

	klog.Infof("Starting statefulset grid controller")
	defer klog.Infof("Shutting down statefulset grid controller")

	if !cache.WaitForNamedCacheSync("statefulset-grid", stopCh,
		ssgc.setGridListerSynced, ssgc.setListerSynced, ssgc.nodeListerSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(ssgc.worker, time.Second, stopCh)
	}
	<-stopCh
}

func (ssgc *StatefulSetGridController) worker() {
	for ssgc.processNextWorkItem() {
	}
}

func (ssgc *StatefulSetGridController) processNextWorkItem() bool {
	key, quit := ssgc.queue.Get()
	if quit {
		return false
	}
	defer ssgc.queue.Done(key)

	err := ssgc.syncHandler(key.(string))
	ssgc.handleErr(err, key)

	return true
}

func (ssgc *StatefulSetGridController) handleErr(err error, key interface{}) {
	if err == nil {
		ssgc.queue.Forget(key)
		return
	}

	if ssgc.queue.NumRequeues(key) < common.MaxRetries {
		klog.V(2).Infof("Error syncing statefulset grid %v: %v", key, err)
		ssgc.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	klog.V(2).Infof("Dropping statefulset grid %q out of the queue: %v", key, err)
	ssgc.queue.Forget(key)
}

func (ssgc *StatefulSetGridController) syncStatefulSetGrid(key string) error {
	startTime := time.Now()
	klog.V(4).Infof("Started syncing statefulset grid %s (%v)", key, startTime)
	defer func() {
		klog.V(4).Infof("Finished syncing statefulset grid %s (%v)", key, time.Since(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	ssg, err := ssgc.setGridLister.StatefulSetGrids(namespace).Get(name)
	if errors.IsNotFound(err) {
		klog.V(2).Infof("statefulset grid %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	if ssg.Spec.GridUniqKey == "" {
		ssgc.eventRecorder.Eventf(ssg, corev1.EventTypeWarning, "Empty", "This statefulset-grid has an empty grid key")
		return nil
	}

	ssgCopy := ssg.DeepCopy()

	if ssgCopy.Spec.DefaultTemplateName == "" {
		ssgCopy.Spec.DefaultTemplateName = common.DefaultTemplateName
	}

	if err := ssgc.templateHasher.RemoveUnusedTemplate(ssgCopy); err != nil {
		klog.Errorf("Failed to remove unused template for statefulsetGrid %s: %v", ssg.Name, err)
		return err
	}

	ssgc.templateHasher.UpdateTemplateHash(ssgCopy)

	if !apiequality.Semantic.DeepEqual(ssg.Spec.Template, ssgCopy.Spec.Template) ||
		!apiequality.Semantic.DeepEqual(ssg.Spec.DefaultTemplateName, ssgCopy.Spec.DefaultTemplateName) ||
		!apiequality.Semantic.DeepEqual(ssg.Spec.TemplatePool, ssgCopy.Spec.TemplatePool) {
		klog.Infof("Updating statefulsetGrid %s/%s template info", ssgCopy.Namespace, ssgCopy.Name)
		_, err = ssgc.crdClient.SuperedgeV1().StatefulSetGrids(ssgCopy.Namespace).Update(context.TODO(), ssgCopy, metav1.UpdateOptions{})
		return err
	}

	// get statefulset workload list of this grid
	setList, err := ssgc.getStatefulSetForGrid(ssg)
	if err != nil {
		return err
	}

	// get all grid labels in all nodes
	gridValues, err := common.GetGridValuesFromNode(ssgc.nodeLister, ssg.Spec.GridUniqKey)
	if err != nil {
		return err
	}

	// sync statefulset grid workload status
	if ssg.DeletionTimestamp != nil {
		return ssgc.syncStatus(ssg, setList, gridValues)
	}

	// sync statefulset grid status and its relevant statefusets workload
	return ssgc.reconcile(ssg, setList, gridValues)
}

func (ssgc *StatefulSetGridController) getStatefulSetForGrid(ssg *crdv1.StatefulSetGrid) ([]*appsv1.StatefulSet, error) {
	setList, err := ssgc.setLister.StatefulSets(ssg.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	labelSelector, err := common.GetDefaultSelector(ssg.Name)
	if err != nil {
		return nil, err
	}
	canAdoptFunc := controller.RecheckDeletionTimestamp(func() (metav1.Object, error) {
		fresh, err := ssgc.crdClient.SuperedgeV1().StatefulSetGrids(ssg.Namespace).Get(context.TODO(), ssg.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		if fresh.UID != ssg.UID {
			return nil, fmt.Errorf("orignal statefulset grid %v/%v is gone: got uid %v, wanted %v", ssg.Namespace,
				ssg.Name, fresh.UID, ssg.UID)
		}
		return fresh, nil
	})

	cm := controller.NewStatefulSetControllerRefManager(ssgc.setClient, ssg, labelSelector, util.ControllerKind, canAdoptFunc)
	return cm.ClaimStatefulSet(setList)
}

func (ssgc *StatefulSetGridController) addStatefulSetGrid(obj interface{}) {
	ssg := obj.(*crdv1.StatefulSetGrid)
	klog.V(4).Infof("Adding statefulset grid %s", ssg.Name)
	ssgc.enqueueStatefulSetGrid(ssg)
}

func (ssgc *StatefulSetGridController) updateStatefulSetGrid(oldObj, newObj interface{}) {
	oldSsg := oldObj.(*crdv1.StatefulSetGrid)
	curSsg := newObj.(*crdv1.StatefulSetGrid)
	klog.V(4).Infof("Updating statefulset grid %s", oldSsg.Name)
	if curSsg.ResourceVersion == oldSsg.ResourceVersion {
		// Periodic resync will send update events for all known StatefulSetGrids.
		// Two different versions of the same StatefulSetGrid will always have different RVs.
		return
	}
	ssgc.enqueueStatefulSetGrid(curSsg)
}

func (ssgc *StatefulSetGridController) deleteStatefulSetGrid(obj interface{}) {
	ssg, ok := obj.(*crdv1.StatefulSetGrid)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		ssg, ok = tombstone.Obj.(*crdv1.StatefulSetGrid)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object that is not a statefulset grid %#v", obj))
			return
		}
	}
	klog.V(4).Infof("Deleting statefulset grid %s", ssg.Name)
	ssgc.enqueueStatefulSetGrid(ssg)
}

func (ssgc *StatefulSetGridController) enqueue(ssg *crdv1.StatefulSetGrid) {
	key, err := controller.KeyFunc(ssg)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", ssg, err))
		return
	}

	ssgc.queue.Add(key)
}
