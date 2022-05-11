# 边缘应用管理利器: ServiceGroup

# 功能背景

## 边缘特点

- 边缘计算场景中，往往会在同一个集群中管理多个边缘站点，每个边缘站点内有一个或多个计算节点。
- 同时希望在每个站点中都运行一组有业务逻辑联系的服务，每个站点内的服务是一套完整的功能，可以为用户提供服务
- 由于受到网络限制，有业务联系的服务之间不希望或者不能跨站点访问

# 操作场景

serviceGroup可以便捷地在共属同一个集群的不同机房或区域中各自部署一组服务，并且使得各个服务间的请求在本机房或本地域内部即可完成，避免服务跨地域访问。

原生 k8s 无法控制deployment的pod创建的具体节点位置，需要通过统筹规划节点的亲和性来间接完成，当边缘站点数量以及需要部署的服务数量过多时，管理和部署方面的极为复杂，乃至仅存在理论上的可能性；

与此同时，为了将服务间的相互调用限制在一定范围，业务方需要为各个deployment分别创建专属的service，管理方面的工作量巨大且极容易出错并引起线上业务异常。

serviceGroup就是为这种场景设计的，客户只需要使用ServiceGroup提供的DeploymentGrid，StatefulSetGrid以及ServiceGrid三种SuperEdge自研的kubernetes 资源，即可方便地将服务分别部署到这些节点组中，并进行服务流量管控，另外，还能保证各区域服务数量及容灾。

本文以详细的案例结合具体的实现原理，来详细说明 ServiceGroup 的使用场景以及需要关注的细节问题

# 关键概念

## 整体架构

<div align="left">
  <img src="../img/serviceGroup-UseCase.png" width=70% title="service-group">
</div>

## 基本概念介绍

