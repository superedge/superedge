package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	commonutil "github.com/superedge/superedge/pkg/application-grid-controller/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	klog "k8s.io/klog/v2"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"

	crdclientset "github.com/superedge/superedge/pkg/application-grid-controller/generated/clientset/versioned"
	crdinformers "github.com/superedge/superedge/pkg/application-grid-controller/generated/informers/externalversions/superedge.io/v1"
	crdv1listers "github.com/superedge/superedge/pkg/application-grid-controller/generated/listers/superedge.io/v1"
)

type FedServiceGridController struct {
	namespace        string
	svcGridLister    crdv1listers.ServiceGridLister
	fedSvcGridLister crdv1listers.ServiceGridLister
	nodeLister       corelisters.NodeLister

	svcGridListerSynced    cache.InformerSynced
	fedSvcGridListerSynced cache.InformerSynced
	nodeListerSynced       cache.InformerSynced

	eventRecorder record.EventRecorder
	queue         workqueue.RateLimitingInterface
	kubeClient    clientset.Interface
	fedCrdClient  crdclientset.Interface
	crdClient     crdclientset.Interface

	// To allow injection of syncServiceGrid for testing.
	syncHandler func(dKey string) error
	// used for unit testing
	enqueueServiceGrid func(serviceGrid *crdv1.ServiceGrid)
}

func NewFedServiceGridController(svcGridInformer crdinformers.ServiceGridInformer,
	fedSvcGridInformer crdinformers.ServiceGridInformer, nodeInformer coreinformers.NodeInformer,
	fedCrdClient, crdClient crdclientset.Interface, kubeClient clientset.Interface, dedicatedNameSpace string) *FedServiceGridController {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1.EventSinkImpl{
		Interface: kubeClient.CoreV1().Events(""),
	})

	fedSgc := &FedServiceGridController{
		namespace:    dedicatedNameSpace,
		kubeClient:   kubeClient,
		crdClient:    crdClient,
		fedCrdClient: fedCrdClient,
		eventRecorder: eventBroadcaster.NewRecorder(scheme.Scheme,
			corev1.EventSource{Component: "fed-service-grid-controller"}),
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(),
			"fed-service-grid-controller"),
	}

	svcGridInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    fedSgc.addServiceGrid,
		UpdateFunc: fedSgc.updateServiceGrid,
		DeleteFunc: fedSgc.deleteServiceGrid,
	})

	fedSvcGridInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    fedSgc.addFedServiceGrid,
		UpdateFunc: fedSgc.updateFedServiceGrid,
		DeleteFunc: fedSgc.deleteFedServiceGrid,
	})

	// TODO: node label changed causing service deletion?
	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    fedSgc.addNode,
		UpdateFunc: fedSgc.updateNode,
		DeleteFunc: fedSgc.deleteNode,
	})

	fedSgc.syncHandler = fedSgc.syncFedServiceGrid
	fedSgc.enqueueServiceGrid = fedSgc.enqueue

	fedSgc.svcGridLister = svcGridInformer.Lister()
	fedSgc.svcGridListerSynced = svcGridInformer.Informer().HasSynced

	fedSgc.nodeLister = nodeInformer.Lister()
	fedSgc.nodeListerSynced = nodeInformer.Informer().HasSynced

	fedSgc.fedSvcGridLister = fedSvcGridInformer.Lister()
	fedSgc.fedSvcGridListerSynced = fedSvcGridInformer.Informer().HasSynced

	return fedSgc
}

func (sgc *FedServiceGridController) addFedServiceGrid(obj interface{}) {
	dg := obj.(*crdv1.ServiceGrid)
	klog.V(4).Infof("Adding fed service grid %s", dg.Name)
	sgc.enqueueServiceGrid(dg)
}

func (sgc *FedServiceGridController) updateFedServiceGrid(oldObj, newObj interface{}) {
	oldSg := oldObj.(*crdv1.ServiceGrid)
	curSg := newObj.(*crdv1.ServiceGrid)
	klog.V(4).Infof("Updating fed service grid %s", oldSg.Name)
	if curSg.ResourceVersion == oldSg.ResourceVersion {
		// Periodic resync will send update events for all known ServiceGrids.
		// Two different versions of the same ServiceGrid will always have different RVs.
		return
	}
	sgc.enqueueServiceGrid(curSg)
}

func (sgc *FedServiceGridController) deleteFedServiceGrid(obj interface{}) {
	sg, ok := obj.(*crdv1.ServiceGrid)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		sg, ok = tombstone.Obj.(*crdv1.ServiceGrid)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object that is not a service grid %#v", obj))
			return
		}
	}
	klog.V(4).Infof("Deleting fed service grid %s", sg.Name)
	sgc.enqueueServiceGrid(sg)
}

func (sgc *FedServiceGridController) enqueue(serivceGrid *crdv1.ServiceGrid) {
	key, err := controller.KeyFunc(serivceGrid)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", serivceGrid, err))
		return
	}

	sgc.queue.Add(key)
}

