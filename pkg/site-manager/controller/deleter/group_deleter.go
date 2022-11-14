package deleter

import (
	"context"

	sitev1alpha2 "github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha2"
	crdClientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	crdv1listers "github.com/superedge/superedge/pkg/site-manager/generated/listers/site.superedge.io/v1alpha2"

	"github.com/superedge/superedge/pkg/site-manager/utils"
	clientset "k8s.io/client-go/kubernetes"
)

// NodeUnitDeleter clean up node label and unit cluster when
// NodeUnit has been deleted
type NodeGroupDeleter struct {
	kubeClient     clientset.Interface
	crdClient      *crdClientset.Clientset
	nodeUnitLister crdv1listers.NodeUnitLister
	finalizerToken string
}

func NewNodeGroupDeleter(
	kubeClient clientset.Interface,
	crdClient *crdClientset.Clientset,
	nodeUnitLister crdv1listers.NodeUnitLister,
	finalizerToken string,
) *NodeGroupDeleter {
	return &NodeGroupDeleter{
		kubeClient,
		crdClient,
		nodeUnitLister,
		finalizerToken,
	}
}

func (d *NodeGroupDeleter) Delete(ctx context.Context, ng *sitev1alpha2.NodeGroup) error {
	_, unitMap, err := utils.GetUnitByGroup(d.nodeUnitLister, ng)
	if err != nil {
		return err
	}

	return utils.DeleteNodeUnitFromSetNode(d.crdClient, ng, unitMap)
}
