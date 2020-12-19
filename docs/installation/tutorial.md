## Tutorial

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

### Use edgeadm to install SuperEdge

+ Convert normal Kubernetes cluster to edge Kubernetes cluster
***edgeadm change --kubeconfig admin.kubeconfig***
+ Revert edge Kubernetes cluster to normal Kubernetes cluster
***edgeadm revert --kubeconfig admin.kubeconfig***
- [**More on edgeadm**](./install_via_edgeadm.md)

### Manual installation
- [**Manual installation**](./install_manually.md)
