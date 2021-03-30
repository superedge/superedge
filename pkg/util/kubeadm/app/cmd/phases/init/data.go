/*
Copyright 2019 The Kubernetes Authors.

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

package phases

import (
	cmd "github.com/superedge/superedge/pkg/edgeadm/cmd"
	"io"

	kubeadmapi "github.com/superedge/superedge/pkg/util/kubeadm/app/apis/kubeadm"
	"k8s.io/apimachinery/pkg/util/sets"
	clientset "k8s.io/client-go/kubernetes"
)

// InitData is the interface to use for init phases.
// The "initData" type from "cmd/init.go" must satisfy this interface.
type InitData interface {
	UploadCerts() bool
	CertificateKey() string
	SetCertificateKey(key string)
	SkipCertificateKeyPrint() bool
	Cfg() *kubeadmapi.InitConfiguration
	DryRun() bool
	SkipTokenPrint() bool
	IgnorePreflightErrors() sets.String
	CertificateWriteDir() string
	CertificateDir() string
	KubeConfigDir() string
	KubeConfigPath() string
	ManifestDir() string
	KubeletDir() string
	ExternalCA() bool
	OutputWriter() io.Writer
	Client() (clientset.Interface, error)
	Tokens() []string
	KustomizeDir() string
	PatchesDir() string
	EdgeadmConf() *cmd.EdgeadmConfig
}