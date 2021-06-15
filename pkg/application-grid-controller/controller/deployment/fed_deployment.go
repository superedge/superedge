package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	commonutil "github.com/superedge/superedge/pkg/application-grid-controller/util"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
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

type FedDeploymentGridController struct {
	namespace       string
	dpGridLister    crdv1listers.DeploymentGridLister
	fedDpGridLister crdv1listers.DeploymentGridLister
	nodeLister      corelisters.NodeLister

	dpGridListerSynced    cache.InformerSynced
	fedDpGridListerSynced cache.InformerSynced
	nodeListerSynced      cache.InformerSynced

	eventRecorder record.EventRecorder
	queue         workqueue.RateLimitingInterface
	kubeClient    clientset.Interface
	fedCrdClient  crdclientset.Interface
	crdClient     crdclientset.Interface

	// To allow injection of syncDeploymentGrid for testing.
	syncHandler func(dKey string) error
	// used for unit testing
	enqueueDeploymentGrid func(deploymentGrid *crdv1.DeploymentGrid)
}

func NewFedDeploymentGridController(dpGridInformer crdinformers.DeploymentGridInformer,
	fedDpGridInformer crdinformers.DeploymentGridInformer, nodeInformer coreinformers.NodeInformer,
	fedCrdClient, crdClient crdclientset.Interface, kubeClient clientset.Interface, dedicatedNameSpace string) *FedDeploymentGridController {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1.EventSinkImpl{
		Interface: kubeClient.CoreV1().Events(""),
	})

	fedDgc := &FedDeploymentGridController{
		namespace:    dedicatedNameSpace,
		kubeClient:   kubeClient,
		crdClient:    crdClient,
		fedCrdClient: fedCrdClient,
		eventRecorder: eventBroadcaster.NewRecorder(scheme.Scheme,
			corev1.EventSource{Component: "fed-deployment-grid-controller"}),
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(),
			"fed-deployment-grid-controller"),
	}

	dpGridInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    fedDgc.addDeploymentGrid,
		UpdateFunc: fedDgc.updateDeploymentGrid,
		DeleteFunc: fedDgc.deleteDeploymentGrid,
	})

	fedDpGridInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    fedDgc.addFedDeploymentGrid,
		UpdateFunc: fedDgc.updateFedDeploymentGrid,
		DeleteFunc: fedDgc.deleteFedDeploymentGrid,
	})

	// TODO: node label changed causing deployment deletion?
	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    fedDgc.addNode,
		UpdateFunc: fedDgc.updateNode,
		DeleteFunc: fedDgc.deleteNode,
	})

	fedDgc.syncHandler = fedDgc.syncFedDeploymentGrid
	fedDgc.enqueueDeploymentGrid = fedDgc.enqueue

	fedDgc.dpGridLister = dpGridInformer.Lister()
	fedDgc.dpGridListerSynced = dpGridInformer.Informer().HasSynced

	fedDgc.nodeLister = nodeInformer.Lister()
	fedDgc.nodeListerSynced = nodeInformer.Informer().HasSynced

	fedDgc.fedDpGridLister = fedDpGridInformer.Lister()
	fedDgc.fedDpGridListerSynced = fedDpGridInformer.Informer().HasSynced

	return fedDgc
}

func (dgc *FedDeploymentGridController) addFedDeploymentGrid(obj interface{}) {
	dg := obj.(*crdv1.DeploymentGrid)
	klog.V(4).Infof("Adding fed deployment grid %s", dg.Name)
	dgc.enqueueDeploymentGrid(dg)
}

func (dgc *FedDeploymentGridController) updateFedDeploymentGrid(oldObj, newObj interface{}) {
	oldDg := oldObj.(*crdv1.DeploymentGrid)
	curDg := newObj.(*crdv1.DeploymentGrid)
	klog.V(4).Infof("Updating fed deployment grid %s %s", oldDg.Namespace, oldDg.Name)
	if curDg.ResourceVersion == oldDg.ResourceVersion {
		// Periodic resync will send update events for all known DeploymentGrids.
		// Two different versions of the same DeploymentGrid will always have different RVs.
		return
	}
	dgc.enqueueDeploymentGrid(curDg)
}

func (dgc *FedDeploymentGridController) deleteFedDeploymentGrid(obj interface{}) {
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
	klog.V(4).Infof("Deleteing fed deployment grid %s %s", dg.Namespace, dg.Name)
	dgc.enqueueDeploymentGrid(dg)
}

func (dgc *FedDeploymentGridController) enqueue(deploymentGrid *crdv1.DeploymentGrid) {
	key, err := controller.KeyFunc(deploymentGrid)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", deploymentGrid, err))
		return
	}

	dgc.queue.Add(key)
}

func (dgc *FedDeploymentGridController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer dgc.queue.ShutDown()

	klog.Infof("Starting fed deployment grid controller")
	defer klog.Infof("Shutting down fed deployment grid controller")

	if !cache.WaitForNamedCacheSync("fed-deployment-grid", stopCh,
		dgc.dpGridListerSynced, dgc.fedDpGridListerSynced, dgc.nodeListerSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(dgc.worker, time.Second, stopCh)
	}
	<-stopCh
}