> 关于 NodeUnit 和 NodeGroup 最新版设计可以参考链接：[边缘节点池和边缘节点池分类设计文档](https://github.com/superedge/superedge/blob/main/docs/components/site-manager_CN.md)

### NodeUnit（边缘节点池）

- NodeUnit通常是位于同一边缘站点内的一个或多个计算资源实例，需要保证同一NodeUnit中的节点内网是通的
- ServiceGroup组中的服务运行在一个NodeUnit之内
- ServiceGroup 允许用户设置服务在一个 NodeUnit中运行的pod数量
- ServiceGroup 能够把服务之间的调用限制在本 NodeUnit 内

### NodeGroup（边缘节点池分类）

- NodeGroup 包含一个或者多个 NodeUnit
- 保证在集合中每个 NodeUnit上均部署ServiceGroup中的服务
- 集群中增加 NodeUnit 时自动将 ServiceGroup 中的服务部署到新增 NodeUnit

### ServiceGroup

- ServiceGroup 包含一个或者多个业务服务:适用场景：1）业务需要打包部署；2）或者，需要在每一个 NodeUnit 中均运行起来并且保证pod数量；3）或者，需要将服务之间的调用控制在同一个 NodeUnit 中，不能将流量转发到其他 NodeUnit。
- 注意：ServiceGroup是一种抽象资源，一个集群中可以创建多个ServiceGroup

**ServiceGroup 涉及的资源类型包括如下三类：**

####  DeploymentGrid

DeploymentGrid的格式与Deployment类似，<deployment-template>字段就是原先deployment的template字段，比较特殊的是gridUniqKey字段，该字段指明了节点分组的label的key值：

```yaml
apiVersion: superedge.io/v1
kind: DeploymentGrid
metadata:
  name:
  namespace:
spec:
  gridUniqKey: <NodeLabel Key>
  <deployment-template>
```

#### StatefulSetGrid

StatefulSetGrid的格式与StatefulSet类似，<statefulset-template>字段就是原先statefulset的template字段，比较特殊的是gridUniqKey字段，该字段指明了节点分组的label的key值：

```yaml
apiVersion: superedge.io/v1
kind: StatefulSetGrid
metadata:
  name:
  namespace:
spec:
  gridUniqKey: <NodeLabel Key>
  <statefulset-template>
```

#### ServiceGrid

ServiceGrid的格式与Service类似，<service-template>字段就是原先service的template字段，比较特殊的是gridUniqKey字段，该字段指明了节点分组的label的key值：

```yaml
apiVersion: superedge.io/v1
kind: ServiceGrid
metadata:
  name:
  namespace:
spec:
  gridUniqKey: <NodeLabel Key>
  <service-template>
```

# 操作步骤

以在边缘部署echo-service为例，我们希望在多个节点组内分别部署echo-service服务，需要做如下事情：

## 确定ServiceGroup唯一标识

这一步是逻辑规划，不需要做任何实际操作。我们将目前要创建的serviceGroup逻辑标记使用的UniqKey为：`location`

## 将边缘节点分组

如下图，我们以一个边缘集群为例，将集群中的节点添加到`边缘节点池`以及`边缘节点池分类`中。SuperEdge开源用户可以参考[边缘节点池和边缘节点池分类设计文档](https://github.com/superedge/superedge/blob/main/docs/components/site-manager_CN.md) 来使用 CRD 进行操作完成下面的步骤

此集群包含 6 个边缘节点，分别位于北京/广州 2 个地域，节点名为`bj-1` `bj-2` `bj-3`  `gz-1` `gz-2`  `gz-3`

```shell
[~]# kubectl get nodes
NAME   STATUS   ROLES    AGE     VERSION
bj-1   Ready    <none>   4d22h   v1.18.2
bj-2   Ready    <none>   4d22h   v1.18.2
bj-3   Ready    <none>   4d22h   v1.18.2
gz-1   Ready    <none>   4d22h   v1.18.2
gz-2   Ready    <none>   4d22h   v1.18.2
gz-3   Ready    <none>   4d22h   v1.18.2
```

然后我们分别创建 2 个NodeUnit（边缘节点池）：`beijing`  `guangzhou` ，分别将相应的节点加入对应的 NodeUnit（边缘节点池）中，如下图

```shell
[~]# kubectl get nodeunit
NAME            TYPE    READY   AGE
beijing         edge    3/3     3d17h
guangzhou       edge    3/3     3d17h
unit-node-all   other   6/6     4d22h
```

```shell
[~]# kubectl describe nodeunit beijing
Spec:
  Nodes:
    bj-1
    bj-2
    bj-3
  Setnode:
    Labels:
      Beijing:    nodeunits.superedge.io
      Location:   beijing
  Type:           edge
  Unschedulable:  false
Status:
  Readynodes:
    bj-1
    bj-2
    bj-3
  Readyrate:  3/3
Events:       <none>

[~]# kubectl describe nodeunit guangzhou
Spec:
  Nodes:
    gz-1
    gz-2
    gz-3
  Setnode:
    Labels:
      Guangzhou:  nodeunits.superedge.io
      Location:   guangzhou
  Type:           edge
  Unschedulable:  false
Status:
  Readynodes:
    gz-1
    gz-2
    gz-3
  Readyrate:  3/3
Events:       <none>
```


最后，我们创建名称为 `location`的 NodeGroup（边缘节点池分类），将`beijing` `guangzhou` 这两个边缘节点池划分到`location`这个分类中，如下图

```shell
[~]# kubectl describe nodegroup location
Spec:
  Nodeunits:
    beijing
    guangzhou
Status:
  Nodeunits:
    beijing
    guangzhou
  Unitnumber:  2
Events:        <none>
```


按照上述界面操作后，其实每个节点上就会打上相应的标签，例如节点 gz-2 上就会打上标签

<div align="left">
  <img src="../img/demo_node_label.jpg" width=50% title="service-group">
</div>

注意：上一步中，label的key 就是 NodeGroup 的名字，同时与ServiceGroup的UniqKey一致，value是NodeUnit的唯一key，value相同的节点表示属于同一个NodeUnit

如果同一个集群中有多个NodeGroup 请为每一个NodeGroup 分配不同的UniqKey，部署ServiceGroup 相关资源的时候会通过 UniqKey 来绑定指定的 NodeGroup 进行部署

## 无状态ServiceGroup

### 部署DeploymentGrid

```yaml
apiVersion: superedge.io/v1
kind: DeploymentGrid
metadata:
  name: deploymentgrid-demo
  namespace: default
spec:
  gridUniqKey: location
  template:
    replicas: 2
    selector:
      matchLabels:
        appGrid: echo
    strategy: {}
    template:
      metadata:
        creationTimestamp: null
        labels:
          appGrid: echo
      spec:
        containers:
        - image: superedge/echoserver:2.2
          name: echo
          ports:
          - containerPort: 8080
            protocol: TCP
          env:
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: POD_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.podIP
          resources: {}
```

### 部署ServiceGrid

```yaml
apiVersion: superedge.io/v1
kind: ServiceGrid
metadata:
  name: servicegrid-demo
  namespace: default
spec:
  gridUniqKey: location
  template:
    selector:
      appGrid: echo
    ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
```

gridUniqKey字段设置为了location，所以我们在将节点分组时采用label的key为location；这时，`bejing` 和 `guangzhou` 的 NodeUnit内都有了echo-service的deployment和对应的pod，在节点内访问统一的service-name也只会将请求发向本组的节点

```shell
[~]# kubectl get dg
NAME                  AGE
deploymentgrid-demo   9s

[~]# kubectl get deployment
NAME                            READY   UP-TO-DATE   AVAILABLE   AGE
deploymentgrid-demo-beijing     2/2     2            2           14s
deploymentgrid-demo-guangzhou   2/2     2            2           14s

[~]# kubectl get pods -o wide
NAME                                             READY   STATUS    RESTARTS   AGE     IP           NODE   NOMINATED NODE   READINESS GATES
deploymentgrid-demo-beijing-65d669b7d-v9zdr      1/1     Running   0          6m51s   10.0.1.72    bj-3   <none>           <none>
deploymentgrid-demo-beijing-65d669b7d-wrx7r      1/1     Running   0          6m51s   10.0.0.70    bj-1   <none>           <none>
deploymentgrid-demo-guangzhou-5d599854c8-hhmt2   1/1     Running   0          6m52s   10.0.0.139   gz-2   <none>           <none>
deploymentgrid-demo-guangzhou-5d599854c8-k9gc7   1/1     Running   0          6m52s   10.0.1.8     gz-3   <none>           <none>

#从上面可以看到，对于一个 deploymentgrid，会分别在每个 nodeunit 下创建一个 deployment，他们会通过 <deployment>-<nodeunit>名称的方式来区分

[~]# kubectl get svc
NAME                   TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
kubernetes             ClusterIP   172.16.33.1     <none>        443/TCP             47h
servicegrid-demo-svc   ClusterIP   172.16.33.231   <none>        80/TCP              3s

[~]# kubectl describe svc servicegrid-demo-svc
Name:              servicegrid-demo-svc
Namespace:         default
Labels:            superedge.io/grid-selector=servicegrid-demo
                   superedge.io/grid-uniq-key=location
Annotations:       topologyKeys: ["location"]
Selector:          appGrid=echo
Type:              ClusterIP
IP Families:       <none>
IP:                172.16.33.231
IPs:               <none>
Port:              <unset>  80/TCP
TargetPort:        8080/TCP
Endpoints:         10.0.0.139:8080,10.0.0.70:8080,10.0.1.72:8080 + 1 more...
Session Affinity:  None
Events:            <none>

#从上面可以看到，对于一个 servicegrid，都会创建一个 <servicename>-svc 的标准 Service；
#！！！注意，这里的 Service 对应的后端 Endpoint 仍然为所有 pod 的 endpoint 地址，这里并不会按照 nodeunit 进行 endpoint 筛选

# 在 beijing 地域的 pod 内执行下面的命令
[~]# curl 172.16.33.231|grep "node name"
        node name:      gz-2
...
# 这里会随机返回 gz-2 或者 gz-3 的 node 名称，并不会跨 NodeUnit 访问到 bj-1或者 bj-3

# 在 guangzhou 地域的 pod 执行下面的命令
[~]# curl 172.16.33.231|grep "node name"
        node name:      bj-3
# 这里会随机返回 bj-1或者 bj-3 的 node 名称，并不会跨 NodeUnit 访问到 gz-2 或者 gz-3
```

另外，对于部署了DeploymentGrid和ServiceGrid后才添加进集群的节点组，该功能会在新的节点组内自动创建指定的deployment

### 原理解析

下面来简单介绍一些 ServiceGroup 如何实现不同 NodeUnit 地域的 Service 访问的流量闭环的，如下图：

<div align="left">
  <img src="../img/demo_deploymentgrid_internal.jpg" width=100% title="deploymentgrid">
</div>
简单的原理剖析：

- 当创建一个 DeploymentGrid 的时候，通过云端的 application-grid-controller 服务，会分别在每个NodeUnit 上生成一个单独的标准 Deployment（例如 deploymentgrid-demo-beijing-XXXXX）
- 当创建相应的 ServiceGrid 的时候，会在集群中创建一个标准的 Service，如上图`servicegrid-demo-svc`
- 这个时候其实都是标准的 Deployment 和标准 Service，这两个行为其实都没有办法根据 NodeUnit 实现流量闭环。这个时候其实就需要`application-grid-wrapper` 这个组件来参与了
- 从上图可以看到`application-grid-wrapper` 组件部署在每一个边缘 node 上，同时边缘侧`kube-proxy` 会通过`application-grid-wrapper`和 apiserver 通信，获取相应资源信息；这里`application-grid-wrapper`会监听 ServiceGrid的CRD 信息，同时在获取到对应的 Service 的 Endpoint 信息后，就会根据所在 NodeUnit 的节点信息进行筛选，将不在同一 NodeUnit 的 Node 上的 Endpoint 剔除，传递给`kube-proxy`更新 iptables 规则。下面就是左侧`bj-3` 北京地域节点上的 iptables 规则：


```shell
-A KUBE-SERVICES -d 172.16.33.231/32 -p tcp -m comment --comment "default/servicegrid-demo-svc: cluster IP" -m tcp --dport 80 -j KUBE-SVC-MLDT4NC26VJPGLP7

-A KUBE-SVC-MLDT4NC26VJPGLP7 -m comment --comment "default/servicegrid-demo-svc:" -m statistic --mode random --probability 0.50000000000 -j KUBE-SEP-VB3QR2E2PUKDLHCW

-A KUBE-SVC-MLDT4NC26VJPGLP7 -m comment --comment "default/servicegrid-demo-svc:" -j KUBE-SEP-U5ZEIIBVDDGER3DI

-A KUBE-SEP-U5ZEIIBVDDGER3DI -p tcp -m comment --comment "default/servicegrid-demo-svc:" -m tcp -j DNAT --to-destination 10.0.1.72:8080

-A KUBE-SEP-VB3QR2E2PUKDLHCW -p tcp -m comment --comment "default/servicegrid-demo-svc:" -m tcp -j DNAT --to-destination 10.0.0.70:8080
```

- 从上面规则分析可以看到，`beijing`地域中，172.16.33.231 的 ClusterIP 只会分流到`10.0.1.72`和`10.0.0.70`两个后端 Endpooint 上，对应两个 pod：`deploymentgrid-demo-beijing-65d669b7d-v9zdr`和`deploymentgrid-demo-beijing-65d669b7d-wrx7r`，而且不会添加上`guangzhou`地域的两个 IP 10.0.0.139 和 10.0.1.8，按照这样的逻辑，就可以在不同的 NodeUnit 中实现流量闭环能力了

> **需要注意以下2个场景：**
>
> - DeploymentGrid + 标准 Service 能否实现流量闭环？
>
> 当然不可以，通过上述的分析，如果是标准的 Service，这个 Service:Endpoint 列表并不会被`application-grid-wrapper` 来监听处理，因此这里会获取全量的 Endpoint 列表，`kube-proxy`更新 iptables 规则的时候就会添加相应规则，将流量导出到其他 NodeUnit 中
>
> - DeploymentGrid + Headless Service 是否可以实现流量闭环？
>
> 可以根据下面的 Yaml 文件部署一个 Headless Service，同样使用 ServiceGrid 的模板来创建：
>
> ```yaml
> apiVersion: superedge.io/v1
> kind: ServiceGrid
> metadata:
>   name: servicegrid-demo
>   namespace: default
> spec:
>   gridUniqKey: location
>   template:
>     clusterIP: None
>     selector:
>       appGrid: echo
>     ports:
>     - protocol: TCP
>       port: 8080
>       targetPort: 8080
> ```
>
> 获取 Service 信息如下：
>
> ```
> Name:              servicegrid-demo-svc
> Namespace:         default
> Labels:            superedge.io/grid-selector=servicegrid-demo
>                    superedge.io/grid-uniq-key=location
> Annotations:       topologyKeys: ["location"]
> Selector:          appGrid=echo
> Type:              ClusterIP
> IP Families:       <none>
> IP:                None
> IPs:               <none>
> Port:              <unset>  8080/TCP
> TargetPort:        8080/TCP
> Endpoints:         10.0.0.139:8080,10.0.0.70:8080,10.0.1.72:8080 + 1 more...
> Session Affinity:  None
> Events:            <none>
> ```
>
> 这里能够看到，仍然能够获取 4 个 Endpoint 的信息，同时 ClusterIP 的地址为空，在集群中的任意 Pod 里通过 nslookup 获取的 Service 地址就是 4 个后端 Endpoint，如下：
>
> ```
> [~]# nslookup servicegrid-demo-svc.default.svc.cluster.local
> Server:         169.254.20.11
> Address:        169.254.20.11#53
> 
> Name:   servicegrid-demo-svc.default.svc.cluster.local
> Address: 10.0.1.8
> Name:   servicegrid-demo-svc.default.svc.cluster.local
> Address: 10.0.1.72
> Name:   servicegrid-demo-svc.default.svc.cluster.local
> Address: 10.0.0.70
> Name:   servicegrid-demo-svc.default.svc.cluster.local
> Address: 10.0.0.139
> ```
>
> 所以如果通过这种形式访问 servicegrid-demo-svc.default.svc.cluster.local 任然会访问到其余地域的 NodeUnit 内，无法实现流量闭环
>
> **同时，由于 Deployment 后端 pod 无状态，会随时重启更新，Pod 名称会随机变化，同时 Deployment 的 Pod 不会自动创建 DNS 地址，因此一般我们不会建议 Deployment + Headless Service 这样配合使用；而是推荐大家使用 StatefulsetGrid + Headless Service 的方式**
>
> - 通过上述的分析，其实可以理解：如果访问 Service 的行为会通过 `kube-proxy`的 iptables 规则去进行转发，同时 Service 是 SerivceGrid 类型，会被`application-grid-wrapper`监听的话，就可以实现区域流量闭环；如果是通过 DNS 获取的实际 Endpoint IP 地址，这样就无法实现流量闭环



## 有状态ServiceGroup

### 部署StatefulSetGrid

```yaml
apiVersion: superedge.io/v1
kind: StatefulSetGrid
metadata:
  name: statefulsetgrid-demo
  namespace: default
spec:
  gridUniqKey: location
  template:
    selector:
      matchLabels:
        appGrid: echo
    serviceName: "servicegrid-demo-svc"
    replicas: 3
    template:
      metadata:
        labels:
          appGrid: echo
      spec:
        terminationGracePeriodSeconds: 10
        containers:
        - image: superedge/echoserver:2.2
          name: echo
          ports:
          - containerPort: 8080
            protocol: TCP
          env:
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: POD_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.podIP
          resources: {}
```

**注意：template中的serviceName设置成即将创建的service名称**

### 部署ServiceGrid

```yaml
apiVersion: superedge.io/v1
kind: ServiceGrid
metadata:
  name: servicegrid-demo
  namespace: default
spec:
  gridUniqKey: location
  template:
    selector:
      appGrid: echo
    ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
```

gridUniqKey字段设置为了location，因此这个 NodeGroup 依然包含`beijing`和`guangzhou`两个 NodeUnit，每个 NodeUnit 内都有了echo-service的statefulset和对应的pod，在节点内访问统一的service-name也只会将请求发向本组的节点

```shell
[~]# kubectl get ssg
NAME                   AGE
statefulsetgrid-demo   31s

[~]# kubectl get statefulset
NAME                             READY   AGE
statefulsetgrid-demo-beijing     3/3     49s
statefulsetgrid-demo-guangzhou   3/3     49s

[~]# kubectl get pods -o wide
NAME                               READY   STATUS    RESTARTS   AGE     IP           NODE   NOMINATED NODE   READINESS GATES
statefulsetgrid-demo-beijing-0     1/1     Running   0          9s      10.0.0.67    bj-1   <none>           <none>
statefulsetgrid-demo-beijing-1     1/1     Running   0          8s      10.0.1.67    bj-3   <none>           <none>
statefulsetgrid-demo-beijing-2     1/1     Running   0          6s      10.0.0.69    bj-1   <none>           <none>
statefulsetgrid-demo-guangzhou-0   1/1     Running   0          9s      10.0.0.136   gz-2   <none>           <none>
statefulsetgrid-demo-guangzhou-1   1/1     Running   0          8s      10.0.1.7     gz-3   <none>           <none>
statefulsetgrid-demo-guangzhou-2   1/1     Running   0          6s      10.0.0.138   gz-2   <none>           <none>

[~]# kubectl get svc
NAME                   TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
kubernetes             ClusterIP   172.16.33.1     <none>        443/TCP             2d2h
servicegrid-demo-svc   ClusterIP   172.16.33.220   <none>        80/TCP              30s

[~]# kubectl describe svc servicegrid-demo-svc
Name:              servicegrid-demo-svc
Namespace:         default
Labels:            superedge.io/grid-selector=servicegrid-demo
                   superedge.io/grid-uniq-key=location
Annotations:       topologyKeys: ["location"]
Selector:          appGrid=echo
Type:              ClusterIP
IP Families:       <none>
IP:                172.16.33.220
IPs:               <none>
Port:              <unset>  80/TCP
TargetPort:        8080/TCP
Endpoints:         10.0.0.136:8080,10.0.0.138:8080,10.0.0.67:8080 + 3 more...
Session Affinity:  None
Events:            <none>

# 在 guangzhou 地域的 pod 中访问 CluserIP，会随机得到 gz-2 gz-3 节点名称，不会访问到 beijing 区域的 Pod；使用 Service 的域名访问效果一致
[~]# curl 172.16.33.220|grep "node name"
        node name:      gz-2
[~]# curl servicegrid-demo-svc.default.svc.cluster.local|grep "node name"
        node name:      gz-3
        
# 在 beijing 地域的 pod 中访问 CluserIP，会随机得到 bj-1 bj-3 节点名称，不会访问到 guangzhou 区域的 Pod；使用 Service 的域名访问效果一致
[~]# curl 172.16.33.220|grep "node name"
        node name:      bj-1
[~]# curl servicegrid-demo-svc.default.svc.cluster.local|grep "node name"
        node name:      bj-3

```

我们在`guangzhou`地域的节点上检查 iptables 规则，如下

```shell
-A KUBE-SERVICES -d 172.16.33.220/32 -p tcp -m comment --comment "default/servicegrid-demo-svc: cluster IP" -m tcp --dport 80 -j KUBE-SVC-MLDT4NC26VJPGLP7

-A KUBE-SVC-MLDT4NC26VJPGLP7 -m comment --comment "default/servicegrid-demo-svc:" -m statistic --mode random --probability 0.33333333349 -j KUBE-SEP-Z2EAS2K37V5WRDQC
  -A KUBE-SVC-MLDT4NC26VJPGLP7 -m comment --comment "default/servicegrid-demo-svc:" -m statistic --mode random --probability 0.50000000000 -j KUBE-SEP-PREBTG6M6AFB3QA4
-A KUBE-SVC-MLDT4NC26VJPGLP7 -m comment --comment "default/servicegrid-demo-svc:" -j KUBE-SEP-URDEBXDF3DV5ITUX

-A KUBE-SEP-Z2EAS2K37V5WRDQC -p tcp -m comment --comment "default/servicegrid-demo-svc:" -m tcp -j DNAT --to-destination 10.0.0.136:8080
-A KUBE-SEP-PREBTG6M6AFB3QA4 -p tcp -m comment --comment "default/servicegrid-demo-svc:" -m tcp -j DNAT --to-destination 10.0.0.138:8080
-A KUBE-SEP-URDEBXDF3DV5ITUX -p tcp -m comment --comment "default/servicegrid-demo-svc:" -m tcp -j DNAT --to-destination 10.0.1.7:8080
```

通过 iptables 规则很明显的可以看到对 `servicegrid-demo-svc`的访问分别 redirect 到了 `10.0.0.136` `10.0.0.138` `10.0.1.7`这 3 个地址，分别对应的就是 guangzhou 地域的 3 个 pod 的 IP 地址

可以看到，如果是StatefulsetGrid + 标准 ServiceGrid 访问方式的话，其原理和上面的 DeploymentGrid 原理一致，都是通过`application-grid-wrapper`配合`kube-proxy`修改 iptables 规则来实现的，没有任何区别

### StatefusetGrid + Headless Service 支持

StatefulSetGrid目前支持使用Headless service**配合Pod FQDN**的方式进行闭环访问，如下所示：

<div align="left">
  <img src="../img/demo_stsgrid_headless.jpg" width=100% title="stsgrid">
</div>

上图中 CoreDNS 里面的两条记录可以暂时不用理解，继续看下文描述

#### 部署 Headless Service

按照下面的模板部署 Headless Service

```yaml
apiVersion: superedge.io/v1
kind: ServiceGrid
metadata:
  name: servicegrid-demo
  namespace: default
spec:
  gridUniqKey: location
  template:
    clusterIP: None
    selector:
      appGrid: echo
    ports:
    - protocol: TCP
      port: 8080
      targetPort: 8080
```

```
[~]# kubectl describe svc servicegrid-demo-svc
Name:              servicegrid-demo-svc
Namespace:         default
Labels:            superedge.io/grid-selector=servicegrid-demo
                   superedge.io/grid-uniq-key=location
Annotations:       topologyKeys: ["location"]
Selector:          appGrid=echo
Type:              ClusterIP
IP Families:       <none>
IP:                None
IPs:               <none>
Port:              <unset>  8080/TCP
TargetPort:        8080/TCP
Endpoints:         10.0.0.136:8080,10.0.0.138:8080,10.0.0.67:8080 + 3 more...
Session Affinity:  None
Events:            <none>
```

这里可以看到 Service 的 ClusterIP 为空，Endpoint 仍然包含 6 个 pod 的IP 地址，同时，如果通过域名查询 `servicegrid-demo-svc.default.svc.cluster.local`会得到下面的信息：

```shell
[~]# nslookup servicegrid-demo-svc.default.svc.cluster.local
Server:         169.254.20.11
Address:        169.254.20.11#53

Name:   servicegrid-demo-svc.default.svc.cluster.local
Address: 10.0.1.7
Name:   servicegrid-demo-svc.default.svc.cluster.local
Address: 10.0.0.136
Name:   servicegrid-demo-svc.default.svc.cluster.local
Address: 10.0.0.138
Name:   servicegrid-demo-svc.default.svc.cluster.local
Address: 10.0.0.69
Name:   servicegrid-demo-svc.default.svc.cluster.local
Address: 10.0.1.67
Name:   servicegrid-demo-svc.default.svc.cluster.local
Address: 10.0.0.67
```

这个时候如果我们通过标准的`servicegrid-demo-svc.default.svc.cluster.local`来访问服务，仍然会跨 NodeUnit 随机访问不同的 Endpoint 地址，无法实现流量闭环。

#### 如何在一个 NodeUnit 内支持 Statefulset 标准访问方式

在一个标准的 K8s 环境中，按照 Statefulset 的标准使用方式，我们会使用一种逻辑来访问 Statefulset 中的 Pod ，类似`Statefulset-0.SVC.NS.svc.cluster.local` 这样的格式，例如我们想要使用`statefulsetgrid-demo-0.servicegrid-demo-svc.default.svc.cluster.local`来访问这个 Statefulset 中的 Pod-0，SuperEdge 针对这个场景进行了多NodeUnit 的能力适配。

由于 Statefulset 的 Pod 都有独立的 DNS 域名，可以通过 FQDN 方式来访问单独的 pod，例如可以查询`statefulsetgrid-demo-beijing-0`域名：

```shell
[~]# nslookup statefulsetgrid-demo-beijing-0.servicegrid-demo-svc.default.svc.cluster.local
Server:         169.254.20.11
Address:        169.254.20.11#53

Name:   statefulsetgrid-demo-beijing-0.servicegrid-demo-svc.default.svc.cluster.local
Address: 10.0.0.67
```

因此可以开始考虑一种方式，是不是可以抛弃掉 NodeUnit 的标记，直接使用`statefulsetgrid-demo-0.servicegrid-demo-svc.default.svc.cluster.local`的域名来访问本地域的 Statefulset 内的 Pod-0 呢，如下：

> 在`beijing`地域访问的就是`statefulsetgrid-demo-beijing-0.servicegrid-demo-svc.default.svc.cluster.local` 这个 Pod 的 IP
>
> 在`guangzhou`地域访问的就是`statefulsetgrid-demo-guangzhou-0.servicegrid-demo-svc.default.svc.cluster.local` 这个 Pod 的 IP

上图中使用 CoreDNS 两条记录指向相同的 Pod IP ，这个能力就可以实现上述的标准访问需求。因此 **SuperEdge 在产品层面提供了相应的能力**，用户需要独立部署下面的Daemonset服务`statefulset-grid-daemon`， 边缘侧节点上都会部署一个服务 `statefulset-grid-daemon`， 这个服务会来更新节点侧的 CoreDNS 信息，参考部署链接：[部署 statefulset-grid-daemon](https://github.com/superedge/superedge/blob/main/deployment/statefulset-grid-daemon.yaml)

```shell
edge-system   statefulset-grid-daemon-8gtrz      1/1     Running   0          7h42m   172.16.35.193   gz-3   <none>           <none>
edge-system   statefulset-grid-daemon-8xvrg      1/1     Running   0          7h42m   172.16.32.211   gz-2   <none>           <none>
edge-system   statefulset-grid-daemon-ctr6w      1/1     Running   0          7h42m   192.168.10.15   bj-3   <none>           <none>
edge-system   statefulset-grid-daemon-jnvxz      1/1     Running   0          7h42m   192.168.10.12   bj-1   <none>           <none>
edge-system   statefulset-grid-daemon-v9llj      1/1     Running   0          7h42m   172.16.34.168   gz-1   <none>           <none>
edge-system   statefulset-grid-daemon-w7lpt      1/1     Running   0          7h42m   192.168.10.7    bj-2   <none>           <none>
```

现在，在某个 NodeUnit 内使用统一headless service访问形式，例如访问如下DNS 获取的 IP 地址：

```
{StatefulSet}-{0..N-1}.SVC.default.svc.cluster.local
```

实际就会访问这个 NodeUnit 下具体pod的 FQDN 地址获取的是同一 Pod IP：

```
{StatefulSet}-{NodeUnit}-{0..N-1}.SVC.default.svc.cluster.local
```

例如，在`beijing`地域访问 `statefulsetgrid-demo-0.servicegrid-demo-svc.default.svc.cluster.local`DNS的 IP 地址和 `statefulsetgrid-demo-beijing-0.servicegrid-demo-svc.default.svc.cluster.local`域名返回的 IP 地址是一样的，如下图

```shell
# 在 guangzhou 地域执行 nslookup 指令
[~]# nslookup statefulsetgrid-demo-0.servicegrid-demo-svc.default.svc.cluster.local
Server:         169.254.20.11
Address:        169.254.20.11#53

Name:   statefulsetgrid-demo-0.servicegrid-demo-svc.default.svc.cluster.local
Address: 10.0.0.136

[~]# nslookup statefulsetgrid-demo-guangzhou-0.servicegrid-demo-svc.default.svc.cluster.local
Server:         169.254.20.11
Address:        169.254.20.11#53

Name:   statefulsetgrid-demo-guangzhou-0.servicegrid-demo-svc.default.svc.cluster.local
Address: 10.0.0.136

# 在 beijing 地域执行 nslookup 指令
[~]# nslookup statefulsetgrid-demo-0.servicegrid-demo-svc.default.svc.cluster.local
Server:         169.254.20.11
Address:        169.254.20.11#53

Name:   statefulsetgrid-demo-0.servicegrid-demo-svc.default.svc.cluster.local
Address: 10.0.0.67

[~]# nslookup statefulsetgrid-demo-beijing-0.servicegrid-demo-svc.default.svc.cluster.local
Server:         169.254.20.11
Address:        169.254.20.11#53

Name:   statefulsetgrid-demo-beijing-0.servicegrid-demo-svc.default.svc.cluster.local
Address: 10.0.0.67
```

每个NodeUnit通过相同的headless service只会访问本 NodeUnit 内的pod

```bash
# 在 guangzhou 区域执行下面的命令
[~]# curl statefulsetgrid-demo-0.servicegrid-demo-svc.default.svc.cluster.local:8080 | grep "pod name:"
        pod name:       statefulsetgrid-demo-guangzhou-0
[~]# curl statefulsetgrid-demo-1.servicegrid-demo-svc.default.svc.cluster.local:8080 | grep "pod name:"
        pod name:       statefulsetgrid-demo-guangzhou-1
[~]# curl statefulsetgrid-demo-2.servicegrid-demo-svc.default.svc.cluster.local:8080 | grep "pod name:"
        pod name:       statefulsetgrid-demo-guangzhou-2

# 在 beijing 区域执行下面的命令
[~]# curl statefulsetgrid-demo-0.servicegrid-demo-svc.default.svc.cluster.local:8080 | grep "pod name:"
        pod name:       statefulsetgrid-demo-beijing-0
[~]# curl statefulsetgrid-demo-1.servicegrid-demo-svc.default.svc.cluster.local:8080 | grep "pod name:"
        pod name:       statefulsetgrid-demo-beijing-1
[~]# curl statefulsetgrid-demo-2.servicegrid-demo-svc.default.svc.cluster.local:8080 | grep "pod name:"
        pod name:       statefulsetgrid-demo-beijing-2
```

#### 实现原理

<div align="left">
  <img src="../img/demo_statefulsetgrid_internal.jpg" width=100% title="statefulsetgrid">
</div>

上图描述了 StatefulsetGrid+Headless Service 的实现原理，主要就是在边缘节点侧部署了`statefulset-grid-daemon`的组件，会监听`StatefulsetGrid`的资源信息；同时刷新边缘侧 CoreDNS 的相关记录，根据所在 NodeUnit 地域，添加`{StatefulSet}-{0..N-1}.SVC.default.svc.cluster.local`域名记录，和标准的 Pod FQDN记录 `{StatefulSet}-{NodeUnit}-{0..N-1}.SVC.default.svc.cluster.local`指向同一 Pod 的 IP 地址。具体如何实现 CoreDNS 域名更新可以参考源代码实现。

> **根据上面的描述，读者应该可以清晰分析清楚DeploymentGrid/StatefulsetGrid 配合 ServiceGrid/Headless Service，在各种搭配使用的场景下具体细节的能力了。**



## 按NodeUnit灰度

DeploymentGrid和StatefulSetGrid均支持按照NodeUnit进行灰度

### 重要字段
和灰度功能相关的字段有这些：

autoDeleteUnusedTemplate，templatePool，templates，defaultTemplateName

templatePool：用于灰度的template集合

templates：NodeUnit和其使用的templatePool中的template的映射关系，如果没有指定，NodeUnit使用defaultTemplateName指定的template

defaultTemplateName：默认使用的template，如果不填写或者使用"default"就采用spec.template

autoDeleteUnusedTemplate：默认为false，如果设置为true，会自动删除templatePool中既不在templates中也不在spec.template中的template模板

### 使用相同的template创建workload
和上面的DeploymentGrid和StatefulsetGrid例子完全一致，如果不需要使用灰度功能，则无需添加额外字段

### 使用不同的template创建workload
```yaml
apiVersion: superedge.io/v1
kind: DeploymentGrid
metadata:
  name: deploymentgrid-demo
  namespace: default
spec:
  defaultTemplateName: test1
  gridUniqKey: zone
  template:
    replicas: 1
    selector:
      matchLabels:
        appGrid: echo
    strategy: {}
    template:
      metadata:
        creationTimestamp: null
        labels:
          appGrid: echo
      spec:
        containers:
        - image: superedge/echoserver:2.2
          name: echo
          ports:
          - containerPort: 8080
            protocol: TCP
          env:
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: POD_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.podIP
          resources: {}
  templatePool:
    test1:
      replicas: 2
      selector:
        matchLabels:
          appGrid: echo
      strategy: {}
      template:
        metadata:
          creationTimestamp: null
          labels:
            appGrid: echo
        spec:
          containers:
          - image: superedge/echoserver:2.2
            name: echo
            ports:
            - containerPort: 8080
              protocol: TCP
            env:
              - name: NODE_NAME
                valueFrom:
                  fieldRef:
                    fieldPath: spec.nodeName
              - name: POD_NAME
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.name
              - name: POD_NAMESPACE
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.namespace
              - name: POD_IP
                valueFrom:
                  fieldRef:
                    fieldPath: status.podIP
            resources: {}
    test2:
      replicas: 3
      selector:
        matchLabels:
          appGrid: echo
      strategy: {}
      template:
        metadata:
          creationTimestamp: null
          labels:
            appGrid: echo
        spec:
          containers:
          - image: superedge/echoserver:2.3
            name: echo
            ports:
            - containerPort: 8080
              protocol: TCP
            env:
              - name: NODE_NAME
                valueFrom:
                  fieldRef:
                    fieldPath: spec.nodeName
              - name: POD_NAME
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.name
              - name: POD_NAMESPACE
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.namespace
              - name: POD_IP
                valueFrom:
                  fieldRef:
                    fieldPath: status.podIP
            resources: {}
  templates:
    zone1: test1
    zone2: test2
```
这个例子中，NodeUnit zone1将会使用test1 template，NodeUnit zone2将会使用test2 template，其余NodeUnit将会使用defaultTemplateName中指定的template，这里
会使用test1

## 多集群分发
支持DeploymentGrid和ServiceGrid的多集群分发，分发的同时也支持多地域灰度，当前基于的多集群管理方案为[clusternet](https://github.com/clusternet/clusternet)

### 特点
- 支持多集群的按NodeUnit灰度
- 保证控制集群和被纳管集群应用的强一致和同步更新/删除，做到一次操作，多集群部署
- 在控制集群可以看到聚合的各分发实例的状态
- 支持节点地域信息更新情况下应用的补充分发：如原先不属于某个NodeGroup的集群，更新节点信息后加入了NodeGroup，控制集群中的应用会及时向该集群补充下发

### 前置条件
- 集群部署了SuperEdge中的组件，如果没有Kubernetes集群，可以通过edgeadm进行创建，如果已有Kubernetes集群，可以通过edageadm的addon部署SuperEdge相关组件，将集群转换为一个SuperEdge边缘集群
- 通过clusternet进行集群的注册和纳管

### 重要字段
如果要指定某个DeploymentGrid或ServiceGrid需要进行多集群的分发，则在其label中添加`superedge.io/fed`，并置为"yes"

### 使用示例
创建3个集群，分别为一个管控集群和2个被纳管的边缘集群A,B，通过clusternet进行注册和纳管

其中A集群中一个节点添加zone: zone1的label，加入NodeUnit zone1；集群B不加入NodeGroup

在管控集群中创建DeploymentGrid，其中labels中添加了superedge.io/fed: "yes"，表示该DeploymentGrid需要进行集群的分发，同时灰度指定分发出去的应用在zone1和zone2中使用不同的副本个数
```yaml
apiVersion: superedge.io/v1
kind: DeploymentGrid
metadata:
  name: deploymentgrid-demo
  namespace: default
  labels:
    superedge.io/fed: "yes"
spec:
  defaultTemplateName: test1
  gridUniqKey: zone
  template:
    replicas: 1
    selector:
      matchLabels:
        appGrid: echo
    strategy: {}
    template:
      metadata:
        creationTimestamp: null
        labels:
          appGrid: echo
      spec:
        containers:
        - image: superedge/echoserver:2.2
          name: echo
          ports:
          - containerPort: 8080
            protocol: TCP
          env:
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: POD_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.podIP
          resources: {}
  templatePool:
    test1:
      replicas: 2
      selector:
        matchLabels:
          appGrid: echo
      strategy: {}
      template:
        metadata:
          creationTimestamp: null
          labels:
            appGrid: echo
        spec:
          containers:
          - image: superedge/echoserver:2.2
            name: echo
            ports:
            - containerPort: 8080
              protocol: TCP
            env:
              - name: NODE_NAME
                valueFrom:
                  fieldRef:
                    fieldPath: spec.nodeName
              - name: POD_NAME
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.name
              - name: POD_NAMESPACE
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.namespace
              - name: POD_IP
                valueFrom:
                  fieldRef:
                    fieldPath: status.podIP
            resources: {}
    test2:
      replicas: 3
      selector:
        matchLabels:
          appGrid: echo
      strategy: {}
      template:
        metadata:
          creationTimestamp: null
          labels:
            appGrid: echo
        spec:
          containers:
          - image: superedge/echoserver:2.2
            name: echo
            ports:
            - containerPort: 8080
              protocol: TCP
            env:
              - name: NODE_NAME
                valueFrom:
                  fieldRef:
                    fieldPath: spec.nodeName
              - name: POD_NAME
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.name
              - name: POD_NAMESPACE
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.namespace
              - name: POD_IP
                valueFrom:
                  fieldRef:
                    fieldPath: status.podIP
            resources: {}
  templates:
    zone1: test1
    zone2: test2
```

创建完成后，可以看到在纳管的A集群中，创建了对应的Deployment，而且依照其NodeUnit信息，有两个实例。
```bash
[root@VM-0-174-centos ~]# kubectl get deploy
NAME                        READY   UP-TO-DATE   AVAILABLE   AGE
deploymentgrid-demo-zone1   2/2     2            2           99s
```
如果在纳管的A集群中手动更改了deployment的相应字段，会以管控集群的为模板更新回来

B集群中的一个节点添加zone: zone2的label，将其加入NodeUnit zone2;管控集群会及时向该集群补充下发zone2对应的应用
```bash
[root@VM-0-42-centos ~]# kubectl get deploy
NAME                        READY   UP-TO-DATE   AVAILABLE   AGE
deploymentgrid-demo-zone2   3/3     3            3           6s
```

在管控集群查看deploymentgrid-demo的状态，可以看到被聚合在一起的各个被纳管集群的应用状态，便于查看
```yaml
status:
  states:
    zone1:
      conditions:
      - lastTransitionTime: "2021-06-17T07:33:50Z"
        lastUpdateTime: "2021-06-17T07:33:50Z"
        message: Deployment has minimum availability.
        reason: MinimumReplicasAvailable
        status: "True"
        type: Available
      readyReplicas: 2
      replicas: 2
    zone2:
      conditions:
      - lastTransitionTime: "2021-06-17T07:37:12Z"
        lastUpdateTime: "2021-06-17T07:37:12Z"
        message: Deployment has minimum availability.
        reason: MinimumReplicasAvailable
        status: "True"
        type: Available
      readyReplicas: 3
      replicas: 3
```

## Refs

* [SEP: ServiceGroup StatefulSetGrid Design Specification](https://github.com/superedge/superedge/issues/26)
