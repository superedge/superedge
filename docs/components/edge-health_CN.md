# edge-health

edge-health运行在每个边缘节点上，用于节点存活性探测

### 作用
- 在节点互通的弱网环境下，互相探测，相互感知，从本边缘节点的角度判定其他边缘节点的健康状况；
- 将探测结果上报给edge-health admission, 影响边缘服务的驱逐，从而保证断网节点Workload依旧正常运行；
- 支持对节点分区域检查，保证每个分区的健康性独立处理；

### 工作原理
- 每个edge-health周期性探测所有边缘节点，并投票确定节点的健康状态；
- edge-health admission综合边缘节点的状态和edge-health的探测结果，影响边缘负载的驱逐，从而保证弱网环境下，断网节点workload的正常运行;
- 支持节点分组/区域检查，解决边缘节点往往跨区域，但各区域间节点互不相通的问题；


### 可选参数
#### 探测相关 
- healthcheckperiod 探测周期，单位为秒，默认为10，
- healthcheckscoreline 判定节点正常所需要达到的分数线，默认为100
- 支持以下参数：
    - timeout指定一次请求的超时时间
    - retrytime 指定探测失败的重试次数
    - weight 指定该探测方式所占探测分数的比重，最高为1
    - port 指定探测的端口
- 当前支持的两个插件：
    - 检测kubelet安全认证端口:
        `--kubeletauthplugin=timeout=5,retrytime=3,weight=1,port=10250`
    - 检测kubelet非安全认证端口：
        `--kubeletplugin=timeout=5,retrytime=3,port=10255,weight=1`

#### 通信交互相关
- communicateperiod 指定交互周期，单位为秒，默认为10
- communicatetimetout 指定发送交互信息超时时间，单位为秒，默认为3
- communicateretrytime 指定发送交互信息失败时的重试次数，默认为1
- communicateserverport 接受和发送交互信息的端口，默认为51005

#### 投票相关
- voteperiod 指定投票周期，单位为秒，默认为10
- votetimeout 指定投票有效时间，超过此时长为无效投票，单位为秒，默认为60

#### 其他：
- masterurl,kubeconfig 用于连接apiserver;当前默认使用InCluster方式,不用填写
- hostname 当前节点的名称;默认自动获取部署yaml中的nodeName字段，不用填写


### 多地域探测
- 开启:
    - 将节点按照地域打上`check-units:<units>`的label, 其中`<units>`可以指定多个站点名称，以半角逗号`,`分割, 例如`check-units: unit1,unit2,unit3`.
    - 在`edge-system`命名空间创建名为`edge-health-config`的configmap，`unit-internal-check`的值指定为`true`,可以直接使用如下yaml进行创建
        ```yaml
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: edge-health-config
          namespace: edge-system
        labels:
          name: edge-health
        data:
          unit-internal-check: true
          check-units: unit1,unit2,unit3
        ```

- 关闭
    - 将`unit-internal-check`的值改为`false`或者删除`edge-health-config`的configmap
    
> 注意：如果开启了多地域但是没有给节点打上地域标签，则该节点探测时只会检测自己