#!/bin/bash
# the script is placed in the edgeadm-linux-amd64-v0.3.0/edge-install/container/ directory of the installation package

set -o errexit
set -o nounset
set -o pipefail

setFileContent() {
	local file=$1
	local pattern=$2
	local content=$3
	grep -Pq "$pattern" $file && sed -i "s;$pattern;$content;g" $file|| echo "content" >> $file
}
command_exists() {
	command -v "$@" > /dev/null 2>&1
}
# 1. clear node
if ! command_exists ifconfig; then
	yum install -y net-tools
fi
if ! command_exists ip; then
	yum install -y iproute
fi
rm -rf /var/lib/cni/
rm -rf /etc/cni/
ifconfig cni0 down || true
ifconfig flannel.1 down || true
ifconfig docker0 down || true
ip link delete cni0 || true
ip link delete flannel.1 || true
ip link delete docker0 || true

# 2. swap off
swapoff -a && sed -i "s/^[^#]*swap/#&/" /etc/fstab

# 3. disable selinux
sed -i 's/SELINUX=enforcing/SELINUX=disabled/g' /etc/sysconfig/selinux /etc/selinux/config && setenforce 0 || true

# 4. enable kubelet
systemctl enable kubelet

# 5. set sysctl
setFileContent /etc/sysctl.conf "^net.ipv4.ip_forward.*" "net.ipv4.ip_forward = 1"
setFileContent /etc/sysctl.conf "^net.bridge.bridge-nf-call-iptables.*" "net.bridge.bridge-nf-call-iptables = 1"
cat <<EOF >/etc/sysctl.d/k8s.conf
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
EOF

sysctl --system

# 6. load-kernel
if ! [ -d /etc/sysconfig/modules/ ]; then
	mkdir -p /etc/sysconfig/modules
fi
cat <<EOF >/etc/sysconfig/modules/ipvs.modules
modprobe -- iptable_nat
modprobe -- ip_vs
modprobe -- ip_vs_sh
modprobe -- ip_vs_rr
modprobe -- ip_vs_wrr
modprobe -- nf_conntrack_ipv4
EOF

if modinfo br_netfilter > /dev/null; then
	echo "modprobe -- br_netfilter" >> /etc/sysconfig/modules/ipvs.modules
fi

chmod 755 /etc/sysconfig/modules/ipvs.modules &&
    source /etc/sysconfig/modules/ipvs.modules &&
    lsmod | grep -e ip_vs -e nf_conntrack_ipv4


