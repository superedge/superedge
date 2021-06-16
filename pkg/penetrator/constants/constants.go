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

package constants

import "time"

// crd
const (
	NodeLabel                      = "app.superedge.io/node-label"
	AnnotationAddNodeJobName       = "app.superedge.io/job-name"
	AnnotationAddNodeConfigmapName = "app.superedge.io/configmap-name"
	SshKey                         = "sshkey"
	PassWd                         = "passwd"
	Expiration                     = 24 * time.Hour
)

// job
const (
	JobConf        = "job.toml"
	JobName        = "JOB_NAME"
	JobNameSpace   = "JOB_NAMESPACE"
	BufferSize     = 1e9
	InstallPackage = "/etc/superedge/penetrator/job/install/edgeadm-"
	AddNodeScript  = "/etc/superedge/penetrator/job/script/addnode.sh"
)

// operator

const (
	EnvOperatorNamespace = "OPERATOR_NAMESPACE"
	EnvOperatorPodName   = "OPERATOR_POD_NAME"
	NameSpaceEdge        = "edge-system"
)

//yaml path

const (
	BootStrapTokenSecert  = "/etc/superedge/penetrator/manifests/bootstraper-token-secret.yaml"
	DirectAddNodeJob      = "/etc/superedge/penetrator/manifests/direct-addnode-job.yaml"
	SpringboardAddNodeJob = "/etc/superedge/penetrator/manifests/springboard-addnode-job.yaml"
)

//arch
const (
	Amd64   = "amd64"
	Arm64   = "arm64"
	Aarch64 = "aarch64"
	X86_64  = "x86_64"
)
