- [Release v0.4.0 / 2021-06-18](#release-v040--2021-06-18)
  - [Features and Enhancements](#features-and-enhancements)
  - [Bug Fixes](#bug-fixes)
  - [Documentation](#documentation)

# Release v0.4.0 / 2021-06-18

## Features and Enhancements

* Penetrator. It is used to batch add and reload edge nodes from the cloud([#121](https://github.com/superedge/superedge/pull/121),[@00pf00](https://github.com/00pf00)). The user submits nodetask, a Kubernetes CRD, and then penetrator batch adds and reloads edge nodes in the background, even if the nodes can only be accessed one-way. [usage](https://github.com/superedge/superedge/blob/main/docs/installation/addnode_via_penetartor.md)
* Support SSH login to edge nodes from cloud, even if the nodes can only be accessed one-way([#140](https://github.com/superedge/superedge/pull/140),[@00pf00](https://github.com/00pf00)). [usage](https://github.com/superedge/superedge/blob/main/docs/components/edge-node-ops.md)
* Servicegroup supports multi cluster application distribution([#139](https://github.com/superedge/superedge/pull/139),[@chenkaiyue](https://github.com/chenkaiyue)). Nodeunits under the same nodegroup can be distributed in different clusters and need to be used together with clusternet. [usage](https://github.com/superedge/superedge/blob/main/docs/components/serviceGroup_CN.md#%E4%BD%BF%E7%94%A8%E7%A4%BA%E4%BE%8B)
* SuperEdge can install into native Kubernetes cluster as add-on([#129](https://github.com/superedge/superedge/pull/129),[@lianghao208](https://github.com/lianghao208)). [usage](https://github.com/superedge/superedge/blob/main/docs/installation/addon_superedge.md)
* Fetch the NewInitNodePhase() function directly into an `init-node.sh` script([#138](https://github.com/superedge/superedge/pull/138),[@k2let](https://github.com/k2let)).
* Upgrade k8s.io/klog into k8s.io/klog/v2([#136](https://github.com/superedge/superedge/pull/136),[@attlee-wang](https://github.com/attlee-wang)).

## Bug Fixes

* Modify edgeadm, to disable selinux([#132](https://github.com/superedge/superedge/pull/132), [@k2let](https://github.com/k2let)).
* Lite-apiserver: add modify-request-accept param([#126](https://github.com/superedge/superedge/pull/126), [@Beckham007](https://github.com/Beckham007)).
* Check node name parameter, the IPv4 format is not supported([#125](https://github.com/superedge/superedge/pull/125), [@zhhray](https://github.com/zhhray)).


## Documentation

* Add hostname requirements to deployment docs([#150](https://github.com/superedge/superedge/pull/150), [@jasine](https://github.com/jasine)).
* Add edge-health deploy([#141](https://github.com/superedge/superedge/pull/141), [@attlee-wang](https://github.com/attlee-wang)).
