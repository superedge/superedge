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
	_, _, err := util.RunLinuxShellFile(constant.PreInstallHook)
	return err
}

func NilFunc(e *initData) error {
	klog.V(5).Infof("I'm nil func")
	return nil
}

func SetNodeHost(e *initData) error {
	klog.V(5).Infof("Start set node host")

	content := "## Start edgeadm hosts config \n"
	for _, host := range e.InitOptions.Hosts {
		content += fmt.Sprintf("%s %s\n", host.IP, host.Domain)
	}
	content += "## End edgeadm hosts config \n"

	if err := util.WriteWithAppend("/etc/hosts", content); err != nil {
		klog.Errorf("Write /etc/hosts error: %v", err)
		return err
	}
	return nil
}

func SetKernelModule(e *initData) error {
	klog.V(5).Infof("Start set kernel module")

	modules := []string{"iptable_nat", "ip_vs", "ip_vs_rr", "ip_vs_wrr", "ip_vs_sh"}
	if _, _, err := util.RunLinuxCommand("modinfo br_netfilter"); err == nil {
		modules = append(modules, "br_netfilter")
	}

	moduleConfig := ""
	for _, m := range modules {
		modprobeCommand := fmt.Sprintf("modprobe %s", m)
		if _, _, err := util.RunLinuxCommand(modprobeCommand); err != nil {
			klog.Errorf("Run linux command: %s, error: %v", modprobeCommand, err)
			return err
		}
		moduleConfig += fmt.Sprintf("%s\n", m)
	}

	if err := util.WriteWithBufio(constant.ModuleFile, moduleConfig); err != nil {
		klog.Errorf("Write file: %s error: %v", constant.ModuleFile, err)
		return err
	}

	return nil
}

func SetSysctl(e *initData) error {
	klog.V(5).Infof("Start set sysctl")

	ipForwardCMD := util.SetFileContent(constant.SysctlFile, "^net.ipv4.ip_forward.*", "net.ipv4.ip_forward = 1")
	if _, _, err := util.RunLinuxCommand(ipForwardCMD); err != nil {
		klog.Errorf("Set sysctl run linux command: %s, error: %s", ipForwardCMD, err)
		return err
	}

	ipTablesCMD := util.SetFileContent(constant.SysctlFile, "^net.bridge.bridge-nf-call-iptables.*", "net.bridge.bridge-nf-call-iptables = 1")
	if _, _, err := util.RunLinuxCommand(ipTablesCMD); err != nil {
		klog.Errorf("Set sysctl run linux command: %s, error: %s", ipTablesCMD, err)
		return err
	}

	workerPath := WorkerHome(e)
	sysctlConf := workerPath + constant.SysctlConf
	if err := util.CopyFile(sysctlConf, constant.SysctlCustomFile); err != nil {
		klog.Errorf("Copy file: %s into %s, error: %s", sysctlConf, constant.SysctlCustomFile, err)
		return err
	}
	return nil
}

func TarInstallMovePackage(e *initData) error {
	workerPath := e.InitOptions.WorkerPath
	tarInstallCmd := fmt.Sprintf("tar -xzvf %s -C %s",
		e.InitOptions.InstallPkgPath, workerPath+constant.EdgeamdDir)
	if _, _, err := util.RunLinuxCommand(tarInstallCmd); err != nil {
		return err
	}
	return nil
}

func InitShellPreInstall(e *initData) error {
	workerPath := WorkerHome(e)
	initShell := workerPath + constant.InitInstallShell
	if _, _, err := util.RunLinuxShellFile(initShell); err != nil {
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

	writeKubeadmConfig := fmt.Sprintf(`echo "%s" > %s/kubeadm-config.yaml`, string(kubeadmConfig), constant.InstallConf)
	if _, _, err := util.RunLinuxCommand(writeKubeadmConfig); err != nil {
		klog.Errorf("Run linux command: %s, error: %v", writeKubeadmConfig, err)
		return err
	}
	return nil
}
