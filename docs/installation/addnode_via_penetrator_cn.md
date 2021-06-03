# 使用penetrator添加边缘节点

## 部署penetrator

直接使用 [penetrator.yaml](../../deployment/penetrator.yaml)部署

## 安装节点

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
* sshCredential：存储SSH登录待添加的节点的密码(passwd)和私钥的secret(sshkey)，密码文件的key值必须为passwd，私钥文件的key值必须为sshkey
* proxyNode: 执行添加节点的job的集群内的节点的节点名，该节点和待安装的节点处于同一个内网（能够SSH登录待安装的节点）

使用SSH的密码文件passwd创建sshCredential

```shell
kubectl -n edge-system create secret generic login-secret --from-file=passwd=./passwd 
```

使用SSH的私钥文件sshkey创建sshCredential

```shell
kubectl -n edge-system create secret generic login-secret --from-file=sshkey=./sshkey 
```

## 重装节点

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


