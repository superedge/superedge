package constant

const (
	SysctlFile       = "/etc/sysctl.conf"
	ModuleFile       = "/etc/modules-load.d/edgeadm.conf"
	SysctlCustomFile = "/etc/sysctl.d/99-edgeadm.conf"
)

const SWAP_OFF = `
swapoff -a && sed -i "s/^[^#]*swap/#&/" /etc/fstab
`

const STOP_FIREWALL = `
systemctl stop firewalld && systemctl disable firewalld
`

const SYS_CONF = `
kernel.sem = "250 32000 32 1024"
net.core.netdev_max_backlog = 20000
net.core.rmem_default = 262144
net.core.rmem_max = 16777216
net.core.somaxconn = 2048
net.core.wmem_default = 262144
net.core.wmem_max = 16777216
net.ipv4.tcp_ﬁn_timeout = 15
net.ipv4.tcp_max_orphans = 131072
net.ipv4.tcp_max_syn_backlog = 16384
net.ipv4.tcp_mem = "786432 2097152 3145728"
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

const KERNEL_MODULE = `
chmod 755 /etc/sysconfig/modules/ipvs.modules && \n
bash /etc/sysconfig/modules/ipvs.modules && \n
lsmod | grep -e ip_vs -e nf_conntrack_ipv4
`
