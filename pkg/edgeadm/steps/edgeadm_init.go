package steps

import (
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/options"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/phases/workflow"
	cmdutil "github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/util"
	"k8s.io/klog/v2"
)

////////////////
var (
	edgeadmInitExample = cmdutil.Examples(`
		# Install docker container runtime.
		kubeadm init phase docker -docker-config docker.json
		`)
)

func NewEdgeadmInitPhase() workflow.Phase {
	return workflow.Phase{
		Name:    "edgeadm-init",
		Short:   "edgeadm init worker",
		Long:    "edgeadm init worker",
		Example: edgeadmInitExample,
		Run:     edgeadmInit,
		InheritFlags: []string{
			options.CfgPath,               //todo
			options.IgnorePreflightErrors, //todo
		},
	}
}

// runPreflight executes preflight checks logic.
func edgeadmInit(c workflow.RunData) error { //todo
	klog.V(5).Infof("Start edgeadm init work")

	return nil
}
