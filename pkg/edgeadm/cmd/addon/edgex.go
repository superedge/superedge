package addon

import (
	"github.com/spf13/cobra"
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

			if err := action.runAddonedgex(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}
		},
	}
	action.flags = cmd.Flags()
	cmd.Flags().StringVar(&action.manifestDir, "manifest-dir", "",
		"Manifests document of edge kubernetes cluster.")

	cmd.Flags().BoolVar(&action.app, "app", false, "Addon the edgex application-services to cluster.")
	cmd.Flags().BoolVar(&action.core, "core", false, "Addon the edgex core-services to cluster.")
	cmd.Flags().BoolVar(&action.device, "device", false, "Addon the edgex device-services to cluster.")
	cmd.Flags().BoolVar(&action.support, "support", false, "Addon the edgex supporting-services to cluster.")
	cmd.Flags().BoolVar(&action.sysmgmt, "sysmgmt", false, "Addon the edgex system management to cluster")
	cmd.Flags().BoolVar(&action.ui, "ui", false, "Addon the edgex ui to cluster.")
	return cmd
}

func NewDetachEdgexCMD() *cobra.Command {
	action := addonAction{}
	cmd := &cobra.Command{
		Use:   "edgex",
		Short: "Detach edgex from Kubernetes cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if err := action.complete(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}

			if err := action.runDetachedgex(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}
		},
	}
	action.flags = cmd.Flags()
	cmd.Flags().StringVar(&action.manifestDir, "manifest-dir", "",
		"Manifests document of edge kubernetes cluster.")

	cmd.Flags().BoolVar(&action.app, "app", false, "Detach the edgex application-services from cluster.")
	cmd.Flags().BoolVar(&action.core, "core", false, "Detach the edgex core-services from cluster.")
	cmd.Flags().BoolVar(&action.device, "device", false, "Detach the edgex device-services from cluster.")
	cmd.Flags().BoolVar(&action.support, "support", false, "Detach the edgex supporting-services from cluster.")
	cmd.Flags().BoolVar(&action.sysmgmt, "sysmgmt", false, "Detach the edgex system management from cluster.")
	cmd.Flags().BoolVar(&action.ui, "ui", false, "Detach the ui from cluster.")
	cmd.Flags().BoolVar(&action.completely, "completely", false, "Detach the configmap and volumes from cluster.")
	return cmd
}

func (a *addonAction) runAddonedgex() error {
	var ser = make([]bool, constant.SerCount+1)
	ser[constant.App] = a.app
	ser[constant.Core] = a.core
	ser[constant.Support] = a.support
	ser[constant.Device] = a.device
	ser[constant.Ui] = a.ui
	ser[constant.Sysmgmt] = a.sysmgmt
	if !(a.app || a.core || a.support || a.device || a.ui || a.sysmgmt) {
		ser[constant.App] = true
		ser[constant.Core] = true
		ser[constant.Support] = true
		ser[constant.Device] = true
		ser[constant.Ui] = true
		ser[constant.Sysmgmt] = true
	}
	return common.DeployEdgex(a.clientSet, a.manifestDir, ser)
}

func (a *addonAction) runDetachedgex() error {
	var ser = make([]bool, constant.SerCount+1)
	ser[constant.App] = a.app
	ser[constant.Core] = a.core
	ser[constant.Support] = a.support
	ser[constant.Device] = a.device
	ser[constant.Ui] = a.ui
	ser[constant.Sysmgmt] = a.sysmgmt
	ser[constant.Completely] = a.completely

	if !(a.app || a.core || a.support || a.device || a.ui || a.sysmgmt || a.completely) {
		ser[constant.App] = true
		ser[constant.Core] = true
		ser[constant.Support] = true
		ser[constant.Device] = true
		ser[constant.Ui] = true
		ser[constant.Sysmgmt] = true
		ser[constant.Completely] = true
	}
	return common.DeleteEdgex(a.clientSet, a.manifestDir, ser)
}
