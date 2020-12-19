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

package cert

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"superedge/pkg/lite-apiserver/config"

	"k8s.io/klog"
)

type CertManager struct {
	config       *config.LiteServerConfig
	caPath       string
	transportMap map[string]*http.Transport

	defaultTr *http.Transport
	lock      sync.Mutex
}

func NewCertManager(config *config.LiteServerConfig) *CertManager {
	return &CertManager{
		config:       config,
		caPath:       config.CAFile,
		transportMap: make(map[string]*http.Transport),
	}
}

func (m *CertManager) Init() error {
	err := m.loadTransport() // TODO reload
	if err != nil {
		return err
	}
	return nil
}

func (m *CertManager) Load(name string) *http.Transport {
	t, e := m.transportMap[name]
	if !e {
		return nil
	}
	return t
}

func (m *CertManager) DefaultTransport() *http.Transport {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.defaultTr == nil {
		caCrt, err := ioutil.ReadFile(m.caPath)
		if err != nil {
			klog.Errorf("read ca file err: %v", err)
			return nil
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caCrt)

		m.defaultTr = makeTransport(&tls.Config{RootCAs: pool})
	}
	return m.defaultTr
}

func (m *CertManager) loadTransport() error {
	for i := range m.config.TLSConfig {
		cert := m.config.TLSConfig[i].CertPath
		key := m.config.TLSConfig[i].KeyPath
		klog.V(4).Infof("")

		tlsCert, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			klog.Errorf("load cert and key error: %v", err)
			return err
		}

		var leaf *x509.Certificate
		if tlsCert.Leaf == nil {
			l, err := x509.ParseCertificate(tlsCert.Certificate[0])
			if err != nil {
				klog.Errorf("Parse cert %s,%s error: %v", cert, key, err)
				return err
			}
			leaf = l
		} else {
			leaf = tlsCert.Leaf
		}
		commonName := leaf.Subject.CommonName

		if len(commonName) == 0 {
			klog.Errorf("cert common name nil")
			return fmt.Errorf("cert common name nil")
		}

		var caCrt []byte
		caCrt, err = ioutil.ReadFile(m.caPath)
		if err != nil {
			klog.Errorf("read ca file err: %v", err)
			return err
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caCrt)

		tr := makeTransport(&tls.Config{RootCAs: pool, Certificates: []tls.Certificate{tlsCert}})
		m.transportMap[commonName] = tr
		klog.Infof("Add common %s in tls map", commonName)
	}
	return nil
}

func makeTransport(tlsClientConfig *tls.Config) *http.Transport {
	// TODO enable http2 if using go1.15
	return &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       tlsClientConfig,
	}
}
