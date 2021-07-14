# 部署SuperEdge - edgeadm方式

- [部署SuperEdge - edgeadm方式](#部署superedge---edgeadm方式)
  - [前提](#前提)
  - [一、edgeadm 简要](#一edgeadm-简要)
    - [1. edgeadm 是什么？](#1-edgeadm-是什么)
    - [2. edgeadm 有什么用？](#2-edgeadm-有什么用)
    - [3. 如何获取 edgeadm 工具](#3-如何获取-edgeadm-工具)
    - [4. 将kubernestes集群转化成边缘集群](#4-将kubernestes集群转化成边缘集群)
  - [二、edgeadm 命令介绍](#二edgeadm-命令介绍)
    - [1. change 命令](#1-change-命令)
    - [2. revert 命令](#2-revert-命令)
    - [3. mainfests 命令](#3-mainfests-命令)

## 前提

- Kubenertes集群的要求
 1. 用 kubeadm 安装好一个原生的kubernestes集群，最低一个 master 节点一个 node 节点，最低配置 2C2G（kubeadm的要求）。
 2. 集群版本推荐v1.18及以上(1.18.2我们做过详细测试)。
 3. 检查kube-api-server 和 kubelet 是否开启特权容器，没有开启的话 kube-api-server 和 kubelet 的启动参数添加 --allow-privileged=true, 并重启。
 4. 确保kubeadm集群健康，各个组件运行正常运行，Node处于Ready状态，能正常下发应用。

>建议使用 kubeadm 安装集群，避免部署失败，kubeadm安装Kubernetes 集群的方法可参考：[kubeadm官方安装指南](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/)

## 一、edgeadm 简要
### 1. edgeadm 是什么？

> edgeadm 是superedge团队推出的管理边缘集群的命令行工具。

### 2. edgeadm 有什么用？

- 能够无缝的将kubernestes集群转化成edge集群, 应用于边缘场景。
- 能够把edgeadm转化成的edge集群恢复至最初的原生kubernestes集群。
- 能够支持用户自定义边缘集群的属性和相关组件的调试开发。

### 3. 如何获取 edgeadm 工具

- 直接获取二进制 [点我获取](https://github.com/superedge/superedge/releases)
- 手动编译工程 [开始编译](../tutorial_CN.md)

### 4. 将kubernestes集群转化成边缘集群

1. 用 kubeadm 部署一个原生的kubernestes集群，要求见上面的 Kubeadm集群的要求。
2. 获取 edgeadm，按下文 change 命令进行。
```
sudo chmod +x edgeadm && ./edgeadm change
```

## 二、edgeadm 命令介绍

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