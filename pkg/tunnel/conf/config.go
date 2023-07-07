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

package conf

import (
	"github.com/BurntSushi/toml"
	"github.com/superedge/superedge/pkg/tunnel/util"
)

var TunnelConf *Tunnel

type Tunnel struct {
	TunnlMode *TunnelMode `toml:"mode"`
}

type TunnelMode struct {
	Cloud *TunnelCloud `toml:"cloud"`
	EDGE  *TunnelEdge  `toml:"edge"`
}

type TunnelCloud struct {
	Stream    *StreamCloud     `toml:"stream"`
	Egress    *EgressServer    `toml:"egress"`
	HttpProxy *HttpProxyServer `toml:"http_proxy"`
	SSH       *SSHServer       `toml:"ssh"`
	TLS       *TLSConfig       `toml:"tls"`
}

type HttpsServer struct {
	Cert string            `toml:"cert"`
	Key  string            `toml:"key"`
	Addr map[string]string `toml:"addr"`
}

type StreamCloud struct {
	Server   *StreamServer `toml:"server"`
	Register *Register     `toml:"register"`
}

type StreamServer struct {
	GrpcPort     int    `toml:"grpc_port"`
	LogPort      int    `toml:"log_port"`
	MetricsPort  int    `toml:"metrics_port"`
	ChannelzAddr string `toml:"channelz_addr"`
}

type TLSConfig struct {
	CipherSuites  string `toml:"tls_cipher_suites"`
	MinTLSVersion string `toml:"tls_min_version"`
}

type EgressServer struct {
	EgressPort int `toml:"port"`
}
type HttpProxyServer struct {
	ProxyPort int `toml:"port"`
}
type SSHServer struct {
	SSHPort int `toml:"port"`
}

type Register struct {
	Service string `toml:"service"`
}

type TunnelEdge struct {
	StreamEdge StreamEdge          `toml:"stream"`
	HttpProxy  HttpProxyEdgeServer `toml:"http_proxy"`
}

type HttpProxyEdgeServer struct {
	ProxyIP   string `toml:"ip"`
	ProxyPort string `toml:"port"`
}

type StreamEdge struct {
	Client *StreamClient `toml:"client"`
}

type StreamClient struct {
	Token        string `toml:"token"`
	Dns          string `toml:"dns"`
	ServerName   string `toml:"server_name"`
	LogPort      int    `toml:"log_port"`
	ChannelzAddr string `toml:"channelz_addr"`
}

type Config struct {
	Mode map[string]map[string]map[string]interface{} `toml:"mode"`
}

func InitConf(mode, path string) error {
	if mode == util.CLOUD {
		TunnelConf = &Tunnel{
			TunnlMode: &TunnelMode{
				Cloud: &TunnelCloud{},
			},
		}
	} else {
		TunnelConf = &Tunnel{
			TunnlMode: &TunnelMode{
				EDGE: &TunnelEdge{},
			},
		}
	}
	_, err := toml.DecodeFile(path, TunnelConf)
	if err != nil {
		return err
	}
	return nil
}
