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

package options

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"github.com/superedge/superedge/pkg/lite-apiserver/config"
)

type RunServerOptions struct {
	KubeApiserverUrl  string
	KubeApiserverPort int
	Port              int
	SyncDuration      int
	BackendTimeout    int
	CAFile            string
	CertFile          string
	KeyFile           string
	ApiserverCAFile   string
	FileCachePath     string
}

func NewRunServerOptions() *RunServerOptions {
	return &RunServerOptions{}
}

// ApplyTo applies the storage options to the method receiver and returns self
func (s *RunServerOptions) ApplyTo(c *config.LiteServerConfig) error {
	c.CAFile = s.CAFile
	c.CertFile = s.CertFile
	c.KeyFile = s.KeyFile
	c.KubeApiserverUrl = s.KubeApiserverUrl
	c.KubeApiserverPort = s.KubeApiserverPort
	c.Port = s.Port
	c.SyncDuration = time.Duration(s.SyncDuration) * time.Second
	c.FileCachePath = s.FileCachePath
	c.BackendTimeout = s.BackendTimeout

	if len(s.ApiserverCAFile) > 0 {
		c.ApiserverCAFile = s.ApiserverCAFile
	} else {
		c.ApiserverCAFile = s.CAFile
	}

	return nil
}

// Validate checks validation of ServerRunOptions
func (s *RunServerOptions) Validate() []error {
	var errors []error

	if len(s.CAFile) == 0 {
		errors = append(errors, fmt.Errorf("CA cannot be empty"))
	}

	if len(s.CertFile) == 0 {
		errors = append(errors, fmt.Errorf("cert cannot be empty"))
	}

	if len(s.KeyFile) == 0 {
		errors = append(errors, fmt.Errorf("key cannot be empty"))
	}

	if len(s.KubeApiserverUrl) == 0 {
		errors = append(errors, fmt.Errorf("kube-apiserver url cannot be empty"))
	}

	if s.Port == 0 {
		errors = append(errors, fmt.Errorf("port cannot be 0"))
	}

	if s.SyncDuration <= 0 {
		s.SyncDuration = 60
	}

	return errors
}

// AddUniversalFlags adds flags for a specific APIServer to the specified FlagSet
func (s *RunServerOptions) AddFlags(fs *pflag.FlagSet) {
	// Note: the weird ""+ in below lines seems to be the only way to get gofmt to
	// arrange these text blocks sensibly.
	fs.StringVar(&s.CAFile, "ca-file", "", "the CA that lite-apiserver used to verify a client certificate")
	fs.StringVar(&s.CertFile, "tls-cert-file", "", "the tls cert of lite-apiserver")
	fs.StringVar(&s.KeyFile, "tls-private-key-file", "", "the tls key of lite-apiserver")
	fs.StringVar(&s.ApiserverCAFile, "apiserver-ca-file", "", "the CA used to verify kube-apiserver server tls")

	fs.StringVar(&s.KubeApiserverUrl, "kube-apiserver-url", "", "the host of kube-apiserver")
	fs.IntVar(&s.KubeApiserverPort, "kube-apiserver-port", 443, "the port of kube-apiserver")

	fs.IntVar(&s.Port, "port", 51003, "the port on the local server to listen on.")
	fs.IntVar(&s.SyncDuration, "sync-duration", 60, "self sync data time(second)")
	fs.IntVar(&s.BackendTimeout, "timeout", 30, "timeout for proxy to backend")
	fs.StringVar(&s.FileCachePath, "file-cache-path", "/data/lite-apiserver/cache", "the path for storage")
}
