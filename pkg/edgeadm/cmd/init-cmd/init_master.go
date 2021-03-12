package init_cmd

import (
	"fmt"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
	"io/ioutil"
	"k8s.io/klog/v2"
)

func WorkerHome(e *initData) string {
	return e.InitOptions.EdgeInitConfig.WorkerPath
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
	workerPath := e.InitOptions.EdgeInitConfig.WorkerPath
	tarInstallCmd := fmt.Sprintf("tar -xzvf %s -C %s",
		e.InitOptions.EdgeInitConfig.InstallPkgPath, workerPath+constant.EdgeamdDir)
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
	var templateData string
	kubeadmConfig := e.InitOptions.KubeadmConfig.KubeadmConfPath
	if kubeadmConfig == "" {

	} else {
		fileData, err := ioutil.ReadFile(KubeadmFile)
		defer func() {
			if r := recover(); r != nil {
				logger.Error("[globals]template file read failed:", err)
			}
		}()
		if err != nil {
			panic(1)
		}
		templateData = string(TemplateFromTemplateContent(string(fileData)))
	}
	///////////////////////
	kubeadmConfigTemplate := string(Template()) //todo: 到填充模板写/root/kubeadm-config.yaml这一步了。。。
	///////////////////////

	cmd := fmt.Sprintf(`echo "%s" > /root/kubeadm-config.yaml`, templateData)
	//cmd := "echo \"" + templateData + "\" > /root/kubeadm-config.yaml"
	_ = SSHConfig.CmdAsync(s.Masters[0], cmd)
	//读取模板数据
	kubeadm := KubeadmDataFromYaml(templateData)
	if kubeadm != nil {
		DnsDomain = kubeadm.Networking.DnsDomain
		ApiServerCertSANs = kubeadm.ApiServer.CertSANs
	} else {
		logger.Warn("decode certSANs from config failed, using default SANs")
		ApiServerCe
	}
	return nil
}
