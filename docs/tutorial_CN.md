## 使用教程

### 源码编译
您可以选择使用SuperEdge Release的版本，也可以根据需要使用源代码编译出符合您需求的版本

#### 1. 选择Release版本
- [版本列表](../CHANGELOG/README.md)

#### 2. 我要编译源代码

- format, build and test: `make`

- build: `make build`

- build some cmd: `make build BINS="lite-apiserver"`

- build multi arch cmd: `make build.multiarch BINS="lite-apiserver" PLATFORMS="linux_amd64 linux_arm64"`

- build docker image: `make image IMAGES="application-grid-controller application-grid-wrapper"`

- build multi arch docker image: `make image.multiarch IMAGES="application-grid-controller application-grid-wrapper" PLATFORMS="linux_amd64 linux_arm64" VERSION="v2.0.0"`

- push mainfest: `make manifest.multiarch IMAGES="application-grid-wrapper application-grid-controller" PLATFORMS="linux_amd64 linux_arm64" REGISTRY_PREFIX="docker.io/superedge" VERSION="v2.0.0"`

- clean: `make clean`

### 我想深入了解原理
#### 1. 组件及特性解析
- [lite-apiserver原理](./components/lite-apiserver_CN.md)
- [service-group原理](./components/serviceGroup_CN.md)
- [edge-health](./components/edge-health_CN.md)
- [使用edgeadm一键安装边缘K8s集群](https://mp.weixin.qq.com/s/zHs_qmD8781r-h4tkie0qQ)
- [打破内网壁垒，从云端一次添加成百上千的边缘节点](https://mp.weixin.qq.com/s/JmzQuiBBkNwS9hpS0hIg7A)
- [SuperEdge 云边隧道新特性：从云端SSH运维边缘节点](https://mp.weixin.qq.com/s/J-sxkiL62FAjGBRHERPbKg)
- [Addon SuperEdge 让原生 K8s 集群可管理边缘应用和节点](https://mp.weixin.qq.com/s/1CnvqASzLnOShj8Hoh-Trw)
- [配置tunnel-cloud HPA](./components/tunnel-cloud-hpa_CN.md)
- [如何使用Prometheus采集边缘metrics](./components/deploy_monitor_CN.md)

#### 2. 技术文章

  - [【从0到1学习边缘容器系列】之 边缘计算与边缘容器的起源](https://mp.weixin.qq.com/s/D0yYtBSAOjJa1LnIr6rTLQ)
  - [【从0到1学习边缘容器系列】之 边缘应用管理](https://mp.weixin.qq.com/s/MUSNACSkeoxAlViltXPO7A)
  - [【从0到1学习边缘容器系列-3】应用容灾之边缘自治](https://mp.weixin.qq.com/s/GbPDdy4u6j5PDrT8Zpr05w)
  - [边缘计算场景下云边端一体化的挑战与实践](https://mp.weixin.qq.com/s/rCA6AKQ7CCZ6Zu81olDVDQ)
  - [一文读懂 SuperEdge 边缘容器架构与原理](https://mp.weixin.qq.com/s/V29ga-fOM2KEq-dlKo-FuA)
  - [一文读懂 SuperEdge 拓扑算法](https://mp.weixin.qq.com/s/oK7E_USE23Hdp5i1fHN_Tw)
  - [一文读懂 SuperEdge 分布式健康检查 (边端)](https://mp.weixin.qq.com/s/E3kBBxfV6_TvNZj5IGkAvQ)
  - [一文读懂 SuperEdge 云边隧道](https://mp.weixin.qq.com/s/5btXwUot0vSGvUlzVcofLg)
  - [SuperEdge 如何支持多地域 StatefulSets 及灰度](https://mp.weixin.qq.com/s/PBGA5Rd-LVKLZawpjHL_Eg)
  - [从 lite-apiserver 看 SuperEdge 边缘节点自治](https://mp.weixin.qq.com/s/kRmkiOVWCwVvhp4veqWWpA)
  - [SuperEdge 高可用云边隧道有哪些特点？](https://mp.weixin.qq.com/s/RId4f-ia326-9wFn4VKE2w)

#### 3. 实践案例
- [完爆！用边缘容器，竟能秒级实现团队七八人一周的工作量](https://mp.weixin.qq.com/s/FMO6V1pvG-Xyi9xfBttCQA)
- [使用TKE Edge部署EdgeX Foundry](https://mp.weixin.qq.com/s/0OOBazTMJQh4SXItNaVIMQ)
- [基于边缘容器技术的工业互联网平台建设](https://mp.weixin.qq.com/s/And8uUFxJZZeTJM_e_7pDA)
- [腾讯WeMake工业互联网平台的边缘容器化实践：打造更高效的工业互联网](https://mp.weixin.qq.com/s/evalqNiqoM2dly57A0Cgrg)

#### 4. 易学易用系列视频
- [云+社区](https://cloud.tencent.com/developer/user/5016738)
- [B站](https://space.bilibili.com/1803883492/channel/detail?cid=191686)
