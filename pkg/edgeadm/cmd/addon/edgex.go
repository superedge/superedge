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

			if err :=  action.runAddonedgex(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}
		},
	}
	action.flags = cmd.Flags()
	cmd.Flags().StringVar(&action.manifestDir, "manifest-dir", "",
		"Manifests document of edge kubernetes cluster.")

	cmd.Flags().BoolVar(&action.app,"app", false, "add the app service.")
	cmd.Flags().BoolVar(&action.core,"core", false, "add the core service.")
	cmd.Flags().BoolVar(&action.support,"support", false, "add the support service.")
	cmd.Flags().BoolVar(&action.device,"device", false, "add the device service.")
	cmd.Flags().BoolVar(&action.ui,"ui", false, "add the ui.")
	cmd.Flags().BoolVar(&action.mqtt,"mqtt", false, "add the mqtt.")
	cmd.Flags().BoolVar(&action.configmap,"configmap", false, "add the configmap. only used when lose configmap")
	return cmd
}

func NewDetachEdgexCMD() *cobra.Command {
	action := addonAction{}
	cmd := &cobra.Command{
		Use:   "edgex",
		Short: "Delete edgex from Kubernetes cluster",
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
	cmd.Flags().BoolVar(&action.app,"app", false, "delete the app service.")
	cmd.Flags().BoolVar(&action.core,"core", false, "delete the core service.")
	cmd.Flags().BoolVar(&action.support,"support", false, "delete the support service.")
	cmd.Flags().BoolVar(&action.device,"device", false, "delete the device service.")
	cmd.Flags().BoolVar(&action.ui,"ui", false, "delete the ui.")
	cmd.Flags().BoolVar(&action.mqtt,"mqtt", false, "delete the mqtt.")
	cmd.Flags().BoolVar(&action.completely,"completely", false, "delete edgex completely.")
	return cmd
}

func  (a *addonAction) runAddonedgex() error {
	var ser map[string]bool
	ser = map[string]bool{constant.App:false,constant.Core:false,constant.Support:false,constant.Device:false,constant.Ui:false,constant.Mqtt:false}
	ser[constant.App]=a.app
	ser[constant.Core]=a.core
	ser[constant.Support]=a.support
	ser[constant.Device]=a.device
	ser[constant.Ui]=a.ui
	ser[constant.Mqtt]=a.mqtt
	if !(a.app||a.core||a.support||a.device||a.ui||a.mqtt||a.configmap) {
		ser[constant.App]=true
		ser[constant.Core]=true
		ser[constant.Support]=true
		ser[constant.Device]=true
		ser[constant.Ui]=true
		ser[constant.Mqtt]=true
	}
	return common.DeployEdgex(a.clientSet, a.manifestDir, ser)
}

func  (a *addonAction) runDetachedgex() error {
	var ser map[string]bool
	ser = map[string]bool{constant.App:false,constant.Core:false,constant.Support:false,constant.Device:false,constant.Ui:false,constant.Mqtt:false,constant.Completely:false}
	ser[constant.App]=a.app
	ser[constant.Core]=a.core
	ser[constant.Support]=a.support
	ser[constant.Device]=a.device
	ser[constant.Ui]=a.ui
	ser[constant.Mqtt]=a.mqtt
	ser[constant.Completely] = a.completely
	if !(a.app||a.core||a.support||a.device||a.ui||a.mqtt||a.completely) {
		ser[constant.App]=true
		ser[constant.Core]=true
		ser[constant.Support]=true
		ser[constant.Device]=true
		ser[constant.Ui]=true
		ser[constant.Mqtt]=true
		ser[constant.Completely]=true
	}
	return common.DeleteEdgex(a.clientSet, a.manifestDir, ser)
}