func (dgc *FedDeploymentGridController) worker() {
	for dgc.processNextWorkItem() {
	}
}

func (dgc *FedDeploymentGridController) processNextWorkItem() bool {
	key, quit := dgc.queue.Get()
	if quit {
		return false
	}
	defer dgc.queue.Done(key)

	err := dgc.syncHandler(key.(string))
	dgc.handleErr(err, key)

	return true
}

func (dgc *FedDeploymentGridController) handleErr(err error, key interface{}) {
	if err == nil {
		dgc.queue.Forget(key)
		return
	}

	if dgc.queue.NumRequeues(key) < common.MaxRetries {
		klog.V(2).Infof("Error syncing fed deployment grid %v: %v", key, err)
		dgc.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	klog.V(2).Infof("Dropping fed deployment grid %q out of the queue: %v", key, err)
	dgc.queue.Forget(key)
}

func (dgc *FedDeploymentGridController) syncFedDeploymentGrid(key string) error {
	startTime := time.Now()
	klog.V(4).Infof("Started syncing fed deployment grid %q (%v)", key, startTime)
	defer func() {
		klog.V(4).Infof("Finished syncing fed deployment grid %q (%v)", key, time.Since(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	fedDg, err := dgc.fedDpGridLister.DeploymentGrids(namespace).Get(name)
	// fedDg has been deleted
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		klog.Errorf("err is %#v", err)
		return err
	}

	if fedDg != nil && fedDg.DeletionTimestamp != nil {
		klog.V(2).Infof("fed deployment grid %v has been deleted", key)
		if existFedDg, _ := dgc.dpGridLister.DeploymentGrids(fedDg.Labels[common.FedTargetNameSpace]).Get(name); existFedDg != nil {
			err = dgc.crdClient.SuperedgeV1().DeploymentGrids(fedDg.Labels[common.FedTargetNameSpace]).Delete(context.TODO(), name, metav1.DeleteOptions{})
			if err != nil && !apierrors.IsNotFound(err) {
				klog.Errorf("delete dg err is %#v", err)
				return err
			}

			patch := `{"metadata":{"finalizers":[]}}`
			if _, err := dgc.fedCrdClient.SuperedgeV1().DeploymentGrids(namespace).Patch(
				context.TODO(), name, types.MergePatchType, []byte(patch), metav1.PatchOptions{}); err != nil {
				return fmt.Errorf("Patching fed DeploymentGrids: %s %s, error: %v\n", namespace, name, err)
			}
		}
		return nil
	}

	if fedDg.Spec.GridUniqKey == "" {
		dgc.eventRecorder.Eventf(fedDg, corev1.EventTypeWarning, "Empty", "This deployment-grid has an empty grid key")
		return nil
	}

	dgCopy := fedDg.DeepCopy()

	var flag = false
	nodes, err := dgc.nodeLister.List(labels.Everything())
	if err != nil {
		return err
	}
	for _, n := range nodes {
		if _, ok := n.Labels[dgCopy.Spec.GridUniqKey]; ok {
			flag = true
			break
		}
	}

	if flag {
		existedDg, err := dgc.dpGridLister.DeploymentGrids(dgCopy.Labels[common.FedTargetNameSpace]).Get(name)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// create deploymentGrid
				dgCopy.Namespace = dgCopy.Labels[common.FedTargetNameSpace]
				delete(dgCopy.Labels, common.FedrationDisKey)
				dgCopy.OwnerReferences = []metav1.OwnerReference{}
				dgCopy.Finalizers = []string{}
				dgCopy.ResourceVersion = ""
				_, err := dgc.crdClient.SuperedgeV1().DeploymentGrids(dgCopy.Labels[common.FedTargetNameSpace]).Create(context.TODO(), dgCopy, metav1.CreateOptions{})
				if err != nil {
					klog.Errorf("create deployment grid %s err: %v", name, err)
					return err
				}
				updateFed := fedDg.DeepCopy()
				if len(updateFed.Finalizers) == 0 {
					patch := `{"metadata":{"finalizers":["dis"]}}`
					if _, err := dgc.fedCrdClient.SuperedgeV1().DeploymentGrids(namespace).Patch(
						context.TODO(), name, types.MergePatchType, []byte(patch), metav1.PatchOptions{}); err != nil {
						return fmt.Errorf("Patching fed DeploymentGrids add finalizers: %s %s, error: %v\n", namespace, name, err)
					}
				}
			} else {
				return err
			}
		} else {
			if !commonutil.DeepContains(existedDg.Spec, dgCopy.Spec) {
				existedDg.Spec = dgCopy.Spec
				klog.Infof("deployment %s template changed", dgCopy.Name)
				out, _ := json.Marshal(dgCopy.Spec)
				klog.V(5).Infof("deploymentGridToUpdate is %s", string(out))
				out, _ = json.Marshal(existedDg.Spec)
				klog.V(5).Infof("existedDeploymentGrid is %s", string(out))
				_, err := dgc.crdClient.SuperedgeV1().DeploymentGrids(existedDg.Namespace).Update(context.TODO(), existedDg, metav1.UpdateOptions{})
				if err != nil {
					return err
				}
			}
		}
	} else {
		if _, err := dgc.dpGridLister.DeploymentGrids(dgCopy.Labels[common.FedTargetNameSpace]).Get(name); err == nil {
			err = dgc.crdClient.SuperedgeV1().DeploymentGrids(dgCopy.Labels[common.FedTargetNameSpace]).Delete(context.TODO(), name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
	}
	return dgc.syncFedStatus(fedDg.DeepCopy())
}

func (dgc *FedDeploymentGridController) addNode(obj interface{}) {
	node := obj.(*corev1.Node)
	if node.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		dgc.deleteNode(node)
		return
	}
	dgs, err := dgc.fedDpGridLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("get fedDpGrid err %v", err)
		return
	}

	for _, dg := range dgs {
		if _, exist := node.Labels[dg.Spec.GridUniqKey]; exist {
			dgc.enqueueDeploymentGrid(dg)
		}
	}
}

func (dgc *FedDeploymentGridController) updateNode(oldObj, newObj interface{}) {
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
		dgs, err := dgc.fedDpGridLister.List(labels.Everything())
		if err != nil {
			klog.Errorf("get fedDpGrid err %v", err)
			return
		}

		for _, dg := range dgs {
			_, oldexist := oldNode.Labels[dg.Spec.GridUniqKey]
			_, newexist := curNode.Labels[dg.Spec.GridUniqKey]
			if oldexist || newexist {
				dgc.enqueueDeploymentGrid(dg)
			}
		}
	}
}

