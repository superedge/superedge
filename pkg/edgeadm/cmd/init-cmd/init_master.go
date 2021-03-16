package init_cmd

import (
	"fmt"

	"k8s.io/klog/v2"

	"github.com/superedge/superedge/pkg/edgeadm/common"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeclient"
)

func WorkerHome(e *initData) string {
	return e.InitOptions.WorkerPath
}

func (e *initData) preInstallHook() error { //todo finish
	klog.Infof("=========, preInstallHook")
	return util.RunLinuxShellFile(constant.PreInstallHook)
}

func NilFunc(e *initData) error {
	klog.V(5).Infof("I'm nil func")
	return nil
}

func TarInstallMovePackage(e *initData) error {
	workerPath := e.InitOptions.WorkerPath
	tarInstallCmd := fmt.Sprintf("tar -xzvf %s -C %s",
		e.InitOptions.InstallPkgPath, workerPath+constant.EdgeamdDir)
	if err := util.RunLinuxCommand(tarInstallCmd); err != nil {
		return err
	}
	return nil
}

func InitShellPreInstall(e *initData) error {
	workerPath := WorkerHome(e)
	initShell := workerPath + constant.InitInstallShell
	if err := util.RunLinuxShellFile(initShell); err != nil {
		return err
	}
	return nil
}

func SetBinExport(e *initData) error {
	// todo
	return nil
}

func CreateKubeadmConfig(e *initData) error {
	initOption := e.InitOptions
	option := map[string]interface{}{ // todo: 填充这些参数
		"VIP":       initOption.VIP,
		"Repo":      initOption.Registry,
		"Version":   initOption.K8sVersion,
		"ApiServer": initOption.ApiServer,
		"PodCIDR":   initOption.PodCIDR,
		"SvcCIDR":   initOption.ServiceCIDR,
		"CertSANS":  initOption.CertSANS,
		"Masters":   initOption.MasterIP,
		"Master0":   initOption.MasterIP,
	}

	isOverV120, err := kubeclient.IsOverK8sVersion("v1.20.00", e.InitOptions.K8sVersion)
	if err != nil {
		klog.Errorf("IsOverK8sVersion baseK8sVersion: %s, k8sVersion: %s, error: %v", err)
		return err
	}

	kubeadmConfigTemplate := constant.KubeadmTemplateV1beta1
	if isOverV120 {
		kubeadmConfigTemplate = constant.KubeadmTemplateV1beta2
	}

	kubeadmConfigYaml := common.ReadYaml("", kubeadmConfigTemplate) //todo: support user config
	kubeadmConfig, err := kubeclient.ParseString(kubeadmConfigYaml, option)
	if err != nil {
		klog.Errorf("Parse kubeadm config yaml: %s, option: %v, error: %v", kubeadmConfigYaml, option, err)
		return err
	}

	writeKubeadmConfig := fmt.Sprintf(`echo "%s" > %skubeadm-config.yaml`, constant.InstallConf, string(kubeadmConfig))
	if err := util.RunLinuxCommand(writeKubeadmConfig); err != nil {
		klog.Errorf("Run linux command: %s, error: %v", writeKubeadmConfig, err)
		return err
	}
	return nil
}
