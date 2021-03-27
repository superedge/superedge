package steps

import (
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/options"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/phases/workflow"
)

func NewInitNodePhase() workflow.Phase {
	return workflow.Phase{
		Name:    "init-node",
		Short:   "Init node",
		Long:    "Init node",
		Example: dockerExample,
		Run:     installDocker,
		InheritFlags: []string{
			options.CfgPath,               //todo
			options.IgnorePreflightErrors, //todo
		},
	}
}
