
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
   - 边缘节点上tunnel-edge主动连接云端tunnel-cloud service,service根据负载均衡策略将请求转到tunnel-cloud的pod。
   - tunnel-edge与tunnel-cloud建立grpc连接后，tunnel-cloud会把自身的podIp和tunnel-edge所在节点的nodeName的映射写入DNS。grpc连接断开之后，tunnel-cloud会删除podIp和节点名映射。

### 请求的代理转发
   - apiserver或者其它云端的应用访问边缘节点上的kubelet或者其它应用时,tunnel-dns通过DNS劫持(将host中的节点名解析为tunnel-cloud的podIp)把请求转发到tunnel-cloud的pod。
   - tunnel-cloud根据节点名把请求信息转发到节点名对应的与tunnel-edge建立的grpc连接。
   - tunnel-edge根据接收的请求信息请求边缘节点上的应用。

## 配置文件
### tunnel cloud
#### stream模块
##### server组件
- grpcport grpc server监听的端口
- logport log的http server的监听地址(curl -X PUT http://podip:logport/debug/flags/v -d "8")
- key grpc server的server端私钥
- cert grpc server的server端证书
- tokenfile token的列表文件(nodename:随机字符串)，用于验证边缘节点tunnel edge发送的token，如果根据节点名验证没有通过，会用default对应的token去验证
- channelzAddr grpc [channlez](https://grpc.io/blog/a-short-introduction-to-channelz/) server的监听地址，用于获取grpc的调试信息
##### dns组件
- configmap coredns hosts插件的配置文件的configmap
- hosts coredns hosts插件的配置文件的configmap在tunnel cloud pod的挂载文件
- service tunnel cloud的service name
- debug dns组件开发，debug = true dns组件关闭，tunnel cloud 内存中的节点名映射不会保存到configmap coredns hosts插件的配置文件的configmap，默认值为false
#### tcp模块
- "0.0.0.0:cloudPort": "EdgeServerIp:EdgeServerPort"  cloudPort为tunnel cloud tcp模块server监听端口，EdgeServerIp和EdgeServerPort为代理转发的边缘节点server的ip和端口
#### https模块
- cert https模块server端证书
- key https模块server端私钥
- addr "httpsServerPort":"EdgeHttpsServerIp:EdgeHttpsServerPort" httpsServerPort为https模块server端监听的地址，EdgeHttpsServerIp:EdgeHttpsServerPort为代理转发边缘节点https server的ip和port，
        https模块的server是跳过验证client端证书的，因此使用curl -k https://podip:httpsServerPort/path 访问https模块监听的端口，addr参数的数据类型为map，可以支持监听多个端口
### tunnel edge
#### https模块
- cert tunnel cloud 代理转发的https server的client端的证书
- key tunnel cloud 代理转发的https server的client端的私钥
#### stream模块
- token 访问tunnel cloud的验证token
- cert tunnel cloud grpc server 的 server端证书的ca证书，用于验证server端证书
- dns tunnel cloud grpc server证书签的ip或域名 
- servername tunnel cloud grpc server的ip和端口
- logport log的http server的监听地址(curl -X PUT http://podip:logport/debug/flags/v -d "8") 
- channelzaddr grpc channlez server的监听地址，用于获取grpc的调试信息
## 本地调试
tunnel支持https和tcp协议分别对应https模块和tcp模块，协议模块的数据是通过grpc长连接传输,即对应的stream模块，可以通过go的testing测试框架
进行本地调试。配置文件的生成可以通过调用[config_test](https://github.com/superedge/superedge/blob/main/pkg/tunnel/conf/config_test.go)的
测试方法Test_Config(其中constant变量config_path是生成的配置文件的路径相对于config_test go 文件的路径，main_path 是配置文件相对testing文件的
路径)，例如stream模块的本地调试:config_path = "../../../conf"生成的配置文件在项目的根目录下的conf文件夹，则
main_path="../../../../"([stream_test](https://github.com/superedge/superedge/blob/main/pkg/tunnel/proxy/stream/stream_test.go)相对
于conf的路径，同时生成配置文件的证书支持配置ca.crt和ca.key(在configpath/certs/ca.crt和configpath/certs/ca.key存在时则使用指定的ca证书签发证书)。
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
设置节点名环境变量->加载配置文件(conf.InitConf)->初始化模块(model.InitMoudule)->初始化stream模块(InitStream)->加载初始化的模块->注册自定义的handler(StreamDebugHandler)
->关闭模块(model.ShutDown)
```
节点名是通过NODE_NAME的环境加载的
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
在tunnel的main的测试文件[tunnel_test](https://github.com/superedge/superedge/blob/main/cmd/tunnel/tunnel_test.go)，需要使用init()
设置参数，同时需要使用TestMain解析参数和调用测试方法
## 使用场景
### tcp转发
tcp模块会把tcp请求转发到[第一个连接云端的边缘节点](https://github.com/superedge/superedge/blob/main/pkg/tunnel/proxy/tcp/tcp.go#L69)
#### tunnel cloud的配置
```toml
[mode]
	[mode.cloud]
		[mode.cloud.stream]
			[mode.cloud.stream.server]
				grpcport = 9000
				logport = 8000
                channelzaddr = "0.0.0.0:6000"
				key = "../../../../conf/certs/cloud.key"
				cert = "../../../../conf/certs/cloud.crt"
				tokenfile = "../../../../conf/token"
			[mode.cloud.stream.dns]
				configmap= "proxy-nodes"
				hosts = "/etc/superedge/proxy/nodes/hosts"
				service = "proxy-cloud-public"
				debug = true
            [mode.cloud.tcp]
                "0.0.0.0:6443" = "127.0.0.1:6443"
            [mode.cloud.https]
                cert ="../../../../conf/certs/kubelet.crt"#kubelet的服务端证书
                key = "../../../../conf/certs/kubelet.key"
			[mode.cloud.https.addr]
				"10250" = "101.206.162.213:10250"
```

### https转发
通过tunnel将云端请求转发到边缘节点，需要使用边缘节点名做为https request的host的域名，域名解析可以复用tunnel cloud，域名解析可以复用[tunnel-coredns](https://github.com/superedge/superedge/blob/main/deployment/tunnel-coredns.yaml)
#### tunnel cloud配置
```toml
[mode]
	[mode.cloud]
		[mode.cloud.stream]
			[mode.cloud.stream.server]
				grpcport = 9000
				logport = 8000
                channelzaddr = "0.0.0.0:6000"
				key = "../../../../conf/certs/cloud.key"
				cert = "../../../../conf/certs/cloud.crt"
				tokenfile = "../../../../conf/token"
			[mode.cloud.stream.dns]
				configmap= "proxy-nodes"
				hosts = "/etc/superedge/proxy/nodes/hosts"
				service = "proxy-cloud-public"
				debug = true
            [mode.cloud.tcp]
                "0.0.0.0:6443" = "127.0.0.1:6443"
            [mode.cloud.https]
                cert ="../../../../conf/certs/kubelet.crt"#kubelet的服务端证书
                key = "../../../../conf/certs/kubelet.key"
			[mode.cloud.https.addr]
				"10250" = "101.206.162.213:10250"
```
#### tunnel edge配置
```toml
[mode]
	[mode.edge]
		[mode.edge.stream]
			[mode.edge.stream.client]
				token = "6ff2a1ea0f1611eb9896362096106d9d"
				cert = "../../../../conf/certs/ca.crt"
  				dns = "localhost"
				servername = "localhost:9000"
				logport = 7000
				channelzaddr = "0.0.0.0:5000"
			[mode.edge.https]
				cert= "../../../../conf/certs/kubelet-client.crt"#apiserver访问kubelet的客户端证书
				key= "../../../../conf/certs/kubelet-client.key"
```
https模块的证书和私钥是tunnel cloud 代理转发的边缘节点的server的server端证书对应的client证书，例如tunnel cloud转发apiserver到kubelet的请求，需要配置kubelet 10250端口server端证书对应的
client端证书。









