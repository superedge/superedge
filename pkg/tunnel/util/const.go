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
)

const (
	TCP_FRONTEND = "frontend"
	TCP_BACKEND  = "backend"
	TCP_CONTROL  = "control"
)

const (
	STREAM = "stream"
	TCP    = "tcp"
	HTTPS  = "https"
)

const (
	CLOUD = "cloud"
	EDGE  = "edge"
)

const (
	CONNECTING    = "connecting"
	CONNECTED     = "connected"
	TRANSNMISSION = "transmission"
	CLOSED        = "closed"
)

const (
	MaxResponseSize = 16384
	TIMEOUT_EXIT    = 180
	MSG_CHANNEL_CAP = 1000
)

const (
	COREFILE_HOSTS_FILE = "hosts"
	NODE_NAME_ENV       = "NODE_NAME"
	POD_IP_ENV          = "POD_IP"
	POD_NAMESPACE_ENV   = "POD_NAMESPACE"
)

const (
	MODULE_DEBUG = "debug"
)
