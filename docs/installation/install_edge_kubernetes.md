# One-click install of edge Kubernetes cluster

* [One\-click installation of edge Kubernetes cluster](#one-click-installation-of-edge-kubernetes-cluster)
    * [1\. Background](#1-background)
    * [2\. Architecture Design](#2-architecture-design)
        * [2\.1 Original intention](#21-original-intention)
        * [2\.2 Goal](#22-goal)
        * [2\.3 Principle](#23-principle)
        * [2\.4 Design and implementation](#24-design-and-implementation)
    * [3\. How to use](#3-how-to-use)
        * [3\.1 Use edgeadm to install edge Kubernetes cluster](#31-use-edgeadm-to-install-edge-kubernetes-cluster)
        * [3\.2 Use edgeadm to install edge high\-availability Kubernetes cluster](#32-use-edgeadm-to-install-edge-high-availability-kubernetes-cluster)
    
## 1. Background

At present, none of the open source projects of edge computing container management such as SuperEdge could provide a way to deploy edge Kubernetes clusters independently. Like SuperEdge, it made a assumption that the user already has a Kubernetes cluster, and then the user can [manually](https://github.com/superedge/superedge/blob/main/docs/installation/install_manually_CN.md) Or borrow the seamless conversion tool provided by SuperEdge [edgeadm](https://github.com/superedge/superedge/blob/main/docs/installation/install_via_edgeadm_CN.md) to seamlessly convert a native Kubernetes cluster into edge Kubernetes Cluster.

Ideal is good, but when we study the [Manual Deployment Document](https://github.com/superedge/superedge/blob/main/docs/installation/install_manually_CN.md), it is found that several certificates need to be signed, many parameters of the original cluster need to be modified, and even the Kube API needs to be restarted... I don't know how many people who want to experience edge Kubernetes clusters are turned away. As for the tool [edgeadm](https://github.com/superedge/superedge/blob/main/docs/installation/install_via_edgeadm_CN.md) that seamlessly converts a native Kubernetes cluster into an edge Kubernetes cluster, the essence is just manual The configuration is performed automatically, which is too restrictive. it is found that the current scheme has the following disadvantages:

-   The Fatal flaw is that there is no way to add new edge nodes.
    -   Adding edge nodes requires the ability to add nodes to the native Kubernetes cluster itself, and then perform conversion, such as kubeadm. That is, first add an ordinary node to the Kubernetes cluster with Kubeadm, and then convert it into an edge node.
    -   However, there is a difference between the edge node and the normal node of the native Kubernetes cluster. Kubernetes requires that the Node node and the Kube-api-servce network be two-way intercommunication. However, in edge scenarios, there are generally more edge nodes and edge devices, and there are not so many public IPs. Only edge nodes can actively access Kube-api-servce, and kube-api-servce cannot directly access edge nodes.  The network is one-way. In this case, it is not feasible to add edge nodes before using Kubernetes.
-   Too restrictive, only for Kubernetes clusters installed in a specific way.
    -   At present, edgeadm only supports seamless conversion of Kubernetes clusters deployed in kubeadm mode. Kubernetes clusters deployed in other ways do not support seamless conversion. Users are required to perform conversion according to manual documents.
    -   edgeadm has very high standardization requirements for conversion clusters. It can only convert the standard Kubernetes cluster built by kubeadm. If the configuration of the cluster is modified or the original configuration file location of the cluster is moved, the conversion may fail.
-   The original intention of the design is different, the seamless conversion is only used to experience the edge ability, and does not have the production capacity.
    -   The original intention of the seamless conversion design is to allow users to experience the edge capability, and does not have the generation capability itself. Especially during the seamless conversion process, kube-api-server and kubelet need to be restarted to make the parameters set by the edge cluster take effect. If the cluster already has business, it may be interrupted.
    -   To truly integrate edge capabilities into production, users need to write the entire process of creating an edge Kubernetes cluster and joining edge nodes in their own internal systems according to manual documents. The implementation threshold is relatively high.

## 2. Architecture Design

In response to the above problems, in order to lower the threshold for user experience of the edge Kubernetes cluster, we and the students in the cloud native community wanted to design a solution that can deploy an edge Kubernetes cluster with one click,completely shielding the installation details, so that users can have zero threshold to experience the edge. For this we have formulated some principles.

### 2.1 Original intention

- Allow users to use edge Kubernetes clusters easily and without barriers;

### 2.2 Goal

- One-click

     - Able to install edge Kubernetes clusters with one click, and add edge nodes with one click;

- Can be installed offline

     - It can be installed online or completely offline, so that the privatization environment can also be very simple;

- Can be used in production

     - Don't encapsulate too much, so that companies that want to use edge Kubernetes clusters can perform simple integration in their internal systems, and they will be available;

- No learning costs

     - Try to be consistent with kubeadm's usage as much as possible, so that users can use it directly without learning costs;

### 2.3 Principle

- Cannot use SSH
  - For one thing, if SSH is used. Intermediate network transmission, remote execution of commands and file transmission will inevitably increase the failure rate;
  - Second, we want users to really use this solution in the production environment, not just experience, but let users do the integration;
- You cannot modify any line of Kubeadm's source code
  - The source code of Kubeadm can be reused and quoted, but the source code of Kubeadm cannot be modified to avoid the troubles of subsequent upgrades;
  - Based on kubeadm but higher than kubeadm, it does not have to be limited by the design of kubeadm, it can be allowed as long as it makes it easier for users;

### 2.4 Design and implementation

---

We carefully studied the source code of Kubeadm and found that we can borrow Kubeadm to create native Kubernetes clusters, Join nodes, and workflow ideas to achieve one-click deployment of edge Kubernetes clusters, and we can perform the installation steps step by step. This is exactly what we want for a simple, flexible, and low learning cost deployment solution. So we stood on the shoulders of giants, reused Kubeadm's source code, and devised the following scheme.

<img src="https://raw.githubusercontent.com/attlee-wang/myimage/master/image/20210419101917.png" alt="image-20210419101917603" style="zoom:50%;" />

>   Among them, the part of `Kubeadm init cluster/join node` completely reuses the source code of kubadm, and all the logic is exactly the same as Kubeadm.

This program has the following advantages:

-   Fully compatible with Kubeadm

    We just stand on the shoulders of Kubeadm, set some configuration parameters required by the edge cluster before Kubeadm init/join, initialize the Master or Node nodes automatically, and install the container runtime. After the completion of Kubeadm init/join, the CNI network plug-in was installed and the corresponding edge capability components were deployed.

    We quoted the Kubeadm source code in Go Mod mode. During the whole process, we did not modify the Kubeadm source code one line. It is completely native and ready to upgrade to a higher version of Kubeadm in the future.

-   Simple, flexible and automated to use

    The edgeadm init cluster and Join node completely retain the original parameters and process of Kubeadm init/join, but automatically initialize the node and install the container when running, you can use the edgeadm --enable-edge=fasle parameter to install the native one-click For Kubernetes clusters, you can also use the edgeadm --enable-edge=true parameter to install an edge Kubernetes cluster with one click.
    You can join any node as long as you can access the node where Kube-api-server is located, or you can join mastter. Join master also continues the Kubeadm approach. When necessary, it can directly expand the master node with the join master to realize high availability.

-   No learning cost, exactly the same as using kubeadm

    Because the Kubeadm init cluster/join node part completely reuses the source code of kubadm, all logic is exactly the same as Kubeadm, completely retains the usage habits of kubeadm and all flag parameters, and the usage is exactly the same as that of kubeadm, without any new learning costs.Users can customize the edge Kubernetes cluster according to the parameters of Kubeadm.

## 3. How to use

### 3.1 Use edgeadm to install edge Kubernetes cluster

- Download edgeadm binary package
```shell
# rm -rf edgeadm && \
wget https://attlee-1251707795.cos.ap-chengdu.myqcloud.com/superedge/v0.3.0/edgeadm &&  \
chmod +x edgeadm
```
- Install edge Kubernetes master node

```shell
[root@centos ~]# ./edgeadm init --kubernetes-version=1.18.2 --image-repository superedge.tencentcloudcr.com/superedge --service-cidr=192.168.11.0/16 --pod-network-cidr=172.22.0.0/16 --apiserver-cert-extra-sans=<Domain/Intranet IP of Master node> --apiserver-advertise-address=<Domain/Public IP of Master node> -v=6
```
> If --image-repository superedge.tencentcloudcr.com/superedge is slower, you can switch to other accelerated mirror warehouses, as long as you can pull down kube-apiserver, kube-controller-manager, kube-scheduler, kube-proxy, etcd, pause and other mirror images. 


If there are no exceptions during execution, and the cluster is successfully initialized, the following content will be output:
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
If there are exceptions during execution, the corresponding error message will be directly returned, and the initialization of the cluster will be interrupted. Use the `./edgeadm reset` command to roll back the initialization operation of the cluster.

To enable non-root users to run kubectl, run the following commands, which are also part of the edgeadm init output:
```shell
# mkdir -p $HOME/.kube
# sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
# sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

If you are the root user, you can run:
```shell
# export KUBECONFIG=/etc/kubernetes/admin.conf
```
Record the `./edgeadm join` command output by `./edgeadm init`. You need this command to join the node to the cluster.

- join to join edge computing nodes
```bash
# ./edgeadm join <Domain/Public IP of Master node>:6443 --token xxxx \
    --discovery-token-ca-cert-hash sha256:xxxxxxxxxx 
```
If there are no exceptions in the execution process,  the new node successfully joins the cluster, and the following will be output:

```shell
This node has joined the cluster:
* Certificate signing request was sent to apiserver and a response was received.
* The Kubelet was informed of the new secure connection details.

Run 'kubectl get nodes' on the control-plane to see this node join the cluster.
```
If there are exceptions in the execution process, the corresponding error message will be directly returned, and the addition of the node will be interrupted. Use the `./edgeadm reset` command to roll back the initialization operation of the cluster.

> If you want to install a native Kubernetes cluster with edgeadm, add the --enable-edge=false parameter to the edgeadm init command

### 3.2 Use edgeadm to install edge high-availability Kubernetes cluster
First, the user needs to have the master VIP as the highly available VIP of the cluster master node
- Install Haproxy

Install Haproxy on the load balancing machine as the main entrance of the cluster:
> Note:
> Replace <master VIP> in the configuration file
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
    server  master-0  <master VIP>:6443 check # Here replace the master VIP with the user's own VIP
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
- Install Keepalived

If the cluster has two masters, install Keepalived on both masters and perform the same operation:
> Note:
>
> 1. Replace <master VIP> in the configuration file
>
> 2. In the keepalived.conf configuration file below, <master's local public network IP> and <another master's public network IP> are opposite in the configuration of the two masters. Don't fill in the error.
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
        <master VIP> # Here replace the master VIP with the user's own VIP
    }
    unicast_src_ip <master Public IP>
    unicast_peer {
      <Public IP of other master nodes>
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
- Install high-availability edge Kubernetes master

Perform cluster initialization operations in one of the masters
```shell
# ./edgeadm init --control-plane-endpoint <master VIP> --upload-certs --kubernetes-version=1.18.2 --image-repository superedge.tencentcloudcr.com/superedge --service-cidr=192.168.11.0/16 --pod-network-cidr=172.22.0.0/16 --apiserver-cert-extra-sans=<Domain/Public IP of Master node> -v=6
```
If there are no exceptions during execution and the cluster is successfully initialized, the following content will be output:
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
If there are exceptions during execution, the corresponding error message will be directly returned, and the initialization of the cluster will be interrupted. Use the `./edgeadm reset` command to roll back the initialization operation of the cluster.

To enable non-root users to run kubectl, run the following commands, which are also part of the edgeadm init output:
```shell
# mkdir -p $HOME/.kube
# sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
# sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

If you are the root user, you can run:
```shell
# export KUBECONFIG=/etc/kubernetes/admin.conf
```
Record the `./edgeadm join` command output by `./edgeadm init`. You need this command to join the node to the cluster.

- join edge master node

Execute the `./edgeadm join` command on another master
```shell
# ./edgeadm join xxx.xxx.xxx.xxx:xxx --token xxxx \
    --discovery-token-ca-cert-hash sha256:xxxxxxxxxx \
    --control-plane --certificate-key xxxxxxxxxx
```
If there are no exceptions in the execution process, the new master successfully joins the cluster, and the following content will be output:
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
If there are exceptions during execution, the corresponding error message will be directly returned, and the initialization of the cluster will be interrupted. Use the `./edgeadm reset` command to roll back the initialization operation of the cluster.

- join edge node

```shell
# ./edgeadm join xxx.xxx.xxx.xxx:xxxx --token xxxx \
    --discovery-token-ca-cert-hash sha256:xxxxxxxxxx
```
If there are no exceptions in the execution process, the new master successfully joins the cluster, and the following content will be output:

```shell
This node has joined the cluster:
* Certificate signing request was sent to apiserver and a response was received.
* The Kubelet was informed of the new secure connection details.

Run 'kubectl get nodes' on the control-plane to see this node join the cluster.
```
If there are exceptions during execution, the corresponding error message will be directly returned, and the initialization of the cluster will be interrupted. Use the `./edgeadm reset` command to roll back the initialization operation of the cluster.

