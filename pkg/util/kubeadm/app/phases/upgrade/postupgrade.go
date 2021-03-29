/*
Copyright 2017 The Kubernetes Authors.

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

package upgrade

import (
	"context"
	"os"

	"github.com/pkg/errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	errorsutil "k8s.io/apimachinery/pkg/util/errors"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	kubeadmapi "github.com/superedge/superedge/pkg/util/kubeadm/app/apis/kubeadm"
	kubeadmconstants "github.com/superedge/superedge/pkg/util/kubeadm/app/constants"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/phases/addons/dns"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/phases/addons/proxy"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/phases/bootstraptoken/clusterinfo"
	nodebootstraptoken "github.com/superedge/superedge/pkg/util/kubeadm/app/phases/bootstraptoken/node"
	kubeletphase "github.com/superedge/superedge/pkg/util/kubeadm/app/phases/kubelet"
	patchnodephase "github.com/superedge/superedge/pkg/util/kubeadm/app/phases/patchnode"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/phases/uploadconfig"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/util/apiclient"
	dryrunutil "github.com/superedge/superedge/pkg/util/kubeadm/app/util/dryrun"
)

// PerformPostUpgradeTasks runs nearly the same functions as 'kubeadm init' would do
// Note that the mark-control-plane phase is left out, not needed, and no token is created as that doesn't belong to the upgrade
func PerformPostUpgradeTasks(client clientset.Interface, cfg *kubeadmapi.InitConfiguration, dryRun bool) error {
	errs := []error{}

	// Upload currently used configuration to the cluster
	// Note: This is done right in the beginning of cluster initialization; as we might want to make other phases
	// depend on centralized information from this source in the future
	if err := uploadconfig.UploadConfiguration(cfg, client); err != nil {
		errs = append(errs, err)
	}

	// Create the new, version-branched kubelet ComponentConfig ConfigMap
	if err := kubeletphase.CreateConfigMap(&cfg.ClusterConfiguration, client); err != nil {
		errs = append(errs, errors.Wrap(err, "error creating kubelet configuration ConfigMap"))
	}

	// Write the new kubelet config down to disk and the env file if needed
	if err := writeKubeletConfigFiles(client, cfg, dryRun); err != nil {
		errs = append(errs, err)
	}

	// Annotate the node with the crisocket information, sourced either from the InitConfiguration struct or
	// --cri-socket.
	// TODO: In the future we want to use something more official like NodeStatus or similar for detecting this properly
	if err := patchnodephase.AnnotateCRISocket(client, cfg.NodeRegistration.Name, cfg.NodeRegistration.CRISocket); err != nil {
		errs = append(errs, errors.Wrap(err, "error uploading crisocket"))
	}

	// Create RBAC rules that makes the bootstrap tokens able to get nodes
	if err := nodebootstraptoken.AllowBoostrapTokensToGetNodes(client); err != nil {
		errs = append(errs, err)
	}

	// Create/update RBAC rules that makes the bootstrap tokens able to post CSRs
	if err := nodebootstraptoken.AllowBootstrapTokensToPostCSRs(client); err != nil {
		errs = append(errs, err)
	}

	// Create/update RBAC rules that makes the bootstrap tokens able to get their CSRs approved automatically
	if err := nodebootstraptoken.AutoApproveNodeBootstrapTokens(client); err != nil {
		errs = append(errs, err)
	}

	// Create/update RBAC rules that makes the nodes to rotate certificates and get their CSRs approved automatically
	if err := nodebootstraptoken.AutoApproveNodeCertificateRotation(client); err != nil {
		errs = append(errs, err)
	}

	// TODO: Is this needed to do here? I think that updating cluster info should probably be separate from a normal upgrade
	// Create the cluster-info ConfigMap with the associated RBAC rules
	// if err := clusterinfo.CreateBootstrapConfigMapIfNotExists(client, kubeadmconstants.GetAdminKubeConfigPath()); err != nil {
	// 	return err
	//}
	// Create/update RBAC rules that makes the cluster-info ConfigMap reachable
	if err := clusterinfo.CreateClusterInfoRBACRules(client); err != nil {
		errs = append(errs, err)
	}

	// If the coredns / kube-dns ConfigMaps are missing, show a warning and assume that the
	// DNS addon was skipped during "kubeadm init", and that its redeployment on upgrade is not desired.
	//
	// TODO: remove this once "kubeadm upgrade apply" phases are supported:
	//   https://github.com/kubernetes/kubeadm/issues/1318
	var missingCoreDNSConfigMap, missingKubeDNSConfigMap bool
	if _, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(
		context.TODO(),
		kubeadmconstants.CoreDNSConfigMap,
		metav1.GetOptions{},
	); err != nil && apierrors.IsNotFound(err) {
		missingCoreDNSConfigMap = true
	}
	if _, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(
		context.TODO(),
		kubeadmconstants.KubeDNSConfigMap,
		metav1.GetOptions{},
	); err != nil && apierrors.IsNotFound(err) {
		missingKubeDNSConfigMap = true
	}
	if missingCoreDNSConfigMap && missingKubeDNSConfigMap {
		klog.Warningf("the ConfigMaps %q/%q in the namespace %q were not found. "+
			"Assuming that a DNS server was not deployed for this cluster. "+
			"Note that once 'kubeadm upgrade apply' supports phases you "+
			"will have to skip the DNS upgrade manually",
			kubeadmconstants.CoreDNSConfigMap,
			kubeadmconstants.KubeDNSConfigMap,
			metav1.NamespaceSystem)
	} else {
		// Upgrade CoreDNS/kube-dns
		if err := dns.EnsureDNSAddon(&cfg.ClusterConfiguration, client); err != nil {
			errs = append(errs, err)
		}
		// Remove the old DNS deployment if a new DNS service is now used (kube-dns to CoreDNS or vice versa)
		if err := removeOldDNSDeploymentIfAnotherDNSIsUsed(&cfg.ClusterConfiguration, client, dryRun); err != nil {
			errs = append(errs, err)
		}
	}

	// If the kube-proxy ConfigMap is missing, show a warning and assume that kube-proxy
	// was skipped during "kubeadm init", and that its redeployment on upgrade is not desired.
	//
	// TODO: remove this once "kubeadm upgrade apply" phases are supported:
	//   https://github.com/kubernetes/kubeadm/issues/1318
	if _, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(
		context.TODO(),
		kubeadmconstants.KubeProxyConfigMap,
		metav1.GetOptions{},
	); err != nil && apierrors.IsNotFound(err) {
		klog.Warningf("the ConfigMap %q in the namespace %q was not found. "+
			"Assuming that kube-proxy was not deployed for this cluster. "+
			"Note that once 'kubeadm upgrade apply' supports phases you "+
			"will have to skip the kube-proxy upgrade manually",
			kubeadmconstants.KubeProxyConfigMap,
			metav1.NamespaceSystem)
	} else {
		// Upgrade kube-proxy
		if err := proxy.EnsureProxyAddon(&cfg.ClusterConfiguration, &cfg.LocalAPIEndpoint, client); err != nil {
			errs = append(errs, err)
		}
	}

	return errorsutil.NewAggregate(errs)
}

func removeOldDNSDeploymentIfAnotherDNSIsUsed(cfg *kubeadmapi.ClusterConfiguration, client clientset.Interface, dryRun bool) error {
	return apiclient.TryRunCommand(func() error {
		installedDeploymentName := kubeadmconstants.KubeDNSDeploymentName
		deploymentToDelete := kubeadmconstants.CoreDNSDeploymentName

		if cfg.DNS.Type == kubeadmapi.CoreDNS {
			installedDeploymentName = kubeadmconstants.CoreDNSDeploymentName
			deploymentToDelete = kubeadmconstants.KubeDNSDeploymentName
		}

		nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
			FieldSelector: fields.Set{"spec.unschedulable": "false"}.AsSelector().String(),
		})
		if err != nil {
			return err
		}

		// If we're dry-running or there are no scheduable nodes available, we don't need to wait for the new DNS addon to become ready
		if !dryRun && len(nodes.Items) != 0 {
			dnsDeployment, err := client.AppsV1().Deployments(metav1.NamespaceSystem).Get(context.TODO(), installedDeploymentName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if dnsDeployment.Status.ReadyReplicas == 0 {
				return errors.New("the DNS deployment isn't ready yet")
			}
		}

		// We don't want to wait for the DNS deployment above to become ready when dryrunning (as it never will)
		// but here we should execute the DELETE command against the dryrun clientset, as it will only be logged
		err = apiclient.DeleteDeploymentForeground(client, metav1.NamespaceSystem, deploymentToDelete)
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		return nil
	}, 10)
}

func writeKubeletConfigFiles(client clientset.Interface, cfg *kubeadmapi.InitConfiguration, dryRun bool) error {
	kubeletDir, err := GetKubeletDir(dryRun)
	if err != nil {
		// The error here should never occur in reality, would only be thrown if /tmp doesn't exist on the machine.
		return err
	}
	errs := []error{}
	// Write the configuration for the kubelet down to disk so the upgraded kubelet can start with fresh config
	if err := kubeletphase.WriteConfigToDisk(&cfg.ClusterConfiguration, kubeletDir); err != nil {
		errs = append(errs, errors.Wrap(err, "error writing kubelet configuration to file"))
	}

	if dryRun { // Print what contents would be written
		dryrunutil.PrintDryRunFile(kubeadmconstants.KubeletConfigurationFileName, kubeletDir, kubeadmconstants.KubeletRunDirectory, os.Stdout)
	}
	return errorsutil.NewAggregate(errs)
}

// GetKubeletDir gets the kubelet directory based on whether the user is dry-running this command or not.
func GetKubeletDir(dryRun bool) (string, error) {
	if dryRun {
		return kubeadmconstants.CreateTempDirForKubeadm("", "kubeadm-upgrade-dryrun")
	}
	return kubeadmconstants.KubeletRunDirectory, nil
}

// moveFiles moves files from one directory to another.
func moveFiles(files map[string]string) error {
	filesToRecover := map[string]string{}
	for from, to := range files {
		if err := os.Rename(from, to); err != nil {
			return rollbackFiles(filesToRecover, err)
		}
		filesToRecover[to] = from
	}
	return nil
}

// rollbackFiles moves the files back to the original directory.
func rollbackFiles(files map[string]string, originalErr error) error {
	errs := []error{originalErr}
	for from, to := range files {
		if err := os.Rename(from, to); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Errorf("couldn't move these files: %v. Got errors: %v", files, errorsutil.NewAggregate(errs))
}
