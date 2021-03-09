package init_cmd

import (
	"fmt"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
)

func (e *initData) preInstallHook() error {
	klog.Infof("=========, preInstallHook")
	return e.execHook(constant.PreInstallHook)
}

func (e *initData) execHook(filename string) error {
	klog.V(5).Info("Execute hook script %s", filename)
	f, err := os.OpenFile(constant.EdgeClusterLogFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0744)
	if err != nil {
		klog.Error(err.Error())
	}
	defer f.Close()

	cmd := exec.Command(filename)
	cmd.Stdout = f
	cmd.Stderr = f
	err = cmd.Run()
	if err != nil {
		return err
	}
	return err
}

func (e *initData) tarInstallMovePackage() error {
	workerPath := e.InitOptions.EdgeInitConfig.WorkerPath
	tarInstallCmd := fmt.Sprintf("tar -xzvf %s -C %s",
		e.InitOptions.EdgeInitConfig.InstallPkgPath, workerPath+constant.EdgeamdDir)
	if err := util.RunLinuxCommand(tarInstallCmd); err != nil {
		return err
	}
	return nil
}
