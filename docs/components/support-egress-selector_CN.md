# 1. 生成证书

## 1.1 生成client证书

```shell
 openssl genrsa -out tunnel-client.key 2048
```

```shell
 openssl req -new  -key tunnel-client.key -subj "/CN=tunnel-client" -out tunnel-client.csr
```

```shell
 openssl x509 -req -CA /etc/kubernetes/pki/ca.crt -CAkey /etc/kubernetes/pki/ca.key -CAcreateserial -in tunnel-client.csr -out tunnel-client.crt
```

将生成的**tunnel-client.crt**和**tunnel-client.key**拷贝到**/etc/kubernetes/pki**

```shell
cp tunnel-client.crt /etc/kubernetes/pki/
cp tunnel-client.key /etc/kubernetes/pki/
```

## 1.2 生成server证书

```shell
 openssl genrsa -out tunnel-server.key 2048
```

```shell
 openssl req -new  -key tunnel-server.key -subj "/CN=tunnel-server" -out tunnel-server.csr
```

```shell
echo "subjectAltName=DNS:tunnel-cloud.edge-system.svc.cluster.local" > tunnel_server_cert_extensions
openssl x509 -req -CA /etc/kubernetes/pki/ca.crt -CAkey /etc/kubernetes/pki/ca.key -CAcreateserial -in tunnel-server.csr -extfile tunnel_server_cert_extensions -out tunnel-server.crt
```

# 2. 部署tunnel-cloud

## 2.1 添加tunnel-server的证书

将生成**tunnel-server.crt**和**tunnel-server.key**的**base64**编码添加到**edge-system/tunnel-cloud-cert**的secret中

```yaml
apiVersion: v1
data:
  tunnel-server.crt: {{tunnel-server.crt}}
  tunnel-client.key: {{tunnel-client.key}}
kind: Secret
metadata:
  name: tunnel-cloud-cert
  namespace: edge-system
type: Opaque
```

## 2.2 添加tunnel-cloud pod的list、watch权限

### 2.2.1 clusterrole

**edge-system/tunnel-cloud**

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: tunnel-cloud
rules:
  - apiGroups: [ "" ]
    resources: [ "configmaps" ]
    verbs: [ "get", "update" ]
  - apiGroups: [ "" ]
    resources: [ "endpoints" ]
    verbs: [ "get" ]
  - apiGroups: [ "" ]
    resources: [ "services" ]
    verbs: [ "get" ]
  - apiGroups: [ "" ]
    resources: [ "pods","nodes" ]
    verbs: [ "get","list","watch" ]
 ```

### 2.2.2 使用ClusterRoleBinding替换Rolebinding

将**edge-system/tunnel-cloud**的**rolebinding**删除,创建**ClusterRoleBinding**

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: tunnel-cloud
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: tunnel-cloud
subjects:
  - kind: ServiceAccount
    name: tunnel-cloud
    namespace: edge-system
```

## 2.3 修改tunnel-cloud的配置文件

**edge-system/tunnel-cloud-conf**

```yaml
apiVersion: v1
data:
  mode.toml: |
    [mode]
        [mode.cloud]
            [mode.cloud.stream]
                [mode.cloud.stream.server]
                    grpcport = 9000
                    logport = 51010
                    metricsport = 6000
                    key = "/etc/superedge/tunnel/certs/tunnel-cloud-server.key"
                    cert = "/etc/superedge/tunnel/certs/tunnel-cloud-server.crt"
                    tokenfile = "/etc/superedge/tunnel/token/token"
                [mode.cloud.stream.dns]
                     configmap="tunnel-nodes"
                     hosts = "/etc/superedge/tunnel/nodes/hosts"
                     service = "tunnel-cloud"
            [mode.cloud.tcp]
                "0.0.0.0:6443" = "127.0.0.1:6443"
            [mode.cloud.https]
                cert ="/etc/superedge/tunnel/certs/apiserver-kubelet-server.crt"
                key = "/etc/superedge/tunnel/certs/apiserver-kubelet-server.key"
                [mode.cloud.https.addr]
                    "10250" = "127.0.0.1:10250"
                    "9100" = "127.0.0.1:9100"
            [mode.cloud.egress]
              servercert="/etc/superedge/tunnel/certs/tunnel-server.crt"
              serverkey="/etc/superedge/tunnel/certs/tunnel-server.key"
              egressport="8000"
kind: ConfigMap
metadata:
  name: tunnel-cloud-conf
  namespace: edge-system
```

## 2.4 在tunnel-cloud的SVC中添加8000端口

```yaml
  - name: egress
    port: 8000
    protocol: TCP
    targetPort: 8000
```

## 2.6 校验egress server

```shell
openssl s_client -cert tunnel-client.crt  -key tunnel-client.key -CAfile /etc/kubernetes/pki/ca.crt  -connect <tunnel-cloud clusterIp>:8000
```

返回结果

```
...
Verify return code: 0 (ok)
```

# 3.修改Apiserver pod的yaml文件

## 3.1 修改apiserver的文件挂载

**egress-selector-configuration.yaml**

```yaml
apiVersion: apiserver.k8s.io/v1beta1
kind: EgressSelectorConfiguration
egressSelections:
  - name: cluster
    connection:
      proxyProtocol: HTTPConnect
      transport:
        tcp:
          url: https://tunnel-cloud.edge-system.svc.cluster.local:8000
          tlsConfig:
            caBundle: /etc/kubernetes/pki/ca.crt
            clientCert: /etc/kubernetes/pki/tunnel-client.crt
            clientKey: /etc/kubernetes/pki/tunnel-client.key
```

将**egress-selector-configuration.yaml**拷贝到master节点的 **/etc/kubernetes/conf/**目录

```yaml
    volumeMounts:
      - mountPath: /etc/kubernetes/conf
        name: k8s-conf
        readOnly: true
      ...
    volumes:
      - hostPath:
          path: /etc/kubernetes/conf
          type: DirectoryOrCreate
        name: k8s-conf

```

## 3.2 修改apiserver的启动参数

```yaml
- --enable-aggregator-routing=true
- --egress-selector-config-file=/etc/kubernetes/conf/egress-selector-configuration.yaml
```

## 4. 测试在边缘端部署webhook server

**推荐在集群中部署nginx ingress controller进行测试**