func (sgc *FedServiceGridController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer sgc.queue.ShutDown()

	klog.Infof("Starting fed service grid controller")
	defer klog.Infof("Shutting down fed service grid controller")

	if !cache.WaitForNamedCacheSync("fed-service-grid", stopCh,
		sgc.svcGridListerSynced, sgc.fedSvcGridListerSynced, sgc.nodeListerSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(sgc.worker, time.Second, stopCh)
	}
	<-stopCh
}

func (sgc *FedServiceGridController) worker() {
	for sgc.processNextWorkItem() {
	}
}

func (sgc *FedServiceGridController) processNextWorkItem() bool {
	key, quit := sgc.queue.Get()
	if quit {
		return false
	}
	defer sgc.queue.Done(key)

	err := sgc.syncHandler(key.(string))
	sgc.handleErr(err, key)

	return true
}

func (sgc *FedServiceGridController) handleErr(err error, key interface{}) {
	if err == nil {
		sgc.queue.Forget(key)
		return
	}

	if sgc.queue.NumRequeues(key) < common.MaxRetries {
		klog.V(2).Infof("Error syncing fed service grid %v: %v", key, err)
		sgc.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	klog.V(2).Infof("Dropping fed service grid %q out of the queue: %v", key, err)
	sgc.queue.Forget(key)
}

func (sgc *FedServiceGridController) syncFedServiceGrid(key string) error {
	startTime := time.Now()
	klog.V(4).Infof("Started syncing fed service grid %q (%v)", key, startTime)
	defer func() {
		klog.V(4).Infof("Finished syncing fed service grid %q (%v)", key, time.Since(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	fedSg, err := sgc.fedSvcGridLister.ServiceGrids(namespace).Get(name)
	// fedSg has been deleted
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		klog.Errorf("err is %#v", err)
		return err
	}

	if fedSg != nil && fedSg.DeletionTimestamp != nil {
		klog.V(2).Infof("fed service grid %v has been deleted", key)
		if existFedDg, _ := sgc.svcGridLister.ServiceGrids(fedSg.Labels[common.FedTargetNameSpace]).Get(name); existFedDg != nil {
			err = sgc.crdClient.SuperedgeV1().ServiceGrids(fedSg.Labels[common.FedTargetNameSpace]).Delete(context.TODO(), name, metav1.DeleteOptions{})
			if err != nil && !apierrors.IsNotFound(err) {
				klog.Errorf("delete sg err is %#v", err)
				return err
			}

			patch := `{"metadata":{"finalizers":[]}}`
			if _, err := sgc.fedCrdClient.SuperedgeV1().ServiceGrids(namespace).Patch(
				context.TODO(), name, types.MergePatchType, []byte(patch), metav1.PatchOptions{}); err != nil {
				return fmt.Errorf("Patching fed ServiceGrids: %s %s, error: %v\n", namespace, name, err)
			}
		}
		return nil
	}

	if fedSg.Spec.GridUniqKey == "" {
		sgc.eventRecorder.Eventf(fedSg, corev1.EventTypeWarning, "Empty", "This service-grid has an empty grid key")
		return nil
	}

	sgCopy := fedSg.DeepCopy()

	var flag = false
	nodes, err := sgc.nodeLister.List(labels.Everything())
	if err != nil {
		return err
	}
	for _, n := range nodes {
		if _, ok := n.Labels[sgCopy.Spec.GridUniqKey]; ok {
			flag = true
			break
		}
	}

	if flag {
		existedSg, err := sgc.svcGridLister.ServiceGrids(sgCopy.Labels[common.FedTargetNameSpace]).Get(name)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// create ServiceGrid
				sgCopy.Namespace = sgCopy.Labels[common.FedTargetNameSpace]
				delete(sgCopy.Labels, common.FedrationDisKey)
				sgCopy.OwnerReferences = []metav1.OwnerReference{}
				sgCopy.Finalizers = []string{}
				sgCopy.ResourceVersion = ""
				_, err := sgc.crdClient.SuperedgeV1().ServiceGrids(sgCopy.Labels[common.FedTargetNameSpace]).Create(context.TODO(), sgCopy, metav1.CreateOptions{})
				if err != nil {
					klog.Errorf("create service grid %s err: %v", name, err)
					return err
				}
				updateFed := fedSg.DeepCopy()
				if len(updateFed.Finalizers) == 0 {
					patch := `{"metadata":{"finalizers":["dis"]}}`
					if _, err := sgc.fedCrdClient.SuperedgeV1().ServiceGrids(namespace).Patch(
						context.TODO(), name, types.MergePatchType, []byte(patch), metav1.PatchOptions{}); err != nil {
						return fmt.Errorf("Patching fed ServiceGrids add finalizers: %s %s, error: %v\n", namespace, name, err)
					}
				}
			} else {
				return err
			}
		} else {
			if !commonutil.DeepContains(existedSg.Spec, sgCopy.Spec) {
				existedSg.Spec = sgCopy.Spec
				klog.Infof("service %s template changed", sgCopy.Name)
				out, _ := json.Marshal(sgCopy.Spec)
				klog.V(5).Infof("ServiceGridToUpdate is %s", string(out))
				out, _ = json.Marshal(existedSg.Spec)
				klog.V(5).Infof("existedServiceGrid is %s", string(out))
				_, err := sgc.crdClient.SuperedgeV1().ServiceGrids(existedSg.Namespace).Update(context.TODO(), existedSg, metav1.UpdateOptions{})
				if err != nil {
					return err
				}
			}
		}
	} else {
		if _, err := sgc.svcGridLister.ServiceGrids(sgCopy.Labels[common.FedTargetNameSpace]).Get(name); err == nil {
			err = sgc.crdClient.SuperedgeV1().ServiceGrids(sgCopy.Labels[common.FedTargetNameSpace]).Delete(context.TODO(), name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sgc *FedServiceGridController) addNode(obj interface{}) {
	node := obj.(*corev1.Node)
	if node.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		sgc.deleteNode(node)
		return
	}
	sgs, err := sgc.fedSvcGridLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("get fedDpGrid err %v", err)
		return
	}

	for _, sg := range sgs {
		if _, exist := node.Labels[sg.Spec.GridUniqKey]; exist {
			sgc.enqueueServiceGrid(sg)
		}
	}
}

func (sgc *FedServiceGridController) updateNode(oldObj, newObj interface{}) {
	oldNode := oldObj.(*corev1.Node)
	curNode := newObj.(*corev1.Node)
	if curNode.ResourceVersion == oldNode.ResourceVersion {
		// Periodic resync will send update events for all known Nodes.
		// Two different versions of the same Node will always have different RVs.
		return
	}
	labelChanged := !reflect.DeepEqual(curNode.Labels, oldNode.Labels)
	// Only handles nodes whose label has changed.
	if labelChanged {
		sgs, err := sgc.fedSvcGridLister.List(labels.Everything())
		if err != nil {
			klog.Errorf("get fedSvcGrid err %v", err)
			return
		}

		for _, sg := range sgs {
			_, oldexist := oldNode.Labels[sg.Spec.GridUniqKey]
			_, newexist := curNode.Labels[sg.Spec.GridUniqKey]
			if oldexist || newexist {
				sgc.enqueueServiceGrid(sg)
			}
		}
	}
}

func (sgc *FedServiceGridController) deleteNode(obj interface{}) {
	node, ok := obj.(*corev1.Node)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		node, ok = tombstone.Obj.(*corev1.Node)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object is not a node %#v", obj))
			return
		}
	}

	sgs, err := sgc.fedSvcGridLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("get fedDpGrid err %v", err)
		return
	}

	for _, sg := range sgs {
		if _, exist := node.Labels[sg.Spec.GridUniqKey]; exist {
			sgc.enqueueServiceGrid(sg)
		}
	}
}

