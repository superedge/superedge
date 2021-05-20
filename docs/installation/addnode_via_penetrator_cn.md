# 使用penetrator添加边缘节点

## 安装节点

### 通过master节点安装节点

需要保证master节点和待添加的节点在同一局域网内

```yaml
apiVersion: nodetask.apps.superedge.io/v1beta1
kind: NodeTask
metadata:
  name: nodes
spec:
  prefixName: "edge"
  secretName: login-secret
  ips:
    - 172.21.3.77
    ...
```

* prefixName: 节点名前缀，节点名的格式: prefixName-随机字符串(6位)
* secretName：存储SSH登录待添加的节点的密码(passwd)和私钥的secret(sshkey),密码文件的key值必须为passwd，私钥文件的key值必须为sshkey
* ips: 待安装的节点的ip列表
```shell
kubectl -n edge create secret generic login-secret --from-file=passwd=./passwd 
```

### 通过work节点安装节点

需要保证work节点和待添加的节点在同一局域网内

```yaml
apiVersion: nodetask.apps.superedge.io/v1beta1
kind: NodeTask
metadata:
  name: nodes
spec:
  prefixName: "edge"
  secretName: login-secret
  nodeName: edge-g5ib9e
  ips:
    - 172.21.1.193
    ...
```

* nodeName: work节点的节点名

## 重装节点

### 通过master节点重装节点

```yaml
apiVersion: nodetask.apps.superedge.io/v1beta1
kind: NodeTask
metadata:
  name: nodes
spec:
  prefixName: "edge"
  secretName: login-secret
  nameIps:
    edge-onq0bw: 172.21.3.77
    ...
```

* nameIps: 重装节点的节点名和IP的map

### 通过work节点重装节点

```yaml
apiVersion: nodetask.apps.superedge.io/v1beta1
kind: NodeTask
metadata:
  name: nodes
spec:
  prefixName: "edge"
  secretName: login-secret
  nodeName: edge-onq0bw
  nameIps:
    edge-6gu07h: 172.21.1.193
    ...
```