func (dgc *FedDeploymentGridController) deleteNode(obj interface{}) {
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

	dgs, err := dgc.fedDpGridLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("get fedDpGrid err %v", err)
		return
	}

	for _, dg := range dgs {
		if _, exist := node.Labels[dg.Spec.GridUniqKey]; exist {
			dgc.enqueueDeploymentGrid(dg)
		}
	}
}

func (dgc *FedDeploymentGridController) addDeploymentGrid(obj interface{}) {
	dg := obj.(*crdv1.DeploymentGrid)
	klog.V(4).Infof("Adding deployment grid %s", dg.Name)
	if _, ok := dg.Labels[common.FedTargetNameSpace]; ok {
		feddg, err := dgc.fedDpGridLister.DeploymentGrids(dgc.namespace).Get(dg.Name)
		if err != nil {
			klog.Errorf("err get feddg: %v", err)
			return
		}
		dgc.enqueueDeploymentGrid(feddg)
	}
}

func (dgc *FedDeploymentGridController) updateDeploymentGrid(oldObj, newObj interface{}) {
	oldDg := oldObj.(*crdv1.DeploymentGrid)
	curDg := newObj.(*crdv1.DeploymentGrid)
	klog.V(4).Infof("Updating fed deployment grid %s", oldDg.Name)
	if curDg.ResourceVersion == oldDg.ResourceVersion {
		// Periodic resync will send update events for all known DeploymentGrids.
		// Two different versions of the same DeploymentGrid will always have different RVs.
		return
	}
	if _, ok := curDg.Labels[common.FedTargetNameSpace]; ok {
		feddg, err := dgc.fedDpGridLister.DeploymentGrids(dgc.namespace).Get(curDg.Name)
		if err != nil {
			klog.Errorf("err get feddg: %v", err)
			return
		}
		dgc.enqueueDeploymentGrid(feddg)
	}
}

func (dgc *FedDeploymentGridController) deleteDeploymentGrid(obj interface{}) {
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
	if _, ok := dg.Labels[common.FedTargetNameSpace]; ok {
		feddg, err := dgc.fedDpGridLister.DeploymentGrids(dgc.namespace).Get(dg.Name)
		if err != nil {
			klog.Errorf("err get feddg: %v", err)
			return
		}
		dgc.enqueueDeploymentGrid(feddg)
	}
}

func (dgc *FedDeploymentGridController) syncFedStatus(dg *crdv1.DeploymentGrid) error {
	disdg, err := dgc.dpGridLister.DeploymentGrids(dg.Labels[common.FedTargetNameSpace]).Get(dg.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if !apiequality.Semantic.DeepEqual(dg.Status.States, disdg.Status.States) {
		dg.Status.States = disdg.Status.States
		klog.V(4).Infof("fed status: old status is %#v", dg.Status.States)
		klog.V(4).Infof("fed status: Updating deployment grid %s/%s status %#v", dg.Namespace, dg.Name, disdg.Status.States)
		_, err = dgc.fedCrdClient.SuperedgeV1().DeploymentGrids(dg.Namespace).UpdateStatus(context.TODO(), dg, metav1.UpdateOptions{})
		if err != nil && errors.IsConflict(err) {
			return nil
		}
		return err
	}
	return nil
}
