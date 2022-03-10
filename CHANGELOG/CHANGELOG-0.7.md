# Release v0.7.0 / 2022-03-04

* [Release v0.7.0 / 2022-03-04](#release-v070--2022-03-04)
    * [1. Features and Enhancements](#1-features-and-enhancements)
    * [2.Bug Fixes](#2bug-fixes)
    * [3.Documentation](#3documentation)
    * [4. Demo Examples](#4-demo-examples)
    * [5. Contributor](#5-contributor)

## 1. Features and Enhancements

-   Support NodeUnit and Nodegroup CRD to manage edge site resources. ([#289](https://github.com/superedge/superedge/pull/289) [@attlee-wang](https://github.com/attlee-wang))
    -   Nodeunit add setnode attribute and nodeunit label to node.  ([#306](https://github.com/superedge/superedge/pull/306) [#327](https://github.com/superedge/superedge/pull/327) [@luhaopei](https://github.com/luhaopei))
    -   Add default apply NodeUnit and NodeGroup crd and default nodeunit. ([#325](https://github.com/superedge/superedge/pull/325) [@attlee-wang](https://github.com/attlee-wang))
    -   support nodegroup autofindnodekeys to automatically create nodeunits in batches  ([#334](https://github.com/superedge/superedge/pull/334)  [@JaneLiuL](https://github.com/JaneLiuL))
-   Support an edge applications can be updated when the cloud edge network is disconnected. ([#343](https://github.com/superedge/superedge/pull/343) [@attlee-wang](https://github.com/attlee-wang)   [@luhaopei](https://github.com/luhaopei))
-   Tunnel supports apiserver EgressSelector feature. ([#314](https://github.com/superedge/superedge/pull/314) [@00pf00](https://github.com/00pf00) )
-   Edge-health supports user-defined edge node health check plugins.  ([#317](https://github.com/superedge/superedge/pull/317)  [@JaneLiuL](https://github.com/JaneLiuL))
-   Add Superedge support k8s v1.20.6 ([#271](https://github.com/superedge/superedge/pull/271) [@attlee-wang](https://github.com/attlee-wang))
-   Support containerd runtime install in edgeadm. ( [#322](https://github.com/superedge/superedge/pull/322) [@malc0lm](https://github.com/malc0lm))
-   Support adding edge nodes and native K8s nodes to SuperEdge edge clusters. ([#282](https://github.com/superedge/superedge/pull/282) [@luhaopei](https://github.com/luhaopei))
-   Lite-apiserver supports specifying multiple network interfaces to establish connections with the cloud. ([#263](https://github.com/superedge/superedge/pull/263)  [@huweihuang](https://github.com/huweihuang))
-   Lite-apiserver support use pebble as storage.  ( [#340](https://github.com/superedge/superedge/pull/340) [#341](https://github.com/superedge/superedge/pull/341) [ctlove0523](https://github.com/ctlove0523))
-   Lite-apiserver support pprof debugging. ([#271](https://github.com/superedge/superedge/pull/272) [@00pf00](https://github.com/00pf00)  [#286](https://github.com/superedge/superedge/pull/286) [wangchenglong01](https://github.com/wangchenglong01))

-   Add SuperEdge unit test e2e test.([#315](https://github.com/superedge/superedge/pull/315)  [#321](https://github.com/superedge/superedge/pull/321)  [#324](https://github.com/superedge/superedge/pull/324)  [@JaneLiuL](https://github.com/JaneLiuL))

## 2.Bug Fixes

* FIX wrapper for different k8s version about EndpointSlice ([#205](https://github.com/superedge/superedge/pull/305) [@chenkaiyue](https://github.com/chenkaiyue))
* FIX: keep annotations when application-grid-controller create service from servicegrid. ([#310](https://github.com/superedge/superedge/pull/310) [@jasine](https://github.com/jasine))
* FIx: edgeadm init and join with --enable-edge, and node label mistaken. ( [#333](https://github.com/superedge/superedge/pull/333) [@malc0lm](https://github.com/malc0lm))
* FIX redefined log_dir panic.  ([#346](https://github.com/superedge/superedge/pull/346)  [@huweihuang](https://github.com/huweihuang))
* FIX: Set_file_content func in init-node.sh. ([#308](https://github.com/superedge/superedge/pull/308) [@jasine](https://github.com/jasine))

## 3.Documentation

* Support nodeunit and Nodegroup CRD to manage edge site resources. ([#289](https://github.com/superedge/superedge/pull/289) [@attlee-wang](https://github.com/attlee-wang))
* Tunnel supports apiserver EgressSelector feature using docs. ([#314](https://github.com/superedge/superedge/pull/314) [@00pf00](https://github.com/00pf00) )

## 4. Demo Examples

-   Superedge and fabedge jointly support native service mutual access and podip pass through in edge k8s clusters([#270](https://github.com/superedge/superedge/pull/270) [@00pf00](https://github.com/00pf00) )
-   Add Use GPU in SuperEdge ( [#291](https://github.com/superedge/superedge/pull/291) [@payall4u](https://github.com/payall4u))
-   Add the WasmEdge runtime to SuperEdge to run WebAssembly applications （[#335](https://github.com/superedge/superedge/pull/335) [@malc0lm](https://github.com/malc0lm))

## 5. Contributor

New outstanding contributors members of the SuperEdge community：

-   [@JaneLiuL](https://github.com/JaneLiuL)
-   [@malc0lm](https://github.com/malc0lm)
-   [@luhaopei](https://github.com/luhaopei)
-   [@huweihuang](https://github.com/huweihuang)

