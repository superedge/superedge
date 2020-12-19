## 开发者指南
### make
+ format, build and test
***make***
+ build
***make build***
+ build some cmd
***make build BINS="lite-apiserver edgeadm"***
+ build multi arch cmd
***make build.multiarch BINS="lite-apiserver edgeadm" PLATFORMS="linux_amd64 linux_arm64"***
+ build docker image
***make image IMAGES="application-grid-controller application-grid-wrapper"***
+ build multi arch docker image
***make image.multiarch IMAGES="application-grid-controller application-grid-wrapper" PLATFORMS="linux_amd64 linux_arm64" VERSION="v2.0.0"***
+ push mainfest
***make manifest.multiarch IMAGES="application-grid-wrapper application-grid-controller" PLATFORMS="linux_amd64 linux_arm64" REGISTRY_PREFIX="docker.io/superedge" VERSION="v2.0.0"***
+ clean
***make clean***

### 使用 edgeadm 部署

+ 将普通集群转换成边缘集群
***edgeadm change --kubeconfig admin.kubeconfig***
+ 将边缘集群回退成普通集群
***edgeadm revert --kubeconfig admin.kubeconfig***
- [**edgeadm 详细使用**](./install_edgeadm_CN.md)

### 手工部署
- [**手工部署**](./install_manual_CN.md)