func (sgc *FedServiceGridController) addServiceGrid(obj interface{}) {
	sg := obj.(*crdv1.ServiceGrid)
	klog.V(4).Infof("Adding service grid %s", sg.Name)
	if _, ok := sg.Labels[common.FedTargetNameSpace]; ok {
		fedsvcgrid, err := sgc.fedSvcGridLister.ServiceGrids(sgc.namespace).Get(sg.Name)
		if err != nil {
			klog.Errorf("err get feddg: %v", err)
			return
		}
		sgc.enqueueServiceGrid(fedsvcgrid)
	}
}

func (sgc *FedServiceGridController) updateServiceGrid(oldObj, newObj interface{}) {
	oldSvcGrid := oldObj.(*crdv1.ServiceGrid)
	curSvcGrid := newObj.(*crdv1.ServiceGrid)
	klog.V(4).Infof("Updating fed service grid %s", oldSvcGrid.Name)
	if curSvcGrid.ResourceVersion == oldSvcGrid.ResourceVersion {
		// Periodic resync will send update events for all known ServiceGrids.
		// Two different versions of the same ServiceGrid will always have different RVs.
		return
	}
	if _, ok := curSvcGrid.Labels[common.FedTargetNameSpace]; ok {
		fedsg, err := sgc.fedSvcGridLister.ServiceGrids(sgc.namespace).Get(curSvcGrid.Name)
		if err != nil {
			klog.Errorf("err get fedsg: %v", err)
			return
		}
		sgc.enqueueServiceGrid(fedsg)
	}
}

func (sgc *FedServiceGridController) deleteServiceGrid(obj interface{}) {
	sg, ok := obj.(*crdv1.ServiceGrid)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		sg, ok = tombstone.Obj.(*crdv1.ServiceGrid)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object that is not a service grid %#v", obj))
			return
		}
	}
	klog.V(4).Infof("Deleting service grid %s", sg.Name)
	if _, ok := sg.Labels[common.FedTargetNameSpace]; ok {
		fedsg, err := sgc.fedSvcGridLister.ServiceGrids(sgc.namespace).Get(sg.Name)
		if err != nil {
			klog.Errorf("err get feddg: %v", err)
			return
		}
		sgc.enqueueServiceGrid(fedsg)
	}
}
