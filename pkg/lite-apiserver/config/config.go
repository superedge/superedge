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

package config

import "time"

type LiteServerConfig struct {
	// lite-server default ca path for kubernetes apiserver
	CAFile   string
	CertFile string
	KeyFile  string

	KubeApiserverUrl  string
	KubeApiserverPort int

	Port int
	// if insecure port not 0, will open http debug server
	// if 0, close
	InsecurePort int

	// timeout for proxy to backend. if <=0, use default timeout
	BackendTimeout int

	TLSConfig []TLSKeyPair

	SyncDuration time.Duration

	FileCachePath string
}

type TLSKeyPair struct {
	CertPath string `json:"cert"`
	KeyPath  string `json:"key"`
}
