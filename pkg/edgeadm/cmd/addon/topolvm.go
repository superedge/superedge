package addon

import (
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/superedge/superedge/pkg/edgeadm/common"
	"github.com/superedge/superedge/pkg/util"
)

func NewInstallTopolvmCMD() *cobra.Command {
	action := addonAction{}
	cmd := &cobra.Command{
		Use:   "topolvm",
		Short: "Addon topolvm local PV Plugin to Kubernetes cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if err := action.complete(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}

			if err := action.runTopolvmAddon(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}
		},
	}
	action.flags = cmd.Flags()
	cmd.Flags().StringVar(&action.manifestDir, "manifest-dir", "",
		"Manifests document of edge kubernetes cluster.")

	return cmd
}

func NewDetachTopolvmCMD() *cobra.Command {
	action := addonAction{}
	cmd := &cobra.Command{
		Use:   "topolvm",
		Short: "Remove topolvm local PV Plugin from Kubernetes cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if err := action.complete(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}

			if err := action.runTopolvmDetach(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}
		},
	}
	action.flags = cmd.Flags()
	cmd.Flags().StringVar(&action.manifestDir, "manifest-dir", "",
		"Manifests document of edge kubernetes cluster.")

	return cmd
}

func (a *addonAction) runTopolvmAddon() error {
	klog.Info("Start install topolvm to your cluster")
	return common.DeployTopolvmAppS(a.kubeConfig, a.manifestDir)
}

func (a *addonAction) runTopolvmDetach() error {
	klog.Info("Start uninstall topolvm from your cluster")
	return common.RemoveTopolvmApps(a.kubeConfig, a.manifestDir)
}
