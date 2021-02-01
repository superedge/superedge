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

package transport

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/superedge/superedge/pkg/lite-apiserver/cert"
	"github.com/superedge/superedge/pkg/lite-apiserver/config"
	"io/ioutil"
	"k8s.io/klog"
	"net"
	"net/http"
	"sync"
	"time"
)

type TransportManager struct {
	config      *config.LiteServerConfig
	certManager *cert.CertManager

	caFile       string
	rootCertPool *x509.CertPool

	defaultTransport *http.Transport

	transportMapLock sync.RWMutex
	transportMap     map[string]*http.Transport
}

func NewTransportManager(config *config.LiteServerConfig, certManager *cert.CertManager) *TransportManager {
	return &TransportManager{
		config:       config,
		certManager:  certManager,
		caFile:       config.ApiserverCAFile,
		transportMap: make(map[string]*http.Transport),
	}
}

func (tm *TransportManager) Init() error {
	// init rootCertPool
	rootCertPool, err := getRootCertPool(tm.caFile)
	if err != nil {
		return err
	}
	tm.rootCertPool = rootCertPool

	// init default transport
	tm.defaultTransport = makeTransport(&tls.Config{RootCAs: tm.rootCertPool})

	// init transportMap
	for commonName, _ := range tm.certManager.GetCertMap() {
		tlsConfig, err := tm.makeTlsConfig(commonName)
		if err != nil {
			klog.Errorf("make tls config error, commonName=%s: %v", commonName, err)
			continue
		}
		tm.updateTransport(commonName, makeTransport(tlsConfig))
	}

	return nil
}

func (tm *TransportManager) Start() {
	// add new cert to create transport

	// inform handler to create new EdgeReverseProxy
}

func (tm *TransportManager) GetTransport(name string) *http.Transport {
	if len(name) == 0 {
		return tm.defaultTransport
	}

	tm.transportMapLock.RLock()
	defer tm.transportMapLock.RUnlock()
	t, e := tm.transportMap[name]
	if !e {
		return tm.defaultTransport
	}

	return t
}

func (tm *TransportManager) GetTransportMap() map[string]*http.Transport {
	tm.transportMapLock.RLock()
	defer tm.transportMapLock.RUnlock()
	return tm.transportMap
}

func (tm *TransportManager) updateTransport(commonName string, transport *http.Transport) {
	tm.transportMapLock.Lock()
	defer tm.transportMapLock.Unlock()
	tm.transportMap[commonName] = transport
}

func (tm *TransportManager) makeTlsConfig(commonName string) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    tm.rootCertPool,
	}

	tlsConfig.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
		currentCert := tm.certManager.GetCert(commonName)
		if currentCert == nil {
			return &tls.Certificate{Certificate: nil}, nil
		}
		return currentCert, nil
	}

	return tlsConfig, nil
}

func getRootCertPool(caFile string) (*x509.CertPool, error) {
	caCrt, err := ioutil.ReadFile(caFile)
	if err != nil {
		klog.Errorf("read ca file %s err: %v", caFile, err)
		return nil, err
	}

	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM(caCrt)
	if !ok {
		klog.Errorf("append ca certs %s error.", caFile)
		return nil, fmt.Errorf("append ca certs %s error.\n", caFile)
	}

	return pool, nil
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
