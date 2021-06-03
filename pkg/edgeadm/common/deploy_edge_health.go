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
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math"
	"math/big"
	"path/filepath"
	"time"

	"k8s.io/client-go/kubernetes"
	k8scert "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"k8s.io/klog/v2"

	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeclient"
)

func DeployEdgeHealth(clientSet kubernetes.Interface, manifestsDir string) error {
	yamlMap, option, err := getEdgeHealthResource(clientSet, manifestsDir)
	if err != nil {
		return err
	}

	for appName, yamlFile := range yamlMap {
		if err := kubeclient.CreateResourceWithFile(clientSet, yamlFile, option); err != nil {
			return err
		}
		klog.Infof("Create %s success!\n", appName)
	}

	return nil
}

func DeleteEdgeHealth(clientSet kubernetes.Interface, manifestsDir string) error {
	yamlMap, option, err := getEdgeHealthResource(clientSet, manifestsDir)
	if err != nil {
		return err
	}
	for appName, yamlFile := range yamlMap {
		if err := kubeclient.DeleteResourceWithFile(clientSet, yamlFile, option); err != nil {
			return err
		}
		klog.Infof("Delete %s success!\n", appName)
	}

	return nil
}

func getEdgeHealthResource(clientSet kubernetes.Interface, manifestsDir string) (map[string]string, interface{}, error) {
	userEdgeHealthWebhook := filepath.Join(manifestsDir, manifests.APP_EDGE_HEALTH_WEBHOOK)
	userEdgeHealthAdmission := filepath.Join(manifestsDir, manifests.APP_EDGE_HEALTH_ADMISSION)
	yamlMap := map[string]string{
		manifests.APP_EDGE_HEALTH_ADMISSION: ReadYaml(userEdgeHealthAdmission, manifests.EdgeHealthAdmissionYaml),
		manifests.APP_EDGE_HEALTH_WEBHOOK:   ReadYaml(userEdgeHealthWebhook, manifests.EdgeHealthWebhookConfigYaml),
	}

	caBundle, ca, caKey, err := GenerateEdgeWebhookCA()
	if err != nil {
		return nil, nil, err
	}
	serverCrt, serverKey, err := GenEdgeWebhookCertAndKey(ca, caKey)
	if err != nil {
		return nil, nil, err
	}

	option := map[string]interface{}{
		"Namespace": constant.NamespaceEdgeSystem,
		"CABundle":  caBundle,
		"ServerCrt": serverCrt,
		"ServerKey": serverKey,
		"HmacKey":   util.GetRandToken(16),
	}

	return yamlMap, option, nil
}

func GenerateEdgeWebhookCA() (caBundle string, caCert *x509.Certificate, caPrivKey *rsa.PrivateKey, err error) {
	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return
	}
	ca := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"superedge"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return
	}
	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	caBundle = base64.StdEncoding.EncodeToString(caPEM.Bytes())
	caCerts, err := util.ParseCertsPEM(caPEM.Bytes())
	if err != nil {
		return
	}
	caCert = caCerts[0]
	return
}

func GenEdgeWebhookCertAndKey(ca *x509.Certificate, key *rsa.PrivateKey) (serverCrt, serverKey string, err error) {
	svCert, svKey, err := util.GenerateCertAndKeyConfig(ca, key, &k8scert.Config{
		CommonName:   "edge-health-admission",
		Organization: []string{"superedge"},
		AltNames: k8scert.AltNames{
			DNSNames: []string{
				fmt.Sprintf("edge-health-admission.%s.svc", constant.NamespaceEdgeSystem),
			},
		},
		Usages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth},
	})

	serverCrt = base64.StdEncoding.EncodeToString(util.EncodeCertPEM(svCert))
	svKeyBytes, err := keyutil.MarshalPrivateKeyToPEM(svKey)
	if err != nil {
		return
	}
	serverKey = base64.StdEncoding.EncodeToString(svKeyBytes)
	return
}
