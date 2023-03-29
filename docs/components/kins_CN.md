# SuperEdge 集群上 Kins 使用指南

## 概述

本文详细介绍 SuperEdge 社区实现的新功能（Kins），一言以蔽之：此功能可以将之前标准的 NodeUnit 边缘节点池，一键提升为独立 K3s 集群，然后此边缘节点池可以和云端控制面断网，并长期离线运行使用，离线过程中可以作为一个标准的K3s 集群进行独立运维；后期如果有运维或者升级的需求，将节点池重新和云端建立连接后，即可从云端实现云上同步升级等远程运维操作。此功能改善之前 NodeUnit 断网后无法独立自治的缺陷。

## 使用流程

### 1. 前置条件

请使用 Edgeadm 的最新 v0.9.0 版本创建集群，并添加一个地域下的边缘节点，如下图：

<img title="" src="https://qcloudimg.tencent-cloud.cn/raw/ec667faa8cb42e3a86119e268b0326e1.jpg" alt="节点列表" width="377">

### 2. 创建 NodeUnit

请使用下面的 NodeUnit 标准 Yaml 文件，创建你需要的 NodeUnit 边缘节点池

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

> 这里普通用户只需要关心下面两个 yaml 变量配置即可：
> 
> autonomyLevel：这里是 NodeUnit 实现断网独立自治的关键设置；其中 L3 表示为标准的 NodeUnit 边缘节点池；L4 表示将 NodeUnit 提升为单 master 的边缘节点池；L5 表示将 NodeUnit 提升为 3 master 的边缘节点池（保证节点数>=3）
> 
> nodes：即该边缘节点池中的节点列表

创建完 NodeUnit 资源后，可以查看相应信息：

<img title="" src="https://qcloudimg.tencent-cloud.cn/raw/5521550e1ed8d4c4ff47de9b9bb6d822.jpg" alt="节点列表" width="478">

### 3. 修改 NodeUnit 提升为独立 K3s 集群

此时可以修改`test`的 nodeunit 配置，将`autonomyLevel`修改为 L4，如下图

<img title="" src="https://qcloudimg.tencent-cloud.cn/raw/eff56348dff846f4d364eb95a36a366e.jpg" alt="节点列表" width="641">

观察相应的 Pod 已经启动：
<img title="" src="https://qcloudimg.tencent-cloud.cn/raw/6580929e7bbe4a8db7593ef4df8ae4e2.jpg" alt="节点列表" width="641">

> 注解：
> 
> test-server-kins：k3s 集群的 master 节点对应的 pod
> 
> test-agent-kins：k3s 集群的 worker 节点对应的 pod
> 
> test-cri-kins：节点上用作 cri 代理的组件

### 4. 访问边缘 K3s 集群

现在边缘独立 K3s 集群已经创建 OK，我们分别可以从云端和边缘侧两个位置，访问这个 K3s集群，使用方式如下：

#### 4.1 边缘侧本地访问 K3s 集群

边缘 K3s 集群的 kubeconfig 默认存储在`kins-system`命名空间下的 configmap 中，本例中名称为`test-cm-kins`，其中前缀`test`根据 NodeUnit 名称动态生成。

可以通过下面的命令获取 kubeconfig 文件：

```shell
kubectl get cm test-cm-kins -n kins-system -o=jsonpath='{.data.kubeconfig\.conf}'
```

获取 kubeconfig 信息如下：
<img title="" src="https://qcloudimg.tencent-cloud.cn/raw/d9a919e83befdae6eb4125be9bbc3abf.jpg" alt="节点列表" width="718">

保存 kubeconfig 文件到需要的边缘节点，例如本例中拷贝到 edge3 节点，即可在 edge3 节点上访问边缘集群，如下：

<img title="" src="https://qcloudimg.tencent-cloud.cn/raw/d96c42bcc9b14292cbfbd88fd10bf4ec.jpg" alt="节点列表" width="565">

后续这个 NodeUnit 节点池，就可以完全和云端 master 断连后，离线提供标准的 K3s 集群服务。

#### 4.2 云端访问 K3s 集群

如果想要从云端访问边缘 K3s 集群，需要用到superedge 标准的`tunnel`的 http/https代理能力，具体操作如下

- 通过集群中 `tunnel-cloud`的 svc 信息，从 tunnel-cloud 的 svc 种获取云端代理配置，如下图：

<img title="" src="https://qcloudimg.tencent-cloud.cn/raw/299a628d7192f7b9939de510c639e623.jpg" alt="节点列表" width="565">

> 记录 http-proxy 的端口，集群内可以用 tunnel-cloud 的svc-ip:8080访问；如果需要从集群外访问，可以如上图，使用 master 的节点 IP:31469访问

- 确认边缘k3s集群的 svc 信息
  
  通过 service 确认需要访问的边缘 k3s 集群在集群内的 svc 访问地址，如下图
  
  <img title="" src="https://qcloudimg.tencent-cloud.cn/raw/1532b1e81b2bc31c76db8ad00a3f324a.jpg" alt="节点列表" width="524">
  
  我们需要访问的 svc 地址即为<NodeUnit>-svc-kins，此例为`test-svc-kins`，可以使用 svc 名称或者 svc ip 访问均可

- 修改 4.1 节边缘侧的 kubeconfig 文件，如下：
  
  ```yaml
  apiVersion: v1
  kind: Config
  clusters:
  - cluster:
      insecure-skip-tls-verify: true
      server: https://test-svc-kins.kins-system:443  #这里可以使用 svc 地址
      proxy-url: http://127.0.0.1:31469  #这里在 master 节点上使用 tunnel-cloud 的 http 代理
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

- 通过上述的 kubeconfig 文件，可以从 master 节点访问边缘侧 K3s 集群：
  
  <img title="" src="https://qcloudimg.tencent-cloud.cn/raw/13fc3f5c41af12103d8339c148737556.jpg" alt="节点列表" width="524">

### 5. 降级/删除边缘侧 K3s 集群

> 注意：如果 NodeUnit 的 autonomyLevel 为 L4/L5，无法直接删除此 NodeUnit，需要首先手动将此 NodeUnit 降级为 L3，然后才能删除；如果直接删除 L4/L5的 NodeUnit，集群会阻塞删除流程，同时显示 NodeUnit 状态为`Deleting`

如果需要将 NodeUnit 独立 K3s 集群恢复为标准 NodeUnit，可以手动修改 NodeUnit 的`autonomyLevel`为 L3 即可；修改后 superedge 集群会回收所有的边缘侧 Pod，集群会被destroy
