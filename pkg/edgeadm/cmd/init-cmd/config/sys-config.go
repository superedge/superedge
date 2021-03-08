/*
 * Tencent is pleased to support the open source community by making TKE
 * available.
 *
 * Copyright (C) 2018 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License"); you may not use this
 * file except in compliance with the License. You may obtain a copy of the
 * License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations under
 * the License.
 */

package config

import (
	"fmt"
	"github.com/jinzhu/configor"
)

type Config struct {
	EdgeConfig    EdgeConfig
	KubeadmConfig KubeadmConfig
}

type EdgeConfig struct {
	WorkerPath     string `yaml:"workerPath"`
	InstallPkgPath string `yaml:"InstallPkgPath"`
}

type KubeadmConfig struct {
	KubeadmConfPath string `yaml:"kubeadmConfPath"`
}

func New(sysConfigPath string) (*Config, error) {
	config := &Config{}
	if err := configor.Load(config, sysConfigPath); err != nil {
		fmt.Printf("sysConfigPath error, err: %v\n", err)
		return nil, err
	}
	return config, nil
}
