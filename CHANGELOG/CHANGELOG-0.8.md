# Release v0.8.0 / 2022-07-20

* [Release v0.8.0 / 2022-07-20](#release-v080--2022-07-20)
    * [1. Features and Enhancements](#1-features-and-enhancements)
    * [2. Bug Fixes](#2bug-fixes)
    * [3. Contributor](#3-contributor)

## 1. Features and Enhancements

- Support Kubernetes 1.22.6 version. ([#368](https://github.com/superedge/superedge/pull/368) [#371](https://github.com/superedge/superedge/pull/371))
-	Separate `edgeadm` from superedge project for easy maintenance. Please refer to the sub-project: [edgeadm](https://github.com/superedge/edgeadm)
- Support SuperEdge cluster in TKEStack project. Please refer to this pr in [TKE](https://github.com/tkestack/tke) ([#1994](https://github.com/tkestack/tke/pull/1994))
-	Tunnel support http_proxy ability to enable edge/edge edge/cloud communication. ([#374](https://github.com/superedge/superedge/pull/374))
-	Lite-apiserver support to cache some critical resource（node, service, etc) at edge node with the purpose of reducing network traffic. ([#396](https://github.com/superedge/superedge/pull/396) [#400](https://github.com/superedge/superedge/pull/400))
-	application-grid-wrapper support to watch `Ingress` resource to enable `nginx-ingress-controller` at NodeUnit. ([#411](https://github.com/superedge/superedge/pull/411))
-	Add parameters to skip the verification of the domain name and ip signed by the apiserver certificate. ([#382](https://github.com/superedge/superedge/pull/382))
-	Tunnel supports ExternalName service forwarding. ([#388](https://github.com/superedge/superedge/pull/388))
-	Lite-apiserver can skip ca verification for serviceaccount. ([#391](https://github.com/superedge/superedge/pull/391))
-	AMR32 Platform building support .([#362](https://github.com/superedge/superedge/pull/362))

## 2.Bug Fixes

* site-manager: FIX NodeUnit and NodeGroup about setnode ([#364](https://github.com/superedge/superedge/pull/364))
* Fix lite-apiserver problem: write cache error ([#366](https://github.com/superedge/superedge/pull/366))
* application-grid-controller: Fix metadata.resourceVersion: Invalid value([#369](https://github.com/superedge/superedge/pull/369))
* edgeadm change kubeproxy configmap namespace([#392](https://github.com/superedge/superedge/pull/392))
* edgeadm should not create endpointslice: 'kubernetes-no-edge'([#397](https://github.com/superedge/superedge/pull/397))
* Solve the problem of network error causing cache return failure ([#402](https://github.com/superedge/superedge/pull/402))
* Kubernetes 1.22 version: the edge node will obtain the fake token when the apiserver is inaccessible ([#419](https://github.com/superedge/superedge/pull/419))

## 3. Contributor

In  v0.8.0 verison, we don't add more big features. The main purpose of this version is to enhance the project's usability and stability. Thanks for the following contributors: 

-   [@00pf00](https://github.com/00pf00)
-   [@malc0lm](https://github.com/malc0lm)
-   [@attlee-wang](https://github.com/attlee-wang)
-   [@dodiadodia](https://github.com/dodiadodia)

