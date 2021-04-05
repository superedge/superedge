package constant

/*
edgeadm/
.
├── bin
│   ├── conntrack
│   ├── kubeadm
│   ├── kubectl
│   └── kubelet
├── conf
│   ├── docker
│   │   ├── daemon.josn
│   │   └── docker.service
│   ├── kubeadm
│   │   ├── 10-kubeadm.conf
│   │   └── kubeadm.yaml
│   ├── kubelet
│   │   └── kubelet.service
│   ├── net
│   │   └── calico.yaml
│   └── node
│       └── sysctl.conf
├── containerd
│   ├── cri-containerd-cni-linux-amd64.tar.gz
│   └── docker-18.06-install-1.4.tgz
├── images
│   ├── application-grid-controller-amd64:pv2.2.0.tar.gz
│   ├── application-grid-wrapper-amd64:pv2.2.0.tar.gz
│   ├── edge-health-admission-amd64:pv2.2.0.tar.gz
│   ├── edge-health-amd64:pv2.2.0.tar.gz
│   ├── flannel-amd64:v0.12.0-edge-1.0.tar.gz
│   ├── hyperkube-amd64:v1.18.2.tar.gz
│   ├── init-dns-amd64:v1.0.0.tar.gz
│   ├── kube-proxy-amd64:v1.18.2.tar.gz
│   ├── pause-amd64:3.2.tar.gz
│   └── tunnel-amd64:pv2.2.0.tar.gz
├── lib64
│   ├── README.md
│   ├── libseccomp.so.2
│   └── libseccomp.so.2.3.1
└── shell
    ├── containerd.sh
    ├── init-install.sh
    ├── init.sh
    ├── kubelet-pre-start.sh
    └── master.sh
*/
const (
	EdgeamdDir         = "/edgeadm/"
	DataDir            = EdgeamdDir + "data/"
	EdgeClusterFile    = DataDir + "edgeadm.json"
	EdgeClusterLogFile = DataDir + "edgeadm.log"

	InstallDir = "/edge-install/"

	InstallBin = InstallDir + "bin/"

	InstallConf = InstallDir + "conf/"
	SysctlConf  = InstallConf + "sysctl.conf"

	UnZipContainerDstPath = InstallDir + "containerd"
	ZipContainerPath      = UnZipContainerDstPath + "/docker.tar.gz"
	DockerInstallShell    = UnZipContainerDstPath + "/docker/install"

	InstallImages = InstallDir + "images/"

	InstallShell     = InstallDir + "shell/"
	InitInstallShell = InstallShell + "init-install.sh"

	HooksDir             = InstallDir + "hooks/"
	PreInstallHook       = HooksDir + "pre-install"
	PostClusterReadyHook = HooksDir + "post-cluster-ready"
	PostInstallHook      = HooksDir + "post-install"
)

const (
	PatchDir = "/patch/"
)

const (
	StatusUnknown = "Unknown"
	StatusDoing   = "Doing"
	StatusSuccess = "Success"
	StatusFailed  = "Failed"
)

const (
	EdgeClusterKubeAPI = "kubeapi.edgeadm.com"
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

const KubeadmTemplateV1beta1 = `
apiVersion: kubeadm.k8s.io/v1beta1
kind: InitConfiguration
localAPIEndpoint:
  advertiseAddress: {{.Master0}}
  bindPort: 6443

---
apiVersion: kubeadm.k8s.io/v1beta1
kind: ClusterConfiguration
kubernetesVersion: {{.Version}}
controlPlaneEndpoint: "{{.ApiServer}}:6443"
imageRepository: {{.Repo}}

networking:
  # dnsDomain: cluster.local
  podSubnet: {{.PodCIDR}}
  serviceSubnet: {{.SvcCIDR}}

apiServer:
  certSANs:
  - 127.0.0.1
  - {{.ApiServer}}
  {{range .Masters -}}
  - {{.}}
  {{end -}}
  {{range .CertSANS -}}
  - {{.}}
  {{end -}}
  - {{.VIP}}
  extraArgs:
    feature-gates: TTLAfterFinished=true
  extraVolumes:
  - name: localtime
    hostPath: /etc/localtime
    mountPath: /etc/localtime
    readOnly: true
    pathType: File

controllerManager:
  extraArgs:
    feature-gates: TTLAfterFinished=true
    experimental-cluster-signing-duration: 876000h
  extraVolumes:
  - hostPath: /etc/localtime
    mountPath: /etc/localtime
    name: localtime
    readOnly: true
    pathType: File

scheduler:
  extraArgs:
    feature-gates: TTLAfterFinished=true
  extraVolumes:
  - hostPath: /etc/localtime
    mountPath: /etc/localtime
    name: localtime
    readOnly: true
    pathType: File

---
apiVersion: kubeproxy.config.k8s.io/v1alpha1
kind: KubeProxyConfiguration
mode: "ipvs"
ipvs:
  excludeCIDRs:
  - "{{.VIP}}/32"
`

const KubeadmTemplateV1beta2 = `
apiVersion: kubeadm.k8s.io/v1beta2
kind: InitConfiguration
localAPIEndpoint:
  advertiseAddress: {{.Master0}}
  bindPort: 6443
nodeRegistration:
  criSocket: /run/containerd/containerd.sock
---

apiVersion: kubeadm.k8s.io/v1beta2
kind: ClusterConfiguration
kubernetesVersion: {{.Version}}
controlPlaneEndpoint: "{{.ApiServer}}:6443"
imageRepository: {{.Repo}}
networking:
  # dnsDomain: cluster.local
  podSubnet: {{.PodCIDR}}
  serviceSubnet: {{.SvcCIDR}}

apiServer:
  certSANs:
  - 127.0.0.1
  - {{.ApiServer}}
  {{range .Masters -}}
  - {{.}}
  {{end -}}
  {{range .CertSANS -}}
  - {{.}}
  {{end -}}
  - {{.VIP}}
  extraArgs:
    feature-gates: TTLAfterFinished=true
  extraVolumes:
  - name: localtime
    hostPath: /etc/localtime
    mountPath: /etc/localtime
    readOnly: true
    pathType: File

controllerManager:
  extraArgs:
    feature-gates: TTLAfterFinished=true
    experimental-cluster-signing-duration: 876000h
  extraVolumes:
  - hostPath: /etc/localtime
    mountPath: /etc/localtime
    name: localtime
    readOnly: true
    pathType: File

scheduler:
  extraArgs:
    feature-gates: TTLAfterFinished=true
  extraVolumes:
  - hostPath: /etc/localtime
    mountPath: /etc/localtime
    name: localtime
    readOnly: true
    pathType: File

---
apiVersion: kubeproxy.config.k8s.io/v1alpha1
kind: KubeProxyConfiguration
mode: "ipvs"
ipvs:
  excludeCIDRs:
  - "{{.VIP}}/32"
`
