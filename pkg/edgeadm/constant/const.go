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
)

const (
	ACTION_CHANGE = "change"
	ACTION_REVERT = "revert"
)

const (
	NODE_ROLE_NODE   = "node"
	NODE_ROLE_MASTER = "master"
)

const (
	NAMESPACE_DEFAULT     = "default"
	NAMESPACE_KUBE_SYSTEM = "kube-system"
)

const (
	SERVICE_KUBERNETES   = "kubernetes"
	SERVICE_TUNNEL_CLOUD = "tunnel-cloud"
)

const (
	CM_KUBECONFIG_CONF = "kubeconfig.conf"
)

const (
	KubeCfgPath   = "/etc/kubernetes/"
	KubeEdgePath  = KubeCfgPath + "edge/"
	KubeadmKey    = KubeCfgPath + "pki/ca.key"
	KubeadmCert   = KubeCfgPath + "pki/ca.crt"
	EdgeManifests = KubeEdgePath + "manifests/"
)

const (
	SystemServiceDir = "/etc/systemd/system/"
	UsrLocalBinDir   = "/usr/local/bin/"
)

const (
	EdgeLocalHost = "superedge.io/local-endpoint"
	EdgeLocalPort = "superedge.io/local-port"

	UpdateKubeProxyTime        = "superedge.update.kube-proxy"
	KubernetesDefaultRoleLabel = "node-role.kubernetes.io/master"
)

const (
	EDGE_CERT_CM            = "edge-cert"
	KUBE_API_CA_CRT         = "kube-api-ca.crt"
	LITE_API_SERVER_CRT     = "lite-apiserver.crt"
	LITE_API_SERVER_KEY     = "lite-apiserver.key"
	LITE_API_SERVER_TLS_CFG = "tls.json"
)

const (
	EDGE_NODE_KEY                 = "superedge.io/edge"
	KUBERNETES_DEFAULT_ROLE_LABEL = "node-role.kubernetes.io/master"
)

const (
	KUBELET_STATUS_CMD         = "systemctl status kubelet.service"
	KUBELET_RESTART_CMD        = "systemctl restart kubelet.service"
	LITE_APISERVER_STATUS_CMD  = "systemctl status lite-apiserver.service"
	LITE_APISERVER_RESTART_CMD = "systemctl restart lite-apiserver.service"
)

const (
	LiteAPIServerAddr = "https://127.0.0.1:51003"
	KubeletHealthzURl = "http://127.0.0.1:10248/healthz"
)

const (
	MasterHostsFilePath    = "/etc/hosts"
	KubeletStartEnvFile    = "/etc/sysconfig/kubelet"
	KubeadmKubeletEdgeCert = "/etc/kubernetes/edge/"
	KubeadmKubeletConfig   = "/etc/kubernetes/kubelet.conf"
	EdgeadmKubeletConfig   = "/etc/kubernetes/edge/kubelet.config"
)

const (
	CHANGE_KUBELET_KUBECONFIG_ARGS = `KUBELET_KUBECONFIG_ARGS="--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/edge/kubelet.config"`
)

const APPLICAION_GRID_WRAPPER_SERVICE_ADDR = "http://127.0.0.1:51006"

const LiteApiServerTlsCfg = `[{"key":"/var/lib/kubelet/pki/kubelet-client-current.pem","cert":"/var/lib/kubelet/pki/kubelet-client-current.pem"}]`
