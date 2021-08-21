package addon

import (
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/superedge/superedge/pkg/edgeadm/common"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
)

func NewInstallEdgexCMD() *cobra.Command {
	action := addonAction{}
	cmd := &cobra.Command{
		Use:   "edgex",
		Short: "Addon edgex to Kubernetes cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if err := action.complete(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}

			if err :=  action.runAddonedgex(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}
		},
	}
	action.flags = cmd.Flags()
	cmd.Flags().StringVar(&action.manifestDir, "manifest-dir", "",
		"Manifests document of edge kubernetes cluster.")

	cmd.Flags().StringVar(&action.caCertFile, "ca.cert", constant.KubeadmCertPath,
		"The root certificate file for cluster.")

	cmd.Flags().StringVar(&action.caKeyFile, "ca.key", constant.KubeadmKeyPath,
		"The root certificate key file for cluster.")
	cmd.Flags().StringVar(&action.masterPublicAddr, "master-public-addr", "",
		"The public IP for control plane")
	cmd.Flags().StringArrayVar(&action.certSANs, "certSANs", []string{""},
		"The cert SAN")
	return cmd
}

func NewDetachEdgexCMD() *cobra.Command {
	action := addonAction{}
	cmd := &cobra.Command{
		Use:   "edgex",
		Short: "Delete edgex to Kubernetes cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if err := action.complete(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}

			if err :=  action.runDetachedgex(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}
		},
	}
	action.flags = cmd.Flags()
	cmd.Flags().StringVar(&action.manifestDir, "manifest-dir", "",
		"Manifests document of edge kubernetes cluster.")

	cmd.Flags().StringVar(&action.caCertFile, "ca.cert", constant.KubeadmCertPath,
		"The root certificate file for cluster. (default \"/etc/kubernetes/pki/ca.crt\")")

	cmd.Flags().StringVar(&action.caKeyFile, "ca.key", constant.KubeadmKeyPath,
		"The root certificate key file for cluster. (default \"/etc/kubernetes/pki/ca.key\")")
	cmd.Flags().StringVar(&action.masterPublicAddr, "master-public-addr", "",
		"The public IP for control plane")
	cmd.Flags().StringArrayVar(&action.certSANs, "certSANs", []string{""},
		"The cert SAN")

	return cmd
}

func  (a *addonAction) runAddonedgex() error {

	klog.Info("Start install addon edgex to your original cluster")
	return common.DeployEdgex(a.clientSet, a.manifestDir, a.caCertFile, a.caKeyFile, a.masterPublicAddr, a.certSANs, a.kubeConfig)
}

func  (a *addonAction) runDetachedgex() error {

	klog.Info("Start uninstall addon edgex from your original cluster")
	return common.DeleteEdgex(a.clientSet, a.manifestDir, a.caCertFile, a.caKeyFile, a.masterPublicAddr, a.certSANs, a.kubeConfig)
}
