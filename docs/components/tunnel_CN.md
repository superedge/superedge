
# tunnel

tunnel是云边端通信的隧道，分为**tunnel-cloud**和**tunnel-edge**，分别承担云边隧道的两端

## 作用
- 代理云端访问边缘节点组件的请求，解决云端无法直接访问边缘节点的问题（边缘节点没有暴露在公网中）

## 架构图
<div align="left">
  <img src="../img/tunnel.png" width=70% title="tunnel Architecture">
</div>

## 实现方案
### 节点注册
   - 边缘节点上**tunnel-edge**主动连接云端**tunnel-cloud** service，service根据负载均衡策略将请求转到**tunnel-cloud**的pod。
   - **tunnel-edge**与**tunnel-cloud**建立gRPC连接后，**tunnel-cloud**会把自身的podIp和**tunnel-edge**所在节点的nodeName的映射写入DNS。gRPC连接断开之后，**tunnel-cloud**会删除podIp和节点名映射。

### 云端请求的转发
   - apiserver或者其它云端的应用访问边缘节点上的kubelet或者其它应用时，tunnel-dns通过DNS劫持(将host中的节点名解析为**tunnel-cloud**的podIp)把请求转发到**tunnel-cloud**的pod。
   - **tunnel-cloud**根据节点名把请求信息转发到节点名对应的与**tunnel-edge**建立的gRPC连接。
   - **tunnel-edge**根据接收的请求信息请求边缘节点上的应用。

## 配置文件
tunnel组件包括**tunnel-cloud**和**tunnel-edge**，运行在边缘节点**tunnel-edge**与运行在云端的**tunnel-cloud**建立gRPC长连接，用于云端转发到边缘节点的隧道。
### tunnel-cloud
**tunnel-cloud**包含**stream**、**TCP**和HTTPS三个模块。其中**stream模块**包括gRPC server和DNS组件，gRPC server用于接收**tunnel-edge**的gRPC长连接请求，DNS组件
用于把**tunnel-cloud**内存中的节点名和IP的映射更新到coredns hosts插件的configmap中。
### tunnel-cloud配置文件
tunnel-cloud-conf.yaml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tunnel-cloud-conf
  namespace: edge-system
data:
  tunnel_cloud.toml: |
    [mode]
      [mode.cloud]
        [mode.cloud.stream]                             # stream模块
          [mode.cloud.stream.server]                    # gRPC server组件
            grpcport = 9000                             # gRPC server监听的端口
            logport = 8000                              # log和健康检查的http server的监听端口，使用(curl -X PUT http://podip:logport/debug/flags/v -d "8")可以设置日志等级
            channelzaddr = "0.0.0.0:6000"               # gRPC [channlez](https://grpc.io/blog/a-short-introduction-to-channelz/) server的监听地址，用于获取gRPC的调试信息
            key = "../../conf/certs/cloud.key"          # gRPC server的server端私钥
            cert = "../../conf/certs/cloud.crt"         # gRPC server的server端证书
            tokenfile = "../../conf/token"              # token的列表文件(nodename:随机字符串)，用于验证边缘节点tunnel-edge发送的token，如果根据节点名验证没有通过，会用default对应的token去验证
          [mode.cloud.stream.dns]                       # DNS组件
            configmap= "proxy-nodes"                    # coredns hosts插件的配置文件的configmap
            hosts = "/etc/superedge/proxy/nodes/hosts"  # coredns hosts插件的配置文件的configmap在tunnel-cloud pod的挂载文件的路径
            service = "proxy-cloud-public"              # tunnel-cloud的service name
            debug = true                                # DNS组件开关，debug=true DNS组件关闭，**tunnel-cloud** 内存中的节点名映射不会保存到coredns hosts插件的配置文件的configmap，默认值为false
        [mode.cloud.tcp]                                # TCP模块
          "0.0.0.0:6443" = "127.0.0.1:6443"             # 参数的格式是"0.0.0.0:cloudPort": "EdgeServerIp:EdgeServerPort"，cloudPort为tunnel-cloud TCP模块server监听端口，EdgeServerIp和EdgeServerPort为代理转发的边缘节点server的IP和端口
        [mode.cloud.https]                              # HTTPS模块
          cert ="../../conf/certs/kubelet.crt"          # HTTPS模块server端证书
          key = "../../conf/certs/kubelet.key"          # HTTPS模块server端私钥
          [mode.cloud.https.addr]                       # 参数的格式是"httpsServerPort":"EdgeHttpsServerIp:EdgeHttpsServerPort"，httpsServerPort为HTTPS模块server端的监听端口，EdgeHttpsServerIp:EdgeHttpsServerPort为代理转发边缘节点HTTPS server的IP和port，HTTPS模块的server是跳过验证client端证书的，因此可以使用(curl -k https://podip:httpsServerPort)访问HTTPS模块监听的端口，addr参数的数据类型为map，可以支持监听多个端口
            "10250" = "101.206.162.213:10250"

