package constant

/* edgeadm kube-linux-*.tar.gz Directory Structure
edge-install/
├── bin
│   ├── conntrack
│   ├── kubectl
│   ├── kubelet
│   └── lite-apiserver
├── cni
│   └── cni-plugins-linux-v0.8.3.tar.gz
└── container
│   └── docker-19.03-linux-arm64.tar.gz
└── script
    └── init-node.sh
*/
const (
	EdgeamdDir         = "edgeadm/"
	EdgeClusterLogFile = EdgeamdDir + "edgeadm.log"
	InstallDir         = EdgeamdDir + "edge-install/"

	InstallBin            = InstallDir + "bin/"
	CNIPluginsDir         = InstallDir + "cni/"
	InstallConfDir        = InstallDir + "conf/"
	UnZipContainerDstPath = InstallDir + "container/"

	// cni plugins pkg dir
	CNIPluginsPKG = CNIPluginsDir + "cni-plugins-*.tgz"

	// install conf dir
	KubeSchedulerConf = InstallConfDir + "kube-scheduler/"

	// docker runtime dir
	ZipContainerPath   = UnZipContainerDstPath + "docker-*.tgz"
	DockerInstallShell = UnZipContainerDstPath + "docker/install"

	// containerd runtime dir
	ContainerdZipPath      = UnZipContainerDstPath + "containerd-*.tgz"
	ContainerdInstallShell = UnZipContainerDstPath + "containerd/install"

	// script dir
	ScriptShellPath = InstallDir + "script/"
	InitNodeShell   = ScriptShellPath + "init-node.sh"
)

const (
	PatchDir = "/patch/"
)

const (
	TMPPackgePath = "/tmp/edgeadm-install.tar.gz"
)

const (
	SchedulerConfigDir = "/etc/kubernetes/kube-scheduler/"
	SchedulerConfig    = SchedulerConfigDir + "kube-scheduler-config.yaml"
	SchedulerPolicy    = SchedulerConfigDir + "kube-scheduler-policy.cfg"
)

const (
	TunnelCoreDNSCIDRIndex = 11
	KubeAPIServerPatch     = "kube-apiserver0+merge.yaml"
)

const KubeAPIServerPatchYaml = `
apiVersion: v1
kind: Pod
metadata:
  name: kube-apiserver
  namespace: kube-system
spec:
  dnsConfig:
    nameservers:
    - {{.TunnelCoreDNSClusterIP}}
  dnsPolicy: None
`

const KubeProxyPatchJson string = `{"spec":{"template":{"spec":{"affinity":{"nodeAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":{"nodeSelectorTerms":[{"matchExpressions":[{"key":"%s","operator":"%s"}]}]}}}}}}}`

const KubeProxyRecoverJson string = `[{"op": "remove", "path": "/spec/template/spec/affinity"}]`
