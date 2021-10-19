[TOC]

# 1.背景

在边缘计算的场景下，边缘节点和云端为单向网络，从云端无法直接方法边缘节点，导致了以下的问题：

- 云端无法访问边缘端的service

- 边访问云端service需要以nodeport的形式

- 云边端podIp直通

为了使用户无感知单向网络带来的使用上的差异，我们与fabedge社区合作，实现在云边podIp直通，来屏蔽单向网络带来的使用上的差异。

# 2. fabedge的原理

<img src="fabric-edge-arch-v2.png" style="zoom:50%;" />

FabEdge在SuperEdge的基础上，建立了一个基于IPSec隧道的，三层的数据转发面，使能云端和边缘端POD通过IP地址直接进行通讯，包括普通的POD和使用了hostnework的POD，以及通过ClusterIP访问Service，不论Service的Endpoint在云端或边缘端。

FabEdge包括三个主要组件：

- Operator： 运行在云端任何节点，监听Node等资源的变化，为其它FabEdge组件维护证书，隧道等配置，并保存到相应configmap/secret；同时负责Agent的生命周期管理，包括创建/删除等。
- Connector： 运行在云端选定节点，使用Operator生成的配置，负责云端隧道的管理，负责在云端和边缘节点之间转发流量。
- Agent： 运行在边缘节点，使用Operator生成的配置，负责本节点的隧道，路由，iptables规则的管理。

# 3. fabedge与SuperEdge结合实现Service互访和podIp直通 方案验证

## 3.1 验证的环境

