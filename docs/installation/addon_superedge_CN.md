简体中文 | [English](./addon_superedge.md)

# Addon SuperEdge 让原生Kuberntes集群具备边缘能力

* [Addon SuperEdge 让原生Kuberntes集群具备边缘能力](#addon-superedge-让原生kuberntes集群具备边缘能力)
   * [1.背景](#1背景)
   * [2. addon SuperEdge的边缘能力](#2-addon-superedge的边缘能力)
      * [&lt;1&gt;. 安装条件](#1-安装条件)
      * [&lt;2&gt;.下载 edgeadm 静态安装包](#2下载-edgeadm-静态安装包)
      * [&lt;3&gt;.addon SuperEdge](#3addon-superedge)
   * [3. Join边缘节点](#3-join边缘节点)
      * [&lt;1&gt;. 安装条件](#1-安装条件-1)
      * [&lt;2&gt;. 创建 Join边缘节点的token](#2-创建-join边缘节点的token)
      * [&lt;3&gt;. 边缘节点加入原有集群](#3-边缘节点加入原有集群)

## 1.背景

目前我们有很多用户已经有一个Kubernetes 集群，Kubernetes集群的搭建方式各种各样，有用公有云厂商的、私有化厂商整体解决方案的、还有用各种开源工具搭建的Kubernetes 集群。他们有一个诉求，**希望他们的Kubernetes集群既能管理云端应用，也能管理边缘应用**，简单整理了下有如下原始的需求：

-   能纳管任意位置的边缘节点，在一个Kubernetes 集群中既能管理云端应用，也能下发和管理边缘应用；
-   边缘节点需要具备边缘自治能力，以应对边缘弱网或者断网，不影响边缘节点上的应用正常提供服务；
-   能有批量添加边缘节点的能力和在原始Kubernetes集群同时管理中心机房和数十个边缘站点应用的能力；

为此我们为edgeadm支持了一个`addon edge-apps`的功能，即把SuperEdge的边缘能力直接以addon的方式添加进用户的Kubernetes集群，为实现上述诉求迈出了第一步。

## 2. addon SuperEdge的边缘能力

### <1>. 安装条件

-    用户已有 Kubernetes 集群，所有kube-controller-manager已开启`--controllers=*,bootstrapsigner,tokencleaner`参数；

     -   目前仅支持通过 kubeadm 搭建的Kubernetes 集群和Kind部署的Kubernetes集群；

     >   如果是用Kind搭建的Kubernetes 集群，部署前需执行以下命令：

     ```powershell
     mkdir -p /etc/kubernetes/pki
     docker ps |grep 'kindest/node'|awk '{print $1}'|while read line;do docker cp $line:/etc/kubernetes/pki /etc/kubernetes/;done
     ```

     -   确保集群中所有 kube-controller-manager 已经开启`BootstrapSignerController`，可参考：[使用 kubeadm 管理令牌](https://kubernetes.io/zh/docs/reference/access-authn-authz/bootstrap-tokens/#token-management-with-kubeadm) 开启;

     >   要是没有Kubernetes 集群，想尝试此功能，可用`edgeadm init`一键创建一个原生的Kubernetes 集群，详细可参考: [一键安装边缘Kubernetes集群](https://github.com/superedge/superedge/blob/main/docs/installation/install_edge_kubernetes_CN.md)。

-    支持的Kubernetes版本：v1.16~v1.19，提供的安装包是Kubernetes v1.18.2版本；

    >   理论上对用户原始Kubernetes版本无限制，但用户需要考虑边缘节点`kubelet`和原始Kubernetes版本的兼容性，统一最好；
    
    >   其他Kubernetes 版本可参考 [一键安装边缘Kubernetes集群](https://github.com/superedge/superedge/blob/main/docs/installation/install_edge_kubernetes_CN.md)中的 5. 自定义Kubernetes静态安装包，自行制作。

### <2>.下载 edgeadm 静态安装包

在任意一个Master 节点上下载 edgeadm 静态安装包，并拷贝到准备加入集群的边缘节点中。

>   注意修改"arch=amd64"参数，目前支持[amd64, arm64], 下载自己机器对应的体系结构，其他参数不变

```powershell
arch=amd64 version=v0.5.0 && rm -rf edgeadm-linux-* && wget https://superedge-1253687700.cos.ap-guangzhou.myqcloud.com/$version/$arch/edgeadm-linux-$arch-$version.tgz && tar -xzvf edgeadm-linux-* && cd edgeadm-linux-$arch-$version && ./edgeadm
```
>   目前支持amd64、arm64两个体系，其他体系可自行编译edgeadm和制作相应体系安装包，具体可参考 [一键安装边缘Kubernetes集群](https://github.com/superedge/superedge/blob/main/docs/installation/install_edge_kubernetes_CN.md)中的5. 自定义Kubernetes静态安装包

安装包大约200M，关于安装包的详细信息可查看 [一键安装边缘Kubernetes集群](https://github.com/superedge/superedge/blob/main/docs/installation/install_edge_kubernetes_CN.md) 中的**5. 自定义Kubernetes静态安装包**。

>   要是下载安装包比较慢，可直接查看相应[SuperEdge相应版本](https://github.com/superedge/superedge/tags), 下载`edgeadm-linux-amd64/arm64-*.0.tgz`，并解压也是一样的。
>
>   `edgeadm addon edge-apps`功能从SuperEdge-v0.4.0-beta.0开始支持，注意下载v0.4.0-beta.0及以后版本。

### <3>.addon SuperEdge

在原有集群任意一个Master节点上addon边缘能力组件：

```powershell
./edgeadm addon edge-apps --ca.cert <集群 CA 证书地址> --ca.key <集群的 CA 证书密钥路径> --master-public-addr <Master节点外网IP/Master节点内网IP/域名>:<Port>
```
其中：
-   --ca.cert: 集群的 CA 证书路径，默认 /etc/kubernetes/pki/ca.crt
-   --ca.key: 集群的 CA 证书密钥路径，默认 /etc/kubernetes/pki/ca.key
-    --master-public-addr：是边缘节点访问 kube-apiserver服务的地址，默认为 <Master节点内网IP>:<端口>

如果`edgeadm addon edge-apps`过程没有问题，终端会输出印如下日志：
``` 
I0606 12:52:50.277493   26593 deploy_edge_preflight.go:377] [upload-config] Uploading the kubelet component config to a ConfigMap
I0606 12:52:50.277515   26593 deploy_edge_preflight.go:393] [kubelet] Creating a ConfigMap "kubelet-config-1.19" in namespace kube-system with the configuration for the kubelets in the cluster
I0606 12:52:51.976165   26593 deploy_tunnel.go:35] Deploy tunnel-coredns.yaml success!
Create tunnel-cloud.yaml success!
I0606 12:52:52.704087   26593 deploy_tunnel.go:44] Deploy tunnel-cloud.yaml success!
I0606 12:52:53.146403   26593 deploy_tunnel.go:60] Deploy tunnel-edge.yaml success!
I0606 12:52:53.146432   26593 common.go:50] Deploy tunnel-edge.yaml success!
I0606 12:52:53.930520   26593 deploy_edge_health.go:54] Create edge-health-admission.yaml success!
I0606 12:52:53.950251   26593 deploy_edge_health.go:54] Create edge-health-webhook.yaml success!
I0606 12:52:53.950275   26593 common.go:57] Deploy edge-health success!
I0606 12:52:54.945275   26593 common.go:64] Deploy service-group success!
I0606 12:52:56.130217   26593 update_config.go:50] Update Kubernetes cluster config support marginal autonomy success
I0606 12:52:56.130243   26593 common.go:71] Update Kubernetes cluster config support marginal autonomy success
I0606 12:52:57.129018   26593 common.go:78] Config lite-apiserver configMap success
```

执行过程中如果出现问题会直接返回相应的错误信息，并中断边缘组件的安装，可使用`./edgeadm detach`命令卸载边缘组件恢复集群。
```powershell
./edgeadm detach edge-apps --ca.cert <集群 CA 证书地址> --ca.key <集群的 CA 证书密钥路径>
```
到此原有的Kubernetes集群就变成了一个既能管理云端应用，也能下发和管理边缘应用的Kubernetes。

## 3. Join边缘节点

### <1>. 安装条件

边缘节点遵循 [kubeadm的最低要求](https://kubernetes.io/zh/docs/setup/production-environment/tools/kubeadm/install-kubeadm/#before-you-begin) ，最低2C2G，磁盘空间不小于1G；

> ⚠️注意：尽可能提供干净的机器，避免其他因素引起安装错误。`要有机器上有容器服务在安装过程中可能会被清理，请在执行之前细心确认`。

### <2>. 创建 Join边缘节点的token

边缘能力组件addon成功后，Join边缘节点和 kubeadm的用法类似，需要在Master节点上使用`edgeadm token create --print-join-command`命令创建出`edgeadm join`的token：

```powershell
./edgeadm token create --print-join-command 
```

> 如果执行过程中没有问题，终端会输出 Join的命令

```powershell
...
edgeadm join <Master节点内网IP>:Port --token xxxx \
     --discovery-token-ca-cert-hash sha256:xxxxxxxxxx 
```

> 提示：创建 token 的有效期和 kubeadm 一样是`24h`，过期之后可以再次执行`./edgeadm token create`创建新的token。
>  --discovery-token-ca-cert-hash 的值生成也同 kubeadm，可在Master节点执行下面命令生成。

```powershell
openssl x509 -pubkey -in /etc/kubernetes/pki/ca.crt | openssl rsa -pubin -outform der 2>/dev/null | openssl dgst -sha256 -hex | sed 's/^.* //'
```

### <3>. 边缘节点加入原有集群

在边缘节点上下载 `2.<2>.中的edgeadm静态安装包`，或者通过其他方式把edgeadm静态安装包上传到边缘节点，然后在边缘节点上执行3.<2>得到的Join命令：

```powershell
./edgeadm join <Master节点外网IP/Master节点内网IP/域名>:Port --token xxxx \
     --discovery-token-ca-cert-hash sha256:xxxxxxxxxx 
     --install-pkg-path <edgeadm Kube-*静态安装包地址/FTP路径> --enable-edge=true
```
>   ⚠️注意：可以把`edgeadm create token --print-join-command`打印的 join 提示命令kube-apiserver的服务地址，视情况换成`Master节点外网IP/Master节点内网IP/域名`，主要取决于想让边缘节点通过外网还是内网访问 kube-apiserver 服务，默认输出的Master节点内网IP。

其中：

-   <Master节点外网IP/Master节点内网IP/域名>:Port 是边缘节点访问 kube-apiserver 服务的地址

-   --enable-edge=true:  加入的节点是否作为边缘节点（是否部署边缘能力组件），默认 true

>   --enable-edge=false 表示 join 原生 Kubernetes 集群节点，和 kubeadm join 的节点完全一样；

-   --install-pkg-path: Kubernetes 静态安装包的路径；

>   --install-pkg-path 的值可以为机器上的路径，也可以为网络地址（比如：http://xxx/xxx/kube-linux-arm64/amd64-*.tar.gz, 能免密wget到就可以），注意用和机器体系匹配的 Kubernetes 静态安装包；


要是执行过程中没有问题，新的 Node 成功加入集群，会输出如下内容：

```shell
This node has joined the cluster:
* Certificate signing request was sent to apiserver and a response was received.
* The Kubelet was informed of the new secure connection details.

Run 'kubectl get nodes' on the control-plane to see this node join the cluster.
```
执行过程中如果出现问题会直接返回相应的错误信息，并中断节点的添加，可使用`./edgeadm reset`命令回滚加入节点的操作，重新 join。

>    提示：边缘节点 join 成功后都会被打上标签: `superedge.io/edge-node=enable`，方便后续应用通过 nodeSelector 选择应用调度到边缘节点；

以上操作如有问题，可以加入到SuperEdge的[Slack](https://join.slack.com/t/superedge-workspace/shared_invite/zt-ldxnm7er-ptdpCXthOct_dYrzyXM3pw)、[Google论坛](https://groups.google.com/g/superedge)、[微信群](https://github.com/superedge/superedge)和我们进行交流，也可在社区提[Issues](https://github.com/superedge/superedge/issues)给我们反馈问题。

