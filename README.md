English | [简体中文](./README_CN.md)

# SuperEdge

## What is SuperEdge?

SuperEdge is an open source **container management system for edge computing** to manage compute resources and container applications in multiple edge regions. These resources and applications, in the current approach, are managed as one single **Kubernetes** cluster. A native Kubernetes cluster can be easily converted to a SuperEdge cluster.

SuperEdge has the following characteristics:

* **Kubernetes-native**: SuperEdge extends the powerful container orchestration and scheduling capabilities of Kubernetes to the edge. It makes nonintrusive enhancements to Kubernetes and is fully compatible with all Kubernetes APIs and resources. Kubernetes users can leverage SuperEdge easily for edge environments with minimal learning.
* **Edge autonomy**: SuperEdge provides L3 edge autonomy. When the network connection between the edge and the cloud is unstable, or the edge node is offline, the node can still work independently. This eliminates the negative impact of unreliable network.
* **Distributed node health monitoring**: SuperEdge provides edge-side health monitoring capabilities. SuperEdge can continue to monitor the processes on the edge side and collect health information for faster and more accurate problem discovery and reporting. In addition, its distributed design can provide multi-region monitoring and management.
* **Built-in edge orchestration capability**: SuperEdge supports automatic deployment of multi-regional microservices. Edge-side services are closed-looped, and it effectively reduces the operational overhead and improves the fault tolerance and availability of the system.
* **Network tunneling**: SuperEdge ensures that Kubernetes nodes can operate under different network situations. It supports network tunnelling using TCP, HTTP and HTTPS.

SuperEdge was initiated by the following companies: Tencent, Intel, VMware, Huya, Cambricon, Captialonline and Meituan.


## Architecture
<div align="center">
  <img src="docs/img/superedge_arch.png" width=80% title="SuperEdge Architecture">
</div>

### Cloud components:
* [**tunnel-cloud**](docs/components/tunnel.md): Maintains a persistent network connection to `tunnel-edge` services. Supports TCP/HTTP/HTTPS network proxies.
* [**application-grid controller**](docs/components/service-group.md): A Kubernetes CRD controller as part of ServiceGroup. It manages DeploymentGrids and ServiceGrids CRDs and control applications and network traffic on edge worker nodes.
* [**edge-
admission**](docs/components/edge-health.md): Assists Kubernetes controllers by providing real-time health check status from `edge-health` services distributed on all edge worker nodes.

### Edge components:
* [**lite-apiserver**](docs/components/lite-apiserver.md): Lightweight kube-apiserver for edge autonomy. It caches and proxies edge components' requests and critical events to cloud kube-apiserver.
* [**edge-health**](docs/components/edge-health.md): Monitors the health status of edge nodes in the same edge region.
* [**tunnel-edge**](docs/components/tunnel.md): Maintains persistent connection to `tunnel-cloud` to retrieve API requests to the controllers on the edge.
* application-grid wrapper: Managed by `application-grid controller` to provide independent internal network space for services within the same ServiceGrid.

## Quickstart Guide
For installation, deployment, and administration, see our [**Tutorial**](docs/installation/tutorial.md).

## Contact
For any question or support, feel free to contact us via:
- [mailing list](https://groups.google.com/g/superedge)
- Slack：[superedge-community](https://join.slack.com/t/superedge-workspace/shared_invite/zt-k1kekpdz-jih6w8RByoylnfTmCTZpkA)

## Contributing
Welcome to [contribute](./CONTRIBUTING.md) and improve SuperEdge

## License

[**LICENSE**](./LICENSE)

