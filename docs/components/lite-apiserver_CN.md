lite-apiserver是运行在边缘节点上的轻量级apiserver，它代理节点上所有组件和Pod访问云端apiserver的请求，并对请求结果做高效缓存。在云边断连的情况下，利用这些缓存提供服务，实现边缘自治的能力。

## 架构图
<div align="left">
  <img src="../img/lite-apiserver.png" width=70% title="lite-apiserver Architecture">
</div>

lite-apiserver具有以下特点:
- 代理所有组件和节点上的pod到apiserver的请求
- 支持访问apisever的证书、token等所有认证和鉴权方式，各组件和pod使用自己的认证方式和权限
- 可处理和缓存所有类型的资源，包括kubernetes内置资源和CRD

## 使用说明
可使用static pod或者systemd在边缘节点上部署lite-apiserver，参见edgeadm或手动部署文档

### 启动参数
#### Server相关参数
- ca-file: 集群或apiserver使用的ca
- tls-cert-file: ca-file签署的lite-apiserver的tls证书
- tls-private-key-file: lite-apiserver的tls证书的私钥
- kube-apiserver-url: 云端apiserver的host
- kube-apiserver-port: 云端apiserver的port
- port: lite-apiserver监听的https端口，默认为51003
- file-cache-path: 本地文件缓存模式下数据保存目录
- timeout: 后端请求 kube-apiserver 默认的超时时间
- sync-duration: 后端数据默认刷新同步时间
#### 证书管理相关参数
- tls-config-file: tls证书配置文件目录，默认为空
