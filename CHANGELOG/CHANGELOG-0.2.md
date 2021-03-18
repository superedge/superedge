- [v0.2.0](#release-v020-2021-03-19)
    - [Features and Enhancements](#features-and-enhancements)
    - [Bug Fixes](#bug-fixes)
    - [Documentation](#documentation)

# Release v0.2.0 / 2021-03-19

## Features and Enhancements

* lite-apiserver now handles certificate rotation automatically. ([#44](https://github.com/superedge/superedge/pull/44), [@Beckham007](https://github.com/Beckham007))
* Added 3 new caching storage options to lite-apiserver, including memory and local KV storage using Badger or Bolt. ([#53](https://github.com/superedge/superedge/pull/53), [@Beckham007](https://github.com/Beckham007))
* Added new StatefulsetGrid ServiceGroup resource for running statefulset workload. Headless service is supported for StatefulsetGrid. ([#37](https://github.com/superedge/superedge/pull/37), [@duyanghao](https://github.com/duyanghao))
* Added canary deployment support for DeploymentGrid and StatefulSetGrid workloads. ([#50](https://github.com/superedge/superedge/pull/50), [@chenkaiyue](https://github.com/chenkaiyue))
* Added reuse policy for lite-apiserver proxy. ([#43](https://github.com/superedge/superedge/pull/43), [@Beckham007](https://github.com/Beckham007))


## Bug Fixes

* Fixed a bug in Tunnel that causes missing request body in HTTP POST forwarding. ([#47](https://github.com/superedge/superedge/pull/47), [@luoyunhe](https://github.com/luoyunhe))
* Fixed TCP proxy forwarding failure. ([#60](https://github.com/superedge/superedge/pull/60), [@00pf00](https://github.com/00pf00))

## Documentation

* Introduction to new StatefulsetGrid resource.
* Introduction to canary deployment for DeploymentGrid and StatefulSetGrid.
* Tunnel configuration guide and sample configs.
* Updated lite-apiserver architecture diagram and added usage example.
