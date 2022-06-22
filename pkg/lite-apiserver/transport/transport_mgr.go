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
	"io/ioutil"
	apimachinerynet "k8s.io/apimachinery/pkg/util/net"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/connrotation"
	"k8s.io/klog/v2"

	"github.com/superedge/superedge/pkg/lite-apiserver/cert"
	"github.com/superedge/superedge/pkg/lite-apiserver/config"
	"github.com/superedge/superedge/pkg/util"
)

const (
	DefaultTimeout          = 10
	DefaultKeepAlive        = 30
	healthCheckDuration     = 1 * time.Second
	defaultNetworkInterface = "default"
)

type TransportManager struct {
	config      *config.LiteServerConfig
	certManager *cert.CertManager

	caFile       string
	rootCertPool *x509.CertPool
	insecure     bool

	timeout int

	certChannel      <-chan string
	transportChannel chan<- string

	defaultTransport *EdgeTransport

	transportMapLock sync.RWMutex
	transportMap     map[string]*EdgeTransport
}

func NewTransportManager(config *config.LiteServerConfig, certManager *cert.CertManager,
	certChannel <-chan string, transportChannel chan<- string) *TransportManager {
	return &TransportManager{
		config:           config,
		certManager:      certManager,
		caFile:           config.ApiserverCAFile,
		insecure:         config.Insecure,
		timeout:          config.BackendTimeout,
		certChannel:      certChannel,
		transportChannel: transportChannel,
		transportMap:     make(map[string]*EdgeTransport),
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
	if tm.insecure {
		tm.defaultTransport = tm.makeTransport(&tls.Config{InsecureSkipVerify: true}, nil)
	} else {
		tm.defaultTransport = tm.makeTransport(&tls.Config{RootCAs: tm.rootCertPool}, nil)
	}

	// init transportMap
	for commonName := range tm.certManager.GetCertMap() {
		t, err := tm.getTransport(commonName)
		if err != nil {
			klog.Errorf("get transport error, commonName=%s: %v", commonName, err)
			continue
		}
		tm.updateTransport(commonName, t)
	}

	return nil
}

func (tm *TransportManager) Start() {
	go func() {
		for {
			select {
			case commonName := <-tm.certChannel:
				// add new cert to create transport
				klog.Infof("receive cert update %s", commonName)

				tm.transportMapLock.RLock()
				old, ok := tm.transportMap[commonName]
				tm.transportMapLock.RUnlock()

				if !ok {
					// new cert
					klog.Infof("receive cert %s update", commonName)
					t, err := tm.getTransport(commonName)
					if err != nil {
						klog.Errorf("get transport error, commonName=%s: %v", commonName, err)
						break
					}
					tm.updateTransport(commonName, t)

					// inform handler to create new EdgeReverseProxy
					tm.transportChannel <- commonName
				} else {
					// cert rotation
					klog.Infof("cert %s rotated, close old connections", commonName)
					old.d.CloseAll()
				}
			}
		}
	}()

	// Check the nic periodically，then update transport
	if tm.config.NetworkInterface != "" {
		go wait.Forever(func() {
			for commonName := range tm.certManager.GetCertMap() {
				old, ok := tm.transportMap[commonName]
				if !ok {
					continue
				}
				t, err := tm.getTransport(commonName)
				if err != nil {
					continue
				}
				// if transport changed, inform handler to create new EdgeReverseProxy
				if t.NetworkInterface != old.NetworkInterface {
					tm.updateTransport(commonName, t)
					tm.transportChannel <- commonName
					klog.V(4).Infof("update transport, commonName [%s], network interface changed from [%s] to [%s]",
						commonName, old.NetworkInterface, t.NetworkInterface)
				}
			}
		}, healthCheckDuration)
	}
}

func (tm *TransportManager) GetTransport(commonName string) *EdgeTransport {
	if len(commonName) == 0 {
		return tm.defaultTransport
	}

	tm.transportMapLock.RLock()
	defer tm.transportMapLock.RUnlock()
	t, ok := tm.transportMap[commonName]
	if !ok {
		klog.V(4).Infof("couldn't get transport for %s, use default transport", commonName)
		return tm.defaultTransport
	}

	return t
}

func (tm *TransportManager) GetTransportMap() map[string]*EdgeTransport {
	tm.transportMapLock.RLock()
	defer tm.transportMapLock.RUnlock()
	return tm.transportMap
}

func (tm *TransportManager) updateTransport(commonName string, transport *EdgeTransport) {
	tm.transportMapLock.Lock()
	defer tm.transportMapLock.Unlock()
	tm.transportMap[commonName] = transport
}

func (tm *TransportManager) makeTlsConfig(commonName string) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if tm.insecure {
		tlsConfig.InsecureSkipVerify = tm.insecure
	} else {
		tlsConfig.RootCAs = tm.rootCertPool
	}

	tlsConfig.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
		klog.V(6).Infof("get cert for %s", commonName)
		currentCert := tm.certManager.GetCert(commonName)
		if currentCert == nil {
			klog.Warningf("cert for %s is nil, use default", commonName)
			return &tls.Certificate{Certificate: nil}, nil
		}
		klog.V(6).Infof("cert for %s is %+v", commonName, currentCert.Leaf)
		return currentCert, nil
	}

	return tlsConfig, nil
}

