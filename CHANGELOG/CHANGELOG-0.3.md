- [v0.3.0 / 2021-05-20](#v030--2021-05-20)
  - [Features and Enhancements](#features-and-enhancements)
  - [Bug Fixes](#bug-fixes)
  - [Documentation](#documentation)

# v0.3.0 / 2021-05-20

## Features and Enhancements

* Edgeadm supported creating native Kubernetes cluster with pre-installed SuperEdge components([#97](https://github.com/superedge/superedge/pull/97), [@attlee-wang](https://github.com/attlee-wang)), [Usage](https://github.com/superedge/superedge/blob/main/docs/installation/install_edge_kubernetes.md) or [Quickstart Guide](https://github.com/superedge/superedge/blob/main/README.md).
* Edgeadm supported adding new nodes to edge Kubernetes cluster([#87](https://github.com/superedge/superedge/pull/87), [@wenhuwang](https://github.com/wenhuwang)).
* Enhanced edge autonomy by detecting failing nodes within the offline regions, so as to avoid routing traffic to abnormal nodes([#86](https://github.com/superedge/superedge/pull/86), [@chenkaiyue](https://github.com/chenkaiyue))
* Supported Golang 1.16([#91](https://github.com/superedge/superedge/pull/91), [@SataQiu](https://github.com/SataQiu)).


## Bug Fixes

* Fixed a bug that caused kubectl cp failing to disconnect after copying([#78](https://github.com/superedge/superedge/pull/78), [@00pf00](https://github.com/00pf00)).


## Documentation

* Added quickstart guide to create edge Kubernetes cluster using Edgeadm.