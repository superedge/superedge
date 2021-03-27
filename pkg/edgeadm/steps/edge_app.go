package steps

import (
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/options"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/phases/workflow"
)

func NewEdgeAppsPhase() workflow.Phase {
	return workflow.Phase{
		Name:    "edge-apps",
		Short:   "Dploy edge apps",
		Long:    "Deploy edge apps",
		Example: dockerExample,
		Run:     installDocker,
		InheritFlags: []string{
			options.CfgPath,               //todo
			options.IgnorePreflightErrors, //todo
		},
	}
}
