package constant

const (
	CNIDir             = "/opt/cni/bin/"
	SysctlFile         = "/etc/sysctl.conf"
	SysctlK8sConf      = "/etc/sysctl.d/k8s.conf"
	KubeletSysConf     = "/etc/sysconfig/kubelet"
	IPvsModulesFile    = "/etc/sysconfig/modules/ipvs.modules"
	KubeletServiceFile = "/usr/lib/systemd/system/kubelet.service"
	KubeadmConfFile    = "/usr/lib/systemd/system/kubelet.service.d/10-kubeadm.conf"
)

const SwapOff = `swapoff -a && sed -i "s/^[^#]*swap/#&/" /etc/fstab`

const StopFireWall = `systemctl stop firewalld && systemctl disable firewalld`

const SysConf = `
kernel.sem = 250 32000 32 1024
net.core.netdev_max_backlog = 20000
net.core.rmem_default = 262144
net.core.rmem_max = 16777216
net.core.somaxconn = 2048
net.core.wmem_default = 262144
net.core.wmem_max = 16777216
net.ipv4.tcp_ﬁn_timeout = 15
net.ipv4.tcp_max_orphans = 131072
net.ipv4.tcp_max_syn_backlog = 16384
net.ipv4.tcp_mem = 786432 2097152 3145728
net.ipv4.tcp_tw_reuse = 1
net.ipv4.ip_forward = 1
net.netﬁlter.nf_conntrack_max = 524288
net.bridge.bridge-nf-call-iptables = 1
net.bridge.bridge-nf-call-ip6tables = 1
fs.inotify.max_user_watches = 1048576
fs.may_detach_mounts = 1
vm.dirty_background_ratio = 5
vm.dirty_ratio = 10
vm.swappiness = 0
vm.max_map_count = 262144
`

const KernelModule = `chmod 755 /etc/sysconfig/modules/ipvs.modules &&
source /etc/sysconfig/modules/ipvs.modules &&
lsmod | grep -e ip_vs -e nf_conntrack_ipv4
`

const KubeletService = `
[Unit]
Description=kubelet: The Kubernetes Node Agent
Documentation=https://kubernetes.io/docs/

[Service]
User=root
ExecStart=/usr/bin/kubelet
Restart=always
StartLimitInterval=0
RestartSec=10

[Install]
WantedBy=multi-user.target
`

const KubeletSys = `
KUBELET_EXTRA_ARGS=
`

const KubeadmConfig = `
# Note: This dropin only works with kubeadm and kubelet v1.11+
[Service]
Environment="KUBELET_KUBECONFIG_ARGS=--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf"
Environment="KUBELET_CONFIG_ARGS=--config=/var/lib/kubelet/config.yaml"
# This is a file that "kubeadm init" and "kubeadm join" generates at runtime, populating the KUBELET_KUBEADM_ARGS variable dynamically
EnvironmentFile=-/var/lib/kubelet/kubeadm-flags.env
# This is a file that the user can use for overrides of the kubelet args as a last resort. Preferably, the user should use
# the .NodeRegistration.KubeletExtraArgs object in the configuration files instead. KUBELET_EXTRA_ARGS should be sourced from this file.
EnvironmentFile=-/etc/sysconfig/kubelet
ExecStart=
ExecStart=/usr/bin/kubelet $KUBELET_KUBECONFIG_ARGS $KUBELET_CONFIG_ARGS $KUBELET_KUBEADM_ARGS $KUBELET_EXTRA_ARGS
`
