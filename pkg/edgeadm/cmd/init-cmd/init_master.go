package init_cmd

import (
	"fmt"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
	"k8s.io/klog/v2"
	"log"
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
		log.Fatal(err.Error())
	}
	defer f.Close()

	cmd := exec.Command(filename)
	cmd.Stdout = f
	cmd.Stderr = f
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (e *initData) tarInstallMovePackage() error {
	tarInstallCmd := fmt.Sprintf("tar -xzvf %s -C %s",
		e.Config.EdgeConfig.InstallPkgPath, e.Config.EdgeConfig.InstallPkgPath)
	if err := util.RunLinuxCommand(tarInstallCmd); err != nil {
		return err
	}

	return nil
}