```
### tunnel-edge
**tunnel-edge**同样包含**stream**、**TCP**和**HTTPS**三个模块。其中**stream模块**包括gRPC client组件，用于向 **tunnel-cloud**发送gRPC长连接的请求。
### tunnel-edge 配置文件
tunnel-edge-conf.yaml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tunnel-edge-conf
  namespace: edge-system
data:
  tunnel_edge.toml: |
    [mode]
      [mode.edge]
        [mode.edge.stream]                              # stream模块
          [mode.edge.stream.client]                     # gRPC client组件
            token = "6ff2a1ea0f1611eb9896362096106d9d"  # 访问tunnel-cloud的验证token
            cert = "../../conf/certs/ca.crt"            # tunnel-cloud的gRPC server 的 server端证书的ca证书，用于验证server端证书
            dns = "localhost"                           # tunnel-cloud的gRPC server证书签的IP或域名
            servername = "localhost:9000"               # tunnel-cloud的gRPC server的IP和端口
            logport = 7000                              # log和健康检查的http server的监听端口，使用(curl -X PUT http://podip:logport/debug/flags/v -d "8")可以设置日志等级
            channelzaddr = "0.0.0.0:5000"               # gRPC channlez server的监听地址，用于获取gRPC的调试信息
        [mode.edge.https]                               # HTTPS模块
          cert= "../../conf/certs/kubelet-client.crt"   # tunnel-cloud 代理转发的HTTPS server的client端的证书
          key= "../../conf/certs/kubelet-client.key"    # **tunnel-cloud** 代理转发的HTTPS server的client端的私钥
```
## tunnel 转发模式
tunnel代理支持**TCP**或**HTTPS**请求转发。
<details><summary>TCP转发 </summary>
<p>

