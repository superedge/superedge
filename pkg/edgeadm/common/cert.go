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

package common

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net"

	"io/ioutil"
	k8scert "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"

	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
)

func GetClientCert(commonName, certPath, keyPath string) ([]byte, []byte, error) {
	caCertBtye, caKeyByte, err := GetRootCartAndKey(certPath, keyPath)
	if err != nil {
		return nil, nil, err
	}

	caCert, caKey, err := ParseCertAndKey(caCertBtye, caKeyByte)
	if err != nil {
		return nil, nil, err
	}

	clientCert, clientKey, err := util.GenerateClientCertAndKey(caCert, caKey, commonName)
	if err != nil {
		return nil, nil, err
	}

	clientCertData := util.EncodeCertPEM(clientCert)
	clientKeyData, err := keyutil.MarshalPrivateKeyToPEM(clientKey)
	if err != nil {
		return nil, nil, err
	}

	return clientCertData, clientKeyData, err
}

func ParseCertAndKey(ca, key []byte) (*x509.Certificate, *rsa.PrivateKey, error) {
	//Transform cacer and key
	caCert, err := util.ParseCertsPEM([]byte(ca))
	if err != nil {
		return nil, nil, err
	}
	caKey, err := util.ParsePrivateKeyPEMRSA([]byte(key))
	if err != nil {
		return nil, nil, err
	}

	if len(caCert) != 1 {
		return nil, nil, fmt.Errorf("CaCert length is not 1")
	}

	return caCert[0], caKey, nil
}

func GetRootCartAndKey(certPath, keyPath string) ([]byte, []byte, error) {
	var caCertFile string
	if util.IsFileExist(certPath) {
		caCertFile = certPath
	} else if util.IsFileExist(constant.KubeadmCertPath) {
		caCertFile = constant.KubeadmCertPath
	} else {
		return nil, nil, fmt.Errorf("Please input root ca.cert file path\n")
	}

	var caKeyFile string
	if util.IsFileExist(keyPath) {
		caKeyFile = keyPath
	} else if util.IsFileExist(constant.KubeadmKeyPath) {
		caKeyFile = constant.KubeadmKeyPath
	} else {
		return nil, nil, fmt.Errorf("Please input root ca.key file path\n")
	}

	ca, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return nil, nil, err
	}
	key, err := ioutil.ReadFile(caKeyFile)
	if err != nil {
		return nil, nil, err
	}

	return ca, key, nil
}

func GetServiceCert(commonName, caCertFile, caKeyFile string, dns []string, ips []string) ([]byte, []byte, error) {
	caCert, caKey, err := GetCertAndKey(caCertFile, caKeyFile)
	if err != nil {
		return nil, nil, err
	}

	certIps := []net.IP{net.ParseIP("127.0.0.1")}
	for _, ip := range ips {
		certIps = append(certIps, net.ParseIP(ip))
	}
	serverCert, serverKey, err := util.GenerateCertAndKeyConfig(caCert, caKey, &k8scert.Config{
		CommonName:   commonName,
		Organization: []string{"superedge"},
		AltNames: k8scert.AltNames{
			DNSNames: dns,
			IPs:      certIps,
		},
		Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	})

	serverCertData := util.EncodeCertPEM(serverCert)
	serverKeyData, err := keyutil.MarshalPrivateKeyToPEM(serverKey)
	if err != nil {
		return nil, nil, err
	}

	return serverCertData, serverKeyData, err
}

func GetCertAndKey(caCertFile, caKeyFile string) (*x509.Certificate, *rsa.PrivateKey, error) {
	caCert, caKey, err := GetRootCartAndKey(caCertFile, caKeyFile)
	if err != nil {
		return nil, nil, err
	}

	cert, key, err := ParseCertAndKey(caCert, caKey)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}
