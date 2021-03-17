
# tunnel

tunnel是云边端通信的隧道，分为tunnel-cloud和tunnel-edge，分别承担云边隧道的两端

## 作用
- 代理云端访问边缘节点组件的请求，解决云端无法直接访问边缘节点的问题（边缘节点没有暴露在公网中）

## 架构图
<div align="left">
  <img src="../img/tunnel.png" width=70% title="tunnel Architecture">
</div>

## 实现方案
### 节点注册
   - 边缘节点上tunnel-edge主动连接云端tunnel-cloud service，service根据负载均衡策略将请求转到tunnel-cloud的pod。
   - tunnel-edge与tunnel-cloud建立grpc连接后，tunnel-cloud会把自身的podIp和tunnel-edge所在节点的nodeName的映射写入DNS。grpc连接断开之后，tunnel-cloud会删除podIp和节点名映射。

### 请求的代理转发
   - apiserver或者其它云端的应用访问边缘节点上的kubelet或者其它应用时，tunnel-dns通过DNS劫持(将host中的节点名解析为tunnel-cloud的podIp)把请求转发到tunnel-cloud的pod。
   - tunnel-cloud根据节点名把请求信息转发到节点名对应的与tunnel-edge建立的grpc连接。
   - tunnel-edge根据接收的请求信息请求边缘节点上的应用。

## 配置文件
tunnel cloud和tunnel edge的启动配置文件
### tunnel cloud

