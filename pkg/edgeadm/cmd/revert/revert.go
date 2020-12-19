package revert

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"superedge/pkg/edgeadm/constant"
	"superedge/pkg/util"
	"superedge/pkg/util/kubeclient"
)

type revertAction struct {
	deployName string
	clientSet  *kubernetes.Clientset

	flags      *pflag.FlagSet
	caKeyFile  string
	caCertFile string
}

func newRevert() revertAction {
	return revertAction{}
}

func NewRevertCMD() *cobra.Command {
	action := newRevert()
	cmd := &cobra.Command{
		Use:   "revert -p DeployName",
		Short: "Revert edge cluster to your original cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if err := action.complete(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}

			if err := action.validate(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}

			if err := action.runRevert(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}
		},
	}

	action.flags = cmd.Flags()
	cmd.Flags().StringVarP(&action.deployName, "deploy", "p", "kubeadm",
		"The mode about deploy k8s cluster, support value: kubeadm")

	cmd.Flags().StringVar(&action.caCertFile, "ca.cert", "",
		"The root certificate file for cluster")

	cmd.Flags().StringVar(&action.caKeyFile, "ca.key", "",
		"The root certificate key file for cluster")

	return cmd
}

func (r *revertAction) complete() error {
	configPath, err := r.flags.GetString("kubeconfig")
	if err != nil {
		klog.Errorf("Get kubeconfig flags error: %v", err)
	}

	r.clientSet, err = kubeclient.GetClientSet(configPath)
	if err != nil {
		klog.Errorf("GetClientSet error: %v", err)
		return err
	}
	if r.clientSet == nil {
		return fmt.Errorf("Please set kubeconfig value!\n")
	}

	//todo: kubectl -n kube-system create cm system-cert  --from-file=/etc/kubernetes/pki
	return nil
}

func (r *revertAction) validate() error {
	return nil
}

func (r *revertAction) runRevert() error {
	fmt.Println("Start revert edge cluster to your original cluster")
	switch r.deployName {
	case constant.DeployModeKubeadm:
		return r.runKubeamdRevert()
	default:
		return fmt.Errorf("Not support %s change to edge cluster\n", r.deployName)
	}

	return nil
}
