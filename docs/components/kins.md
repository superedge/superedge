# SuperEdge-Kins Manual

## Overview

This document will describe SuperEdge's new ability: Kins, which can promote the  standard edge nodes pool `NodeUnit` to an independent K3s cluster by just one click. Then this edge nodes pool can be disconnected from the cloud control plane and run totally offline for a long time. During the offline time, it can be operated as a standard K3s cluster for independent operation and maintenance. If there is a need for operation or upgrade in the later stage, the edge node pool can reconnect to the cloud control plane to achieve the goal. This feature improves the shortcomings of the previous `NodeUnit` which can't be operated offline.

## Goal

In the cloud native && edge computing scenario, most users have two basic scene demands for the edge side

- Manage the independent edge node through K8s control plane on cloud side

- The platform manages the independent K8s cluster on the edge side, also known
  as edge cloud

SuperEdge frameworks previously focused on the scenario of processing management of edge node. Although we provide similar NodeUnit capabilities and have certain edge regional autonomy capabilities, they cannot achieve complete offline independence and autonomy on the edge side, due to the architecture. It is limited in many scenarios such as intelligent industries. After communicating with many industry customers, we found that most industry customers are more likely to accept the edge side lightweight independent cluster, and these customers basically choose K3s as the solution of edge side products.

However, at the present time, this scheme can only be deployed and implemented in a way similar to privatization delivery. In the edge computing scenaro of multi regions/ few nodes/ scattered distribution, "cloud" + "edge cloud" project will encounter the difficulty of delivery and high cost.

Therefore, in response to this scenario, SuperEdge proposes this Kins architecture. Users can join the edge nodes to the cloud K8s control plane, and simply operate the governance level of `NodeUnit` in the cloud, then SuperEdge will provision K3s cluster on the edge nodes. The K3s cluster can be remotely controlled from the cloud side, and it also can be completely offline as a standalone cluster. It can be connected to the cloud control plane as needed, and remotely upgraded through cloud-edge tunnel for remote operation and maintance, greatly reducing delivery and maintance costs.

Through this new feature, it's possilbe to cover the majority of the basic requirements in edge computing scenarios.

## Architecture

<img src="https://qcloudimg.tencent-cloud.cn/raw/88fcfd8adf1de8177c77dd4d230ecfc8.png" width="100%" alt="">

The basic principle is show in the figure above:

Multiple nodes on the edge side can be grouped as `NodeUnit`. Through relevant operations on the cloud, SuperEdge will pull up the corresponding master and agent componentes of K3s to form a single master or 3 master K3s cluster. This edge K3s cluster can be accessed from the cloud through tunnel, and can also be directly accessed on the edge nodes.

## Usage steps

### 1. Prerequisite

