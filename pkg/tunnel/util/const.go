/*
Copyright 2020 The SuperEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

const (
	STREAM_HEART_BEAT = "heartbeat"
	STREAM_TRACE_ID   = "Traceid"
)

const (
	TCP_FORWARD = "tcp-forward"
)

const (
	STREAM     = "stream"
	TCP        = "tcp"
	SSH        = "ssh"
	EGRESS     = "egress"
	HTTP_PROXY = "httpProxy"
)

const (
	PODIP_INDEXER     = "PodIP"
	METANAME_INDEXER  = "MeatName"
	SERVICEIP_INDEXER = "ServiceIP"
)

const (
	CLOUD = "cloud"
	EDGE  = "edge"
)

const (
	CLOSED = "closed"
)

const (
	TIMEOUT_EXIT    = 180
	MSG_CHANNEL_CAP = 1000
)

const (
	COREFILE_HOSTS_FILE     = "hosts"
	NODE_NAME_ENV           = "NODE_NAME"
	POD_IP_ENV              = "POD_IP"
	POD_NAME                = "POD_NAME"
	POD_NAMESPACE_ENV       = "POD_NAMESPACE"
	USER_NAMESPACE_ENV      = "USER_NAMESPACE"
	PROXY_AUTHORIZATION_ENV = "PROXY_AUTHORIZATION"
	EdgeNoProxy             = "EDGE_NO_PROXY"
	CloudProxy              = "CLOUD_PROXY"
)

const (
	ConnectMsg = "HTTP/1.1 200 Connection established\r\n\r\n"
)

const (
	HostsConfig       = "tunnel-nodes"
	CacheConfig       = "tunnel-cache"
	CachePath         = "/etc/tunnel/cache"
	CertsPath         = "/etc/tunnel/certs"
	AuthorizationPath = "/etc/tunnel/auth"
)

const (
	EdgeNodesFile   = "edge_nodes"
	CloudNodesFile  = "cloud_nodes"
	ServicesFile    = "services"
	UserServiceFile = "user_services"
	TunnelCloudCert = "cloud.crt"
	TunnelCloudKey  = "cloud.key"
	EgressCert      = "egress.crt"
	EgressKey       = "egress.key"
	TunnelEdgeCA    = "ca.crt"
)

const (
	HostsPath            = "/etc/tunnel/nodes/hosts"
	TunnelCloudTokenPath = "/etc/tunnel/token/token"
	TunnelCloudCertPath  = CertsPath + "/" + TunnelCloudCert
	TunnelCloudKeyPath   = CertsPath + "/" + TunnelCloudKey
	EgressCertPath       = CertsPath + "/" + EgressCert
	EgressKeyPath        = CertsPath + "/" + EgressKey
	EdgeNodesFilePath    = CachePath + "/" + EdgeNodesFile
	CloudNodesFilePath   = CachePath + "/" + CloudNodesFile
	ServicesFilePath     = CachePath + "/" + ServicesFile
	UserServiceFilepath  = CachePath + "/" + UserServiceFile
	TunnelEdgeCAPath     = CertsPath + "/" + TunnelEdgeCA
)
