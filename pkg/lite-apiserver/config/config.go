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

type LiteServerConfig struct {
	// lite-server default ca path for kubernetes apiserver
	// CAFile defines the certificate authority
	// that servers use if required to verify a client certificate(e.g. kubelet, kube-proxy).
	// the same with client-ca-file of kube-apiserver
	CAFile string
	// CertFile the tls cert file for lite-apiserver's https server
	CertFile string
	// KeyFile  the tls key file for lite-apiserver's https server
	KeyFile string

	// ApiserverCAFile used to verify kube-apiserver server tls.
	// use CAFile if ApiserverCAFile is not assigned.
	ApiserverCAFile string

	KubeApiserverUrl  string
	KubeApiserverPort int

	// the address list of lite-apiserver listen
	ListenAddress []string
	// Port the https port of lite-apiserver
	Port int

	// BackendTimeout timeout for the request from lite-apiserver to kube-apiserver.
	// if <=0, use default timeout
	BackendTimeout int

	// Profiling pprof for the lite-apiserver.
	// default false
	Profiling bool

	TLSConfig []TLSKeyPair

	ModifyRequestAccept bool

	CacheType         string
	FileCachePath     string
	BadgerCachePath   string
	BoltCacheFile     string
	PebbleCachePath   string
	NetworkInterface  string
	Insecure          bool
	URLMultiplexCache []string
}

type TLSKeyPair struct {
	CertPath string `json:"cert"`
	KeyPath  string `json:"key"`
}
