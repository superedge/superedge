/*
Copyright 2020 The SuperEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"bytes"
	"fmt"
	"k8s.io/klog/v2"
	"os/exec"
)

// SetFileContent generates cmd for set file content.
func SetFileContent(file, pattern, content string) string {
	return fmt.Sprintf(`grep -Pq "%s" %s && sed -i "s;%s;%s;g" %s|| echo "%s" >> %s`,
		pattern, file,
		pattern, content, file,
		content, file)
}

func RunLinuxCommand(command string) (string, string, error) {
	var outBuff, errBuff bytes.Buffer
	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.Stdout, cmd.Stderr = &outBuff, &errBuff

	defer func() {
		klog.V(4).Infof("Run command: '%s' \n "+
			"stdout: %s\n stderr: %s\n", command, outBuff.String(), errBuff.String())
	}()

	//Run cmd
	if err := cmd.Start(); err != nil {
		klog.Warningf("Exec command: %s, error: %v", command, err)
		return "", "", err
	}

	//Wait cmd run finish
	if err := cmd.Wait(); err != nil {
		klog.Warningf("Wait command: %s exec finish error: %v", command, err)
		return "", "", err
	}

	return outBuff.String(), errBuff.String(), nil
}

func RunLinuxShellFile(filename string) (string, string, error) {
	klog.V(5).Infof("Run shell script %s", filename)

	cmd := exec.Command(filename)
	var outBuff, errBuff bytes.Buffer
	cmd.Stdout, cmd.Stderr = &outBuff, &errBuff

	defer func() {
		klog.V(4).Infof("Run shell script %s \n"+
			"stdout: %s\n stderr: %s\n", filename, outBuff.String(), errBuff.String())
	}()

	err := cmd.Run()
	if err != nil {
		return "", "", err
	}
	return outBuff.String(), errBuff.String(), nil
}
