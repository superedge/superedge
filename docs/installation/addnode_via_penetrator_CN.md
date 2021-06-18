# 使用Penetrator通过云端添加边缘节点

## 1. 用edgeadm搭建SuperEdge Kubernetes边缘集群

如何搭建: [用edgeadm一键安装边缘独立Kubernetes 集群](../../README_CN.md)

## 2. 部署Penetrator

直接使用 [penetrator.yaml](../../deployment/penetrator.yaml)部署

```shell
kubectl apply -f https://raw.githubusercontent.com/superedge/superedge/main/deployment/penetrator.yaml
```

## 3. 操作节点的前置条件

使用SSH的密码文件passwd创建sshCredential

```shell
kubectl -n edge-system create secret generic login-secret --from-file=passwd=./passwd 
```

或者使用SSH的私钥文件sshkey创建sshCredential

```shell
kubectl -n edge-system create secret generic login-secret --from-file=sshkey=./sshkey 
```

## 4.1 安装节点

```yaml
apiVersion: nodetask.apps.superedge.io/v1beta1
kind: NodeTask
metadata:
  name: nodes
spec:
  nodeNamePrefix: "edge"
  targetMachines:
    - 172.21.3.194
    ...
  sshCredential: login-secret
  proxyNode: vm-2-117-centos
```

* nodeNamePrefix: 节点名前缀，节点名的格式: nodeNamePrefix-随机字符串(6位)
* targetMachines: 待安装的节点的ip列表
* sshCredential：存储SSH登录待添加的节点的密码(passwd)和私钥(sshkey)的secret，密码文件的key值必须为passwd，私钥文件的key值必须为sshkey
* proxyNode: 执行添加节点的job的集群内的节点的节点名，该节点和待安装的节点处于同一个内网（能够SSH登录待安装的节点）

## 4.2 重装节点

```yaml
apiVersion: nodetask.apps.superedge.io/v1beta1
kind: NodeTask
metadata:
  name: nodes
spec:
  nodeNamesOverride:
    edge-1mokvl: 172.21.3.194
    ...
  sshCredential: login-secret
  proxyNode: vm-2-117-centos
```

* nodeNamesOverride: 重装节点的节点名和IP

## 5. 状态查询

NodeTask的Status中包含任务的执行状态(creating和ready)和未安装完成节点的节点名和IP，可以使用命令查看：

```shell
kubectl get nt NodeTaskName -o custom-columns='STATUS:status.nodetaskStatus' 
```

任务在执行过程的成功和错误信息以事件的形式上报到apiserver，可以使用命令查看：

```shell
kubectl -n edge-system get event
```

