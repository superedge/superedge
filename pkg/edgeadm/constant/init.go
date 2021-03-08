package constant

/*
edgeadm/
├── data
│   ├── edgeadm.json
│   └── edgeadm.log
└── edge-install
    ├── bin
    │   ├── conntrack
    │   ├── kubeadm
    │   ├── kubectl
    │   └── kubelet
    ├── conf
    │   ├── 10-kubeadm.conf
    │   ├── calico.yaml
    │   ├── kubeadm.yaml
    │   ├── kubelet.service
    │   └── net
    │       └── calico.yaml
    ├── container
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
        ├── init.sh
        ├── kubelet-pre-start.sh
        └── master.sh
*/
const (
	EdgeamdDir         = "/edgeadm/"
	DataDir            = EdgeamdDir + "data/"
	EdgeClusterFile    = DataDir + "edgeadm.json"
	EdgeClusterLogFile = DataDir + "edgeadm.log"

	InstallDir = EdgeamdDir + "edge-install/"

	InstallBin       = InstallDir + "bin/"
	InstallConf      = InstallDir + "conf/"
	InstallContainer = InstallDir + "container/"
	InstallImages    = InstallDir + "images/"
	InstallShell     = InstallDir + "shell/"

	HooksDir             = InstallDir + "hooks/"
	PreInstallHook       = HooksDir + "pre-install"
	PostClusterReadyHook = HooksDir + "post-cluster-ready"
	PostInstallHook      = HooksDir + "post-install"
)

const (
	StatusUnknown = "Unknown"
	StatusDoing   = "Doing"
	StatusSuccess = "Success"
	StatusFailed  = "Failed"
)
