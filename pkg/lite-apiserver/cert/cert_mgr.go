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
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"
	"time"

	"k8s.io/klog/v2"

	"github.com/superedge/superedge/pkg/lite-apiserver/config"
	"github.com/superedge/superedge/pkg/util"
)

const (
	fastReloadDuration = 1 * time.Second
	reloadDuration     = 30 * time.Minute
)

type CertManager struct {
	tlsConfig []config.TLSKeyPair

	fastReload bool

	certChannel chan<- string

	certMapLock sync.RWMutex
	certMap     map[string]*tls.Certificate
}

func NewCertManager(config *config.LiteServerConfig, certChannel chan<- string) *CertManager {
	return &CertManager{
		tlsConfig:   config.TLSConfig,
		certChannel: certChannel,
		certMap:     make(map[string]*tls.Certificate),
	}
}

func (cm *CertManager) Init() error {
	fastReload := false

	for i := range cm.tlsConfig {
		cert := cm.tlsConfig[i].CertPath
		key := cm.tlsConfig[i].KeyPath

		tlsCert, commonName, err := loadCert(cert, key)
		if err != nil {
			// load one cert error.
			fastReload = true
			continue
		}

		if util.CertHasExpired(tlsCert.Leaf) {
			klog.Infof("cert %s,%s has expired.", cert, key)
			fastReload = true
		}

		cm.updateCert(commonName, tlsCert)
	}

	cm.fastReload = fastReload
	return nil
}

func (cm *CertManager) Start() {
	t := time.NewTimer(cm.getReloadDuration())

	go func() {
		for {
			fastReload := false

			select {
			case <-t.C:
				for i := range cm.tlsConfig {
					cert := cm.tlsConfig[i].CertPath
					key := cm.tlsConfig[i].KeyPath

					klog.V(4).Infof("handling cert=%s, key=%s", cert, key)
					tlsCert, commonName, err := loadCert(cert, key)
					if err != nil {
						// load one cert error.
						fastReload = true
						continue
					}
					if util.CertHasExpired(tlsCert.Leaf) {
						klog.Infof("cert %s,%s has expired.", cert, key)
						fastReload = true
					}

					cm.handleCertUpdate(tlsCert, commonName, cert, key)
				}
			}

			// reset timer
			cm.fastReload = fastReload
			t.Reset(cm.getReloadDuration())
		}
	}()
}

func (cm *CertManager) handleCertUpdate(tlsCert *tls.Certificate, commonName string, cert, key string) {
	cm.certMapLock.RLock()
	oldCert, ok := cm.certMap[commonName]
	cm.certMapLock.RUnlock()
	if !ok {
		// new cert
		klog.Infof("add new cert CN=%s cert=%s, key=%s", commonName, cert, key)
		cm.updateCert(commonName, tlsCert)

		// inform to create new transport
		cm.certChannel <- commonName
	} else {
		// check cert changed
		if !bytes.Equal(oldCert.Leaf.Signature, tlsCert.Leaf.Signature) {
			// update cert
			klog.Infof("update cert CN=%s cert=%s, key=%s", commonName, cert, key)
			cm.updateCert(commonName, tlsCert)

			// inform to update transport
			cm.certChannel <- commonName
		}
	}
}

func (cm *CertManager) GetCert(commonName string) *tls.Certificate {
	cm.certMapLock.RLock()
	defer cm.certMapLock.RUnlock()
	return cm.certMap[commonName]
}

func (cm *CertManager) GetCertMap() map[string]*tls.Certificate {
	cm.certMapLock.RLock()
	defer cm.certMapLock.RUnlock()
	return cm.certMap
}

func (cm *CertManager) updateCert(commonName string, tlsCert *tls.Certificate) {
	cm.certMapLock.Lock()
	defer cm.certMapLock.Unlock()
	cm.certMap[commonName] = tlsCert
}

func (cm *CertManager) getReloadDuration() time.Duration {
	if cm.fastReload {
		return fastReloadDuration
	} else {
		return reloadDuration
	}
}

func loadCert(cert string, key string) (*tls.Certificate, string, error) {
	tlsCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		klog.Errorf("load cert and key error: %v", err)
		return nil, "", err
	}

	var leaf *x509.Certificate
	if tlsCert.Leaf == nil {
		l, err := x509.ParseCertificate(tlsCert.Certificate[0])
		if err != nil {
			klog.Errorf("parse cert %s,%s error: %v", cert, key, err)
			return nil, "", err
		}
		leaf = l
		tlsCert.Leaf = l
	} else {
		leaf = tlsCert.Leaf
	}
	commonName := leaf.Subject.CommonName

	if len(commonName) == 0 {
		klog.Errorf("cert common name nil")
		return nil, "", fmt.Errorf("cert common name nil")
	}

	return &tlsCert, commonName, nil
}
