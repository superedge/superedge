# 一键安装边缘独立Kubernetes 集群

* [一键安装边缘 Kubernetes 集群](#一键化安装边缘-kubernetes-集群)
    * [1\. 背景](#1-背景)
    * [2\. 架构设计](#2-架构设计)
        * [2\.1 初衷](#21-初衷)
        * [2\.2 目标](#22-目标)
        * [2\.3 原则](#23-原则)
        * [2\.4 设计方案](#24-设计方案终版)
    * [3\. 使用方法](#3-使用方法)
        * [3\.1 用 edgeadm 安装边缘 Kubernetes 集群](#31-用-edgeadm-安装边缘-kubernetes-集群)
        * [3\.2 用 edgeadm 安装边缘高可用 Kubernetes 集群](#32-用-edgeadm-安装边缘高可用-kubernetes-集群)

## 1. 背景

目前，所有的边缘计算容器管理开源项目在使用上均存在一个默认的前提，即用户需要提前准备好一个Kubernetes集群，然后再通过工具或者其他方式在集群中[部署](https://github.com/superedge/superedge/blob/main/docs/installation/install_via_edgeadm_CN.md)边缘容器组件，该前提降低了用户幸福感，为了提升用户幸福感，edgeadm打算提供一个一步到位的能力，将搭建集群和部署边缘容器组件集成到一个流程中。

是什么原因促使SuperEdge团队做出此决定呢？主要是业界现有方案均有如下缺点：

-   使用要求高

    - 要求用户必须提前准备一个Kubernetes集群，对于许多搭建Kubernetes集群不太熟练的用户来说不太友好
    - 即使是能熟练搭建的用户也需要消耗搭建集群和部署边缘容器组件两份时间和精力
-   不方便添加新边缘节点

    -   目前添加边缘节点需要分为两步：1）使用Kubeadm等工具添加节点；2）使用Edgeadm等工具将新节点转化成边缘节点。
    -   过程即麻烦又容易出错，操作不当甚至会造成集群中已有节点和服务发生异常。

---

## 2. 架构设计

针对上述问题，为了降低用户使用边缘 Kubernetes 集群的门槛，让边缘 Kubernetes 集群具备生产能力，我们和云原生社区的同学一道设计了一种一键就可以部署出来一个边缘 Kubernetes 集群的方案，完全屏蔽安装细节，让用户可以零门槛的体验边缘能力。

### 2.1 初衷

-   让用户很简单，无门槛的使用边缘 Kubernetes 集群，并能在生产环境真正把边缘能力用起来；

### 2.2 目标

-   一键化使用

    -   能够一键搭建起一个边缘 Kubernetes 集群
    -   能够一键添加边缘节点；

-   两种安装方式

    -   支持在线安装，
    -   支持离线安装，让私有化环境也能很简单；

-   可生产使用

    不要封装太多，可以让想使用边缘 Kubernetes 集群的企业能在内部系统进行简单集成，就生成可用；

-   零学习成本

    我们提供的工具Edgeadm尽可能的和 Kubeadm 的使用方式保持一致，让用户无额外学习成本，就可以直接使用；

### 2.3 原则

-   不修改 kubeadm 源码
    -   尽量复用和引用 kubeadm 的源码，尽量不修改 kubeadm 的源码，避免后面升级的隐患；
    -   基于kubeadm但又高于 kubeadm，不必被 kubeadm 的设计所局限，只要能让用户更简单都可以被允许；
-   允许用户选择是否部署边缘组件
    
- 允许用户定制边缘容器组件yaml配置

### 2.4 设计与实现

---

我们仔细研究了Kubeadm的源码，发现可以借用Kubeadm创建原生 Kubernetes集群、Join节点、workflow思想来达到一键部署边缘 Kubernetes集群，并且可以分步去执行安装步骤。这正是我们想要的简单、灵活、低学习成本的部署方案。于是我们站在巨人的肩膀上，复用Kubeadm的源码，设计出了如下的方案。

<img src="https://raw.githubusercontent.com/attlee-wang/myimage/master/image/20210419101917.png" alt="image-20210419101917603" style="zoom:50%;" />

>   其中 `Kubeadm init cluster/join node`部分完全复用了kubadm的源码，所有逻辑和Kubeadm完全相同。

这个方案有如下几个优点：

-   完全兼容Kubeadm

    我们只是站在Kubeadm的肩膀上，在Kubeadm init/join之前设置了一些边缘集群需要的配置参数，将初始化Master或Node节点自动化，安装了容器运行时。在Kubeadm init/join完成之后，安装了CNI网络插件和部署了相应的边缘能力组件。

    我们以Go Mod方式引用了Kubeadm源码，整个过程中并未对Kubeadm的源码修改过一行，完全的原生，为后面升级更高版本的Kubeadm做好了准备。

-   用起来简单、灵活、自动化

    edgeadm init集群和Join节点完全保留了Kubeadm init/join原有的参数和流程，只是自动了初始化节点和安装容器运行时，可以用`edgeadm --enable-edge=fasle`参数来一键化安装原生Kubernetes集群， 也可以用`edgeadm --enable-edge=true`参数一键化来安装边缘Kubernetes集群。

    可以Join任何只要能够访问到Kube-api-server位于任何位置的节点, 也可以Join mastter。Join master也延续了Kubeadm的的方式，搭建高可用的节点可以在需要的时候，直接用Join master去扩容Master节点，实现高可用。

-   无学习成本，和kubeadm的使用完全相同

    因为`Kubeadm init cluster/join node`部分完全复用了kubadm的源码，所有逻辑和Kubeadm完全相同，完全保留了kubeadm的使用习惯和所有flag参数，用法和kubeadm使用完全一样，没有任何新的学习成本，用户可以按Kubeadm的参数去自定义边缘 Kubernetes 集群。

---

## 3. 使用方法

### 3.1 用 edgeadm 安装边缘 Kubernetes 集群

- 下载 edgeadm 二进制包
```shell
rm -rf edgeadm && \
wget https://attlee-1251707795.cos.ap-chengdu.myqcloud.com/superedge/v0.3.0/edgeadm &&  \
chmod +x edgeadm
```
- 安装边缘 Kubernetes master 节点

```shell
[root@centos ~]# ./edgeadm init --kubernetes-version=1.18.2 --image-repository superedge.tencentcloudcr.com/superedge --service-cidr=192.168.11.0/16 --pod-network-cidr=172.22.0.0/16 --apiserver-cert-extra-sans=<Master节点内网IP/域名> --apiserver-advertise-address=<Master节点公网IP/域名> --install-pkg-path <Kube静态安装包地址/FTP路径> -v=6
```
>   要是--image-repository superedge.tencentcloudcr.com/superedge 比较慢，可换成其他加速镜像仓库，只要能Pull下来kube-apiserver，kube-controller-manager，kube-scheduler，kube-proxy，etcd， pause等镜像就可以。

要是执行过程中没有问题，集群成功初始化，会输出如下内容：

```shell
Your Kubernetes control-plane has initialized successfully!

To start using your cluster, you need to run the following as a regular user:

  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config

You should now deploy a pod network to the cluster.
Run "kubectl apply -f [podnetwork].yaml" with one of the options listed at:
  https://kubernetes.io/docs/concepts/cluster-administration/addons/

Then you can join any number of worker nodes by running the following on each as root:

edgeadm join xxx.xxx.xxx.xxx:xxx --token xxxx \
    --discovery-token-ca-cert-hash sha256:xxxxxxxxxx 
```
执行过程中如果出现问题会直接返回相应的错误信息，并中断集群的初始化，使用`./edgeadm reset`命令回滚集群的初始化操作。

要使非 root 用户可以运行 kubectl，请运行以下命令，它们也是 edgeadm init 输出的一部分：
```shell
# mkdir -p $HOME/.kube
# sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
# sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

或者，如果你是 root 用户，则可以运行：
```shell
# export KUBECONFIG=/etc/kubernetes/admin.conf
```

记录`./edgeadm init`输出的`./edgeadm join`命令，你需要此命令将节点加入集群。

- join 加入边缘计算节点
```shell
[root@centos ~]# ./edgeadm join <域名/Master节点公网IP/Master节点内网IP>:6443 --token xxxx \
     --discovery-token-ca-cert-hash sha256:xxxxxxxxxx --install-pkg-path <Kube静态安装包地址/FTP路径>
```
> 执行join的时候可以把Kube-api的地址换成--apiserver-cert-extra-sans指定的master节点公网IP/自定义域名，默认打印的Kube-api的内网IP

要是执行过程中没有问题，新的 node 成功加入集群，会输出如下内容：

```shell
This node has joined the cluster:
* Certificate signing request was sent to apiserver and a response was received.
* The Kubelet was informed of the new secure connection details.

Run 'kubectl get nodes' on the control-plane to see this node join the cluster.
```
执行过程中如果出现问题会直接返回相应的错误信息，并中断节点的添加，使用`./edgeadm reset`命令回滚集群的初始化操作。

> 如果想用 edgeadm 安装原生 Kubernetes 集群，在 edgeadm init 命令中加上 --enable-edge=false 参数即可

### 3.2 用 edgeadm 安装边缘高可用 Kubernetes 集群
首先用户需要有 master VIP 作为集群 master 节点高可用的 VIP
- 安装 Haproxy

在负载均衡的机器上安装 Haproxy 作为集群总入口
> 注意：替换配置文件中的 < master VIP >
```shell
# yum install -y haproxy
# cat << EOF >/etc/haproxy/haproxy.cfg
global
    log         127.0.0.1 local2

    chroot      /var/lib/haproxy
    pidfile     /var/run/haproxy.pid
    maxconn     4000
    user        haproxy
    group       haproxy
    daemon
    stats socket /var/lib/haproxy/stats
defaults
    mode                    http
    log                     global
    option                  httplog
    option                  dontlognull
    option http-server-close
    option forwardfor       except 127.0.0.0/8
    option                  redispatch
    retries                 3
    timeout http-request    10s
    timeout queue           1m
    timeout connect         10s
    timeout client          1m
    timeout server          1m
    timeout http-keep-alive 10s
    timeout check           10s
    maxconn                 3000
frontend  main *:5000
    acl url_static       path_beg       -i /static /images /javascript /stylesheets
    acl url_static       path_end       -i .jpg .gif .png .css .js

    use_backend static          if url_static
    default_backend             app

frontend kubernetes-apiserver
    mode                 tcp
    bind                 *:16443
    option               tcplog
    default_backend      kubernetes-apiserver
backend kubernetes-apiserver
    mode        tcp
    balance     roundrobin
    server  master-0  <master VIP>:6443 check # 这里替换 master VIP 为用户自己的 VIP
backend static
    balance     roundrobin
    server      static 127.0.0.1:4331 check
backend app
    balance     roundrobin
    server  app1 127.0.0.1:5001 check
    server  app2 127.0.0.1:5002 check
    server  app3 127.0.0.1:5003 check
    server  app4 127.0.0.1:5004 check
EOF
```
- 安装 Keepalived

如果集群有两台 master，在两台 master 都安装 Keepalived，执行同样操作：
> 注意：
>
> 1.  替换配置文件中的 < master VIP >
>
> 2.  下面的 keepalived.conf 配置文件中 < master 本机公网 IP > 和 < 另一台 master 公网 IP > 在两台 master 的配置中是相反的，不要填错。
```shell
# yum install -y keepalived
# cat << EOF >/etc/keepalived/keepalived.conf 
! Configuration File for keepalived

global_defs {
   smtp_connect_timeout 30
   router_id LVS_DEVEL_EDGE_1
}
vrrp_script checkhaproxy{
script "/etc/keepalived/do_sth.sh"
interval 5
}
vrrp_instance VI_1 {
    state BACKUP
    interface eth0
    nopreempt
    virtual_router_id 51
    priority 100
    advert_int 1
    authentication {
        auth_type PASS
        auth_pass aaa
    }
    virtual_ipaddress {
        <master VIP> # 这里替换 master VIP 为用户自己的 VIP
    }
    unicast_src_ip <master 本机公网 IP>
    unicast_peer {
      <另一台 master 公网 IP>
    }
notify_master "/etc/keepalived/notify_action.sh MASTER"
notify_backup "/etc/keepalived/notify_action.sh BACKUP"
notify_fault "/etc/keepalived/notify_action.sh FAULT"
notify_stop "/etc/keepalived/notify_action.sh STOP"
garp_master_delay 1
garp_master_refresh 5
   track_interface {
     eth0
   }
   track_script {
     checkhaproxy 
   }
}
EOF
```
- 安装高可用边缘 Kubernetes master

在其中一台 master 中执行集群初始化操作
```shell
./edgeadm init --control-plane-endpoint <Master VIP> --upload-certs --kubernetes-version=1.18.2 --image-repository superedge.tencentcloudcr.com/superedge --service-cidr=192.168.11.0/16 --pod-network-cidr=172.22.0.0/16 --apiserver-cert-extra-sans=<域名/Master节点公网IP/Master节点内网IP> --install-pkg-path <Kube静态安装包地址/FTP路径> -v=6
```
要是执行过程中没有问题，集群成功初始化，会输出如下内容：
```shell
Your Kubernetes control-plane has initialized successfully!

To start using your cluster, you need to run the following as a regular user:

  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config

You should now deploy a pod network to the cluster.
Run "kubectl apply -f [podnetwork].yaml" with one of the options listed at:
  https://kubernetes.io/docs/concepts/cluster-administration/addons/

You can now join any number of the control-plane node running the following command on each as root:

  edgeadm join xxx.xxx.xxx.xxx:xxx --token xxxx \
    --discovery-token-ca-cert-hash sha256:xxxxxxxxxx \
    --control-plane --certificate-key xxxxxxxxxx

Please note that the certificate-key gives access to cluster sensitive data, keep it secret!
As a safeguard, uploaded-certs will be deleted in two hours; If necessary, you can use
"edgeadm init phase upload-certs --upload-certs" to reload certs afterward.

Then you can join any number of worker nodes by running the following on each as root:

edgeadm join xxx.xxx.xxx.xxx:xxxx --token xxxx \
    --discovery-token-ca-cert-hash sha256:xxxxxxxxxx  
```
执行过程中如果出现问题会直接返回相应的错误信息，并中断集群的初始化，使用`./edgeadm reset`命令回滚集群的初始化操作。

要使非 root 用户可以运行 kubectl，请运行以下命令，它们也是 edgeadm init 输出的一部分：
```shell
# mkdir -p $HOME/.kube
# sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
# sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

或者，如果你是 root 用户，则可以运行：
```shell
# export KUBECONFIG=/etc/kubernetes/admin.conf
```
记录`./edgeadm init`输出的`./edgeadm join`命令，你需要此命令将节点加入集群。

- join master 边缘节点

在另一台 master 执行`./edgeadm join`命令
```shell
# ./edgeadm join xxx.xxx.xxx.xxx:xxx --token xxxx    \
    --discovery-token-ca-cert-hash sha256:xxxxxxxxxx \
    --control-plane --certificate-key xxxxxxxxxx     \
    --install-pkg-path <Kube静态安装包地址/FTP路径> 
```
要是执行过程中没有问题，新的 master 成功加入集群，会输出如下内容：
```shell
This node has joined the cluster and a new control plane instance was created:

* Certificate signing request was sent to apiserver and approval was received.
* The Kubelet was informed of the new secure connection details.
* Control plane (master) label and taint were applied to the new node.
* The Kubernetes control plane instances scaled up.
* A new etcd member was added to the local/stacked etcd cluster.

To start administering your cluster from this node, you need to run the following as a regular user:

        mkdir -p $HOME/.kube
        sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
        sudo chown $(id -u):$(id -g) $HOME/.kube/config

Run 'kubectl get nodes' to see this node join the cluster.
```
执行过程中如果出现问题会直接返回相应的错误信息，并中断节点的添加，使用`./edgeadm reset`命令回滚集群的初始化操作。
- join node 边缘节点

```shell
# ./edgeadm join xxx.xxx.xxx.xxx:xxxx --token xxxx \
    --discovery-token-ca-cert-hash sha256:xxxxxxxxxx --install-pkg-path <Kube静态安装包地址/FTP路径>
```
要是执行过程中没有问题，新的 node 成功加入集群，会输出如下内容：
```shell
This node has joined the cluster:
* Certificate signing request was sent to apiserver and a response was received.
* The Kubelet was informed of the new secure connection details.

Run 'kubectl get nodes' on the control-plane to see this node join the cluster.
```
执行过程中如果出现问题会直接返回相应的错误信息，并中断节点的添加，使用`./edgeadm reset`命令回滚集群的初始化操作。

