package constant

/*
edgeadm-init/
.
├── bin
│   ├── kubectl
│   ├── kubelet
│   └── lite-apiserver
├── cni
│   └── cni-plugins-linux-v0.8.3.tar.gz
├── conf
│   ├── docker
│   │   ├── daemon.josn
│   │   └── docker.service
│   ├── kubeadm
│   │   ├── 10-kubeadm.conf
│   │   └── kubeadm.yaml
│   ├── kubelet
│   │   └── kubelet.service
│   ├── lite-apiserver
│   │   ├── lite-apiserver
│   │   └── lite-apiserver.service
│   └── node
│       └── sysctl.conf
├── container
│   └── docker-18.06-linux-amd64.tar.gz
└── shell
    └── init.sh
*/
const (
	EdgeamdDir         = "edgeadm/"
	EdgeClusterLogFile = EdgeamdDir + "edgeadm.log"
	InstallDir         = EdgeamdDir + "edge-install/"

	InstallBin    = InstallDir + "bin/"
	CNIPluginsDir = InstallDir + "cni/"
	CNIPluginsPKG = CNIPluginsDir + "cni-plugins-*.tar.gz"

	UnZipContainerDstPath = InstallDir + "container/"
	ZipContainerPath      = UnZipContainerDstPath + "docker-*.tar.gz"
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

const KubeProxyPatchJson string = `{"spec":{"affinity":{"nodeAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":{"nodeSelectorTerms":[{"matchExpressions":[{"key":"%s","operator":"DoesNotExist","values":["%s"]}]}]}}}}}`
