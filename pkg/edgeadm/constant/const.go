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

package constant

const (
	DeployModeKubeadm = "kubeadm"
	ModeKubeProxy     = "kube-proxy"
)

const (
	ActionChange = "change"
	ActionRevert = "revert"
)

const (
	NodeRoleNode   = "node"
	NodeRoleMaster = "master"
)

const (
	NamespaceDefault    = "default"
	NamespaceEdgeSystem = "edge-system"
	NamespcaeKubeSystem = "kube-system"
)

const (
	ServiceKubernetes  = "kubernetes"
	ServiceTunnelCloud = "tunnel-cloud"
)

const (
	CMKubeConfig               = "kubeconfig.conf"
	CMKubeProxy                = "kube-proxy"
	CMKubeProxyNoEdge          = "kube-proxy-no-edge"
	KubernetesEndpoint         = "kubernetes"
	KubernetesEndpointNoEdge   = "kubernetes-no-edge"
	ConfigMapClusterInfoNoEdge = "cluster-info-no-edge"
)

const (
	KubeCfgPath             = "/etc/kubernetes/"
	KubeEdgePath            = KubeCfgPath + "edge/"
	KubePkiPath             = KubeCfgPath + "pki/"
	KubeadmKeyPath          = KubeCfgPath + "pki/ca.key"
	KubeadmCertPath         = KubeCfgPath + "pki/ca.crt"
	LiteAPIServerCACertPath = KubeCfgPath + "pki/lite-apiserver-ca.crt"
	LiteAPIServerCrtPath    = KubeEdgePath + "lite-apiserver.crt"
	LiteAPIServerKeyPath    = KubeEdgePath + "lite-apiserver.key"
	LiteAPIServerTLSPath    = KubeEdgePath + "tls.json"
)

const (
	SystemServiceDir = "/etc/systemd/system/"
)

// label
const (
	EdgeNodeLabelKey   = "superedge.io/edge-node"
	EdgeMasterLabelKey = "superedge.io/edge-master"
	EdgeChangeLabelKey = "superedge.io/change"
	EdgeLocalHost      = "superedge.io/local-endpoint"
	EdgeLocalPort      = "superedge.io/local-port"

	EdgeNodeLabelValueEnable   = "enable"
	EdgeMasterLabelValueEnable = "enable"
	EdgeChangeLabelValueEnable = "enable"

	UpdateKubeProxyTime        = "superedge.update.kube-proxy"
	KubernetesDefaultRoleLabel = "node-role.kubernetes.io/master"
)

const (
	EdgeCertCM           = "edge-cert"
	KubeAPIClusterIP     = "kube-api-cluster-ip"
	KubeAPICACrt         = "kube-api-ca.crt"
	LiteAPIServerCrt     = "lite-apiserver.crt"
	LiteAPIServerKey     = "lite-apiserver.key"
	LiteAPIServerTLSJSON = "tls.json"
)

const (
	LiteAPIServerStatusCmd  = "systemctl status lite-apiserver.service"
	LiteAPIServerRestartCmd = "systemctl daemon-reload && systemctl restart lite-apiserver.service"
)

const (
	LiteAPIServerAddr = "https://127.0.0.1:51003"
)

const ApplicationGridWrapperServiceAddr = "http://127.0.0.1:51006"

const LiteAPIServerTLSCfg = `[{"key":"/var/lib/kubelet/pki/kubelet-client-current.pem","cert":"/var/lib/kubelet/pki/kubelet-client-current.pem"}]`
