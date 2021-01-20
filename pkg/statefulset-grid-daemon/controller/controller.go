package controller

import (
	"fmt"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	crdclientset "github.com/superedge/superedge/pkg/application-grid-controller/generated/clientset/versioned"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
	"k8s.io/klog"
	"time"
)

var controllerKind = appsv1.SchemeGroupVersion.WithKind("StatefulSet")

type StatefulSetController struct {
	hostName   string

	nodeLister         corelisters.NodeLister
	nodeListerSynced   cache.InformerSynced

	podLister         corelisters.PodLister
	podListerSynced    cache.InformerSynced

	setLister           appslisters.StatefulSetLister
	setListerSynced     cache.InformerSynced

	eventRecorder record.EventRecorder
	queue         workqueue.RateLimitingInterface
	kubeClient    clientset.Interface
	crdClient     crdclientset.Interface

	syncHandler func(dKey string) error
	enqueueStatefulset func(statefulset *appsv1.StatefulSet)
}

func NewStatefulSetController(nodeInformer coreinformers.NodeInformer, podInformer coreinformers.PodInformer,
	statefulSetInformer appsinformers.StatefulSetInformer, kubeClient clientset.Interface,
	hostName string) *StatefulSetController {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1.EventSinkImpl{
		Interface: kubeClient.CoreV1().Events(""),
	})

	setc := &StatefulSetController{
		hostName: hostName,
		kubeClient: kubeClient,
		eventRecorder: eventBroadcaster.NewRecorder(scheme.Scheme,
			corev1.EventSource{Component: "statefulset-grid-daemon"}),
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(),
			"statefulset-grid-daemon"),
	}

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    setc.addNode,
		UpdateFunc: setc.updateNode,
		DeleteFunc: setc.deleteNode,
	})

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    setc.addPod,
		UpdateFunc: setc.updatePod,
		DeleteFunc: setc.deletePod,
	})

	statefulSetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    setc.addStatefulset,
		UpdateFunc: setc.updateStatefulset,
		DeleteFunc: setc.deleteStatefulset,
	})

	setc.syncHandler = setc.syncStatefulSet
	setc.enqueueStatefulset = setc.enqueue

	setc.nodeLister = nodeInformer.Lister()
	setc.nodeListerSynced = nodeInformer.Informer().HasSynced

	setc.podLister = podInformer.Lister()
	setc.podListerSynced = podInformer.Informer().HasSynced

	setc.setLister = statefulSetInformer.Lister()
	setc.setListerSynced = statefulSetInformer.Informer().HasSynced

	return setc
}

func (setc *StatefulSetController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer setc.queue.ShutDown()

	klog.Infof("Starting statefulset grid daemon")
	defer klog.Infof("Shutting down statefulset grid daemon")

	if !cache.WaitForNamedCacheSync("deployment-grid", stopCh,
		setc.nodeListerSynced, setc.podListerSynced, setc.setListerSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(setc.worker, time.Second, stopCh)
	}
	<-stopCh
}

func (setc *StatefulSetController) worker() {
	for setc.processNextWorkItem() {
	}
}

func (setc *StatefulSetController) processNextWorkItem() bool {
	key, quit := setc.queue.Get()
	if quit {
		return false
	}
	defer setc.queue.Done(key)

	err := setc.syncHandler(key.(string))
	setc.handleErr(err, key)

	return true
}

func (setc *StatefulSetController) handleErr(err error, key interface{}) {
	if err == nil {
		setc.queue.Forget(key)
		return
	}

	if setc.queue.NumRequeues(key) < common.MaxRetries {
		klog.V(2).Infof("Error syncing statefulset %v: %v", key, err)
		setc.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	klog.V(2).Infof("Dropping statefulset %q out of the queue: %v", key, err)
	setc.queue.Forget(key)
}

func (setc *StatefulSetController)syncStatefulSet(key string) error{

}

func (setc *StatefulSetController) enqueue(set *appsv1.StatefulSet) {
	key, err := controller.KeyFunc(set)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", set, err))
		return
	}

	setc.queue.Add(key)
}