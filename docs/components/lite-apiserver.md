# lite-apiserver

`lite-apiserver` is a light-weight version of the kube-apiserver running on the edge nodes. It acts as a proxy for requests from all components and pods on the edge node to the cloud apiserver, and caches the responses to achieve edge autonomy in case of the disconnected cloud -edge network.

## Architecture
<div align="left">
  <img src="../img/lite-apiserver.png" width=70% title="lite-apiserver Architecture">
</div>

`lite-apiserver` has the following functionalities:
- Caches request for all edge components (kubelet, kube-proxy, etc.) and pods running on edge nodes
- Provides various authentication mechanism for edge components and pods including X509 Client Certs, Bearer Token, etc.
- Caches requests from all kind of Kubernetes resources, including build-in Kubernetes resource and custom CRDs

## Installation
`lite-apiserver` can be run at the edge as Kubernetes pod or systemd service. See [**Installation Guide**]() to get more detail

## Options

#### Server
- ca-file: kubernetes cluster CA or CA used by cloud kube-apiserver
- tls-cert-file: TLS cert for lite-apiserver
- tls-private-key-file: TLE key for lite-apiserver
- kube-apiserver-url: URL of cloud kube-apiserver
- kube-apiserver-port: Port of cloud kube-apiserver
- port: https listener port of lite-apiserver, default to 51003
- file-cache-path: file path of local cache
- timeout: default request timeout for backend request to cloud kube-apiserver
- sync-duration: default resync duration for backend data

#### Certification Management
- tls-config-file: path to the TLS secrets, default to empty
