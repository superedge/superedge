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

type UserConfig struct {
	BaseConfig BaseConfig `yaml:"baseConfig"`
	SelfConfig SelfConfig `yaml:"selfConfig"`
}

// BaseConfig
type BaseConfig struct {
	Organization  string        `yaml:"organization"`
	Etcd          ETCD          `yaml:"etcd"`
	ImageRegistry ImageRegistry `yaml:"imageRegistry"`
}

type ETCD struct {
	ETCDServers string `yaml:"etcdServers"  required:"true"`
}

type ImageRegistry struct {
	RegistrySecret    string `yaml:"registrySecret"`
	IsSecureRegistry  bool   `yaml:"isSecureRegistry"  default:false required:"true"`
	RegistryNamespace string `yaml:"registryNamespace" required:"true"`
	RegistryServer    string `yaml:"registryServer"    required:"true"`
	RegistryUsername  string `yaml:"registryUsername"`
	RegistryPassword  string `yaml:"registryPassword"`
}

// SelfConfig
type SelfConfig struct {
	NodeHosts    []NodeHosts  `yaml:"nodeHosts"`
	DockerConfig DockerConfig `yaml:"dockerConfig"`
}

type NodeHosts struct {
	IP     string `yaml:"ip"     required:"true"`
	Domain string `yaml:"domain" required:"true"`
}

type DockerConfig struct {
	RegistryMirrors    []string `yaml:"registryMirrors"`
	InsecureRegistries []string `yaml:"insecureRegistries"`
}
