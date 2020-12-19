<!-- BEGIN MUNGE: GENERATED_TOC -->
- [v0.1.0](#v010)
  - [Downloads for v0.1.0](#downloads-for-v010)
    - [Binaries](#binaries)
  - [Changelog](#changelog)
    - [Other notable changes](#other-notable-changes)
<!-- END MUNGE: GENERATED_TOC -->

<!-- NEW RELEASE NOTES ENTRY -->
## Downloads for v0.1.0

### Binaries

filename | md5
-------- | -----------
[superedge-linux-amd64.tar.gz](https://github.com/superedge/superedge/releases/download/v0.1.0/edgeadm-linux-amd64-v0.1.0.tgz) | `c65ed328e437360aecb0752cd80421e5`
[superedge-linux-arm.tar.gz](https://github.com/superedge/superedge/releases/download/v0.1.0/edgeadm-linux-arm64-v0.1.0.tgz) | `b3ae69b7cad8fded74f8cb2e9d7fbe25`

## SuperEdge v0.1.0 Release Notes

### 0.1.0 What's New

This is the first release of SuperEdge, which includes a series of enhancements for transforming Kubernetes cluster to an edge-enabled cluster.

**Tunnel**

* Proxies requests from master nodes to edge worker nodes, which allows accessing edge nodes with no public IP address.
* It is enabled by default for request from kube-apiserver to Kubelet on the Edge Worker ndoes.
* Supports TCP/HTTP/HTTPS

**ServiceGroup**

* Enhanced application management on the edge by grouping related applications and edge worker nodes.   
* Offers two new CRDs for managing serviceGroup - DeploymentGrid & ServiceGrid.
* Provides isolated internal network for applications inside the same serviceGroup.

**Edge Autonomy**

* Keeps services on the edge nodes up and running stably despite of network disconnection between edge and cloud.
* Allows pods on the edge to restart correctly even though losing connection to the master kube-apiserver.  

**EdgeHealth**

* Enhanced monitoring of edge worker nodes.
* Distributed health check on groups of edge ndoes.
* Determines status of the node based on the status checks from nodes within the same health check group.
* Eliminates false alert in case of losing cloud-to-edge network connection.

**EdgeAdm**

* EdgeAdm helps users to deploy SuperEdge to native Kubernetes in 1-click.
