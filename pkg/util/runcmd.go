package util

import (
	"bytes"
	"k8s.io/klog/v2"
	"os/exec"
)

func RunLinuxCommand(command string) error {
	var outBuff bytes.Buffer
	cmd := exec.Command("/bin/bash", "-c", command)

	cmd.Stdout = &outBuff
	cmd.Stderr = &outBuff
	defer func() {
		defer klog.V(4).Infof("Run command: '%s' output: \n %s", command, outBuff.String())
	}()

	//Run cmd
	if err := cmd.Start(); err != nil {
		klog.Errorf("Exec command: %s, error: %v", command, err)
		return err
	}

	//Wait cmd run finish
	if err := cmd.Wait(); err != nil {
		klog.Errorf("Wait command: %s exec finish error: %v", command, err)
		return err
	}

	return nil
}

func RunLinuxShellFile(filename string) error {
	klog.V(5).Infof("Run shell script %s", filename)

	var outBuff bytes.Buffer
	cmd := exec.Command(filename)
	cmd.Stdout = &outBuff
	cmd.Stderr = &outBuff

	defer func() {
		defer klog.V(4).Infof("Run shell script %s output:\n %s", filename, outBuff.String())
	}()

	err := cmd.Run()
	if err != nil {
		return err
	}
	return err
}
