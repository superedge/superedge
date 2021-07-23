- [Release v0.5.0 / 2021-07-23](#release-v050--2021-07-23)
  - [Features and Enhancements](#features-and-enhancements)
  - [Bug Fixes](#bug-fixes)
  - [Documentation](#documentation)

# Release v0.5.0 / 2021-07-23

## Features and Enhancements

* Edgeadm: add kube-vip as default HA component support([#186](https://github.com/superedge/superedge/pull/186),[@lianghao208](https://github.com/lianghao208))
* TunnelCloud HPA: add tunnel-cloud metrics for HPA([#187](https://github.com/superedge/superedge/pull/187),[@00pf00](https://github.com/00pf00)), support tunnel-cloud HPA based on the number of connected edge nodes([#189](https://github.com/superedge/superedge/pull/189),[@00pf00](https://github.com/00pf00)). [Usage](https://github.com/superedge/superedge/blob/main/docs/components/tunnel-cloud-hpa_CN.md)
* Deploy Prometheus([#189](https://github.com/superedge/superedge/pull/189),[@00pf00](https://github.com/00pf00)), [Usage](https://github.com/superedge/superedge/blob/main/docs/components/deploy-monitor_CN.md)
* TunnelCloud DNS: Separate the synchronization of podIP and Endpoints([#177](https://github.com/superedge/superedge/pull/177),[@00pf00](https://github.com/00pf00))
* Edgeadm: Deploy coredns for every edge node by servicegroup([#185](https://github.com/superedge/superedge/pull/185),[@attlee-wang](https://github.com/attlee-wang))

## Bug Fixes

* Edgeadm: Fix x509 certificate bug when joining edge nodes([#174](https://github.com/superedge/superedge/pull/174),[@lianghao208](https://github.com/lianghao208))
* Edgeadm: Fix tunnel-coredns addr when join master([#176](https://github.com/superedge/superedge/pull/176),[@attlee-wang](https://github.com/attlee-wang))




## Documentation

* Add technical articles related to SuprEdge([#188](https://github.com/superedge/superedge/pull/188),[@neweyes](https://github.com/neweyes))
* Add proposal template([#199](https://github.com/superedge/superedge/pull/199),[@neweyes](https://github.com/neweyes))
