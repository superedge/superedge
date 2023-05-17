/*
Copyright 2021 The SuperEdge Authors.

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
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/superedge/superedge/pkg/site-manager/constant"
	deleter "github.com/superedge/superedge/pkg/site-manager/controller/deleter"

	crdClientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	crdinformers "github.com/superedge/superedge/pkg/site-manager/generated/informers/externalversions/site.superedge.io/v1alpha2"
	crdv1listers "github.com/superedge/superedge/pkg/site-manager/generated/listers/site.superedge.io/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"

	"k8s.io/apimachinery/pkg/util/wait"
	appinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	applisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	sitev1alpha2 "github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha2"
	"github.com/superedge/superedge/pkg/site-manager/controller/unitcluster"
	"github.com/superedge/superedge/pkg/site-manager/utils"

	"github.com/superedge/superedge/pkg/util"
)

type NodeUnitController struct {
	nodeLister       corelisters.NodeLister
	nodeListerSynced cache.InformerSynced

	dsListter      applisters.DaemonSetLister
	dsListerSynced cache.InformerSynced

	nodeUnitLister       crdv1listers.NodeUnitLister
	nodeUnitListerSynced cache.InformerSynced

	eventRecorder record.EventRecorder
	queue         workqueue.RateLimitingInterface
	kubeClient    clientset.Interface
	crdClient     *crdClientset.Clientset

	syncHandler     func(key string) error
	enqueueNodeUnit func(name string)
	nodeUnitDeleter *deleter.NodeUnitDeleter
	kinsController  *unitcluster.KinsController
}

func NewNodeUnitController(
	nodeInformer coreinformers.NodeInformer,
	dsInformer appinformers.DaemonSetInformer,
	nodeUnitInformer crdinformers.NodeUnitInformer,
	nodeGroupInformer crdinformers.NodeGroupInformer,
	kubeClient clientset.Interface,
	crdClient *crdClientset.Clientset,
) *NodeUnitController {

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1.EventSinkImpl{
		Interface: kubeClient.CoreV1().Events(""),
	})

	err := sitev1alpha2.AddToScheme(scheme.Scheme)
	if err != nil {
		klog.Error(err)
	}
	nodeUnitController := &NodeUnitController{
		kubeClient:    kubeClient,
		crdClient:     crdClient,
		eventRecorder: eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "nodeunit-controller"}),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "nodeunit-controller"),
	}

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    nodeUnitController.addNode,
		UpdateFunc: nodeUnitController.updateNode,
		DeleteFunc: nodeUnitController.deleteNode,
	})

	dsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    nodeUnitController.addDaemonSet,
		UpdateFunc: nodeUnitController.updateDaemonSet,
		DeleteFunc: nodeUnitController.deleteDaemonSet,
	})

	nodeUnitInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    nodeUnitController.addNodeUnit,
		UpdateFunc: nodeUnitController.updateNodeUnit,
		DeleteFunc: nodeUnitController.deleteNodeUnit,
	})

	nodeUnitController.syncHandler = nodeUnitController.syncUnit
	nodeUnitController.enqueueNodeUnit = nodeUnitController.enqueue

	nodeUnitController.nodeLister = nodeInformer.Lister()
	nodeUnitController.nodeListerSynced = nodeInformer.Informer().HasSynced

	nodeUnitController.nodeUnitLister = nodeUnitInformer.Lister()
	nodeUnitController.nodeUnitListerSynced = nodeUnitInformer.Informer().HasSynced

	nodeUnitController.nodeUnitDeleter = deleter.NewNodeUnitDeleter(
		kubeClient,
		crdClient,
		nodeInformer.Lister(),
		nodeInformer.Informer().HasSynced,
		NodeUnitFinalizerID,
		workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "nodeunit-deleter"),
	)
	nodeUnitController.kinsController = unitcluster.NewKinsController(
		kubeClient,
		crdClient,
		nodeInformer.Lister(),
		dsInformer.Lister(),
		nodeUnitInformer.Lister(),
	)
	klog.V(4).Infof("Site-manager set handler success")

	return nodeUnitController
}

func (c *NodeUnitController) Run(workers, syncPeriodAsWhole int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.V(1).Infof("Starting site-manager daemon")
	defer klog.V(1).Infof("Shutting down site-manager daemon")

	if !cache.WaitForNamedCacheSync("site-manager-daemon", stopCh,
		c.nodeListerSynced, c.nodeUnitListerSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
		klog.V(4).Infof("Site-manager set worker-%d success", i)
	}
	// run deleter
	c.nodeUnitDeleter.Run(stopCh)

	<-stopCh
}

func (c *NodeUnitController) worker() {
	for c.processNextWorkItem() {
	}
}

func (c *NodeUnitController) processNextWorkItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)
	klog.V(4).Infof("Get siteManager queue key: %s", key)
	err := c.syncHandler(key.(string))
	c.handleErr(err, key)

	return true
}

func (c *NodeUnitController) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < constant.MaxRetries {
		klog.V(2).Infof("Error syncing siteManager %v: %v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	klog.V(2).Infof("Dropping siteManager %q out of the queue: %v", key, err)
	c.queue.Forget(key)
}

func (c *NodeUnitController) addDaemonSet(obj interface{}) {
}
func (c *NodeUnitController) updateDaemonSet(oldObj interface{}, newObj interface{}) {
}
func (c *NodeUnitController) deleteDaemonSet(obj interface{}) {
}

func (c *NodeUnitController) addNodeUnit(obj interface{}) {
	nu := obj.(*sitev1alpha2.NodeUnit)
	klog.V(5).InfoS("Adding NodeUnit", "node unit", klog.KObj(nu))
	c.enqueueNodeUnit(nu.Name)
}

func (c *NodeUnitController) updateNodeUnit(oldObj interface{}, newObj interface{}) {
	oldNu, newNu := oldObj.(*sitev1alpha2.NodeUnit), newObj.(*sitev1alpha2.NodeUnit)
	// check if old nodeunit setNode update, and should not in node
	needClearNu := oldNu.DeepCopy()
	clearLabel := make(map[string]string)
	clearAnno := make(map[string]string)
	clearTaint := make([]corev1.Taint, 0, 2)
	for k, v := range oldNu.Spec.SetNode.Labels {
		if _, ok := newNu.Spec.SetNode.Labels[k]; !ok {
			clearLabel[k] = v
		}
	}
	for k, v := range oldNu.Spec.SetNode.Annotations {
		if _, ok := newNu.Spec.SetNode.Annotations[k]; !ok {
			clearAnno[k] = v
		}
	}
	for _, t := range oldNu.Spec.SetNode.Taints {
		if !utils.TaintInSlices(newNu.Spec.SetNode.Taints, t) {
			clearTaint = append(clearTaint, t)
		}
	}
	needClearNu.Spec.SetNode.Labels = clearLabel
	needClearNu.Spec.SetNode.Annotations = clearAnno
	needClearNu.Spec.SetNode.Taints = clearTaint
	c.nodeUnitDeleter.Delete(needClearNu)

	klog.V(5).InfoS("Updating NodeUnit", "old node unit", klog.KObj(oldNu), "new node unit", klog.KObj(newNu), "need clear setnode", util.ToJson(needClearNu.Spec.SetNode))
	c.enqueueNodeUnit(newNu.Name)
}

func (c *NodeUnitController) deleteNodeUnit(obj interface{}) {

	nu, ok := obj.(*sitev1alpha2.NodeUnit)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		nu, ok = tombstone.Obj.(*sitev1alpha2.NodeUnit)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not a NodeUnit %#v", obj))
			return
		}
	}
	klog.V(5).InfoS("Deleting NodeUnit", "node unit", klog.KObj(nu))
	c.enqueueNodeUnit(nu.Name)
}

func (c *NodeUnitController) addNode(obj interface{}) {
	node := obj.(*corev1.Node)
	if node.DeletionTimestamp != nil {
		c.deleteNode(obj)
		return
	}
	klog.V(5).InfoS("Adding Node", "node", klog.KObj(node))

	_, unitList, err := utils.GetUnitsByNode(c.nodeUnitLister, node)
	if err != nil {
		klog.V(2).ErrorS(err, "utils.GetUnitsByNode error", "node", node.Name)
		return
	}
	for _, nuName := range unitList {
		c.enqueueNodeUnit(nuName)
	}
	return
}

func (c *NodeUnitController) updateNode(oldObj, newObj interface{}) {
	oldNode, newNode := oldObj.(*corev1.Node), newObj.(*corev1.Node)
	if oldNode.ResourceVersion == newNode.ResourceVersion {
		// Periodic resync will send update events for all known nodes.
		return
	}
	klog.V(5).InfoS("Updating Node", "old node", klog.KObj(oldNode), "new node", klog.KObj(newNode))

	oldUnitLabel := make(map[string]string)
	newUnitLabel := make(map[string]string)
	for k, v := range oldNode.Labels {
		if v == constant.NodeUnitSuperedge {
			oldUnitLabel[k] = v
		}
	}
	for k, v := range newNode.Labels {
		if v == constant.NodeUnitSuperedge {
			newUnitLabel[k] = v
		}
	}
	// maybe update node unit label manual, recover it
	if !reflect.DeepEqual(oldUnitLabel, newUnitLabel) {
		_, unitList, err := utils.GetUnitsByNode(c.nodeUnitLister, oldNode)
		if err != nil {
			klog.V(2).ErrorS(err, "utils.GetUnitsByNode error", "node", oldNode.Name)
			return
		}
		for _, nuName := range unitList {
			c.enqueueNodeUnit(nuName)
		}
	}
	// current unit enqueqe
	_, unitList, err := utils.GetUnitsByNode(c.nodeUnitLister, newNode)
	if err != nil {
		klog.V(2).ErrorS(err, "utils.GetUnitsByNode error", "node", newNode.Name)
		return
	}
	for _, nuName := range unitList {
		c.enqueueNodeUnit(nuName)
	}
	return
}

func (c *NodeUnitController) deleteNode(obj interface{}) {

	node, ok := obj.(*corev1.Node)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		node, ok = tombstone.Obj.(*corev1.Node)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not a Node %#v", obj))
			return
		}
	}
	klog.V(5).InfoS("Deleting Node", "node", klog.KObj(node))

	_, unitList, err := utils.GetUnitsByNode(c.nodeUnitLister, node)
	if err != nil {
		klog.V(2).ErrorS(err, "utils.GetUnitsByNode error", "node", node.Name)
		return
	}
	for _, nuName := range unitList {
		c.enqueueNodeUnit(nuName)
	}
}

func (c *NodeUnitController) syncUnit(key string) error {
	startTime := time.Now()
	klog.V(4).InfoS("Started syncing nodeunit", "nodeunit", key, "startTime", startTime)
	defer func() {
		klog.V(4).InfoS("Finished syncing nodeunit", "nodeunit", key, "duration", time.Since(startTime))
	}()

	n, err := c.nodeUnitLister.Get(key)
	if errors.IsNotFound(err) {
		klog.V(2).InfoS("NodeUnit has been deleted", "nodeunit", key)
		// deal with node unit delete

		return nil
	}
	if err != nil {
		return err
	}

	nu := n.DeepCopy()

	if nu.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(nu, NodeUnitFinalizerID) {
			controllerutil.AddFinalizer(nu, NodeUnitFinalizerID)
			if _, err := c.crdClient.SiteV1alpha2().NodeUnits().Update(context.TODO(), nu, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(nu, NodeUnitFinalizerID) {
			if nu.Spec.AutonomyLevel != "L3" {
				c.eventRecorder.Event(nu, corev1.EventTypeWarning, fmt.Sprintf("The nodeunit whose autonomyLevel is %s is not allowed to be deleted", nu.Spec.AutonomyLevel), "Before deleting, please adjust the autonomyLevel of nodeunit to L3")
				return nil
			}

			err = c.reconcileNodeUnit(nu)
			if err != nil {
				return err
			}
			// our finalizer is present, so lets handle any external dependency
			if err := c.nodeUnitDeleter.Delete(nu); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(nu, NodeUnitFinalizerID)
			if _, err := c.crdClient.SiteV1alpha2().NodeUnits().Update(context.TODO(), nu, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}
		// Stop reconciliation as the item is being deleted
		return nil
	}
	// reconcile
	return c.reconcileNodeUnit(nu)
}

func (c *NodeUnitController) enqueue(name string) {
	c.queue.Add(name)
}

func (c *NodeUnitController) reconcileNodeUnit(nu *sitev1alpha2.NodeUnit) error {

	// 0. list nodemap and nodeset belong to current node unit
	unitNodeSet, nodeMap, err := utils.GetNodesByUnit(c.nodeLister, nu)
	if err != nil {
		return err
	}
	// 1. check nodes which should not belong to this unit, clear them(this will use gc)

	currentNodeSet := sets.NewString()

	currentNodeMap := make(map[string]*corev1.Node)
	gcNodeMap := make(map[string]*corev1.Node)
	currentNodeSelector := labels.NewSelector()
	nRequire, err := labels.NewRequirement(
		nu.Name,
		selection.In,
		[]string{constant.NodeUnitSuperedge},
	)
	if err != nil {
		return err
	}
	currentNodeSelector = currentNodeSelector.Add(*nRequire)
	utils.ListNodeFromLister(c.nodeLister, currentNodeSelector, func(n interface{}) {
		node, ok := n.(*corev1.Node)
		if !ok {
			return
		}
		currentNodeMap[node.Name] = node
		currentNodeSet.Insert(node.Name)
	})

	// find need gc node set
	needGCNodes := currentNodeSet.Difference(unitNodeSet)
	for _, gcNode := range needGCNodes.UnsortedList() {
		gcNodeMap[gcNode] = currentNodeMap[gcNode]
	}
	klog.V(5).InfoS("get node after node selector",
		"ensure nodes", unitNodeSet.UnsortedList(),
		"current nodes", currentNodeSet.UnsortedList(),
		"need gc nodes", needGCNodes.UnsortedList(),
	)

	if err := utils.DeleteNodesFromSetNode(c.kubeClient, nu, gcNodeMap); err != nil {
		return err
	}

	// 2. check node which should belong to this unit, ensure setNode(default is label)
	// 2.1 set node unit default value
	if v, ok := nu.Spec.SetNode.Labels[nu.Name]; !ok || v != constant.NodeUnitSuperedge {
		klog.V(5).InfoS("NodeUnit init, setnode default label", "node unit", nu.Name)
		if nu.Spec.SetNode.Labels == nil {
			nu.Spec.SetNode.Labels = make(map[string]string)
		}
		nu.Spec.SetNode.Labels[nu.Name] = constant.NodeUnitSuperedge
		_, err = c.crdClient.SiteV1alpha2().NodeUnits().Update(context.TODO(), nu, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("Update nodeUnit: %s error: %#v", nu.Name, err)
			return err
		}
		// first set default label,when nodeunit created, reconcile next.
		return nil
	}

	// 2.2 setnode to node
	if err := utils.SetNodeToNodes(c.kubeClient, nu, nodeMap); err != nil {
		return err
	}

	// 2.3 check node unit autonomy level if need install/uninstall unit cluster
	ucerr := c.kinsController.ReconcileUnitCluster(nu)
	// 3. caculate node unit status
	newStatus, err := utils.CaculateNodeUnitStatus(nodeMap, nu)
	if err != nil {
		return nil
	}
	// 3.1 update node unit cluster status
	ucStatus, err := c.kinsController.UpdateUnitClusterStatus(nu)
	if err != nil {
		klog.ErrorS(err, "Update node unit cluster status error", "node unit", nu.Name)
		return err
	}
	if ucerr != nil {
		klog.ErrorS(ucerr, "ReconcileUnitCluster error", "node unit", nu.Name)
		ucStatus.Phase = sitev1alpha2.ClusterFailed
		ucStatus.Conditions = []sitev1alpha2.ClusterCondition{
			{
				Type:          "Init",
				Status:        sitev1alpha2.ConditionFalse,
				LastProbeTime: metav1.Now(),
				Message:       ucerr.Error(),
			},
		}
	}
	newStatus.UnitCluster = *ucStatus

	if !reflect.DeepEqual(newStatus, &nu.Status) || !reflect.DeepEqual(newStatus, &nu.Status) {
		nu.Status = *newStatus
		// update node unit status only when status changed
		_, err = c.crdClient.SiteV1alpha2().NodeUnits().UpdateStatus(context.TODO(), nu, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("Update nodeUnit status: %s error: %#v", nu.Name, err)
			return err
		}
	}

	klog.V(4).InfoS("NodeUnit update success", "node unit", nu.Name)

	return nil
}
