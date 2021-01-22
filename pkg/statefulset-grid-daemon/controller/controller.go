package controller

import (
	"fmt"
	"github.com/superedge/superedge/pkg/statefulset-grid-daemon/common"
	"github.com/superedge/superedge/pkg/statefulset-grid-daemon/util"
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
	"k8s.io/klog"
	"strings"
	"time"
)

var controllerKind = appsv1.SchemeGroupVersion.WithKind("StatefulSet")

var (
	KeyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc
)

type StatefulSetController struct {
	hostName   string
	hosts   *util.Hosts

	nodeLister         corelisters.NodeLister
	nodeListerSynced   cache.InformerSynced

	podLister         corelisters.PodLister
	podListerSynced    cache.InformerSynced

	setLister           appslisters.StatefulSetLister
	setListerSynced     cache.InformerSynced

	eventRecorder record.EventRecorder
	queue         workqueue.RateLimitingInterface
	kubeClient    clientset.Interface

	syncHandler func(dKey string) error
	enqueueStatefulset func(statefulset *appsv1.StatefulSet)
}

func NewStatefulSetController(nodeInformer coreinformers.NodeInformer, podInformer coreinformers.PodInformer,
	statefulSetInformer appsinformers.StatefulSetInformer, kubeClient clientset.Interface,
	hostName string, hosts *util.Hosts) *StatefulSetController {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1.EventSinkImpl{
		Interface: kubeClient.CoreV1().Events(""),
	})

	setc := &StatefulSetController{
		hostName: hostName,
		hosts: hosts,
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
	startTime := time.Now()
	klog.V(4).Infof("Started syncing statefulset %q (%v)", key, startTime)
	defer func() {
		klog.V(4).Infof("Finished syncing statefulset %q (%v)", key, time.Since(startTime))
	}()

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	set, err := setc.setLister.StatefulSets(ns).Get(name)
	if errors.IsNotFound(err) {
		klog.V(2).Infof("statefulset %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	setcopy := set.DeepCopy()

	if rel, err := setc.IsConcernedStatefulSet(setcopy); err != nil || !rel {
		return nil
	}

	podToHost := []*corev1.Pod{}

	podlist, err :=  setc.podLister.Pods(setcopy.Namespace).List(labels.Everything())
	if err !=nil {
		klog.Errorf("get podlist err %v", err)
		return err
	}

	for _, p := range podlist {
		if len(p.OwnerReferences) != 0 && p.OwnerReferences[0].Name == setcopy.Name && p.OwnerReferences[0].Kind == controllerKind.Kind {
			podToHost = append(podToHost, p)
		}
	}
	var PodInfoToHost = make(map[string]string)
	//  <ControllerRef>-<gridvalue>-<num>.<ns>.<svc>
	ControllerRef := metav1.GetControllerOf(setcopy)
	if ControllerRef != nil {
		gridvalue := setcopy.Name[len(ControllerRef.Name)+1:]
		for _, p := range podToHost{
			index := strings.Index(p.Name, gridvalue)
			if index == -1 {
				return nil
			}
			//ip: <ControllerRef>-<num>.<ns>.<svc>
			podNameTohost := p.Name[0:index]+p.Name[index+len(gridvalue)+1:] + "."+ setcopy.Spec.ServiceName
			PodInfoToHost[p.Status.PodIP] = podNameTohost

			if err := setc.hosts.UpdateHosts(PodInfoToHost, p.Namespace, ControllerRef.Name, setcopy.Spec.ServiceName); err != nil{
				klog.Errorf("update err: %v", err)
				return err
			}
		}
	}
	return nil
}

func (setc *StatefulSetController) enqueue(set *appsv1.StatefulSet) {
	key, err := KeyFunc(set)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", set, err))
		return
	}

	setc.queue.Add(key)
}