# 部署SuperEdge - 纯手工方式

- [1. 部署Tunnel](#1-%E9%83%A8%E7%BD%B2tunnel)
  - [1.1 部署Tunnel Coredns](#11-%E9%83%A8%E7%BD%B2tunnel-coredns)
  - [1.2 部署Tunnel Cloud](#12-%E9%83%A8%E7%BD%B2tunnel-cloud)
    - [1.2.1 需要补全的参数](#121-%E9%9C%80%E8%A6%81%E8%A1%A5%E5%85%A8%E7%9A%84%E5%8F%82%E6%95%B0)
    - [1.2.2 TunnelPersistentConnnectionServerKey和TunnelPersistentConnnectionServerCrt的生成举例](#122-tunnelpersistentconnnectionserverkey%E5%92%8Ctunnelpersistentconnnectionservercrt%E7%9A%84%E7%94%9F%E6%88%90%E4%B8%BE%E4%BE%8B)
    - [1.2.3 TunnelProxyServerKey和TunnelProxyServerCrt的生成举例](#123-tunnelproxyserverkey%E5%92%8Ctunnelproxyservercrt%E7%9A%84%E7%94%9F%E6%88%90%E4%B8%BE%E4%BE%8B)
  - [1.3 Kube-apiserver使用Tunnel隧道](#13-kube-apiserver%E4%BD%BF%E7%94%A8tunnel%E9%9A%A7%E9%81%93)
  - [1.4 部署Tunnel Edge](#14-%E9%83%A8%E7%BD%B2tunnel-edge)
- [2. 部署lite-apiserver](#2-%E9%83%A8%E7%BD%B2lite-apiserver)
  - [2.1 部署lite-apiserver](#21-%E9%83%A8%E7%BD%B2lite-apiserver)
  - [2.2 node上组件使用lite-apiserver](#22-node%E4%B8%8A%E7%BB%84%E4%BB%B6%E4%BD%BF%E7%94%A8lite-apiserver)
- [3 部署Application Grid](#3-%E9%83%A8%E7%BD%B2application-grid)
  - [3.1 部署Application Grid Controller](#31-%E9%83%A8%E7%BD%B2application-grid-controller)
  - [3.2 Add Annotate Endpoint Kubernetes](#32-add-annotate-endpoint-kubernetes)
  - [3.3 部署Application Grid Wrapper](#33-%E9%83%A8%E7%BD%B2application-grid-wrapper)
  - [3.4 Kube-proxy使用Application Grid Wrapper](#34-kube-proxy%E4%BD%BF%E7%94%A8application-grid-wrapper)
- [4 部署Edge Health](#4-%E9%83%A8%E7%BD%B2edge-health)
  - [4.1 部署edge-health-admission](#41-%E9%83%A8%E7%BD%B2edge-health-admission)
  - [4.2 部署Edge Health](#42-%E9%83%A8%E7%BD%B2edge-health)


## 1. 部署Tunnel
### 1.1 部署Tunnel Coredns
使用Deployment方式，将tunnel-coredns部署在*云端control plane节点*中
```bash
$ kubectl apply -f deployment/tunnel-coredns.yaml
```

### 1.2 部署Tunnel Cloud

#### 1.2.1 需要补全的参数

-   TunnelCloudEdgeToken：tunnel-cloud和tunnel-edge的认证token，至少随机32位字符串；

-   TunnelPersistentConnectionServerKey: tunnel-cloud service端证书key的base64加密, 用于tunnel-cloud和tunnel-edge之间的认证

-   TunnelPersistentConnectionServerCrt: tunnel-cloud service端证书crt的base64加密，可用openssl等工具生成，注意签tunnel-cloud的service name: "tunnelcloud.io"；

-   TunnelProxyServerKey: 集群ca签的server端证书key的base64加密；

-   TunnelProxyServerCrt: 集群ca签的server端证书crt的base64加密；

使用Deployment方式，将tunnel-cloud部署在*云端control plane节点*中。

`````bash
$ kubectl apply -f deployment/tunnel-cloud.yaml
`````

#### 1.2.2 TunnelPersistentConnnectionServerKey和TunnelPersistentConnnectionServerCrt的生成举例

-   生成tunnel的CA

    ```bash
    # Generate CA private key 
    openssl genrsa -out tunnel-ca.key 2048
    
    # Generate CSR 
    openssl req -new -key tunnel-ca.key -out tunnel-ca.csr
    
    # Add DNS and IP
    echo "subjectAltName=DNS:superedge.io,IP:127.0.0.1" > tunnel_ca_cert_extensions
    
    # Generate Self Signed certificate
    openssl x509 -req -days 365 -in tunnel-ca.csr -signkey tunnel-ca.key -extfile tunnel_ca_cert_extensions -out tunnel-ca.crt
    ```

-   生成TunnelPersistentConnectionServerKey和TunnelPersistentConnectionServerCrt

    ```bash
    # private key
    openssl genrsa -des3 -out tunnel_persistent_connectiong_server.key 2048
    
    # generate csr
    openssl req -new -key tunnel_persistent_connectiong_server.key -subj "/CN=tunnel-cloud" -out tunnel_persistent_connectiong_server.csr
    
    # Add DNS and IP, 必须填写 "DNS:tunnelcloud.io"
    echo "subjectAltName=DNS:tunnelcloud.io,IP:127.0.0.1" > tunnel_persistent_connectiong_server_cert_extensions
    
    # Generate Self Signed certificate
    openssl x509 -req -days 365 -in tunnel_persistent_connectiong_server.csr -CA tunnel-ca.crt -CAkey tunnel_ca.key -CAcreateserial  -extfile tunnel_persistent_connectiong_server_cert_extensions -out tunnel_persistent_connectiong_server.crt
    ```

-   TunnelPersistentConnectionServerKey和TunnelPersistentConnectionServerCrt的生成

    ```bash
    # generate TunnelPersistentConnectionServerKey
    cat tunnel_persistent_connectiong_server.key | base64 --wrap=0
    #generate TunnelPersistentConnectionServerCrt
    cat tunnel_persistent_connectiong_server.crt | base64 --wrap=0
    ```

#### 1.2.3 TunnelProxyServerKey和TunnelProxyServerCrt的生成举例

生成TunnelProxyServerKey和TunnelProxyServerCrt（用于kube-apiserver和tunnel-cloud之间的认证）

```bash
# private key
openssl genrsa -des3 -out tunnel_proxy_server.key 2048

# generate csr
openssl req -new -key tunnel_proxy_server.key -subj "/CN=tunnel-cloud" -out tunnel_proxy_server.csr

# Add DNS and IP
echo "subjectAltName=DNS:superedge.io,IP:127.0.0.1" > cert_extensions

# Generate Self Signed certificate（注意ca.crt和ca.key为集群的证书, Kubeadm部署的集群中，CA是/etc/kubernetes/pki下的ca.crt和ca.key）
openssl x509 -req -days 365 -in tunnel_proxy_server.csr -CA ca.crt -CAkey ca.key -CAcreateserial  -extfile cert_extensions -out tunnel_proxy_server.crt
```

Base64加密同TunnelPersistentConnectionServerKey和TunnelPersistentConnectionServerCrt

### 1.3 Kube-apiserver使用Tunnel隧道

将kube-apiserver的DNS解析指向tunnel-coredns，通过dns劫持，将kube-apiserver发送给边缘节点的请求，通过tunnel隧道代为请求。（边缘场景下，Cloud端无法直接访问Edge端。）

```bash
#获取tunnel-coredns的Cluster IP
$ kubectl get service tunnel-coredns -n edge-system
NAME             TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)                  AGE
tunnel-coredns   ClusterIP   10.10.47.74   <none>        53/UDP,53/TCP,9153/TCP   140m
#修改kube-apierver的DNS，使用tunnel-coredns
...
dnsConfig:
    nameservers:
    - 10.10.47.74 #修改为tunnel-cloud的CLUSTER-IP；  
...
```

注意：通过DNS劫持进行请求重定向，边缘节点Name不能为IP，因为IP不经过DNS解析。

### 1.4 部署Tunnel Edge

要填充的参数：

-   MasterIP：kube-api-server的master节点内网IP，填一个就可；       

-   TunnelCloudEdgeToken：tunnel-cloud和tunnel-edge的认证token；

    >  至少随机32位字符串，tunnel-cloud和tunnel-edge必须为同一个，需要完全相同；

-   TunnelPersistentConnectionPort： tunnel-cloud的NodePort端口；

-   KubernetesCaCert：kube-apiserver的ca.crt的base64加密；
    
> 用于验证tunnel cloud的server端证书

-   KubeletClientKey：集群ca签的client端证书key的base64加密;

-   KubeletClientCrt：集群ca签的client端证书crt的base64加密;


使用DaemonSet方式，将tunnel-edge部署在*边缘Node节点*中
```bash
$ kubectl apply -f deployment/tunnel-edge.yaml
```

KubeletClientKey和KubeletClientCrt的生成举例:

```bash
# private key
openssl genrsa -des3 -out kubelet_client.key 1024
# generate csr
openssl req -new -key kubelet_client.key -out kubelet_client.csr
# Generate Self Signed certificate（注意ca.crt和ca.key为集群的证书, Kubeadm部署的集群中，CA是/etc/kubernetes/pki下的ca.crt和ca.key）
openssl ca -in kubelet_client.csr -out kubelet_client.crt -cert ca.crt -keyfile ca.key
```

Base64加密KubeletClientKey和KubeletClientCrt, 方式同TunnelPersistentConnectionServerKey和TunnelPersistentConnectionServerCrt

## 2. 部署lite-apiserver

### 2.1 部署lite-apiserver

使用集群CA（Kubeadm部署的集群中，CA是/etc/kubernetes/pki下的ca.crt和ca.key）生成lite-apiserver的https tls证书（lite-apiserver.crt和lite-apiserver.key）。
```bash
#获取service 'kubernetes'的ClusterIP
$ kubectl get service kubernetes
NAME         TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
kubernetes   ClusterIP   10.10.0.1    <none>        443/TCP   23d

#生成lite-apiserver.key
$ openssl genrsa -out lite-apiserver.key 2048

#创建lite-apiserver.csr
$ cat << EOF >lite-apiserver.conf
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
[req_distinguished_name]
CN = lite-apiserver
[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names
[alt_names]
DNS.1 = localhost
IP.1 = 127.0.0.1
IP.2 = 10.10.0.1 # 请改成对应kubernetes的ClusterIP
EOF

$ openssl req -new -key lite-apiserver.key -subj "/CN=lite-apiserver" -config lite-apiserver.conf -out lite-apiserver.csr

#生成lite-apiserver.crt
openssl x509 -req -in lite-apiserver.csr -CA ca.crt -CAkey ca.key -CAcreateserial -days 5000 -extensions v3_req -extfile lite-apiserver.conf -out lite-apiserver.crt
```
* 分发lite-apiserver.crt和lite-apiserver.key到边缘节点的/etc/kubernetes/edge/下；

* 修改deployment/lite-apiserver.yaml中的--kube-apiserver-url和--kube-apiserver-port指向apiserver的host和port；

* 配置--tls-config-file=/etc/kubernetes/edge/tls.json， 并在边缘节点上创建/etc/kubernetes/edge/tls.json文件，写入如下内容：

    ```json
    [
        {
            "key":"/var/lib/kubelet/pki/kubelet-client-current.pem", #内容由kubelet生成，只用引用机器上面的绝对地址便可
            "cert":"/var/lib/kubelet/pki/kubelet-client-current.pem" #因为*-key.pem和*-crt.pem在同一个文件，所以引用了同一个文件
        }
    ]
    ```

    kubelet-client-current-key.pem的内容： kubelet访问kube-apiserver的key；

    kubelet-client-current-cert.pem的内容：kubelet访问kube-apiserver的crt；

    > 因为lite-apiserver需要代理kubelet的请求，所以要把kubelet访问kube-apiserver的证书配置给lite-apiserver，让lite-apiserver代kubelet访问kube-apiserver。

* 使用Static Pod方式将lite-apiserver部署在*边缘Node节点*中, 分发deployment/lite-apiserver.yaml到边缘kubelet的manifests下（kubeadm集群位于/etc/kubernetes/manifests/）。

### 2.2 node上组件使用lite-apiserver

lite-apiserver默认监听51003端口(可在deployment/lite-apiserver.yaml的--port中指定)，可使用 https://127.0.0.1:51003 替换原kube-apiserver地址
* kubelet: 修改kubelet.conf中的cluster.server为 https://127.0.0.1:51003，重启kubelet。

## 3 部署Application Grid

### 3.1 部署Application Grid Controller

使用Deployment方式，将application-grid-controller部署在*云端control plane节点*中
```bash
$ kubectl apply -f deployment/application-grid-controller.yaml
```

### 3.2 Add Annotate Endpoint Kubernetes

```bash
kubectl annotate endpoints kubernetes superedge.io/local-endpoint=127.0.0.1
kubectl annotate endpoints kubernetes superedge.io/local-port=51003
```

> 让kubernestes endpoints通过lite-apiserver访问kube-apiserver

### 3.3 部署Application Grid Wrapper

使用DaemonSet方式，将application-grid-wrapper部署在*边缘Node节点*中
```bash
$ kubectl apply -f deployment/application-grid-wrapper.yaml
```
Application-grid-wrapper通过lite-apiserver请求kube-apiserver。

### 3.4 Kube-proxy使用Application Grid Wrapper

修改kube-proxy配置文件的cluster.server为 http://127.0.0.1:51006 （kube-proxy的配置文件位于kube-system namespace下的 kube-proxy的configMap中）

> 51006为application-grid-wrapper的默认监听的端口，application-grid-wrappe通过影响kube-proxy的endpoint筛选，控制service的请求只在一个Unit内。

## 4 部署Edge Health

### 4.1 部署edge-health-admission

使用Deployment方式，将edge-health-admission部署在*云端control plane节点*中
```bash
$ kubectl apply -f deployment/edge-health-admission.yaml
```

使用Deployment方式，将edge-health-webhook部署在*云端control plane节点*中
```bash
$ kubectl apply -f deployment/edge-health-webhook.yaml
```

> 目前webhook中的证书是预先生成的，用户可以替换成自己生成的证书。
>
> `deployment/edge-health-webhook.yaml`中的`caBundle`填写CA证书。
>
> `deployment/edge-health-admission.yaml`中`validate-admission-control-server-certs Secret`的`server.crt`和`server.key`分别填写CA颁发的证书和私钥。

### 4.2 部署Edge Health

需要填充的参数：

-   HmacKey：分布式健康检查，edge-health的消息验证key，至少16位随机字符串；

使用DaemonSet方式，将edge-health部署在*边缘Node节点*中
```bash
$ kubectl apply -f deployment/edge-health.yaml
```
