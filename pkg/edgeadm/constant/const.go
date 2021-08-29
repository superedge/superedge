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

import "github.com/superedge/superedge/pkg/edgeadm/constant/manifests/edgex"

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
	NamespaceKubeSystem = "kube-system"
	NamespaceKubePublic = "kube-public"
	NamespaceEdgex      = "edgex"
)

const (
	ServiceKubernetes    = "kubernetes"
	ServiceTunnelCloud   = "tunnel-cloud"
	ServiceEdgeCoreDNS   = "edge-coredns"
	ServiceTunnelCoreDNS = "tunnel-coredns"
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
	KubeManifestsPath       = KubeCfgPath + "manifests/"
	KubeVIPPath             = KubeManifestsPath + "vip.yaml"
)

const (
	SystemServiceDir = "/etc/systemd/system/"
)

// label
const (
	EdgeNodeLabelKey     = "superedge.io/edge-node"
	EdgeMasterLabelKey   = "superedge.io/edge-master"
	EdgeChangeLabelKey   = "superedge.io/change"
	EdgehostnameLabelKey = "superedge.io.hostname"
	EdgeLocalHost        = "superedge.io/local-endpoint"
	EdgeLocalPort        = "superedge.io/local-port"

	EdgeNodeLabelValueEnable   = "enable"
	EdgeMasterLabelValueEnable = "enable"
	EdgeChangeLabelValueEnable = "enable"

	UpdateKubeProxyTime        = "superedge.update.kube-proxy"
	KubernetesDefaultRoleLabel = "node-role.kubernetes.io/master"
)

const (
	EdgeCertCM             = "edge-info"
	LiteAPIServerTLSJSON   = "tls.json"
	KubeAPIClusterIP       = "kube-api-cluster-ip"
	KubeAPICACrt           = "kube-api-ca.crt"
	LiteAPIServerCrt       = "lite-apiserver.crt"
	LiteAPIServerKey       = "lite-apiserver.key"
	EdgeCoreDNSClusterIP   = "edge-coredns-cluster-ip"
	TunnelCoreDNSClusterIP = "tunnel-coredns-cluster-ip"
)

const (
	HostDNSBeginMark        = "# begin (generated by SuperEdge)"
	HostDNSEndMark          = "# end (generated by SuperEdge)"
	ResetDNSCmd             = "sed -i '/" + HostDNSBeginMark + "/,/" + HostDNSEndMark + "/d' " + HostsFilePath
	HostsFilePath           = "/etc/hosts"
	LiteAPIServerStatusCmd  = "systemctl status lite-apiserver.service"
	LiteAPIServerRestartCmd = "systemctl daemon-reload && systemctl restart lite-apiserver.service && systemctl enable lite-apiserver.service"
)

const (
	LiteAPIServerAddr    = "https://127.0.0.1:51003"
	AddonAPIServerDomain = "kubernetes.default"
)

const (
	TunnelNodePortNameGRPG = "grpc"
)

const ApplicationGridWrapperServiceAddr = "http://127.0.0.1:51006"

const LiteAPIServerTLSCfg = `[{"key":"/var/lib/kubelet/pki/kubelet-client-current.pem","cert":"/var/lib/kubelet/pki/kubelet-client-current.pem"}]`

const ImageRepository = "superedge.tencentcloudcr.com/superedge"

const (
	Configmap   = 0
	App         = 1
	Core        = 2
	Device      = 3
	Support     = 4
	Sysmgmt     = 5
	Ui          = 6
	Mqtt        = 7
	Completely  = 8
)

const (
	SerCount   =  9

	Sername0 = edgex.EDGEX_CONFIGMAP
	Sername1 = edgex.EDGEX_APP
	Sername2 = edgex.EDGEX_CORE
	Sername3 = edgex.EDGEX_DEVICE
	Sername4 = edgex.EDGEX_SUPPORT
	Sername5 = edgex.EDGEX_SYS_MGMT
	Sername6 = edgex.EDGEX_UI
	Sername7 = edgex.EDGEX_MQTT
	Sername8 = edgex.EDGEX_CONFIGMAP

	Seryaml0 = edgex.EDGEX_CONFIGMAP_YAML
	Seryaml1 = edgex.EDGEX_APP_YAML
	Seryaml2 = edgex.EDGEX_CORE_YAML
	Seryaml3 = edgex.EDGEX_DEVICE_YAML
	Seryaml4 = edgex.EDGEX_SUPPORT_YAML
	Seryaml5 = edgex.EDGEX_SYS_MGMT_YAML
	Seryaml6 = edgex.EDGEX_UI_YAML
	Seryaml7 = edgex.EDGEX_MQTT_YAML
	Seryaml8 = edgex.EDGEX_CONFIGMAP_YAML
)