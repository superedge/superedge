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

package conf

import (
	"github.com/BurntSushi/toml"
	"github.com/superedge/superedge/pkg/penetrator/constants"
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
)

var JobConf *JobConfig

type JobConfig struct {
	NodesIps      map[string]string `toml:"nodesIps"`
	SSHPort       int               `toml:"sshPort"`
	NodeLabel     string            `toml:"nodeLabel"`
	AdmToken      string            `toml:"admToken"`
	ApiserverAddr string            `toml:"apiserverAddr"`
	ApiserverPort string            `toml:"apiserverPort"`
	CaHash        string            `toml:"caHash"`
	Secret        JobSecret
}

type JobSecret struct {
	PassWd  string
	PwPath  string
	SshKey  []byte
	KeyPath string
}

func InitJobConfig(jobpath, secretpath string) error {
	// Load sshkey or passwd
	secret := JobSecret{}
	_, keyerr := os.Stat(secretpath + constants.SshKey)
	if keyerr != nil {
		if os.IsNotExist(keyerr) {
			passwd, err := ioutil.ReadFile(secretpath + constants.PassWd)
			if err != nil {
				klog.Errorf("failed to read passwd file, error: %v", err)
				return err
			}
			secret.PassWd = string(passwd)
			secret.PwPath = secretpath + constants.PassWd
		} else {
			return keyerr
		}
	} else {
		sshkey, err := ioutil.ReadFile(secretpath + constants.SshKey)
		if err != nil {
			klog.Errorf("failed to read sshkey file, error: %v", err)
			return err
		}
		secret.SshKey = sshkey
		secret.KeyPath = secretpath + constants.SshKey
	}

	//Load job config
	job := &JobConfig{}
	conf, err := ioutil.ReadFile(jobpath + constants.JobConf)
	if err != nil {
		klog.Errorf("failed to read job config file, error: %v", err)
		return err
	}
	err = toml.Unmarshal(conf, job)
	if err != nil {
		klog.Errorf("failed to load job config file, error: %v", err)
		return err
	}

	job.Secret = secret
	JobConf = job

	return nil
}
