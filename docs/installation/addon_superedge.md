English|[简体中文](./addon_superedge_CN.md) 

# Addon SuperEdge enables native Kuberntes clusters with edge capabilities

* [Addon SuperEdge enables native Kuberntes clusters with edge capabilities](#addon-superedge-enables-native-kuberntes-clusters-with-edge-capabilities)
   * [1.background](#1background)
   * [2. addon SuperEdge of Edge capabilities](#2-addon-superedge-of-edge-capabilities)
      * [&lt;1&gt;. Installation conditions](#1-installation-conditions)
      * [&lt;2&gt;.Download edgeadm static installation package](#2download-edgeadm-static-installation-package)
      * [&lt;3&gt;.addon SuperEdge](#3addon-superedge)
   * [3. Join edge node](#3-join-edge-node)
      * [&lt;1&gt;. Join conditions](#1-join-conditions)
      * [&lt;2&gt;. Create the token of the Join edge node](#2-create-the-token-of-the-join-edge-node)
      * [&lt;3&gt;. Join edge node into riginal Kubernetes cluster](#3-join-edge-node-into-riginal-kubernetes-cluster)

## 1.background

At present, many of our users already have a Kubernetes cluster. There are various ways to build Kubernetes clusters. There are overall solutions from public cloud vendors, privatized vendors, and Kubernetes clusters built with various open source tools. They have a request, **I hope their Kubernetes cluster can manage both cloud applications and edge applications**. Simply put together the following original requirements:

-   Able to accommodate edge nodes at any location, and manage cloud applications in a Kubernetes cluster, as well as distribute and manage edge applications;
-   Edge nodes need to have edge autonomy capabilities to cope with weak or disconnected edge networks without affecting the normal provision of services by applications on edge nodes;
-   Ability to add edge nodes in batches and the ability to simultaneously manage the central computer room and dozens of edge site applications in the original Kubernetes cluster;

For this reason, we support an `addon edge-apps` function for edgeadm, that is, adding the edge capabilities of SuperEdge directly into the user's Kubernetes cluster in the form of addon, which is the first step to achieve the above requirements.

## 2. addon SuperEdge of Edge capabilities

### <1>. Installation conditions

-    The user has a Kubernetes cluster, and all kube-controller-managers have enabled the `--controllers=*,bootstrapsigner,tokencleaner` parameters;

     -   Currently, Kubernetes clusters built only by kubeadm and Kubernetes clusters deployed by Kind are supported;

     >   If you are using Kind to build a Kubernetes cluster, you need to execute the following commands before deployment:

     ```powershell
     mkdir -p /etc/kubernetes/pki
     docker ps |grep 'kindest/node'|awk '{print $1}'|while read line;do docker cp $line:/etc/kubernetes/pki /etc/kubernetes/;done
     ```

     -   Make sure that all kube-controller-managers in the cluster have enabled `BootstrapSignerController`, refer to: [Use kubeadm to manage tokens](https://kubernetes.io/zh/docs/reference/access-authn-authz/bootstrap-tokens/ #token-management-with-kubeadm) open;

     >   If you don’t have a Kubernetes cluster and want to try this feature, you can use `edgeadm init` to create a native Kubernetes cluster with one click. For details, please refer to: [One-click installation of edge Kubernetes cluster](https://github.com/superedge/superedge/blob /main/docs/installation/install_edge_kubernetes_CN.md).

-    Supported Kubernetes version: v1.16~v1.19, the provided installation package is Kubernetes v1.18.2 version;

    >   Theoretically, there is no limit to the user's original Kubernetes version, but users need to consider the compatibility of the edge node `kubelet` and the original Kubernetes version, and the consistency is the best;
    
    >   For other Kubernetes versions, please refer to [One-click installation of edge Kubernetes cluster](https://github.com/superedge/superedge/blob/main/docs/installation/install_edge_kubernetes_CN.md) in 5. Customize the Kubernetes static installation package, by yourself Make.

### <2>.Download edgeadm static installation package

Download the edgeadm static installation package on any Master node, and copy it to the edge node that is ready to join the cluster.

>   Pay attention to modify the "arch=amd64" parameter, currently supports [amd64, arm64], download the corresponding architecture of your own machine, and other parameters remain unchanged

```powershell
arch=amd64 version=v0.5.0 && rm -rf edgeadm-linux-* && wget https://superedge-1253687700.cos.ap-guangzhou.myqcloud.com/$version/$arch/edgeadm-linux-$arch-$version.tgz && tar -xzvf edgeadm-linux-* && cd edgeadm-linux-$arch-$version && ./edgeadm
```
>   Currently supports amd64 and arm64 systems. For other systems, you can compile edgeadm and make corresponding system installation packages. For details, please refer to [One-click installation of edge Kubernetes cluster](https://github.com/superedge/superedge/blob/main/ docs/installation/install_edge_kubernetes_CN.md) 5. Customize the Kubernetes static installation package

The installation package is about 200M. For detailed information about the installation package, please refer to the **in [One-click installation of edge Kubernetes cluster](https://github.com/superedge/superedge/blob/main/docs/installation/install_edge_kubernetes_CN.md) 5. Customize the Kubernetes static installation package**.

>   If downloading the installation package is slow, you can directly check the corresponding [SuperEdge version](https://github.com/superedge/superedge/tags), download `edgeadm-linux-amd64/arm64-*.0.tgz`, and Decompression is the same.

>   The `edgeadm addon edge-apps` function is supported starting from SuperEdge-v0.4.0-beta.0, pay attention to download v0.4.0-beta.0 and later versions.

### <3>.addon SuperEdge

Addon edge capability components on any Master node of the original cluster:

```powershell
./edgeadm addon edge-apps --ca.cert <Cluster CA certificate path> --ca.key <Cluster CA certificate key path> --master-public-addr <Master public/Intranet IP or domain>:<Port>
```
among them:：
-   --ca.cert: CA certificate path of the cluster, default /etc/kubernetes/pki/ca.crt
-   --ca.key: The CA certificate key path of the cluster, default /etc/kubernetes/pki/ca.key
-    --master-public-addr：It is the address of the edge node to access the kube-apiserver service, the default is <Master node intranet IP>:<port>

If there is no problem with the `edgeadm addon edge-apps` process, the terminal will output the following log:

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

If there is a problem during the execution, the corresponding error message will be returned directly and the installation of the edge component will be interrupted. You can use the `./edgeadm detach` command to uninstall the edge component to restore the cluster.

```powershell
./edgeadm detach edge-apps --ca.cert <Cluster CA certificate path> --ca.key <Cluster CA certificate key path>
```
The original Kubernetes cluster here has become a Kubernetes that can manage cloud applications as well as distribute and manage edge applications.

## 3. Join edge node

### <1>. Join conditions

Edge nodes follow [kubeadm's minimum requirements](https://kubernetes.io/zh/docs/setup/production-environment/tools/kubeadm/install-kubeadm/#before-you-begin), minimum 2C2G, no disk space Less than 1G;

> ⚠️Note: Provide clean machines as much as possible to avoid installation errors caused by other factors. `If there is a container service on the machine, it may be cleaned up during the installation process, please confirm carefully before executing`.

### <2>. Create the token of the Join edge node

After the edge capability component addon is successful, the usage of Join edge node is similar to kubeadm. You need to use the `edgeadm token create --print-join-command` command on the Master node to create the `edgeadm join` token:

```powershell
./edgeadm token create --print-join-command 
```

> If there is no problem during execution, the terminal will output the Join command

```powershell
...
edgeadm join <Master Intranet IP>:Port --token xxxx \
     --discovery-token-ca-cert-hash sha256:xxxxxxxxxx 
```

> Tip: The validity period of the created token is `24h`, the same as kubeadm. After the expiration, you can execute `./edgeadm token create` again to create a new token.
>
> The value generation of --discovery-token-ca-cert-hash is also the same as kubeadm, which can be generated by executing the following command on the Master node.

```powershell
openssl x509 -pubkey -in /etc/kubernetes/pki/ca.crt | openssl rsa -pubin -outform der 2>/dev/null | openssl dgst -sha256 -hex | sed 's/^.* //'
```

### <3>. Join edge node into riginal Kubernetes cluster

Download the edgeadm static installation package in `2.<2>. on the edge node, or upload the edgeadm static installation package to the edge node by other means, and then execute the Join command obtained in 3.<2> on the edge node:

```powershell
./edgeadm join <Master public/Intranet IP or domain>:Port --token xxxx \
     --discovery-token-ca-cert-hash sha256:xxxxxxxxxx 
     --install-pkg-path <edgeadm Kube-* static installation package address/FTP path> 
     --enable-edge=true
```
>   ⚠️Note: You can change the service address of kube-apiserver in the join prompt command kube-apiserver printed by `edgeadm create token --print-join-command` to `Master node external network IP/Master node internal network IP/domain name`, depending on the situation. Yu want to let the edge node access the kube-apiserver service through the external network or the internal network, the default output of the internal network IP of the Master node.

among them:

-   <Master node external network IP/Master node internal network IP/domain name>: Port is the address of the edge node to access the kube-apiserver service

-   --enable-edge=true:  Whether the added node is used as an edge node (whether to deploy edge capability components), the default is true

>   --enable-edge=false Means to join the native Kubernetes cluster node, which is exactly the same as the node of kubeadm join;

-   --install-pkg-path: The path of the Kubernetes static installation package

>   The value of --install-pkg-path can be the path on the machine or the network address (for example: http://xxx/xxx/kube-linux-arm64/amd64-*.tar.gz, which can be encrypted without wget You can), pay attention to use the Kubernetes static installation package that matches the machine system;

If there is no problem in the execution process, the new Node successfully joins the cluster, and the following output will be output:

```shell
This node has joined the cluster:
* Certificate signing request was sent to apiserver and a response was received.
* The Kubelet was informed of the new secure connection details.

Run 'kubectl get nodes' on the control-plane to see this node join the cluster.
```
If there is a problem in the execution process, the corresponding error message will be directly returned, and the addition of the node will be interrupted. You can use the `./edgeadm reset` command to roll back the operation of joining the node and rejoin.

>    Tips: After successful join, the edge node will be labeled: `superedge.io/edge-node=enable`, which is convenient for subsequent applications to select application scheduling to the edge node through nodeSelector;

If you have any problems with the above operations, you can join SuperEdge's [Slack](https://join.slack.com/t/superedge-workspace/shared_invite/zt-ldxnm7er-ptdpCXthOct_dYrzyXM3pw), [Google Forum](https://groups .google.com/g/superedge), WeChat group to communicate with us, and you can also submit [Issues](https://github.com/superedge/superedge/issues) to give us feedback on issues in the community.