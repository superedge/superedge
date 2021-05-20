- [Release v0.3.0 / 2021-05-20](#release-v030--2021-05-20)
  - [Features and Enhancements](#features-and-enhancements)
  - [Bug Fixes](#bug-fixes)
  - [Documentation](#documentation)

# Release v0.3.0 / 2021-05-20

## Features and Enhancements

* edgeadm tool supports to create a new edge Kubernetes cluster([#97](https://github.com/superedge/superedge/pull/97), [@attlee-wang](https://github.com/attlee-wang)), and its [usage](https://github.com/superedge/superedge/blob/main/docs/installation/install_edge_kubernetes.md) is consistent with Kubedm, or create an edge Kubernetes cluster follow by [Quickstart Guide](https://github.com/superedge/superedge/blob/main/README.md).
* edgeadm tool supports adding new nodes to the edge kubernetes cluster([#87](https://github.com/superedge/superedge/pull/87), [@wenhuwang](https://github.com/wenhuwang)). Before superedge v0.3.0, we can use `edgeadm change` to convert the existing original kubernetes cluster to the edge kubernetes cluster, but we can't add new nodes after change.
* when the node is offline, it can also sense the state of other nodes and identify whether the back-end pod is reachable, so as to avoid routing traffic to abnormal nodes([#86](https://github.com/superedge/superedge/pull/86), [@chenkaiyue](https://github.com/chenkaiyue))
* support golang 1.16([#91](https://github.com/superedge/superedge/pull/91), [@SataQiu](https://github.com/SataQiu)).


## Bug Fixes

* Fixed `kubectl cp` connection can not be disconnected after copy file over([#78](https://github.com/superedge/superedge/pull/78), [@00pf00](https://github.com/00pf00)).


## Documentation

* update roadmap([roadmap](https://github.com/superedge/superedge/blob/main/docs/roadmap.md), [@ruyingzhe](https://github.com/ruyingzhe))
* Introduction to use edgeadm to create an edge Kubernetes cluster.
* Fixed wrrong information in install_manually docs([#73](https://github.com/superedge/superedge/pull/73), [@huweihuang](https://github.com/huweihuang)).