**TCP模块**会把**TCP**请求转发到[第一个连接云端的边缘节点](https://github.com/superedge/superedge/blob/main/pkg/tunnel/proxy/tcp/tcp.go#L69), 当**tunnel-cloud**只有一个**tunnel-edge**连接时，
请求会转发到**tunnel-edge**所在的节点
#### tunnel-cloud 配置文件
tunnel-cloud-conf.yaml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tunnel-cloud-conf
  namespace: edge-system
data:
  tunnel_cloud.toml: |
    [mode]
      [mode.cloud]
        [mode.cloud.stream]
          [mode.cloud.stream.server]
            grpcport = 9000
            key = "/etc/superedge/tunnel/certs/tunnel-cloud-server.key"
            cert = "/etc/superedge/tunnel/certs/tunnel-cloud-server.crt"
            tokenfile = "/etc/superedge/tunnel/token/token"
            logport = 51000
          [mode.cloud.stream.dns]
            debug = true
        [mode.cloud.tcp]
          "0.0.0.0:6443" = "127.0.0.1:6443"
          [mode.cloud.https]

```
**tunnel-cloud** 的gRPC server监听在9000端口，等待**tunnel-edge**建立gRPC长连接。访问**tunnel-cloud**的6443的请求会被转发到边缘节点的访问地址127.0.0.1:6443的server
#### tunnel-cloud.yaml

```yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: tunnel-cloud-conf
  namespace: edge-system
data:
  tunnel_cloud.toml: |
    {{tunnel_cloud.toml}}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: tunnel-cloud-token
  namespace: edge-system
data:
  token: |
    default:{{.TunnelCloudEdgeToken}}
---
apiVersion: v1
data:
  tunnel-cloud-server.crt: '{{tunnel-cloud-server.crt}}'
  tunnel-cloud-server.key: '{{tunnel-cloud-server.key}}'
kind: Secret
metadata:
  name: tunnel-cloud-cert
  namespace: edge-system
type: Opaque
---
apiVersion: v1
kind: Service
metadata:
  name: tunnel-cloud
  namespace: edge-system
spec:
  ports:
    - name: proxycloud
      port: 9000
      protocol: TCP
      targetPort: 9000
  selector:
    app: tunnel-cloud
  type: NodePort
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: tunnel-cloud
  name: tunnel-cloud
  namespace: edge-system
spec:
  selector:
    matchLabels:
      app: tunnel-cloud
  template:
    metadata:
      labels:
        app: tunnel-cloud
    spec:
      serviceAccount: tunnel-cloud
      serviceAccountName: tunnel-cloud
      containers:
        - name: tunnel-cloud
          image: superedge/tunnel:v0.2.0
          imagePullPolicy: IfNotPresent
          livenessProbe:
            httpGet:
              path: /cloud/healthz
              port: 51010
            initialDelaySeconds: 10
            periodSeconds: 60
            timeoutSeconds: 3
            successThreshold: 1
            failureThreshold: 1
          command:
            - /usr/local/bin/tunnel
          args:
            - --m=cloud
            - --c=/etc/superedge/tunnel/conf/mode.toml
            - --log-dir=/var/log/tunnel
            - --alsologtostderr
          volumeMounts:
            - name: token
              mountPath: /etc/superedge/tunnel/token
            - name: certs
              mountPath: /etc/superedge/tunnel/certs
            - name: conf
              mountPath: /etc/superedge/tunnel/conf
          ports:
            - containerPort: 9000
              name: tunnel
              protocol: TCP
            - containerPort: 6443
              name: apiserver
              protocol: TCP
          resources:
            limits:
              cpu: 50m
              memory: 100Mi
            requests:
              cpu: 10m
              memory: 20Mi
      volumes:
        - name: token
          configMap:
            name: tunnel-cloud-token
        - name: certs
          secret:
            secretName: tunnel-cloud-cert
        - name: conf
          configMap:
                  name: tunnel-cloud-conf
      nodeSelector:
              node-role.kubernetes.io/master: ""
      tolerations:
              - key: "node-role.kubernetes.io/master"
                operator: "Exists"
                effect: "NoSchedule"
```

tunnel-cloud-token的configmap中的TunnelCloudEdgeToken为随机字符串，用于验证**tunnel-edge**；tunnel-cloud-cert的secret对应的gRPC server的server端证书和私钥。

#### tunnel-edge 配置文件
tunnel-edge-conf.yaml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tunnel-edge-conf
  namespace: edge-system
data:
  tunnel_edge.toml: |
    [mode]
      [mode.edge]
        [mode.edge.stream]
          [mode.edge.stream.client]
            token = "{{.TunnelCloudEdgeToken}}"
            cert = "/etc/superedge/tunnel/certs/tunnel-ca.crt"
            dns = "{{ServerName}}"
            servername = "{{.MasterIP}}:9000"
            logport = 51000
```

**tunnel-edge**使用MasterIP:9000访问云端**tunnel-cloud**，使用TunnelCloudEdgeToken做为验证token，发向云端进行验证； token为**tunnel-cloud**的部署deployment的tunnel-cloud-token的configmap中的TunnelCloudEdgeToken；DNS为**tunnel-cloud**的gRPC
server的证书签的域名或IP；MasterIP为云端**tunnel-cloud** 所在节点的IP，9000为 **tunnel-cloud** service的nodePort
#### tunnel-edge.yaml
```yaml
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: tunnel-edge
rules:
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: tunnel-edge
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: tunnel-edge
subjects:
  - kind: ServiceAccount
    name: tunnel-edge
    namespace: edge-system
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tunnel-edge
  namespace: edge-system
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: tunnel-edge-conf
  namespace: edge-system
data:
  mode.toml: |
    {{tunnel-edge-conf}}
---
apiVersion: v1
data:
  tunnel-ca.crt: '{{.tunnel-ca.crt}}'
kind: Secret
metadata:
  name: tunnel-edge-cert
  namespace: edge-system
type: Opaque
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tunnel-edge
  namespace: edge-system
spec:
  selector:
    matchLabels:
      app: tunnel-edge
  template:
    metadata:
      labels:
        app: tunnel-edge
    spec:
      hostNetwork: true
      containers:
        - name: tunnel-edge
          image: superedge/tunnel:v0.2.0
          imagePullPolicy: IfNotPresent
          livenessProbe:
            httpGet:
              path: /edge/healthz
              port: 51010
            initialDelaySeconds: 10
            periodSeconds: 180
            timeoutSeconds: 3
            successThreshold: 1
            failureThreshold: 3
          resources:
            limits:
              cpu: 20m
              memory: 20Mi
            requests:
              cpu: 10m
              memory: 10Mi
          command:
            - /usr/local/bin/tunnel
          env:
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: spec.nodeName
          args:
            - --m=edge
            - --c=/etc/superedge/tunnel/conf/tunnel_edge.toml
            - --log-dir=/var/log/tunnel
            - --alsologtostderr
          volumeMounts:
            - name: certs
              mountPath: /etc/superedge/tunnel/certs
            - name: conf
              mountPath: /etc/superedge/tunnel/conf
      volumes:
        - secret:
            secretName: tunnel-edge-cert
          name: certs
        - configMap:
            name: tunnel-edge-conf
          name: conf
```
tunnel-edge-cert的secret对应的验证gRPC server证书的ca证书；**tunnel-edge**是以deployment的形式部署的，副本数为1，**TCP**转发现在只支持转发到单个节点。
</p>
</details>

<details><summary>HTTPS转发</summary>
<p>

通过tunnel将云端请求转发到边缘节点，需要使用边缘节点名做为**HTTPS** request的host的域名，域名解析可以复用[**tunnel-coredns**](https://github.com/superedge/superedge/blob/main/deployment/tunnel-coredns.yaml) 。使用**HTTPS**转发需要部署[**tunnel-cloud**](https://github.com/superedge/superedge/blob/main/deployment/tunnel-cloud.yaml) 、[**tunnel-edge**](https://github.com/superedge/superedge/blob/main/deployment/tunnel-edge.yaml) 和**tunnel-coredns**三个模块。
#### tunnel-cloud 配置文件
tunnel-cloud-conf.yaml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tunnel-cloud-conf
  namespace: edge-system
data:
  tunnel_cloud.toml: |
    [mode]
      [mode.cloud]
        [mode.cloud.stream]
          [mode.cloud.stream.server]
            grpcport = 9000
            logport = 51010
            key = "/etc/superedge/tunnel/certs/tunnel-cloud-server.key"
            cert = "/etc/superedge/tunnel/certs/tunnel-cloud-server.crt"
            tokenfile = "/etc/superedge/tunnel/token/token"
          [mode.cloud.stream.dns]
            configmap= "tunnel-nodes"
            hosts = "/etc/superedge/tunnel/nodes/hosts"
            service = "tunnel-cloud"
        [mode.cloud.https]
          cert = "/etc/superedge/tunnel/certs/apiserver-kubelet-server.crt"
          key = "/etc/superedge/tunnel/certs/apiserver-kubelet-server.key"
        [mode.cloud.https.addr]
          "10250" = "127.0.0.1:10250"
```
**tunnel-cloud** 的gRPC server监听在9000端口，等待**tunnel-edge**建立gRPC长连接。访问**tunnel-cloud**的10250的请求会被转发到边缘节点的访问地址127.0.0.1:10250的server。
#### tunel-cloud.yaml
[tunnel-cloud.yaml](https://github.com/superedge/superedge/blob/main/deployment/tunnel-cloud.yaml)

#### tunnel-edge 配置文件
tunnel-edge-conf.yaml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tunnel-edge-conf
  namespace: edge-system
data:
  tunnel_edge.toml: |
    [mode]
      [mode.edge]
        [mode.edge.stream]
          [mode.edge.stream.client]
            token = "{{.TunnelCloudEdgeToken}}"
            cert = "/etc/superedge/tunnel/certs/cluster-ca.crt"
            dns = "tunnel.cloud.io"
            servername = "{{.MasterIP}}:9000"
            logport = 51000
        [mode.edge.https]
          cert= "/etc/superedge/tunnel/certs/apiserver-kubelet-client.crt"
          key= "/etc/superedge/tunnel/certs/apiserver-kubelet-client.key"
```
**HTTPS模块**的证书和私钥是**tunnel-cloud**代理转发的边缘节点的server的server端证书对应的client证书，例如**tunnel-cloud**转发apiserver到kubelet的请求，需要配置kubelet 10250端口server端证书对应的client证书和私钥。
#### tunnel-edge.yaml
[tunnel-edge.yaml](https://github.com/superedge/superedge/blob/main/deployment/tunnel-edge.yaml)
</p>
</details>

## 本地调试
tunnel支持**HTTPS**(**HTTPS模块**)和**TCP**协议(**TCP模块**)，协议模块的数据是通过gRPC长连接传输(**stream模块**)，因此可以分模块进行本地调试。本地调试可以使用go的testing测试框架。配置文件的生成可以通过调用[config_test](https://github.com/superedge/superedge/blob/main/pkg/tunnel/conf/config_test.go)的
测试方法Test_Config(其中constant变量config_path是生成的配置文件的路径相对于config_test go 文件的路径，main_path 是配置文件相对testing文件的
路径)，例如**stream模块**的本地调试:config_path = "../../../conf"(生成的配置文件在项目的根目录下的conf文件夹)，则
main_path="../../../../conf"(([stream_test](https://github.com/superedge/superedge/blob/main/pkg/tunnel/proxy/stream/stream_test.go)相对
于conf的路径)，同时生成配置文件支持配置ca.crt和ca.key(在configpath/certs/ca.crt和configpath/certs/ca.key存在时则使用指定的ca签发证书)。
### stream模块调试
#### stream server的启动

```go
func Test_StreamServer(t *testing.T) {
	err := conf.InitConf(util.CLOUD, "../../../../../conf/cloud_mode.toml")
	if err != nil {
		t.Errorf("failed to initialize stream server configuration file err = %v", err)
		return
	}
	model.InitModules(util.CLOUD)
	InitStream(util.CLOUD)
	model.LoadModules(util.CLOUD)
	context.GetContext().RegisterHandler(util.MODULE_DEBUG, util.STREAM, StreamDebugHandler)
	model.ShutDown()

}
```
```
加载配置文件(conf.InitConf)->初始化模块(model.InitMoudule)->初始化stream模块(InitStream)->加载初始化的模块->注册自定义的handler(StreamDebugHandler)->关闭模块(model.ShutDown)
```
StreamDebugHandler是调试云边消息收发的自定义handler
#### stream client的启动
```go
func Test_StreamClient(t *testing.T) {
	os.Setenv(util.NODE_NAME_ENV, "node1")
	err := conf.InitConf(util.EDGE, "../../../../../conf/edge_mode.toml")
	if err != nil {
		t.Errorf("failed to initialize stream client configuration file err = %v", err)
		return
	}
	model.InitModules(util.EDGE)
	InitStream(util.EDGE)
	model.LoadModules(util.EDGE)
	context.GetContext().RegisterHandler(util.MODULE_DEBUG, util.STREAM, StreamDebugHandler)
	go func() {
		running := true
		for running {
			node := context.GetContext().GetNode(os.Getenv(util.NODE_NAME_ENV))
			if node != nil {
				node.Send2Node(&proto.StreamMsg{
					Node:     os.Getenv(util.NODE_NAME_ENV),
					Category: util.STREAM,
					Type:     util.MODULE_DEBUG,
					Topic:    uuid.NewV4().String(),
					Data:     []byte{'c'},
				})
			}
			time.Sleep(10 * time.Second)
		}
	}()
	model.ShutDown()

}
```
```
设置节点名环境变量->加载配置文件(conf.InitConf)->初始化模块(model.InitMoudule)->初始化stream模块(InitStream)->加载初始化的模块->注册自定义的handler(StreamDebugHandler)->关闭模块(model.ShutDown)
```

节点名是通过NODE_NAME的环境变量加载的
### TCP模块调试
#### TCP server的调试
```go
func Test_TcpServer(t *testing.T) {
	err := conf.InitConf(util.CLOUD, "../../../../../conf/cloud_mode.toml")
	if err != nil {
		t.Errorf("failed to initialize stream server configuration file err = %v", err)
		return
	}
	model.InitModules(util.CLOUD)
	InitTcp()
	stream.InitStream(util.CLOUD)
	model.LoadModules(util.CLOUD)
	model.ShutDown()
}
```
需要同时初始化**TCP模块**(InitTcp)和**stream模块**(stream.InitStream)
#### TCP client的调试
```go
func Test_TcpClient(t *testing.T) {
	os.Setenv(util.NODE_NAME_ENV, "node1")
	err := conf.InitConf(util.EDGE, "../../../../../conf/edge_mode.toml")
	if err != nil {
		t.Errorf("failed to initialize stream client configuration file err = %v", err)
		return
	}
	model.InitModules(util.EDGE)
	InitTcp()
	stream.InitStream(util.EDGE)
	model.LoadModules(util.EDGE)
	model.ShutDown()
}
```
### HTTPS模块调试
和**TCP模块**调试类似，需要同时加载**HTTPS模块**和**stream模块**
### tunnel main()函数的调试
在tunnel的main的测试文件[tunnel_test](https://github.com/superedge/superedge/blob/main/cmd/tunnel/tunnel_test.go)，需要使用init()设置参数，同时需要使用TestMain解析参数和调用测试方法









