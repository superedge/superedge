package steps

import (
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/options"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/phases/workflow"
)

func NewKubeletPhase() workflow.Phase {
	return workflow.Phase{
		Name:    "kubelet",
		Short:   "Install kubelet",
		Long:    "Install kubelet",
		Example: dockerExample,
		Run:     installDocker,
		InheritFlags: []string{
			options.CfgPath,               //todo
			options.IgnorePreflightErrors, //todo
		},
	}
}