#### stream模块
##### server组件
- grpcport: grpc server监听的端口
- logport: log和健康检查的http server的监听端口，使用(curl -X PUT http://podip:logport/debug/flags/v -d "8")可以设置日志等级
- key: grpc server的server端私钥
- cert: grpc server的server端证书
- tokenfile: token的列表文件(nodename:随机字符串)，用于验证边缘节点tunnel edge发送的token，如果根据节点名验证没有通过，会用default对应的token去验证
- channelzAddr: grpc [channlez](https://grpc.io/blog/a-short-introduction-to-channelz/) server的监听地址，用于获取grpc的调试信息
##### dns组件

- configmap: coredns hosts插件的配置文件的configmap
- hosts: coredns hosts插件的配置文件的configmap，在tunnel cloud pod的挂载文件
- service: tunnel cloud的service name
- debug: dns组件开关，debug=true dns组件关闭，tunnel cloud 内存中的节点名映射不会保存到coredns hosts插件的配置文件的configmap，默认值为false
#### tcp模块
- 参数的格式是"0.0.0.0:cloudPort": "EdgeServerIp:EdgeServerPort"，cloudPort为tunnel cloud tcp模块server监听端口，EdgeServerIp和EdgeServerPort为代理转发的边缘节点server的ip和端口
#### https模块
- cert: https模块server端证书
- key: https模块server端私钥
- addr: 参数的格式是"httpsServerPort":"EdgeHttpsServerIp:EdgeHttpsServerPort"
  ，httpsServerPort为https模块server端的监听端口，EdgeHttpsServerIp:EdgeHttpsServerPort为代理转发边缘节点https server的ip和port，
  https模块的server是跳过验证client端证书的，因此可以使用(curl -k https://podip:httpsServerPort)访问https模块监听的端口，addr参数的数据类型为map，可以支持监听多个端口
### tunnel edge
#### https模块
- cert: tunnel cloud 代理转发的https server的client端的证书
- key: tunnel cloud 代理转发的https server的client端的私钥
#### stream模块

- token: 访问tunnel cloud的验证token
- cert: tunnel cloud的grpc server 的 server端证书的ca证书，用于验证server端证书
- dns: tunnel cloud的grpc server证书签的ip或域名
- servername: tunnel cloud的grpc server的ip和端口
- logport: log和健康检查的http server的监听端口，使用(curl -X PUT http://podip:logport/debug/flags/v -d "8")可以设置日志等级
- channelzaddr: grpc channlez server的监听地址，用于获取grpc的调试信息
## 使用场景
### tcp转发
tcp模块会把tcp请求转发到[第一个连接云端的边缘节点](https://github.com/superedge/superedge/blob/main/pkg/tunnel/proxy/tcp/tcp.go#L69)，当连接tunnel cloud只有一个tunnel edge时，
默认转发到tunnel edge所在的节点
#### tunnel cloud
##### 配置文件(tunnel_cloud.toml)

```toml
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
tunnel cloud 的grpc server监听在9000端口，等待tunnel edge建立grpc长连接。访问tunnel cloud的6443的请求会被转发到边缘节点的访问地址127.0.0.1:6443的server
##### tunnel-cloud.yaml

```yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: tunnel-cloud-conf
  namespace: kube-system
data:
  mode.toml: |
    {{tunnel-cloud-tcp.toml}}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: tunnel-cloud-token
  namespace: kube-system
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
  namespace: kube-system
type: Opaque
---
apiVersion: v1
kind: Service
metadata:
  name: tunnel-cloud
  namespace: kube-system
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
  namespace: kube-system
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

部署yaml中的tunnel-cloud-conf的configmap对应的就是tunnel
cloud的配置文件；tunnel-cloud-token的configmap中的TunnelCloudEdgeToken为随机字符串，用于验证tunnel edge； tunnel-cloud-cert的secret对应的grpc
server的server端证书和私钥。

#### tunnel edge

##### 配置文件(tunnel_edge.toml)

```toml
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

tunnel edge使用MasterIP:9000访问云端tunnel cloud，使用TunnelCloudEdgeToken做为验证token，发向云端进行验证。 token为tunnel
cloud的部署deployment的tunnel-cloud-token的configmap中的TunnelCloudEdgeToken；dns为tunnel cloud的grpc
server的证书签的域名或ip；MasterIP为云端tunnel cloud 所在节点的ip，9000为 tunnel-cloud service的nodePort
##### tunnel-edge.yaml
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
    namespace: kube-system
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tunnel-edge
  namespace: kube-system
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: tunnel-edge-conf
  namespace: kube-system
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
  namespace: kube-system
type: Opaque
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tunnel-edge
  namespace: kube-system
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
部署yaml中的tunnel-edge-conf的configmap对应的就是tunnel edge的配置文件；tunnel-edge-cert的secret对应的验证grpc server证书的ca证书；tunnel edge是以deployment的形式部署的，副本数为1，tcp转发现在只支持转发到单个节点。
### https转发
通过tunnel将云端请求转发到边缘节点，需要使用边缘节点名做为https request的host的域名，域名解析可以复用[tunnel-coredns](https://github.com/superedge/superedge/blob/main/deployment/tunnel-coredns.yaml)。使用https转发需要部署[tunnel-cloud](https://github.com/superedge/superedge/blob/main/deployment/tunnel-cloud.yaml)、[tunnel-edge](https://github.com/superedge/superedge/blob/main/deployment/tunnel-edge.yaml)和[tunnel-coredns](https://github.com/superedge/superedge/blob/main/deployment/tunnel-coredns.yaml)三个模块。
#### tunnel cloud配置
```toml
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
tunnel cloud 的grpc server监听在9000端口，等待tunnel edge建立grpc长连接。访问tunnel cloud的10250的请求会被转发到边缘节点的访问地址127.0.0.1:10250的server。
tunnel-cloud配置对应的是tunnel-cloud的部署yaml中[tunnel-cloud-conf](https://github.com/superedge/superedge/blob/main/deployment/tunnel-cloud.yaml#L41)configmap对应的内容

#### tunnel edge配置
```toml
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
https模块的证书和私钥是tunnel cloud代理转发的边缘节点的server的server端证书对应的client证书，例如tunnel cloud转发apiserver到kubelet的请求，需要配置kubelet 10250端口server端证书对应的client证书和私钥。
tunnel-edge配置对应的tunnel-edge的部署yaml中[tunnel-edge-conf](https://github.com/superedge/superedge/blob/main/deployment/tunnel-edge.yaml#L33)configmap对应的内容。
## 本地调试
tunnel支持https(https模块)和tcp协议(tcp模块)，协议模块的数据是通过grpc长连接传输(stream模块)，因此可以分模块进行本地调试。本地调试可以使用go的testing测试框架。配置文件的生成可以通过调用[config_test](https://github.com/superedge/superedge/blob/main/pkg/tunnel/conf/config_test.go)的
测试方法Test_Config(其中constant变量config_path是生成的配置文件的路径相对于config_test go 文件的路径，main_path 是配置文件相对testing文件的
路径)，例如stream模块的本地调试:config_path = "../../../conf"(生成的配置文件在项目的根目录下的conf文件夹)，则
main_path="../../../../conf"(([stream_test](https://github.com/superedge/superedge/blob/main/pkg/tunnel/proxy/stream/stream_test.go)相对
于conf的路径)，同时生成配置文件支持配置ca.crt和ca.key(在configpath/certs/ca.crt和configpath/certs/ca.key存在时则使用指定的ca签发证书)。
### stream模块调试
#### stream server的启动
```go
func Test_StreamServer(t *testing.T) {
	err := conf.InitConf(util.CLOUD, "../../../../conf/cloud_mode.toml")
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
	err := conf.InitConf(util.EDGE, "../../../../conf/edge_mode.toml")
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
### tcp模块调试
#### tcp server的调试
```go
func Test_TcpServer(t *testing.T) {
	err := conf.InitConf(util.CLOUD, "../../../../conf/cloud_mode.toml")
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
需要同时初始化tcp模块(InitTcp)和stream模块(stream.InitStream)
#### tcp client的调试
```go
func Test_TcpClient(t *testing.T) {
	os.Setenv(util.NODE_NAME_ENV, "node1")
	err := conf.InitConf(util.EDGE, "../../../../conf/edge_mode.toml")
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
### https模块调试
和tcp模块调试类似，需要同时加载https模块和stream模块
### tunnel main()函数的调试
在tunnel的main的测试文件[tunnel_test](https://github.com/superedge/superedge/blob/main/cmd/tunnel/tunnel_test.go)，需要使用init()设置参数，同时需要使用TestMain解析参数和调用测试方法









