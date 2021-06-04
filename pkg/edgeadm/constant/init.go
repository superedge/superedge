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
    └── docker-19.03-linux-arm64.tar.gz
*/
const (
	EdgeamdDir         = "edgeadm/"
	EdgeClusterLogFile = EdgeamdDir + "edgeadm.log"
	InstallDir         = EdgeamdDir + "edge-install/"

	InstallBin    = InstallDir + "bin/"
	CNIPluginsDir = InstallDir + "cni/"
	CNIPluginsPKG = CNIPluginsDir + "cni-plugins-*.tgz"

	UnZipContainerDstPath = InstallDir + "container/"
	ZipContainerPath      = UnZipContainerDstPath + "docker-*.tgz"
	DockerInstallShell    = UnZipContainerDstPath + "docker/install"
)

const (
	PatchDir = "/patch/"
)

const (
	TMPPackgePath = "/tmp/edgeadm-install.tar.gz"
)

const (
	TunnelCoreDNSCIDRIndex = 12
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