![img](https://raw.githubusercontent.com/00pf00/superedge/fabedge-pr/docs/img/Architecture.png)
在SuperEdge边缘独立集群中添加4个节点，2个节点（cloud-1和cloud-2）在云端和master节点在同一内网，2个节点（edge-1和edge-2）在边缘端。将cloud-1节点作为connector节点，将edge-1和edge-2加入community。具体的搭建过程，请参照[fabedge文档](https://github.com/FabEdge/fabedge/blob/Release-v0.3-beta/docs/integrate-with-superedge.md)。

## 3.2 验证场景

### 3.2.1云端pod访问边缘端pod

![img](https://raw.githubusercontent.com/00pf00/superedge/fabedge-pr/docs/img/fabedge%20cloud-pod%20access%20edge-pod.png)

#### 3.2.1.1 cloud-2上的pod访问边缘端edge-1上的pod

fabedge在edge-1节点的node资源写入cloud-1的节点的内网ip和flannnel.1网卡的mac地址，将edge-1伪装成cloud-1节点。

```
metadata:
  annotations:
    flannel.alpha.coreos.com/backend-data: '{"VtepMAC":"cloud-1 flannel.1 mac"}'
    flannel.alpha.coreos.com/backend-type: vxlan
    flannel.alpha.coreos.com/kube-subnet-manager: "true"
    flannel.alpha.coreos.com/public-ip: cloud-1 内网ip
```

cloud-2上的pod访问cloud-1上的pod请求，首先经过cni0网桥，根据路由规则，将请求转发到flannel.1上，由flannel.1对请求信息进行封包，由于fabedge将edge-1节点伪装成cloud-1节点，因此flannel.1会将封包之后的请求信息发送到cloud-1节点。cloud-1节点在接收到请求包之后，会在本节点的flannel.1对请求包进行解包，然后将请求通过ipsec
vpn隧道将请求转发到edge-1节点。edge-1节点在收到包之后根据路由规则将请求包发送br-fabedge网桥，然后再转发到pod中。
回包路径与请求包路径一样，响应消息到达cloud-1之后，先在flannel.1上进行封包，然后发送到cloud-2上，在flannel.1上进行解包

#### 3.2.1.2 cloud-1上的pod访问边缘端edge-1上的pod

cloud-1上的pod，由于不需要通过flannel的网络将请求转发到cloud-1，因此pod的请求不会经过cloud-1的flannel.1

### 3.2.2 edge-1上的pod访问edge-2上的pod

由于edge-1和edge-2在同一个community，fabedge会在节点之间建立ipsec vpn隧道，边缘节点pod之间的请求，会通过ipsec vpn隧道进行转发。

## 3.3 验证结果

### 3.3.1 云访边

| 边缘端pod的部署方式 | 云端pod的部署方式 |         测试项        | 测试结果 |
|:-------------------:|:-----------------:|:---------------------:|:--------:|
|     hostnetwork     |       podIp       | cloud-1 访问edge-1    | 通过     |
|                     |                   | cloud-1 访问edge-2    | 通过     |
|                     |                   | cloud-1 访问clusterIp | 通过     |
|        podIp        |    hostnetwork    | apiserver 访问service | 通过     |
|                     |                   | apiserver 访问edge-1  | 通过     |
|                     |                   | apiserver 访问edge-2  | 通过     |
|                     |       podIp       | cloud-1访问edge-1     | 通过     |
|                     |                   | cloud-1访问edge-2     | 通过     |
|                     |                   | cloud-1访问clusterIp  | 通过     |
|                     |                   | cloud-2 访问edge-1    | 通过     |
|                     |                   | cloud-2访问edge-2     | 通过     |
|                     |                   | cloud-2访问clusterIp  | 通过     |
|                     |    hostnetwork    | cloud-1 访问edge-1    | 通过     |
|                     |                   | cloud-1 访问edge-2    | 通过     |
|                     |                   | cloud-1 访问clusterIp | 通过     |
|                     |                   | cloud-2 访问edge-1    | 通过   |
|                     |                   | cloud-2 访问edge-2    | 通过   |
|                     |                   | cloud-2 访问clusterIp |  通过  |

### 3.3.2 边访云

| 云端pod的部署方式 | 边缘端pod的部署方式 | 测试项               | 测试结果 |
|-------------------|---------------------|----------------------|----------|
| podIp             | hostnetwork         | edge-1 访问cloud-1   | 通过     |
|                   |                     | edge-1 访问cloud-2   | 通过     |
|                   |                     | edge-1 访问clusterIp | 通过     |
|                   |                     | edge-2 访问cloud-1   | 通过     |
|                   |                     | edge-2 访问cloud-2   | 通过     |
|                   |                     | edge-2 访问clusterIp | 通过     |
|                   | podIp               | edge-1 访问cloud-1   | 通过     |
|                   |                     | edge-1 访问cloud-2   | 通过     |
|                   |                     | edge-1 访问clusterIp | 通过     |
|                   |                     | edge-2 访问cloud-1   | 通过     |
|                   |                     | edge-2 访问cloud-2   | 通过     |
|                   |                     | edge-2 访问clusterIp | 通过     |
| hostNetwork       | podIp               | edge-1 访问cloud-1   | 通过     |
|                   |                     | edge-1 访问cloud-2   | 通过     |
|                   |                     | edge-1 访问clusterIp | 通过     |
|                   |                     | edge-2 访问cloud-1   | 通过     |
|                   |                     | edge-2 访问cloud-2   | 通过     |
|                   |                     | edge-2 访问clusterIp | 通过     |

### 3.3.3 边访边

| 被访问的边缘端pod部署方式 | 发起请求的pod的部署方式 | 测试项               | 测试结果 |
|---------------------------|-------------------------|----------------------|----------|
| hostNetwok                | podIp                   | edge-2 访问edge-1    | 通过     |
|                           |                         | edge-2 访问clusterIp | 通过     |
|                           |                         | edge-1 访问edge-2    | 通过     |
|                           |                         | edge-1 访问clusterIp | 通过     |
| podIp                     | podIp                   | edge-2 访问edge-1    | 通过     |
|                           |                         | edge-2 访问clusterIp | 通过     |
|                           |                         | edge-1 访问edge-2    | 通过     |
|                           |                         | edge-1 访问clusterIp | 通过     |
|                           | hostnetwork             | edge-1 访问edge-2    | 通过     |
|                           |                         | edge-1 访问clusterIp | 通过     |
|                           |                         | edge-2 访问edge-1    | 通过     |
|                           |                         | edge-2 访问clusterIp | 通过     |

### 结论

根据以上的测试结果可以得出以下的结论：

- 使用fabedge可以实现云边端service互访
- 使用fabedge可以实现云边podIp直通
- 使用fabedge不影响边缘节点间pod的通信

# 4. 展望

- 支持更多的CNI，包括calico等
- 自动同步SuperEdge NodeUnit和FabEdge Community标签，简化边边通讯
- 支持FabEdge Connector的HA/HPA
