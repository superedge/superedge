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
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	sitev1alpha2 "github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha2"
	deleter "github.com/superedge/superedge/pkg/site-manager/controller/deleter"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/superedge/superedge/pkg/site-manager/constant"
	crdClientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	crdinformers "github.com/superedge/superedge/pkg/site-manager/generated/informers/externalversions/site.superedge.io/v1alpha2"
	crdv1listers "github.com/superedge/superedge/pkg/site-manager/generated/listers/site.superedge.io/v1alpha2"
	"github.com/superedge/superedge/pkg/site-manager/utils"
)

type NodeGroupController struct {
	nodeLister       corelisters.NodeLister
	nodeListerSynced cache.InformerSynced

	nodeUnitLister       crdv1listers.NodeUnitLister
	nodeUnitListerSynced cache.InformerSynced

	nodeGroupLister       crdv1listers.NodeGroupLister
	nodeGroupListerSynced cache.InformerSynced

	eventRecorder record.EventRecorder
	queue         workqueue.RateLimitingInterface
	kubeClient    clientset.Interface
	crdClient     *crdClientset.Clientset

	syncHandler      func(key string) error
	enqueueNodeGroup func(name string)
	nodeGroupDeleter *deleter.NodeGroupDeleter
}

func NewNodeGroupController(
	nodeInformer coreinformers.NodeInformer,
	nodeUnitInformer crdinformers.NodeUnitInformer,
	nodeGroupInformer crdinformers.NodeGroupInformer,
	kubeClient clientset.Interface,
	crdClient *crdClientset.Clientset,
) *NodeGroupController {

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1.EventSinkImpl{
		Interface: kubeClient.CoreV1().Events(""),
	})

	groupController := &NodeGroupController{
		kubeClient:    kubeClient,
		crdClient:     crdClient,
		eventRecorder: eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "nodegroup-controller"}),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "nodegroup-controller"),
	}

	nodeUnitInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    groupController.addNodeUnit,
		UpdateFunc: groupController.updateNodeUnit,
		DeleteFunc: groupController.deleteNodeUnit,
	})

	nodeGroupInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    groupController.addNodeGroup,
		UpdateFunc: groupController.updateNodeGroup,
		DeleteFunc: groupController.deleteNodeGroup,
	})

	groupController.syncHandler = groupController.syncGroup
	groupController.enqueueNodeGroup = groupController.enqueue

	groupController.nodeLister = nodeInformer.Lister()
	groupController.nodeListerSynced = nodeInformer.Informer().HasSynced

	// malc0lm TODO: add node informer to auto find node key immediately
	groupController.nodeUnitLister = nodeUnitInformer.Lister()
	groupController.nodeUnitListerSynced = nodeUnitInformer.Informer().HasSynced

	groupController.nodeGroupLister = nodeGroupInformer.Lister()
	groupController.nodeGroupListerSynced = nodeGroupInformer.Informer().HasSynced

	groupController.nodeGroupDeleter = deleter.NewNodeGroupDeleter(kubeClient, crdClient, nodeUnitInformer.Lister(), NodeGroupFinalizerID)
	return groupController
}

func (c *NodeGroupController) Run(workers, syncPeriodAsWhole int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.V(1).Infof("Starting NodeGroupController")
	defer klog.V(1).Infof("Shutting down NodeGroupController")

	if !cache.WaitForNamedCacheSync("NodeGroupController", stopCh,
		c.nodeListerSynced, c.nodeUnitListerSynced, c.nodeGroupListerSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
		klog.V(4).Infof("NodeGroupController set worker-%d success", i)
	}

	<-stopCh
}

func (c *NodeGroupController) worker() {
	for c.processNextWorkItem() {
	}
}

