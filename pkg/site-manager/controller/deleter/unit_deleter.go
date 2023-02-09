package deleter

import (
	"time"

	sitev1 "github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha2"
	"github.com/superedge/superedge/pkg/site-manager/constant"
	crdClientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	"github.com/superedge/superedge/pkg/site-manager/utils"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

// NodeUnitDeleter clean up node label and unit cluster when
// NodeUnit has been deleted
type NodeUnitDeleter struct {
	kubeClient       clientset.Interface
	crdClient        *crdClientset.Clientset
	nodeLister       corelisters.NodeLister
	nodeListerSynced cache.InformerSynced
	finalizerToken   string
	deleteQueue      workqueue.RateLimitingInterface
}

func NewNodeUnitDeleter(
	kubeClient clientset.Interface,
	crdClient *crdClientset.Clientset,
	nodeLister corelisters.NodeLister,
	nodeListerSynced cache.InformerSynced,
	finalizerToken string,
	deleteQueue workqueue.RateLimitingInterface,
) *NodeUnitDeleter {
	nud := &NodeUnitDeleter{
		kubeClient,
		crdClient,
		nodeLister,
		nodeListerSynced,
		finalizerToken,
		deleteQueue,
	}
	return nud
}

func (c *NodeUnitDeleter) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.deleteQueue.ShutDown()

	klog.V(1).Infof("Starting NodeGroupController")
	defer klog.V(1).Infof("Shutting down NodeGroupController")

	if !cache.WaitForNamedCacheSync("NodeGroupController", stopCh,
		c.nodeListerSynced) {
		return
	}
	// just 5 worker to clear node label
	for i := 0; i < 5; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
	}

	<-stopCh
}

func (d *NodeUnitDeleter) Delete(nu *sitev1.NodeUnit) error {
	d.deleteQueue.Add(nu)
	return nil
}

func (d *NodeUnitDeleter) sync(nu *sitev1.NodeUnit) error {
	_, nodeMap, err := utils.GetNodesByUnit(d.nodeLister, nu)
	if err != nil {
		return err
	}
	return utils.DeleteNodesFromSetNode(d.kubeClient, nu, nodeMap)

}

func (c *NodeUnitDeleter) worker() {
	for c.processNextWorkItem() {
	}
}

func (c *NodeUnitDeleter) processNextWorkItem() bool {
	key, quit := c.deleteQueue.Get()
	if quit {
		return false
	}
	defer c.deleteQueue.Done(key)
	klog.V(4).Infof("Get NodeUnitDeleter queue key unit name: %s", key.(*sitev1.NodeUnit).Name)
	err := c.sync(key.(*sitev1.NodeUnit))
	c.handleErr(err, key)

	return true
}

func (c *NodeUnitDeleter) handleErr(err error, key interface{}) {
	if err == nil {
		c.deleteQueue.Forget(key)
		return
	}

	if c.deleteQueue.NumRequeues(key) < constant.MaxRetries {
		klog.V(2).Infof("Error syncing NodeUnit %v: %v", key, err)
		c.deleteQueue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	klog.V(2).Infof("Dropping NodeUnit %q out of the queue: %v", key, err)
	c.deleteQueue.Forget(key)
}
