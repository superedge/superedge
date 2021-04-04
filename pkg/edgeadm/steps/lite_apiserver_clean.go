package steps

import (
	"fmt"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/options"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/phases/workflow"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/util/initsystem"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
)

func NewCleanupLiteApiServerPhase() workflow.Phase {
	return workflow.Phase{
		Name:  "lite-apiserver clean up",
		Short: "Clean up lite-apiserver on edge node",
		Long:  "Clean up lite-apiserver on edge node",
		Run:   runCleanupLiteAPIServer,
		InheritFlags: []string{
			options.IgnorePreflightErrors, //todo
			options.CfgPath,
			options.NodeCRISocket,
			options.NodeName,
			options.FileDiscovery,
			options.TokenDiscovery,
			options.TokenDiscoveryCAHash,
			options.TokenDiscoverySkipCAHash,
			options.TLSBootstrapToken,
			options.TokenStr,
		},
	}
}

// runPreflight executes preflight checks logic.
func runCleanupLiteAPIServer(c workflow.RunData) error {
	// Try to stop the lite-apiserver service
	klog.V(1).Infoln("[reset] Getting init system")
	initSystem, err := initsystem.GetInitSystem()
	if err != nil {
		klog.Warningln("[reset] The lite-apiserver service could not be stopped by edgeadm. Unable to detect a supported init system!")
		klog.Warningln("[reset] Please ensure lite-apiserver is stopped manually")
	} else {
		fmt.Println("[reset] Stopping the lite-apiserver service")
		if err := initSystem.ServiceStop("lite-apiserver"); err != nil {
			klog.Warningf("[reset] The lite-apiserver service could not be stopped by edgeadm: [%v]\n", err)
			klog.Warningln("[reset] Please ensure lite-apiserver is stopped manually")
		}
	}
	resetConfigDir(constant.KubeEdgePath, constant.LiteApiServerCACert)
	return nil
}

// resetConfigDir is used to cleanup the files kubeadm writes in /etc/kubernetes/.
func resetConfigDir(configPathDir, pkiPathDir string) {
	filesToClean := []string{
		filepath.Join(configPathDir, constant.LITE_API_SERVER_KEY),
		filepath.Join(configPathDir, constant.LITE_API_SERVER_CRT),
		filepath.Join(configPathDir, constant.LiteApiserverTLS),
		pkiPathDir,
	}
	fmt.Printf("[reset] Deleting files: %v\n", filesToClean)
	for _, path := range filesToClean {
		if err := os.RemoveAll(path); err != nil {
			klog.Warningf("[reset] Failed to remove file: %q [%v]\n", path, err)
		}
	}
}