func (c *NodeGroupController) processNextWorkItem() bool {
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

func (c *NodeGroupController) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < constant.MaxRetries {
		klog.V(2).Infof("Error syncing NodeGroup %v: %v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	klog.V(2).Infof("Dropping NodeGroup %q out of the queue: %v", key, err)
	c.queue.Forget(key)
}

func (c *NodeGroupController) syncGroup(key string) error {
	startTime := time.Now()
	klog.V(4).InfoS("Started syncing nodegroup", "nodegroup", key, "startTime", startTime)
	defer func() {
		klog.V(4).InfoS("Finished syncing nodegroup", "nodegroup", key, "duration", time.Since(startTime))
	}()

	n, err := c.nodeGroupLister.Get(key)
	if errors.IsNotFound(err) {
		klog.V(2).InfoS("NodeGroup has been deleted", "nodegroup", key)
		// deal with node unit delete

		return nil
	}
	if err != nil {
		return err
	}

	ng := n.DeepCopy()

	if ng.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(ng, NodeGroupFinalizerID) {
			controllerutil.AddFinalizer(ng, NodeGroupFinalizerID)
			if _, err := c.crdClient.SiteV1alpha2().NodeGroups().Update(context.TODO(), ng, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(ng, NodeGroupFinalizerID) {
			// our finalizer is present, so lets handle any external dependency
			if err := c.nodeGroupDeleter.Delete(context.TODO(), ng); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(ng, NodeGroupFinalizerID)
			if _, err := c.crdClient.SiteV1alpha2().NodeGroups().Update(context.TODO(), ng, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}
		// Stop reconciliation as the item is being deleted
		return nil
	}

	// reconcile

	return c.reconcileNodeGroup(ng)
}
func (c *NodeGroupController) enqueue(name string) {
	c.queue.Add(name)
}

func (c *NodeGroupController) reconcileNodeGroup(ng *sitev1alpha2.NodeGroup) error {
	unitSet, unitMap, err := utils.GetUnitByGroup(c.nodeUnitLister, ng)
	if err != nil {
		klog.ErrorS(err, "GetUnitByGroup error")
		return err
	}

	// 1. ensure  node unit which not belong to this group
	currentUnitSet := sets.NewString()

	currentUnitMap := make(map[string]*sitev1alpha2.NodeUnit)
	gcUnitMap := make(map[string]*sitev1alpha2.NodeUnit)
	unitLabelSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{ng.Name: constant.NodeGroupSuperedge},
	}
	unitSelector, err := metav1.LabelSelectorAsSelector(unitLabelSelector)
	if err != nil {
		return err
	}
	utils.ListNodeUnitFromLister(c.nodeUnitLister, unitSelector, func(n interface{}) {
		nu, ok := n.(*sitev1alpha2.NodeUnit)
		if !ok {
			return
		}
		currentUnitMap[nu.Name] = nu
		currentUnitSet.Insert(nu.Name)
	})
	needGCUnits := currentUnitSet.Difference(unitSet)
	for _, gcNode := range needGCUnits.UnsortedList() {
		gcUnitMap[gcNode] = currentUnitMap[gcNode]
	}
	if err := utils.DeleteNodeUnitFromSetNode(c.crdClient, ng, gcUnitMap); err != nil {
		klog.ErrorS(err, "DeleteNodeUnitFromSetNode error")
		return err
	}

	// 2. ensure autoFindKey node will work fine
	if len(ng.Spec.AutoFindNodeKeys) > 0 {
		if err := c.autoFindNodeKeysByNodeGroup(ng); err != nil {
			klog.ErrorS(err, "autoFindNodeKeysByNodeGroup error")
			return err
		}
	}

	// 3. ensure node group child node unit has property label
	if err := utils.SetNodeToNodeUnits(c.crdClient, ng, unitMap); err != nil {
		klog.ErrorS(err, "SetNodeToNodeUnits error")
		return err
	}
	// 4. caculate status and update group status
	newStatus, err := utils.CaculateNodeGroupStatus(unitSet, ng)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(newStatus, &ng.Status) {
		ng.Status = *newStatus

		// update node unit status only when status changed
		_, err = c.crdClient.SiteV1alpha2().NodeGroups().UpdateStatus(context.TODO(), ng, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("Update NodeGroup=(%s) error: %#v", ng.Name, err)
			return err
		}
	}
	klog.V(4).InfoS("NodeGroup update success", "node group", ng.Name)
	return nil
}

func (c *NodeGroupController) addNodeUnit(obj interface{}) {
	nu := obj.(*sitev1alpha2.NodeUnit)
	if nu.DeletionTimestamp != nil {
		c.deleteNodeUnit(obj)
		return
	}
	klog.V(5).InfoS("Adding NodeUnit", "nodeunit", klog.KObj(nu))

	_, groupList, err := utils.GetGroupsByUnit(c.nodeGroupLister, nu)
	if err != nil {
		klog.V(2).ErrorS(err, "GetGroupsByUnit error", "nodeunit", klog.KObj(nu))
		return
	}
	for _, ng := range groupList {
		c.enqueueNodeGroup(ng)
	}
}

func (c *NodeGroupController) updateNodeUnit(oldObj interface{}, newObj interface{}) {
	oldNu, newNu := oldObj.(*sitev1alpha2.NodeUnit), newObj.(*sitev1alpha2.NodeUnit)
	if oldNu.ResourceVersion == newNu.ResourceVersion {
		// Periodic resync will send update events for all known nodes.
		return
	}
	klog.V(5).InfoS("Updating NodeUnit", "old node unit", klog.KObj(oldNu), "new node unit", klog.KObj(newNu))

	oldGroupLabel := make(map[string]string)
	newGroupLabel := make(map[string]string)
	for k, v := range oldNu.Labels {
		if v == constant.NodeGroupSuperedge {
			oldGroupLabel[k] = v
		}
	}
	for k, v := range newNu.Labels {
		if v == constant.NodeGroupSuperedge {
			newGroupLabel[k] = v
		}
	}
	// maybe update node group label manual, recover it
	if !reflect.DeepEqual(oldGroupLabel, newGroupLabel) {
		_, groupList, err := utils.GetGroupsByUnit(c.nodeGroupLister, oldNu)
		if err != nil {
			klog.V(2).ErrorS(err, "GetGroupsByUnit error", "node unit", oldNu.Name)
			return
		}
		for _, ngName := range groupList {
			c.enqueueNodeGroup(ngName)
		}
	}
	// current unit enqueqe
	_, groupList, err := utils.GetGroupsByUnit(c.nodeGroupLister, newNu)
	if err != nil {
		klog.V(2).ErrorS(err, "GetGroupsByUnit error", "node unit", newNu.Name)
		return
	}
	for _, ngName := range groupList {
		c.enqueueNodeGroup(ngName)
	}
	return

}

func (c *NodeGroupController) deleteNodeUnit(obj interface{}) {
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

	_, groupList, err := utils.GetGroupsByUnit(c.nodeGroupLister, nu)
	if err != nil {
		klog.V(2).ErrorS(err, "GetGroupsByUnit error", "node unit", nu.Name)
		return
	}
	for _, ngName := range groupList {
		c.enqueueNodeGroup(ngName)
	}
}

func (c *NodeGroupController) addNodeGroup(obj interface{}) {
	ng := obj.(*sitev1alpha2.NodeGroup)
	klog.V(5).InfoS("Adding NodeGroup", "node group", klog.KObj(ng))
	c.enqueueNodeGroup(ng.Name)
}

func (c *NodeGroupController) updateNodeGroup(oldObj interface{}, newObj interface{}) {
	oldNg, newNg := oldObj.(*sitev1alpha2.NodeGroup), newObj.(*sitev1alpha2.NodeGroup)
	klog.V(5).InfoS("Updating NodeGroup", "old node group", klog.KObj(oldNg), "new node group", klog.KObj(newNg))
	c.enqueueNodeGroup(newNg.Name)
}

func (c *NodeGroupController) deleteNodeGroup(obj interface{}) {
	ng, ok := obj.(*sitev1alpha2.NodeGroup)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		ng, ok = tombstone.Obj.(*sitev1alpha2.NodeGroup)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not a NodeGroup %#v", obj))
			return
		}
	}
	klog.V(5).InfoS("Deleting NodeGroup", "node group", klog.KObj(ng))
	c.enqueueNodeGroup(ng.Name)
}

func (c *NodeGroupController) autoFindNodeKeysByNodeGroup(nodeGroup *sitev1alpha2.NodeGroup) error {
	// find nodes by keys
	allnodes, err := c.nodeLister.List(labels.Everything())
	if err != nil {
		return err
	}
	var matchNodes []*corev1.Node

	for _, node := range allnodes {
		if len(node.Labels) == 0 {
			continue
		}
		match, newUnitName, newUnitSelector := checkIfContains(node.Labels, nodeGroup.Spec.AutoFindNodeKeys)
		if match && len(newUnitSelector) > 0 {
			matchNodes = append(matchNodes, node)
			// auto find key match, create a net node unit
			c.newNodeUnit(nodeGroup, newUnitName, nodeGroup.Spec.AutoFindNodeKeys, newUnitSelector)
		}
	}
	return nil
}

func filterString(name string) string {
	if withCheckContains(name) || withCheckSize(name) {
		return hashString(name)
	}
	return name
}

func hashString(name string) string {
	h := sha1.New()
	h.Write([]byte(name))
	sha1_hash := hex.EncodeToString(h.Sum(nil))
	return sha1_hash
}

func withCheckContains(name string) bool {
	if strings.Contains(name, "/") {
		return true
	}
	return false
}

// check size is it more than 64
func withCheckSize(name string) bool {
	if len(name) >= 64 {
		return true
	}
	return false
}

func (c *NodeGroupController) newNodeUnit(ng *sitev1alpha2.NodeGroup, newname string, keyslices []string, sel map[string]string) error {

	klog.V(4).Infof("prepare to ceate nodeUnite: %s, selector: %s", newname, sel)
	nuLabel := make(map[string]string)
	nuLabel[constant.NodeUnitAutoFindLabel] = utils.HashAutoFindKeys(keyslices)
	nuLabel[ng.Name] = constant.NodeGroupSuperedge

	newNodeUnit := &sitev1alpha2.NodeUnit{
		ObjectMeta: metav1.ObjectMeta{
			Name:        newname,
			Annotations: sel,
			Labels:      nuLabel,
		},
		Spec: sitev1alpha2.NodeUnitSpec{
			Type: utils.EdgeNodeUnit,
			Selector: &sitev1alpha2.Selector{
				MatchLabels: sel,
			},
			SetNode: sitev1alpha2.SetNode{
				Labels: map[string]string{ng.Name: newname, newname: constant.NodeUnitSuperedge},
			},
		},
	}

	// check if any exist generated node unit and update it
	currentNu, err := c.crdClient.SiteV1alpha2().NodeUnits().Get(context.TODO(), newname, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		klog.Warning("obj not found, will create nodeunit now")
		_, err = c.crdClient.SiteV1alpha2().NodeUnits().Create(context.TODO(), newNodeUnit, metav1.CreateOptions{})
		if err != nil {
			klog.ErrorS(err, "error to create node unit")
			return err
		}

	} else if err == nil {
		if !reflect.DeepEqual(currentNu.Spec.Selector.MatchLabels, sel) || !reflect.DeepEqual(currentNu.Labels, nuLabel) {
			currentNu.Spec.Selector.MatchLabels = sel
			currentNu.Labels = nuLabel
			_, err := c.crdClient.SiteV1alpha2().NodeUnits().Update(context.TODO(), currentNu, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
		}
	} else {
		return err
	}

	return nil
}

func checkOwnerReferenceContains(owner metav1.OwnerReference, tmpSlice []metav1.OwnerReference) bool {
	for _, value := range tmpSlice {
		if value == owner {
			return true
		}
	}
	return false
}

func checkIfContains(nodelabel map[string]string, keyslices []string) (bool, string, map[string]string) {
	var res string
	sel := make(map[string]string)
	sort.Strings(keyslices)
	for _, value := range keyslices {
		if _, ok := nodelabel[value]; ok {
			sel[value] = nodelabel[value]
			if res == "" {
				res = nodelabel[value]
			} else {
				res = res + "-" + nodelabel[value]
			}

			continue
		} else {
			return false, "", sel
		}
	}
	// check new unit name is vaild for Name field
	return true, filterString(res), sel
}
