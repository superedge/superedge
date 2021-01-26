# 部署SuperEdge - edgeadm方式

* [部署SuperEdge \- edgeadm方式](#部署superedge---edgeadm方式)
    * [前提](#前提)
    * [一、edgeadm 简要](#一edgeadm-简要)
        * [1\. edgeadm 是什么？](#1-edgeadm-是什么)
        * [2\. edgeadm 有什么用？](#2-edgeadm-有什么用)
        * [3\. 如何获取 edgeadm 工具](#3-如何获取-edgeadm-工具)
    * [二、edgeadm 搭建集群](#二edgeadm-搭建集群)
        * [1\. 将 kubernestes 集群转化成边缘集群](#1-将-kubernestes-集群转化成边缘集群)
        * [2\. 使用 edgeadm 创建集群](#2-使用-edgeadm-创建集群)
            * [2\.1 创建普通 Kubernetes 集群](#21-创建普通-kubernetes-集群)
                * [2\.1\.1 初始化普通 Kubernetes 集群](#211-初始化普通-kubernetes-集群)
                * [2\.1\.2 将节点作为普通 node 加入 Kubernetes 集群](#212-将节点作为普通-node-加入-kubernetes-集群)
            * [2\.2 创建边缘 Kubernetes 集群](#22-创建边缘-kubernetes-集群)
                * [2\.2\.2 初始化边缘 Kubernetes 集群](#222-初始化边缘-kubernetes-集群)
                * [2\.2\.2 将节点作为边缘 node 加入 Kubernetes 集群](#222-将节点作为边缘-node-加入-kubernetes-集群)
    * [三、edgeadm 命令介绍](#三edgeadm-命令介绍)
        * [1\. change 命令](#1-change-命令)
        * [2\. revert 命令](#2-revert-命令)
        * [3\. mainfests 命令](#3-mainfests-命令)
        * [4\. init 命令](#4-init-命令)
            * [4\.1 单 master 集群初始化](#41-单-master-集群初始化)
            * [4\.2 高可用 master 集群初始化](#42-高可用-master-集群初始化)
        * [5\. join 命令](#5-join-命令)
            * [5\.1 worker 节点加入集群](#51-worker-节点加入集群)
            * [5\.2 master 节点加入集群](#52-master-节点加入集群)
        * [6\. reset 命令](#6-reset-命令)


## 前提

- 如果使用 edgeadm 将已有的普通 Kubernetes 集群转化为边缘 Kubernetes 集群，则对已有的 Kubernetes 集群有如下要求
    - 用 kubeadm 安装好一个原生的 kubernetes 集群，最低一个 master 节点一个 node 节点，最低配置 2C2G（kubeadm的要求）。
    - 集群版本推荐v1.18及以上(1.18.2我们做过详细测试)。
    - 检查kube-api-server 和 kubelet 是否开启特权容器，没有开启的话 kube-api-server 和 kubelet 的启动参数添加 --allow-privileged=true, 并重启。
    - 确保 Kubernetes集群健康，各个组件运行正常运行，Node处于Ready状态，能正常下发应用。

>建议使用 kubeadm 安装集群，避免部署失败，kubeadm安装Kubernetes 集群的方法可参考：[kubeadm 官方安装指南](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/)

- 如果使用 edgeadm 从零开始创建边缘 Kubernetes 集群，则对组建集群的机器有如下要求
    - 最低一个 master 节点一个 node 节点，最低配置 2C2G
    - `edgeadm init`命令的`--kubernetes-version`参数指定集群版本推荐v1.18及以上(1.18.2我们做过详细测试)。
    - 确保集群网络正常可用，边缘节点和云端节点可正常通信。
## 一、edgeadm 简要
### 1. edgeadm 是什么？

> edgeadm 是superedge团队推出的管理边缘集群的命令行工具。

### 2. edgeadm 有什么用？

-   能够一键创建出一个独立的边缘Kubernestes集群，随时随地的添加边缘节点；
-   能够一键创建出一个原生的Kubernestes集群，将原生的Kubernestes集群change成边缘Kubernestes集群；

- 能够无缝的将kubernestes集群转化成edge集群, 应用于边缘场景。
- 能够把edgeadm转化成的edge集群恢复至最初的原生kubernestes集群。
- 能够支持用户自定义边缘集群的属性和相关组件的调试开发。

### 3. 如何获取 edgeadm 工具

- 直接获取二进制 [点我获取](https://github.com/superedge/superedge/releases)
- 手动编译工程 [开始编译](./tutorial_CN.md)

## 二、edgeadm 搭建集群

edgeadm 底层基于 kubeadm，因此不仅可以使用 edgeadm 搭建边缘集群，同时也支持搭建普通 Kubernetes 集群。同时，也支持将已有的 Kubernetes 集群转化为边缘集群。
### 1. 将 kubernestes 集群转化成边缘集群

1. 用 kubeadm 部署一个原生的 kubernestes 集群，要求见上面的 kubeadm 集群的要求。
2. 获取 edgeadm，按下文[ change 命令 ](#1-change-%E5%91%BD%E4%BB%A4)进行转化。
```
sudo chmod +x edgeadm && ./edgeadm change
```

### 2. 使用 edgeadm 创建集群

使用 edgeadm，能够创建一个符合最佳实践的最小化 Kubernetes 集群（底层完全基于 kubeadm 实现），同时也能创建一个符合最佳实践的最小化 Kubernetes 边缘集群。

#### 2.1 创建普通 Kubernetes 集群

##### 2.1.1 初始化普通 Kubernetes 集群

使用 edgeadm 初始化普通 Kubernetes 集群的命令和 kubeadm 的初始化方式非常相似，只需要在执行初始化 `init` 命令时加上 `--kubeadm-auto` 参数指定使用 kubeadm 的方式安装普通 Kubernetes 集群即可。与 kubeadm 不同的是，edgeadm 不需要用户提前做安装容器运行时、设置系统内核参数等一系列节点的初始化操作，初始化完成后默认安装 CNI 插件，只需要一条命令即可完成集群的初始化操作，做到了开箱即用。

下面是初始化一个普通 Kubernetes 集群 master 节点的例子：
```
[root@master01 ~]# edgeadm init --apiserver-advertise-address [master 节点 IP] --kubeadm-auto
```
由于 edgeadm 底层基于 kubeadm 的方式搭建 Kubernetes 集群，edgeadm 的 `init` 命令完全兼容 kubeadm 的 `init` 命令，因此 `init` 命令的其它详细参数可参考[ kubeadm 官方文档 ](https://kubernetes.io/zh/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/)

##### 2.1.2 将节点作为普通 node 加入 Kubernetes 集群
和上文中初始化普通 Kubernetes 集群的方式相同，只需要在执行加入集群 `join` 命令时加上 `--kubeadm-auto` 参数指定使用 kubeadm 的方式将节点作为普通 Kubernetes node 加入集群即可。join命令的详细用法参考：[ join 命令](#5-join-%E5%91%BD%E4%BB%A4) 。

#### 2.2 创建边缘 Kubernetes 集群

##### 2.2.2 初始化边缘 Kubernetes 集群
使用 edgeadm 初始化边缘 Kubernetes 集群的命令和 kubeadm 的初始化方式完全一致，不加上 `--kubeadm-auto` 参数则默认初始化一个边缘 Kubernetes 集群的 master 节点。与 kubeadm 不同的是，edgeadm 不需要用户提前做安装容器运行时、设置系统内核参数等一系列节点的初始化操作，初始化完成后默认安装 CNI 插件，只需要一条命令即可完成集群的初始化操作，做到了开箱即用。

下面是初始化一个边缘 Kubernetes 集群 master 节点的例子：
```
[root@master01 ~]# edgeadm init --apiserver-advertise-address [master 节点 IP]
```
由于 edgeadm 底层基于 kubeadm 的方式搭建 Kubernetes 集群，edgeadm 的 `init` 命令完全兼容 kubeadm 的 `init` 命令，因此 `init` 命令的其它详细参数可参考[ kubeadm 官方文档 ](https://kubernetes.io/zh/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/) 。

`init` 命令不仅支持初始化单 master 的边缘 Kubernetes 集群，还支持初始化高可用 master 的边缘 Kuberentes 集群，`init` 命令的更多用法参考：[ init 命令](#4-init-%E5%91%BD%E4%BB%A4) 。

##### 2.2.2 将节点作为边缘 node 加入 Kubernetes 集群
和上文中初始化边缘 Kubernetes 集群的方式相同，在执行加入集群 `join` 命令时不加上 `--kubeadm-auto` 参数则默认将节点作为边缘 Kubernetes node 加入集群。join命令的详细用法参考：[ join 命令](#5-join-%E5%91%BD%E4%BB%A4) 。


## 三、edgeadm 命令介绍

### 1. change 命令

<div align="left">
  <img src="../img/edgeadm-change.png" width=70% title="edgeadm change output">
</div>

- 含义
> 将kubernestes原生集群转化成edge集群。

- 最简执行
```
 [root@master01 ~]# edgeadm change
```
edgeadm默认读取${home}/.kube/config的kubeconfig文件和/etc/kubernetes/pki/ca.* 证书
要是 kubeconfig 和 ca.* 不在默认路径请按全参执行。

- 全参数执行
```
[root@master01 ~]# edgeadm change -p [集群的部署方式] --kubeconfig  [kubeconfig文件路径] --ca.cert [集群根证书路径] --ca.key [集群根证书key路径]
```

要是执行过程中没有问题，会输出如下内容：
```
[root@master01 ~]# edgeadm change 
Create tunnel-coredns.yaml success!
...
Deploy helper-job-master* success!
Kubeadm Cluster Change To Edge cluster Success!
```
要有问题会直接返回相应的错误，并中断集群的转换。

-   注意点：

    <1>. 转化的镜像默认是从docker hub superedge 仓库拉取的，目前支持amd64和arm64体系，其他体系可自行编译，按mainfests命令方式替换执行。

    <2>. 默认读取kubeconfig的顺序是：--kubeconfig > Env KUBECONFIG > ~/.kube/config

---
### 2. revert 命令

- 含义
>  将edge集群恢复成最初的kubernestes原生集群。

- 最简执行

```
[root@master01 ~]# edgeadm revert 
```

edgeadm默认读取${home}/.kube/config的kubeconfig文件和/etc/kubernetes/pki/ca.* 证书
要是kubeconfig 和 ca.* 不在默认路径请按全参执行。

- 全参数执行
```
[root@master01 ~]# edgeadm revert -p kubeadm --kubeconfig  [kubeconfig文件路径] --ca.cert [集群根证书路径] --ca.key [集群根证书key路径]
```

要是执行过程中没有问题，会输出如下内容：
```
[root@master01 ~]# edgeadm revert 
Deploy helper-job-node* success!
...
Deploy helper-job-master* success!
Kubeadm Cluster Revert To Edge Cluster Success!
```
要有问题会直接返回相应的错误，并中断集群的恢复。

---
### 3. mainfests 命令

- 含义
> 输出edge集群所有的yaml文件到特定文件下

- 最简执行
```
[root@master01 ~]# edgeadm manifests 
```
默认将edge集群所需要的yaml文件全部输出到./manifests/ 文件夹下

- 全参执行
```
[root@master01 ~]# edgeadm manifests -m  /目标文件夹
```
- 输出这些yaml文件有什么用？
> 可以根据实际情况修改yaml内容，然后用change命令部署

比如：修改了edge-health的代码，先将编译出的镜像推送到私有镜像仓库，然后部署自己编译出来的镜像 edge-health:0.1.0
1. 修改 ./manifests/edge-health.yaml， 将镜像换成修改后的
<div align="left">
  <img src="../img/edit-edge-health.png" width=70% title="edit dege health">
</div>

2. 然后用change 命令重新转化edge cluster
```
[root@master01 ~]# edgeadm change -m ./manifests/
```
3. 查看edge-health的pod, 镜像已经变成自定义的：
<div align="left">
  <img src="../img/view-edge-health.png" width=70% title="view edge health">
</div>


> **注意：**
manifests/下生成的yaml模板，所有参数都可以更改和自定义，更改请遵守kubernetes规范。对于带 {{.*}} 的参数可直接赋值，没有赋值的 {{.*}} edgeadm 工具在chenge时会自动填充。

---
### 4. init 命令

- 含义
> 初始化 edge 集群的一个 master 节点，用法和 kubeadm 保持一致

#### 4.1 单 master 集群初始化
- 最简执行
```
[root@master01 ~]# edgeadm init --apiserver-advertise-address [master 节点 IP]
```
默认初始化一个 master 节点的 edge 集群

- 全参执行
```
[root@master01 ~]# edgeadm init --kubeconfig [指定使用该路径下的 kubeconfig 和集群通信] --kubernetes-version [kubernetes 版本] --image-repository [镜像仓库前缀] --service-cidr [service 网段] --pod-network-cidr [pod 网段] --apiserver-advertise-address [master 节点 IP]
```
要是执行过程中没有问题，集群成功初始化，会输出如下内容：
```
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
执行过程中如果出现问题会直接返回相应的错误信息，并中断集群的初始化，使用[ edgeadm reset ](#6-reset-%E5%91%BD%E4%BB%A4)命令回滚集群的初始化操作。
根据上面输出的`edgeadm join`命令，进行后续的 worker 节点添加操作。

#### 4.2 高可用 master 集群初始化
初始化高可用 master 的集群需要先为 kube-apiserver 创建负载均衡器，如：使用 Haproxy + Keepalived 的方案为 kube-apiserver 提供一个 VIP，使用该 VIP 作为`init`命令的`--control-plane-endpoint`参数值传入，即可初始化一个高可用 master 的集群。创建 kube-apiserver 负载均衡器的方式可参考 [ 利用 kubeadm 创建高可用集群 ](https://kubernetes.io/zh/docs/setup/production-environment/tools/kubeadm/high-availability/) 。
- 最简执行
```
[root@master01 ~]# edgeadm init --control-plane-endpoint [master VIP] --upload-certs
```
初始化一个高可用 edge 集群

- 全参执行
```
[root@master01 ~]# edgeadm init --kubeconfig [指定使用该路径下的 kubeconfig 和集群通信] --kubernetes-version [kubernetes 版本] --image-repository [镜像仓库前缀] --service-cidr [service 网段] --pod-network-cidr [pod 网段] --control-plane-endpoint [master VIP] --upload-certs
```
要是执行过程中没有问题，集群成功初始化，会输出如下内容：
```
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
执行过程中如果出现问题会直接返回相应的错误信息，并中断集群的初始化，使用[ edgeadm reset ](#6-reset-%E5%91%BD%E4%BB%A4)命令回滚集群的初始化操作。

要使非 root 用户可以运行 kubectl，请运行以下命令，它们也是 edgeadm init 输出的一部分：
```
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

或者，如果你是 root 用户，则可以运行：
```
export KUBECONFIG=/etc/kubernetes/admin.conf
```

记录 edgeadm init 输出的 [ edgeadm join ](#5-join-%E5%91%BD%E4%BB%A4) 命令。 你需要此命令将节点加入集群。

---
### 5. join 命令

- 含义
> 将一个节点加入到 edge 集群中，用法和 kubeadm 保持一致

#### 5.1 worker 节点加入集群
- 在 worker 节点上执行[ 单 master 集群初始化 ](#4.1-%E5%8D%95-master-%E9%9B%86%E7%BE%A4%E5%88%9D%E5%A7%8B%E5%8C%96)或[ 高可用 master 集群初始化 ](#4.2-%E9%AB%98%E5%8F%AF%E7%94%A8-master-%E9%9B%86%E7%BE%A4%E5%88%9D%E5%A7%8B%E5%8C%96)命令初始化成功后输出的[ edgeadm join ](#5-join-%E5%91%BD%E4%BB%A4)命令，将该 worker 节点加入集群
```
[root@worker01 ~]# edgeadm join xxx.xxx.xxx.xxx:16443 --token xxxx \
    --discovery-token-ca-cert-hash sha256:xxxxxxxxxx
```

#### 5.2 master 节点加入集群
- 在新的 master 节点上执行[ 高可用 master 集群初始化 ](#4.2-%E9%AB%98%E5%8F%AF%E7%94%A8-master-%E9%9B%86%E7%BE%A4%E5%88%9D%E5%A7%8B%E5%8C%96)命令初始化成功后输出的[ edgeadm join ](#5-join-%E5%91%BD%E4%BB%A4)命令，将该 master 节点加入集群
```
[root@master02 ~]# edgeadm join xxx.xxx.xxx.xxx:16443 --token xxxx \
    --discovery-token-ca-cert-hash sha256:xxxxxxxxxx \
    --control-plane --certificate-key xxxxxxxxxx
```
如果没有 token，可以通过在 master 节点上运行以下命令来获取：
```
[root@master01 ~]# edgeadm token list
```
如果没有 --discovery-token-ca-cert-hash 的值，则可以通过在 master 节点上执行以下命令链来获取：
```
[root@master01 ~]# openssl x509 -pubkey -in /etc/kubernetes/pki/ca.crt | openssl rsa -pubin -outform der 2>/dev/null | \
[root@master01 ~]# openssl dgst -sha256 -hex | sed 's/^.* //'
```
---

### 6. reset 命令

- 含义
> 重置集群，将节点重置成集群搭建前的状态，用法和 kubeadm 保持一致

- 最简执行

```
[root@master01 ~]# edgeadm reset 
```