- [v0.1.0](#release-v010-2020-12-19)
  - [Features](#features)

# Release v0.1.0 / 2020-12-19

🎉🎉🎉 First Release!

This is the first release of SuperEdge, which includes a series of features to transform a Kubernetes cluster to an edge-native container management system for edge computing.

## Features

**Tunnel**

* Proxies requests from master nodes to edge worker nodes, which allows accessing edge nodes with no public IP address.
* It is enabled by default for request from kube-apiserver to Kubelet on the Edge Worker nodes.
* Supports TCP/HTTP/HTTPS.

**ServiceGroup**

* Enhanced application management on the edge by grouping related applications and edge worker nodes.   
* Offers two new CRDs for managing serviceGroup - DeploymentGrid & ServiceGrid.
* Provides isolated internal network for applications inside the same serviceGroup.

**Edge Autonomy**

* Keeps services on the edge nodes up and running stably despite of network disconnection between edge and cloud.
* Allows pods on the edge to restart correctly even though losing connection to the master kube-apiserver.  

**EdgeHealth**

* Enhanced monitoring of edge worker nodes.
* Distributed health check on groups of edge nodes.
* Determines status of the node based on the status checks from nodes within the same health check group.
* Eliminates false alert in case of losing cloud-to-edge network connection.

**EdgeAdm**

* EdgeAdm is a command-line tool to help users to deploy SuperEdge to native Kubernetes in 1-click.
