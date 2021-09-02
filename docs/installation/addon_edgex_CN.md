简体中文

# 一键部署EdgeX Foundry到边缘集群  

* [一键部署EdgeX Foundry到边缘集群](#一键部署edgex-foundry到边缘集群)
  * [1\. 背景](#1-背景)
  * [2\. 方案设计](#2-方案设计)
  * [3\. EdgeX Foundry组件的安装](#3-edgex-foundry组件的安装)
    * [&lt;1&gt; 准备条件](#1-准备条件)
    * [&lt;2&gt; 安装EdgeX Foundry的组件](#2-安装edgex-foundry的组件)
  * [4\. EdgeX Foundry的测试](#4-edgex-foundry的测试)
    * [&lt;1&gt; 访问consul](#1-访问consul)
    * [&lt;2&gt; 访问ui](#2-访问ui)
  * [5\. EdgeX Foundry的验证](#5-edgex-foundry的验证)
    * [&lt;1&gt; 连接设备](#1-连接设备)
    * [&lt;2&gt; 数据访问](#2-数据访问)
    * [&lt;3&gt; 设备控制](#3-设备控制)
      * [(1) 查看可用命令](#1-查看可用命令)
      * [(2) get命令](#2-get命令)
      * [(3) put命令](#3-put命令)
    * [&lt;4&gt; 数据导出](#4-数据导出)
  * [6\. EdgeX Foundry的卸载](#6-edgex-foundry的卸载)
  * [7\. 补充](#7-补充)
  

## 1. 背景

随着物联网的发展，连接云的设备种类和数量大大增加， EdgeX Foundry是一个开源的可拓展的软件平台，可以部署在网络边缘连接各设备和云端，对设备进行管理和控制。部署EdgeX Foundry在边缘集群，可以进一步增强边缘集群的功能，同时相比将EdgeX Foundry部署在中心云集群，可以利用边缘集群的优势，更大发挥EdgeX Foundry的功能。

-   EdgeX Foundry官网没有直接提供适用于kubernetes的yaml文件，需要自己搜索或编写各组件的文件，大大增加了部署成本。
-   各个组件没有按层级分类，没有提供有效的分层部署方式，不能支持自定义安装需要的部分组件。
-   没有经过测试和调整直接部署各组件容易遇到各种错误。

## 2. 方案设计

为了能让用户快速在边缘集群使用EdgeX Foundry的功能，我们提供关于EdgeX Foundry在边缘集群的一键部署操作，通过提前配置相关文件，按层级分类，将命令集成到edgeadm的 addon命令下，并进行相应测试，减少可能的错误，使用户只需简单几步操作，就可以轻松实现EdgeX Foundry的部署、使用、卸载，降低用户的部署成本。

-    面对没有官方适用kubernetes的yaml文件的问题，我们通过网络搜索其他相关开源项目，选择有效的yaml文件，以及自己手动编写部分缺失文件，来集齐各个组件的yaml文件。

-    为了增强部署的扩展性，支持用户根据输入的不同参数，进行EdgeX Foundry的分层部署。我们将所有组件按EdgeX Foundry的服务层次分为不同文件，并且支持在安装命令后跟随flag指定安装的服务层级，根据需要部署特定层级的组件即可。默认情况下安装所有的组件满足完整的功能。相关组件的分层如下所示  

<div align="center">
  <img src="/docs/img/edgex-layer.png" width=80% title="Edgex-layer">
</div>
  


-    我们已经在边缘集群上对部署的各个组件进行了相应的测试和调整，减少了可能的错误。

## 3. EdgeX Foundry组件的安装

### <1> 准备条件

执行以下命令下载edgeadm文件和k8s安装包  

```shell
arch=amd64 version=v0.5.0 && rm -rf edgeadm-linux-* && wget https://superedge-1253687700.cos.ap-guangzhou.myqcloud.com/$version/$arch/edgeadm-linux-$arch-$version.tgz && tar -xzvf edgeadm-linux-* && cd edgeadm-linux-$arch-$version && ./edgeadm
```

安装一个边缘集群,具体参考以下链接(上面已经下载了最新的安装文件，下面链接内安装边缘集群**无需**再安装edgeadm的安装包)  

[一键安装边缘独立Kubernetes 集群](https://github.com/superedge/superedge/blob/main/docs/installation/install_edge_kubernetes_CN.md)  


### <2> 安装EdgeX Foundry的组件
执行以下命令，即可一键安装EdgeX Foundry的所有组件  

```shell
./edgeadm addon edgex
```  


如果得到以下成功提示，说明部署成功 

```shell
Start install edgex-application-services.yml to your cluster
Deploy edgex-application-services.yml success!
Start install edgex-core-services.yml to your cluster
Deploy edgex-core-services.yml success!
Start install edgex-device-services.yml to your cluster
Deploy edgex-device-services.yml success!
Start install edgex-support-services.yml to your cluster
Deploy edgex-support-services.yml success!
Start install edgex-system-management.yml to your cluster
Deploy edgex-system-management.yml success!
Start install edgex-ui.yml to your cluster
Deploy edgex-ui.yml success!

```  

也可以通过以下命令添加所需组件到集群  

```shell
./edgeadm addon edgex [flag]
```  
可以通过`./edgeadm addon edgex --help`命令查看可以使用的flag  

具体flag细节如下  

```shell
--app           Addon the edgex application-services to cluster.
--core          Addon the edgex core-services to cluster.
--device        Addon the edgex device-services to cluster.
--support       Addon the edgex supporting-services to cluster.
--sysmgmt       Addon the edgex system management to cluster
--ui            Addon the edgex ui to cluster.
```  
例如只安装core服务层的相关组件，请运行  

```shell
./edgeadm addon edgex --core
```  
其他组件同上安装，替换flag即可。如需同时安装多个层级组件，可以同时添加多个flag。  
  
部署成功后，可以通过以下命令查看svc和pod的启动情况  

```shell
kubectl get svc,pods -n edgex
```  
  

**注意**: 如果出现同一层级的组件部分安装成功，部分安装失败，可直接重新执行安装命令进行更新和安装。如果已安装的组件出现异常无法运行，可以使用`./edgeadm detach edgex [flag]`对特定层级的组件进行卸载重装。卸载操作具体参考 [6\. EdgeX Foundry的卸载](#6-edgex-foundry的卸载)
  
## 4. EdgeX Foundry的测试  
### <1> 访问consul  
从网页访问core-consul的服务的端口可以查看各组件的部署情况，其中`30850`是core-consul服务暴露的端口号  
```shell
curl http://localhost:30850/ui/dc1/services
```  
<div align="center">
  <img src="/docs/img/edgex-consul.png" width=80% title="Edgex-consul">
</div>
  

如果显示红色叉号，说明组件安装失败，如果刷新仍然无效，请试图对该组件所在层级进行卸载和安装。
  
### <2> 访问ui
从网页通过访问ui服务的端口同样可以查看各组件是否正常部署，其中`30040``是ui服务暴露的端口号  

```shell
curl http://localhost:30040/
```  
<div align="center">
  <img src="/docs/img/edgex-ui.png" width=80% title="Edgex-ui">
</div>  



如果部署成功，则各项会有相应的条目生成
  
## 5. EdgeX Foundry的验证
### <1> 连接设备
利用下面的命令,打开一个新的yaml文件  

```shell
vim edgex-device-random.yaml
```  
并将以下内容复制粘贴到文件中然后保存退出  
```shell 
apiVersion: v1
kind: Service
metadata:
  name: edgex-device-random
  namespace: edgex
spec:
  type: NodePort
  selector:
    app: edgex-device-random
  ports:
  - name: http
    port: 49988
    protocol: TCP
    targetPort: 49988
    nodePort: 30088
---
apiVersion: apps/v1
kind: Deployment
metadata: 
  name: edgex-device-random
  namespace: edgex
spec:
  selector:
    matchLabels: 
      app: edgex-device-random
  template:
    metadata:
      labels: 
        app: edgex-device-random
    spec:
      hostname: edgex-device-random
      containers:
      - name: edgex-device-random
        image: EdgeX Foundry/docker-device-random-go:1.3.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          protocol: TCP
          containerPort: 49988
        envFrom: 
        - configMapRef:
            name: common-variables
        env:
          - name: Service_Host
            value: "edgex-device-random"
```  
通过以下命令启动该组件  
```shell
kubectl apply -f edgex-device-random.yaml
```  
该命令会启动一个随机整数生成器的虚拟设备连接到EdgeX Foundry，该设备会向core-data发送随机数，同时接收core-command的命令控制。  
### <2> 数据访问
通过以下命令从网页访问core-data的服务的端口查看上一步启动的随机数设备向core服务发送的最近10条数据，其中`30080`是core-data服务的端口号，`Random-Integer-Generator01`是以上文件安装的虚拟设备  
```shell
curl http://localhost:30080/api/v1/event/device/Random-Integer-Generator01/10
```  
<div align="center">
  <img src="/docs/img/edgex-data.png" width=80% title="Edgex-data">
</div>  


### <3> 设备控制
#### (1) 查看可用命令
网页访问core-command服务的端口查看可以对虚拟设备进行的命令,包括get和put，其中get用于获取数据，put用于下发命令，其中`30082`是core-command服务的端口号  

```shell
curl http://localhost:30082/api/v1/device/name/Random-Integer-Generator01
```  
<div align="center">
  <img src="/docs/img/edgex-command.png" width=80% title="Edgex-command">
</div>  


#### (2) get命令
从上面的网页内容中可以看到get命令的url，使用get的url可以获取随机数设备发送的数据(**此处仅为例子,具体url根据显示获取,并请记得将`edgex-core-command:48082`字段改为`localhost:30082`**)，其中`30082`是core-command服务的端口号  
```shell
curl http://localhost:30082/api/v1/device/4a602dc3-afd5-4c76-9d72-de02407e80f8/command/5353248d-8006-4b01-8250-a07cb436aeb1
```  
<div align="center">
  <img src="/docs/img/edgex-get.png" width=80% title="Edgex-get">
</div>  


#### (3) put命令
执行put命令可以对虚拟设备进行控制，这里以修改其产生的随机数的范围为例，从网页中找到put命令的url，并执行以下命令：(**此处仅为例子,具体url由显示的put命令的url得到,并请记得将`edgex-core-command:48082`字段改为`localhost:30082`,将`{}`内的内容改为可用的参数,该可修改参数也由之前查询命令的显示中得到**),其中`30082`是core-command服务的端口号  
```shell
curl -X PUT -d '{"Min_Int8": "0", "Max_Int8": "100"}' http://localhost:30082/api/v1/device/4a602dc3-afd5-4c76-9d72-de02407e80f8/command/5353248d-8006-4b01-8250-a07cb436aeb1
```  
这里将虚拟设备的生成数范围改为0到100，执行put命令无输出，可通过get命令查看新产生的数据是否在范围0-100内。
### <4> 数据导出  
通过一下命令打开一个新的yaml文件  
```shell
vim mqtt.yaml
```  
将以下内容复制粘贴到文件中  

```shell
apiVersion: v1
kind: Service
metadata:
  name: edgex-app-service-configurable-mqtt
  namespace: edgex
spec:
  type: NodePort 
  selector:
    app: edgex-app-service-configurable-mqtt
  ports:
  - name: http
    port: 48101
    protocol: TCP
    targetPort: 48101
    nodePort: 30200
---
apiVersion: apps/v1
kind: Deployment
metadata: 
  name: edgex-app-service-configurable-mqtt
  namespace: edgex
spec:
  selector:
    matchLabels: 
      app: edgex-app-service-configurable-mqtt
  template:
    metadata:
      labels: 
        app: edgex-app-service-configurable-mqtt
    spec:
      hostname: edgex-app-service-configurable-mqtt
      containers:
      - name: edgex-app-service-configurable-mqtt
        image: EdgeX Foundry/docker-app-service-configurable:1.1.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          protocol: TCP
          containerPort: 48101
        envFrom: 
        - configMapRef:
            name: common-variables
        env:
          - name: edgex_profile
            value: "mqtt-export"
          - name: Service_Host
            value: "edgex-app-service-configurable-mqtt"
          - name: Service_Port
            value: "48101"
          - name: MessageBus_SubscribeHost_Host
            value: "edgex-core-data"
          - name: Binding_PublishTopic
            value: "events"
          - name: Writable_Pipeline_Functions_MQTTSend_Addressable_Address
            value: "broker.mqttdashboard.com"
          - name: Writable_Pipeline_Functions_MQTTSend_Addressable_Port
            value: "1883"
          - name: Writable_Pipeline_Functions_MQTTSend_Addressable_Protocol
            value: "tcp"
          - name: Writable_Pipeline_Functions_MQTTSend_Addressable_Publisher
            value: "edgex"
          - name: Writable_Pipeline_Functions_MQTTSend_Addressable_Topic
            value: "EdgeXEvents"
```  
执行以下命令  
```shell
kubectl apply -f mqtt.yaml
```  

启动该组件，该组件可以将core-data中的数据导出到HiveMQ的公开的MQTT broker上。可以通过网页访问该代理查看数据是否成功导出到云端。
访问以下网址进入网页  
```shell
http://www.hivemq.com/demos/websocket-client/
```  
<div align="center">
  <img src="/docs/img/edgex-hivemq-connect.png" width=80% title="Edgex-hivemq-connect">
</div>  

点击connect进行连接，填写主题为EdgeXEvents  

<div align="center">
  <img src="/docs/img/edgex-hivemq-create.png" width=80% title="Edgex-hivemq-create">
</div>  

即可看到message一栏出现虚拟设备向EdgeX Foundry发送的数据  

<div align="center">
  <img src="/docs/img/edgex-hivemq-message.png" width=80% title="Edgex-hivemq-message">
</div>  

但是，由于这是公有的broker，多方多次上传的数据都会保留并共存在相应的主题下，所以即使message一栏有数据显示，可能是之前导出操作遗留的数据，要想真正验证是否导出成功，可以在connect后尝试创建一个新主题，该主题尚无message显示，再修改mqtt.yaml中`env`下的`Writable_Pipeline_Functions_MQTTSend_Addressable_Topic`的值为该主题，部署后查看broker网页中是否有数据出现，若有，说明真正导出成功。

**注意**: 如果上述操作中出现网页无法访问等异常，请重新查看pod情况，必要时进行卸载重装。
  
## 6. EdgeX Foundry的卸载
如果是执行`./edgeadm addon edgex`安装了所有组件或者自定义安装了所有层级组件的，可以执行以下命令将所有EdgeX Foundry卸载，同时卸载在主机上产生的挂载数据。如果是只安装了部分层级或者有部分组件缺失的，请根据后文中的通过添加flag的方式逐个层级卸载  

```shell
./edgeadm detach edgex
```  
出现以下成功显示，说明卸载完成。  

```shell
Start uninstall edgex-application-services.yml to your cluster
Detach edgex-application-services.yml success!
Start uninstall edgex-application-services.yml from your cluster
Detach edgex-application-services.yml success!
Start uninstall edgex-core-services.yml from your cluster
Detach edgex-core-services.yml success!
Start uninstall edgex-device-services.yml from your cluster
Detach edgex-device-services.yml success!
Start uninstall edgex-support-services.yml from your cluster
Detach edgex-support-services.yml success!
Start uninstall edgex-system-management.yml from your cluster
Detach edgex-system-management.yml success!
Start uninstall edgex-ui.yml from your cluster
Detach edgex-ui.yml success!
Start uninstall edgex-configmap.yml from your cluster
Detach edgex-configmap.yml success!
Start uninstall edgex completely.
Delete edgex completely success!
```  
也可执行`./edgeadm detach edgex [flag]`对EdgeX Foundry进行卸载，可以通过以下命令查看可以使用的flag  
```shell
./edgeadm detach edgex –-help
```  

可用的flag显示如下  

```shell
--app          Detach the edgex application-services from cluster.
--core          Detach the edgex core-services from cluster.
--device       Detach the edgex device-services from cluster.
--support       Detach the edgex supporting-services from cluster.
--sysmgmt         Detach the edgex system management from cluster.
--ui              Detach the ui from cluster.
--completely       Detach the configmap and volumes from cluster.
```
  

如需卸载core服务的相关组件，请运行  

```shell
./edgeadm detach edgex –-core
```  

其他组件删除操作同上，替换flag即可，支持多个flag同时删除多个层级的组件。
可以通过以下命令查看所有pod是否已删除。  

```shell
kubectl get pods -n edgex  
```  

**注意**:  

-    如果删除中出现错误，导致某一层级的组件部分已删除，部分未删除，则对该层级重新执行删除操作将失败，需要用addon对该层级所有组件重装，再进行删除
如：删除core层级的过程中遇到失败，导致core-data的组件已删除而core-consul的组件未删除，则`./edgeadm detach edgex –-core`命令无法再次正常重新执行，需要用`./edgeadm addon edgex –-core`补充缺失的core-data组件，再使用`./edgeadm detach edgex –-core`删除core层级。
-    `./edgeadm detach edgex`仅适用于所有组件都存在的情况，如仅存在部分组件，请对相应层级进行独立删除。
  
## 7. 补充  
-   以上提供的安装版本为haoni版本，如需安装其他版本的组件，请拉取仓库源码，并在`/pkg/edgeadm/constant/manifests/edgex`目录下修改对应组件的相关细节。
-   以上安装不包含serity的相关组件和配置，后期版本可能添加相关功能，也可在项目源文件中自行配置。
-   如果使用中遇到相关问题或有改进意见，也可以在SuperEdge社区提Issues一块来修复。
