* [Release v0.6.0 / 2021-09-28](#release-v060--2021-09-28)
   * [Features and Enhancements](#features-and-enhancements)
   * [Bug Fixes](#bug-fixes)
   * [Documentation](#documentation)

# Release v0.6.0 / 2021-09-28

## Features and Enhancements

* EdgeX Foundry: Integrate one-click deployment of native EdgeX Foundry components in edgeadm addon to support the access of edge devices([#229](https://github.com/superedge/superedge/pull/229),[@OmigaXm](https://github.com/OmigaXm))
* Integrate one-click deployment topolvm in edgeadm addon, users can experience the CSI capabilities of Local PV dynamic configuration PV and dynamic expansion PV([#247](https://github.com/superedge/superedge/pull/247),[@attlee-wang](https://github.com/attlee-wang))
* Add the feedback of ServiceGroup deployment status and events to increase the convenience for users to operate and maintain the application status of edge sites;([#240](https://github.com/superedge/superedge/pull/240),[@chenkaiyue](https://github.com/chenkaiyue))
* Add 3 demos that use ecological components in SuperEdge
    * Deploy Tars demo on SuperEdge to help users use Tars development framework on edge sites;([#243](https://github.com/superedge/superedge/pull/243),[@XIAOYUELIN](https://github.com/XIAOYUELIN))
    * Add an demo of using Tengine + SuperEdge to deploy edge AI applications across platforms to help users use AI-related frameworks in SuperEdge([#243](https://github.com/superedge/superedge/pull/243),[@XIAOYUELIN](https://github.com/XIAOYUELIN))
    * Add an demo of collecting edge application monitoring data so that users can access edge application monitoring([#232](https://github.com/superedge/superedge/pull/232),[@XIAOYUELIN](https://github.com/XIAOYUELIN))

## Bug Fixes

* ServiceGroup: use template templateHasher modified to reconcile;([#240](https://github.com/superedge/superedge/pull/240),[@chenkaiyue](https://github.com/chenkaiyue))
* ServiceGroup: fix event scheme used in  ServiceGroup;([#240](https://github.com/superedge/superedge/pull/240),[@chenkaiyue](https://github.com/chenkaiyue))
* Use multiple pipes to read an io.ReaderClose at the same time;;([#252](https://github.com/superedge/superedge/pull/252),[@00pf00](https://github.com/00pf00))


## Documentation

* Added the use document of edgex found on Superedge([#234](https://github.com/superedge/superedge/pull/234),[@OmigaXm](https://github.com/OmigaXm))
* Added topolvm use document([#254](https://github.com/superedge/superedge/pull/254),[@OmigaXm](https://github.com/OmigaXm))
* Added MAINTAINERS and SECURITY file ([#244](https://github.com/superedge/superedge/pull/244),[@attlee-wang](https://github.com/attlee-wang))
