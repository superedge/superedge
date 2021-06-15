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
	"fmt"
	"github.com/superedge/superedge/pkg/penetrator/job/conf"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"
	"os/exec"
	"strconv"
	"strings"
)

func SShConnectNode(ip string, port int, secret conf.JobSecret) (*ssh.Client, error) {
	_, err := exec.Command("/bin/sh", "-c", "rm -f /root/.ssh/known_hosts").CombinedOutput()
	if err != nil {
		klog.Errorf("failed to remove ssh private key, error:%v", err)
		return nil, err
	}
	cfg := &ssh.ClientConfig{
		User:            "root",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	if secret.SshKey != nil {
		singer, err := ssh.ParsePrivateKey(secret.SshKey)
		if err != nil {
			klog.Errorf("failed to parse key, error:%v", err)
			return nil, err
		}
		cfg.Auth = []ssh.AuthMethod{ssh.PublicKeys(singer)}
	} else {
		if secret.PassWd != "" {
			cfg.Auth = []ssh.AuthMethod{ssh.Password(strings.Replace(secret.PassWd, "\n", "", -1))}
		}
	}

	return ssh.Dial("tcp", ip+":"+strconv.Itoa(port), cfg)
}

func ScpFile(ip, file string, port int, secret conf.JobSecret) error {
	if secret.KeyPath != "" {
		_, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("scp -P %d -i %s %s root@%s:/root", port, secret.KeyPath, file, ip)).CombinedOutput()
		if err != nil {
			klog.Errorf("failed to scp file,error: %v", err)
			return err
		}
	} else if secret.PassWd != "" {
		_, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("sshpass -f %s scp  -P %d  %s root@%s:/root", secret.PwPath, port, file, ip)).CombinedOutput()
		if err != nil {
			klog.Errorf("failed to scp file, error: %v", err)
			return err
		}
	}
	return nil
}
