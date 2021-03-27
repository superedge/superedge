package steps

import (
	"fmt"
	phases "github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/phases/init"

	"github.com/pkg/errors"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/options"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/phases/workflow"
	cmdutil "github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/util"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/preflight"
	utilsexec "k8s.io/utils/exec"
)

var (
	preflightExample = cmdutil.Examples(`
		# Run pre-flight checks for kubeadm init using a config file.
		kubeadm init phase preflight --config kubeadm-config.yml
		`)
)

// NewPreflightPhase creates a kubeadm workflow phase that implements preflight checks for a new control-plane node.
func NewPreflightPhase() workflow.Phase {
	return workflow.Phase{
		Name:    "preflight-01",
		Short:   "Run pre-flight checks",
		Long:    "Run pre-flight checks for kubeadm init.",
		Example: preflightExample,
		Run:     runPreflight,
		InheritFlags: []string{
			options.CfgPath,
			options.IgnorePreflightErrors,
		},
	}
}

// runPreflight executes preflight checks logic.
func runPreflight(c workflow.RunData) error {
	data, ok := c.(phases.InitData)
	if !ok {
		return errors.New("preflight phase invoked with an invalid data struct")
	}

	fmt.Println("[preflight] Running pre-flight checks")
	if err := preflight.RunInitNodeChecks(utilsexec.New(), data.Cfg(), data.IgnorePreflightErrors(), false, false); err != nil {
		return err
	}

	if !data.DryRun() {
		fmt.Println("[preflight] Pulling images required for setting up a Kubernetes cluster")
		fmt.Println("[preflight] This might take a minute or two, depending on the speed of your internet connection")
		fmt.Println("[preflight] You can also perform this action in beforehand using 'kubeadm config images pull'")
		if err := preflight.RunPullImagesCheck(utilsexec.New(), data.Cfg(), data.IgnorePreflightErrors()); err != nil {
			return err
		}
	} else {
		fmt.Println("[preflight] Would pull the required images (like 'kubeadm config images pull')")
	}

	return nil
}
