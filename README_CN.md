简体中文 | [English](./README.md)

# SuperEdge

## 什么是SuperEdge?

SuperEdge是Kubernetes原生的边缘容器方案，它将Kubernetes强大的容器管理能力扩展到边缘计算场景中，针对边缘计算场景中常见的技术挑战提供了解决方案，如：单集群节点跨地域、云边网络不可靠、边缘节点位于NAT网络等。这些能力可以让应用很容易地部署到边缘计算节点上，并且可靠地运行。

SuperEdge可以帮助您很方便地把分布在各处的计算资源放到一个Kubernetes集群中管理，包括但不限于：边缘云计算资源、私有云资源、现场设备，打造属于您的边缘PaaS平台。 

SuperEdge支持所有Kubernetes资源类型、API接口、使用方式、运维工具，无额外的学习成本。也兼容其他云原生项目，如：Promethues，使用者可以结合其他所需的云原生项目一起使用。

SuperEdge项目由以下公司共同发起：腾讯、Intel、VMware、虎牙直播、寒武纪、首都在线和美团。


## 项目特性
SuperEdge具有如下特性:

- **Kubernetes 原生**：SuperEdge 以无侵入的方式将 Kubernetes 强大的容器编排、调度能力拓展到边缘端，其原生支持 Kubernetes，完全兼容 Kubernetes 所有 API 及资源，无额外学习成本
- **边缘自治**：SuperEdge 提供 L3 级边缘自治能力，当边缘节点与云端网络连接不稳定或处于离线状态时，边缘节点可以自主工作，化解了网络不可靠所带来的不利影响
- **分布式节点健康监测**：SuperEdge 是业内首个提供边缘侧健康监测能力的开源容器管理系统。SuperEdge 能在边缘侧持续守护进程，并收集节点的故障信息，实现更加快速和精准的问题发现与报告。此外，其分布式的设计还可以实现多区域、多范围的监测和管理
- **内置边缘编排能力**：SuperEdge 能够自动部署多区域的微服务，方便管理运行于多个地区的微服务。同时，网格内闭环服务可以有效减少运行负载，提高系统的容错能力和可用性
- **内网穿透**：SuperEdge 能够保证 Kubernetes 节点在有无公共网络的情况下都可以连续运行和维护，并且同时支持传输控制协议（TCP）、超文本传输协议（HTTP）和超文本传输安全协议（HTTPS）


## 体系架构
<div align="center">
  <img src="docs/img/superedge_arch.png" width=80% title="SuperEdge Architecture">
</div>

云端组件：
- [**tunnel-cloud**](docs/components/tunnel_CN.md): 云端tunnel服务组件，用于建立云边长连接隧道，支持代理tcp/http/https流量
- [**application-grid controller**](docs/components/service-group_CN.md): 应用网络（serviceGroup）控制器
- [**edge-health admission**](docs/components/edge-health_CN.md): 分布式节点健康检查机制云端组件，辅助Kubernetes控制器工作

边端组件:
- [**lite-apiserver**](docs/components/lite-apiserver_CN.md): 节点侧轻量版apiserver shadow，代理节点组件到云端apiserver的请求，缓存关键数据以用于边缘自治
- [**edge-health**](docs/components/edge-health_CN.md): 分布式节点健康检查，用于感知边缘节点状态，支持对节点分区域检查能力
- [**tunnel-edge**](docs/components/tunnel_CN.md): 边缘tunnel服务组件，主动与tunnel-cloud建立长连接，将云端请求代理到对应的边缘服务，如：kubelet、业务pod等
- [**application-grid wrapper**](docs/components/service-group_CN.md): 应用网格流量控制组件，可将svc之间的流量闭环在同一个应用网格之中，避免跨网格访问

## 快速入门指南
关于安装、部署和管理，请参见[**教程**](docs/installation/tutorial_CN.md)。

## 联系
如果您有任何疑问或需要支持，请随时与我们联系：
- [**邮件列表**](https://groups.google.com/g/superedge)
- Slack：[superedge-community](https://join.slack.com/t/superedge-workspace/shared_invite/zt-k1kekpdz-jih6w8RByoylnfTmCTZpkA)

## 参与项目
欢迎您[参与](./CONTRIBUTING.md)完善项目

## 许可证
[**许可证**](./LICENSE)
