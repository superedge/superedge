package statefulset

import (
	"fmt"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	"time"

	"k8s.io/klog"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

type StatefulSetGridController struct {
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
	enqueueStatefulSetGrid func(setGrid *crdv1.StatefulSetGrid)
}

func NewStatefulSetGridController(setGridInformer crdinformers.StatefulSetGridInformer, setInformer appsinformers.StatefulSetInformer,
	nodeInformer coreinformers.NodeInformer, kubeClient clientset.Interface, crdClient crdclientset.Interface) *StatefulSetGridController {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1.EventSinkImpl{
		Interface: kubeClient.CoreV1().Events(""),
	})

	setGridController := &StatefulSetGridController{
		kubeClient: kubeClient,
		crdClient:  crdClient,
		eventRecorder: eventBroadcaster.NewRecorder(scheme.Scheme,
			corev1.EventSource{Component: "statefulset-grid-controller"}),
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "statefulset-grid-controller"),
	}

	// TODO
	/*setGridInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    setGridController.addStatefulSetGrid,
		UpdateFunc: setGridController.updateStatefulSetGrid,
		DeleteFunc: setGridController.deleteStatefulSetGrid,
	})

	setInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    setGridController.addStatefulSet,
		UpdateFunc: setGridController.updateStatefulSet,
		DeleteFunc: setGridController.deleteStatefulSet,
	})

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    setGridController.addNode,
		UpdateFunc: setGridController.updateNode,
		DeleteFunc: setGridController.deleteNode,
	})*/

	setGridController.syncHandler = setGridController.syncStatefulSetGrid
	setGridController.enqueueStatefulSetGrid = setGridController.enqueue

	setGridController.setGridLister = setGridInformer.Lister()
	setGridController.setGridListerSynced = setGridInformer.Informer().HasSynced

	setGridController.setLister = setInformer.Lister()
	setGridController.setListerSynced = setInformer.Informer().HasSynced

	setGridController.nodeLister = nodeInformer.Lister()
	setGridController.nodeListerSynced = nodeInformer.Informer().HasSynced

	return setGridController
}

func (setGridController *StatefulSetGridController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer setGridController.queue.ShutDown()

	klog.Infof("Starting statefulset grid controller")
	defer klog.Infof("Shutting down statefulset grid controller")

	if !cache.WaitForNamedCacheSync("statefulset-grid", stopCh,
		setGridController.setGridListerSynced, setGridController.setListerSynced, setGridController.nodeListerSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(setGridController.runWorker, time.Second, stopCh)
	}
	<-stopCh
}

func (setGridController *StatefulSetGridController) runWorker() {
	for setGridController.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the queue and
// attempt to process it, by calling the syncHandler.
func (setGridController *StatefulSetGridController) processNextWorkItem() bool {
	obj, shutdown := setGridController.queue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.queue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the queue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the queue and attempted again after a back-off
		// period.
		defer setGridController.queue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the queue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// queue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// queue.
		if key, ok = obj.(string); !ok {
			// As the item in the queue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			setGridController.queue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in queue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// statefulSetGrid resource to be synced.
		if err := setGridController.syncHandler(key); err != nil {
			// Put the item back on the queue to handle any transient errors.
			if setGridController.queue.NumRequeues(key) < common.MaxRetries {
				setGridController.queue.AddRateLimited(key)
				return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
			} else {
				setGridController.queue.Forget(obj)
				return fmt.Errorf("stop syncing '%s': %s, dropping", key, err.Error())
			}
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		setGridController.queue.Forget(obj)
		klog.V(2).Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (setGridController *StatefulSetGridController) syncStatefulSetGrid(key string) error {
	startTime := time.Now()
	klog.V(4).Infof("Started syncing statefulset-grid %s (%v)", key, startTime)
	defer func() {
		klog.V(4).Infof("Finished syncing statefulset-grid %s (%v)", key, time.Since(startTime))
	}()

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Invalid resource key: %s", key))
		return err
	}

	setGrid, err := setGridController.setGridLister.StatefulSetGrids(ns).Get(name)
	if err != nil {
		// The statefulSetGrid resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("statefulset-grid '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	if setGrid.Spec.GridUniqKey == "" {
		setGridController.eventRecorder.Eventf(setGrid, corev1.EventTypeWarning, "Empty", "This statefulset-grid has an empty grid key")
		return nil
	}

	// TODO
	// get statefulset workload list of this grid
	/*setList, err := setGridController.getStatefulSetForGrid(setGrid)
	if err != nil {
		return err
	}

	// get all grid labels in all nodes
	gridValues, err := util.GetGridValuesFromNode(setGridController.nodeLister, setGrid.Spec.GridUniqKey)
	if err != nil {
		return err
	}

	// sync statefulset-grid workload status
	if setGrid.DeletionTimestamp != nil {
		return setGridController.syncStatus(setGrid, setList, gridValues)
	}

	// sync statefulset-grid workload relevant statefusets
	return setGridController.reconcile(setGrid, setList, gridValues)*/
	return nil
}

func (setGridController *StatefulSetGridController) enqueue(setGrid *crdv1.StatefulSetGrid) {
	key, err := controller.KeyFunc(setGrid)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", setGrid, err))
		return
	}

	setGridController.queue.Add(key)
}
