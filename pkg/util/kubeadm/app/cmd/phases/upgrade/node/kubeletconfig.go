/*
Copyright 2019 The Kubernetes Authors.

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

package node

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/options"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/phases/workflow"
	cmdutil "github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/util"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/constants"
	kubeletphase "github.com/superedge/superedge/pkg/util/kubeadm/app/phases/kubelet"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/phases/upgrade"
	dryrunutil "github.com/superedge/superedge/pkg/util/kubeadm/app/util/dryrun"
)

var (
	kubeletConfigLongDesc = cmdutil.LongDesc(`
		Download the kubelet configuration from a ConfigMap of the form "kubelet-config-1.X" in the cluster,
		where X is the minor version of the kubelet. kubeadm uses the KuberneteVersion field in the kubeadm-config
		ConfigMap to determine what the _desired_ kubelet version is.
		`)
)

// NewKubeletConfigPhase creates a kubeadm workflow phase that implements handling of kubelet-config upgrade.
func NewKubeletConfigPhase() workflow.Phase {
	phase := workflow.Phase{
		Name:  "kubelet-config",
		Short: "Upgrade the kubelet configuration for this node",
		Long:  kubeletConfigLongDesc,
		Run:   runKubeletConfigPhase(),
		InheritFlags: []string{
			options.DryRun,
			options.KubeconfigPath,
			options.KubeletVersion,
		},
	}
	return phase
}

func runKubeletConfigPhase() func(c workflow.RunData) error {
	return func(c workflow.RunData) error {
		data, ok := c.(Data)
		if !ok {
			return errors.New("kubelet-config phase invoked with an invalid data struct")
		}

		// otherwise, retrieve all the info required for kubelet config upgrade
		cfg := data.Cfg()
		dryRun := data.DryRun()

		// Set up the kubelet directory to use. If dry-running, this will return a fake directory
		kubeletDir, err := upgrade.GetKubeletDir(dryRun)
		if err != nil {
			return err
		}

		// TODO: Checkpoint the current configuration first so that if something goes wrong it can be recovered

		// Store the kubelet component configuration.
		// By default the kubelet version is expected to be equal to cfg.ClusterConfiguration.KubernetesVersion, but
		// users can specify a different kubelet version (this is a legacy of the original implementation
		// of `kubeadm upgrade node config` which we are preserving in order to not break the GA contract)
		if data.KubeletVersion() != "" && data.KubeletVersion() != cfg.ClusterConfiguration.KubernetesVersion {
			fmt.Printf("[upgrade] Using kubelet config version %s, while kubernetes-version is %s\n", data.KubeletVersion(), cfg.ClusterConfiguration.KubernetesVersion)
			if err := kubeletphase.DownloadConfig(data.Client(), data.KubeletVersion(), kubeletDir); err != nil {
				return err
			}

			// WriteConfigToDisk is what we should be calling since we already have the correct component config loaded
		} else if err = kubeletphase.WriteConfigToDisk(&cfg.ClusterConfiguration, kubeletDir); err != nil {
			return err
		}

		// If we're dry-running, print the generated manifests
		if dryRun {
			if err := printFilesIfDryRunning(dryRun, kubeletDir); err != nil {
				return errors.Wrap(err, "error printing files on dryrun")
			}
			return nil
		}

		fmt.Println("[upgrade] The configuration for this node was successfully updated!")
		fmt.Println("[upgrade] Now you should go ahead and upgrade the kubelet package using your package manager.")
		return nil
	}
}

// printFilesIfDryRunning prints the Static Pod manifests to stdout and informs about the temporary directory to go and lookup
func printFilesIfDryRunning(dryRun bool, kubeletDir string) error {
	if !dryRun {
		return nil
	}

	// Print the contents of the upgraded file and pretend like they were in kubeadmconstants.KubeletRunDirectory
	fileToPrint := dryrunutil.FileToPrint{
		RealPath:  filepath.Join(kubeletDir, constants.KubeletConfigurationFileName),
		PrintPath: filepath.Join(constants.KubeletRunDirectory, constants.KubeletConfigurationFileName),
	}
	return dryrunutil.PrintDryRunFiles([]dryrunutil.FileToPrint{fileToPrint}, os.Stdout)
}