func (tm *TransportManager) makeTransport(tlsClientConfig *tls.Config, localAddr net.Addr) *EdgeTransport {
	if tm.timeout == 0 {
		tm.timeout = DefaultTimeout
	}

	dialer := &net.Dialer{
		Timeout:   time.Duration(tm.timeout) * time.Second,
		KeepAlive: DefaultKeepAlive * time.Second,
	}

	if localAddr != nil {
		dialer.LocalAddr = localAddr
	}

	d := connrotation.NewDialer(dialer.DialContext)

	t1 := &http.Transport{
		DialContext:           d.DialContext,
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       tlsClientConfig,
	}
	t1 = apimachinerynet.SetTransportDefaults(t1)

	return &EdgeTransport{
		d: d,
		// TODO enable http2 if using go1.15
		// the params are same with http.DefaultTransport
		Transport: t1,
	}
}

func (tm *TransportManager) getTransport(commonName string) (*EdgeTransport, error) {
	// default transport
	tlsConfig, err := tm.makeTlsConfig(commonName)
	if err != nil {
		return nil, fmt.Errorf("make tls config error, commonName=%s: %v", commonName, err)
	}
	defaultTransport := tm.makeTransport(tlsConfig, nil)

	// get healthy transport
	if tm.config.NetworkInterface != "" {
		url := tm.config.KubeApiserverUrl
		port := tm.config.KubeApiserverPort

		isHealthy, err := tm.checkApiserverHealth(defaultTransport.Transport, url, port)
		if err != nil {
			klog.Errorf("failed to check apiserver health by default interface, err: %v", err)
		}
		if isHealthy {
			defaultTransport.NetworkInterface = defaultNetworkInterface
			return defaultTransport, nil
		}

		netIfList := strings.Split(tm.config.NetworkInterface, ",")
		if len(netIfList) == 1 && netIfList[0] == "" {
			return nil, fmt.Errorf("the network interface invalid")
		}
		for _, netIf := range netIfList {
			localAddr, err := util.GetLocalAddrByInterface(netIf)
			if err != nil {
				klog.Errorf("failed to get localAddr by interface [%s], err: %v", netIf, err)
				continue
			}

			healthyTransport := tm.makeTransport(tlsConfig, localAddr)
			isHealthy, err := tm.checkApiserverHealth(healthyTransport.Transport, url, port)
			if err != nil {
				klog.Errorf("failed to check apiserver health by interface [%s], err: %v", netIf, err)
				continue
			}
			klog.V(8).Infof("check apiserver health by interface [%s]", netIf)
			if isHealthy {
				healthyTransport.NetworkInterface = netIf
				return healthyTransport, nil
			}
		}
	}

	return defaultTransport, nil
}

func (tm *TransportManager) checkApiserverHealth(transport *http.Transport, url string, port int) (bool, error) {
	if transport == nil {
		return false, fmt.Errorf("http client is invalid")
	}

	client := &http.Client{Transport: transport}
	resp, err := client.Get(fmt.Sprintf("https://%s:%d/healthz", url, port))
	if err != nil {
		return false, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return false, fmt.Errorf("failed to read response of cluster healthz, %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("response status code is %d", resp.StatusCode)
	}

	if strings.ToLower(string(b)) != "ok" {
		return false, fmt.Errorf("cluster healthz is %s", string(b))
	}

	return true, nil
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
