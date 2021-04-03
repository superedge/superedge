package steps

import (
	"fmt"
	"github.com/superedge/superedge/pkg/edgeadm/cmd"
	"github.com/superedge/superedge/pkg/edgeadm/common"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
	phases "github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/phases/init"
	"k8s.io/klog"

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

var (
	edgeConfig    *cmd.EdgeadmConfig
	dockerExample = cmdutil.Examples(`
		# Install docker container runtime.
		kubeadm init phase docker -docker-config docker.json
		`)
)

func NewContainerPhase(config *cmd.EdgeadmConfig) workflow.Phase {
	edgeConfig = config
	return workflow.Phase{
		Name:    "container",
		Short:   "Install container runtime",
		Long:    "Install container runtime",
		Example: dockerExample,
		Run:     installContainer,
		InheritFlags: []string{
			options.CfgPath,               //todo
			options.IgnorePreflightErrors, //todo
		},
	}
}

//install container runtime (docker | containerd | CRI-O)
func installContainer(c workflow.RunData) error {

	err := installDocker()
	if err != nil {
		return err
	}

	fmt.Println("Has been successfully installed edge container")
	return nil
}

func installDocker() error {
	klog.V(5).Infof("====Start install docker container runtime====")

	//unzip Docker Package
	if err := common.UnzipPackage(edgeConfig.WorkerPath+constant.ZipContainerPath, edgeConfig.WorkerPath+constant.UnZipContainerDstPath); err != nil {
		klog.Errorf("Unzip Docker container runtime Package: %s, error: %v", edgeConfig.WorkerPath+constant.UnZipContainerDstPath, err)
		return err
	}

	if _, _, err := util.RunLinuxCommand(fmt.Sprintf(`sh %s`, edgeConfig.WorkerPath+constant.DockerInstallShell)); err != nil {
		klog.Errorf("Run Docker install shell: %s, error: %v", edgeConfig.WorkerPath+constant.UnZipContainerDstPath, err)
		return err
	}

	klog.V(5).Infof("====Stop install docker container runtime====")
	return nil
}