Please create a SuperEdge cluster with the latest version of [Edgeadm](https://github.com/superedge/edgeadm/releases/tag/v0.9.0), and join the edge nodes as below:

<img title="" src="https://qcloudimg.tencent-cloud.cn/raw/ec667faa8cb42e3a86119e268b0326e1.jpg" alt="节点列表" width="377">

### 2. Create NodeUnit

Please use the standard `NodeUnit` CRD Yaml to create the edge nodes pool, as below:

```yaml
apiVersion: site.superedge.io/v1alpha2
kind: NodeUnit
metadata:
  name: test
spec:
  autonomyLevel: L3
  nodes:
  - edge1
  - edge2
  - edge3
  type: edge
  unschedulable: false
```

> Users only need to focus on these 2 varibles in configuration:
> 
> - autonomyLevel: This is the key point of Kins. L3 refert to the standard NodeUnit, and L4 refer to the single master K3s cluster, L5 means the 3 master K3s cluster
> 
> - nodes: refer to the list of nodes on the edge node pool

After creating the `NodeUnit`, you can review the corresponding information:

<img title="" src="https://qcloudimg.tencent-cloud.cn/raw/5521550e1ed8d4c4ff47de9b9bb6d822.jpg" alt="节点列表" width="478">

### 3. Modify NodeUnit to promote to K3s cluster

Modify the **test** NodeUnit's `autonomyLevel` to L4, as below:

<img title="" src="https://qcloudimg.tencent-cloud.cn/raw/eff56348dff846f4d364eb95a36a366e.jpg" alt="节点列表" width="641">

Observe the corresponding pods has been activated:

<img title="" src="https://qcloudimg.tencent-cloud.cn/raw/6580929e7bbe4a8db7593ef4df8ae4e2.jpg" alt="节点列表" width="641">

> Annotations:
> 
> test-server-kins: the pod referring to master node of K3s 
> 
> test-agnet-kins: the pod referring to worker node of K3s
> 
> test-cri-kins: components of cri proxy on the nodes

### 4. Access Edge K3s cluster

Now we can access this K3s cluster from both cloud and edge sides, as follows

#### 4.1 Local access on edge side

The `kubeconfig` of edge cluster is stored in the configmap of `kins-system` namespace. In this example, it's `test-cm-kins`, `test` prefix is the name of the NodeUnit.

The kubeconfig file can be acquired by the following command:

```shell
kubectl get cm test-cm-kins -n kins-system -o=jsonpath='{.data.kubeconfig\.conf}'
```

The kubeconfig file is show as follow:

<img title="" src="https://qcloudimg.tencent-cloud.cn/raw/d9a919e83befdae6eb4125be9bbc3abf.jpg" alt="节点列表" width="718">

Save the kubeconfig file to the desired edge node, for example, copy it to edge3 node and access the K3s cluster on this node:

<img title="" src="https://qcloudimg.tencent-cloud.cn/raw/d96c42bcc9b14292cbfbd88fd10bf4ec.jpg" alt="节点列表" width="565">

In the future, this NodeUnit node pool can be completely disconnected from the cloud and operated totally offline.

#### 4.2 Cloud access on the cloud side

If you want to access the edge K3s cluster from the cloud side, you need to use the proxy function of `Tunnel`, following the specific operations as follows:

- Obtain the svc information of `tunnel-cloud` and get the proxy configuration:

<img title="" src="https://qcloudimg.tencent-cloud.cn/raw/299a628d7192f7b9939de510c639e623.jpg" alt="节点列表" width="565">

> Record the endpoint the `http-proxy`, which can be accessed through the svc ip:8080. If you need to access the proxy out of the cluster, please use the master-ip:31469

- Confirm svc information of the edge K3s cluster

Get the K3s cluster's service address in the cluster, and you can use the service name or the service ip:

<img title="" src="https://qcloudimg.tencent-cloud.cn/raw/1532b1e81b2bc31c76db8ad00a3f324a.jpg" alt="节点列表" width="524">

- Modify configuration of kubeconfig of chapter 4.1

```yaml
  apiVersion: v1
  kind: Config
  clusters:
  - cluster:
      insecure-skip-tls-verify: true
      server: https://test-svc-kins.kins-system:443  #service name or service ip
      proxy-url: http://127.0.0.1:31469  # use the tunnel-cloud 的 http-proxy on master node
    name: default
  contexts:
  - context:
      cluster: default
      namespace: default
      user: default
    name: default
  current-context: default
  users:
  - name: default
    user:
      token: rfj9s2bhpcs6fm9xxxxxxxxxxxxxxxxx
```

- The result of accessing edge cluster from the cloud side:

<img title="" src="https://qcloudimg.tencent-cloud.cn/raw/13fc3f5c41af12103d8339c148737556.jpg" alt="节点列表" width="524">

### 5 Degrade/Delete edge K3s cluster

> Notice: If the autonomyLevel of NodeUnit is L4/L5, it can't be deleted directly. You need to dowgrade this NodeUnit to L3 before deleting. If the NodeUnit of L4/L5 is directly deleted, the cluster will block the deletion process and disply NodeUnit status as `Deleting`

If you need to downgrade the K3s cluster to standard NodeUnit, you can manually modify the NodeUnit to L3. After modification, the SuperEdge cluster will recycle the edge side pods, and the edge K3s cluster will be destroyed.
