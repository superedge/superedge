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

serviceGroup就是为这种场景设计的，客户只需要使用ServiceGroup提供的DeploymentGrid和ServiceGrid两种superedge自研的kubernetes 资源，即可方便地将服务分别部署到这些节点组中，并进行服务流量管控，另外，还能保证各区域服务数量及容灾。


# 关键概念
## 整体架构
<div align="left">
  <img src="../img/serviceGroup-UseCase.png" width=70% title="service-group">
</div>

## NodeUnit
- NodeUnit通常是位于同一边缘站点内的一个或多个计算资源实例，需要保证同一NodeUnit中的节点内网是通的
- ServiceGroup组中的服务运行在一个NodeUnit之内
- ServiceGroup 允许用户设置服务在一个 NodeUnit中运行的pod数量
- ServiceGroup 能够把服务之间的调用限制在本 NodeUnit 内

## NodeGroup
- NodeGroup 包含一个或者多个 NodeUnit
- 保证在集合中每个 NodeUnit上均部署ServiceGroup中的服务
- 集群中增加 NodeUnit 时自动将 ServiceGroup 中的服务部署到新增 NodeUnit

## ServiceGroup
- ServiceGroup 包含一个或者多个业务服务:适用场景：1）业务需要打包部署；2）或者，需要在每一个 NodeUnit 中均运行起来并且保证pod数量；3）或者，需要将服务之间的调用控制在同一个 NodeUnit 中，不能将流量转发到其他 NodeUnit。
- 注意：ServiceGroup是一种抽象资源，一个集群中可以创建多个ServiceGroup

## 涉及的资源类型
### DepolymentGrid
DeploymentGrid的格式与Deployment类似，<deployment-template>字段就是原先deployment的template字段，比较特殊的是gridUniqKey字段，该字段指明了节点分组的label的key值；
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

### ServiceGrid
ServiceGrid的格式与Service类似，<service-template>字段就是原先service的template字段，比较特殊的是gridUniqKey字段，该字段指明了节点分组的label的key值；
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
以在边缘部署nginx为例，我们希望在多个节点组内分别部署nginx服务，需要做如下事情：

## 确定ServiceGroup唯一标识
这一步是逻辑规划，不需要做任何实际操作。我们将目前要创建的serviceGroup逻辑标记使用的 UniqKey为：zone。

## 将边缘节点分组
这一步需要使用kubectl对边缘节点打 label

例如，我们选定 Node12、Node14，打上label，zone=nodeunit1；Node21、Node23 打上label，zone=nodeunit2。

注意：上一步中 label的key与ServiceGroup 的UniqKey一致，value是NodeUnit的唯一key，value相同的节点表示属于同一个NodeUnit

如果同一个集群中有多个ServiceGroup请为每一个ServiceGroup分配不同的Uniqkey

## 部署deploymentGrid
```yaml
apiVersion: superedge.io/v1
kind: DeploymentGrid
metadata:
  name: deploymentgrid-demo
  namespace: default
spec:
  gridUniqKey: zone
  template:
    selector:
      matchLabels:
        appGrid: nginx
    replicas: 2
    template:
      metadata:
        labels:
          appGrid: nginx
      spec:
        containers:
        - name: nginx
          image: nginx:1.7.9
          ports:
          - containerPort: 80
            protocol: TCP
```

### 部署serviceGrid
```yaml
apiVersion: superedge.io/v1
kind: ServiceGrid
metadata:
  name: servicegrid-demo
  namespace: default
spec:
  gridUniqKey: zone
  template:
    selector:
      appGrid: nginx
    ports:
    - protocol: TCP
      port: 80
      targetPort: 80
```
gridUniqKey字段设置为了zone，所以我们在将节点分组时采用的label的key为zone，如果有三组节点，分别为他们添加zone: zone-0, zone: zone-1 ,zone: zone-2的label即可；这时，每组节点内都有了nginx的deployment和对应的pod，在节点内访问统一的service-name也只会将请求发向本组的节点。

```
[~]# kubectl get deploy
NAME                         READY   UP-TO-DATE   AVAILABLE   AGE
deploymentgrid-demo-zone-0   2/2     2            2           85s
deploymentgrid-demo-zone-1   2/2     2            2           85s
deploymentgrid-demo-zone-2   2/2     2            2           85s

[~]# kubectl get svc
NAME                   TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
kubernetes             ClusterIP   172.19.0.1     <none>        443/TCP   87m
servicegrid-demo-svc   ClusterIP   172.19.0.177   <none>        80/TCP    80s
```

另外，对于部署了DeploymentGrid和ServiceGrid后才添加进集群的节点组，该功能会在新的节点组内自动创建指定的deployment和service。