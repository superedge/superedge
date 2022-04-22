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
	"bytes"
	"crypto"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"github.com/superedge/superedge/pkg/util"
	"io"
	"io/ioutil"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"k8s.io/klog/v2"
	"math"
	"math/big"
	"net"
	"os"
	"testing"
	"text/template"
	"time"
)

const duration365d = time.Hour * 24 * 365
const (
	TUNNEL = "tunnel"
	K8S    = "kubernetes"
)
const (
	config_path = "../../../conf"
	main_path   = "../../../../conf" /*The .Path parameter of the matching template*/
)
const token = "6ff2a1ea0f1611eb9896362096106d9d"

const edge_mode = `
[mode]
	[mode.edge]
		[mode.edge.stream]
			[mode.edge.stream.client]
				token = "{{.Token}}"
				cert = "{{.Path}}/certs/ca.crt"
  				dns = "localhost"
				servername = "localhost:9000"
				logport = 7000
				channelzaddr = "0.0.0.0:5000"
			[mode.edge.https]
				cert= "{{.Path}}/certs/kubelet-client.crt"#apiserver访问kubelet的客户端证书
				key= "{{.Path}}/certs/kubelet-client.key"`
const cloud_toml = `
[mode]
	[mode.cloud]
		[mode.cloud.stream]
			[mode.cloud.stream.server]
				grpcport = 9000
				logport = 8000
                channelzaddr = "0.0.0.0:6000"
				key = "{{.Path}}/certs/cloud.key"
				cert = "{{.Path}}/certs/cloud.crt"
				tokenfile = "{{.Path}}/token"
			[mode.cloud.stream.dns]
				configmap= "proxy-nodes"
				hosts = "/etc/superedge/proxy/nodes/hosts"
				service = "proxy-cloud-public"
				debug = true
            [mode.cloud.tcp]
                "0.0.0.0:6443" = "127.0.0.1:6443"
            [mode.cloud.https]
                cert ="{{.Path}}/certs/kubelet.crt"#kubelet的服务端证书
                key = "{{.Path}}/certs/kubelet.key"
			[mode.cloud.https.addr]
				"10250" = "101.206.162.213:10250"`

type CertConfig struct {
	CommonName   string
	Organization []string
	AltNames     AltNames
	Usages       []x509.ExtKeyUsage
}

type AltNames struct {
	DNSNames []string
	IPs      []net.IP
}

func Test_Config(t *testing.T) {
	var ca *x509.Certificate
	var key *rsa.PrivateKey
	_, err := os.Stat(config_path + "/certs/ca.crt")
	if err == nil {
		cb, err := ioutil.ReadFile(config_path + "/certs/ca.crt")
		if err != nil {
			t.Errorf("load cert file fail! err = %v", err)
			return
		}
		ck, err := ioutil.ReadFile(config_path + "/certs/ca.key")
		if err != nil {
			t.Errorf("load key  file fail! err = %v", err)
			return
		}
		crt, err := tls.X509KeyPair(cb, ck)
		if err != nil {
			t.Errorf("parase cert fail! err = %v", err)
			return
		}
		certs, err := x509.ParseCertificates(crt.Certificate[0])
		if err != nil {
			t.Errorf("get cert fail err = %v", err)
			return
		}
		ca = certs[0]
		key = crt.PrivateKey.(*rsa.PrivateKey)
	} else if os.IsNotExist(err) {
		ca, key, err = GenerateCa(TUNNEL)
		if err != nil {
			t.Errorf("failed to get ca and key err = %v", err)
			return
		}
		SaveCertAndKey(ca, key, config_path+"/certs", "ca")
	}

	serverCert, serverKey, err := GenerateServerCertAndKey(ca, key, TUNNEL, []string{"127.0.0.1"}, []string{"localhost"})
	if err != nil {
		t.Errorf("failed to get server ca and key CN = %s err = %v", TUNNEL, err)
		return
	}
	SaveCertAndKey(serverCert, serverKey, config_path+"/certs", "cloud")
	kubeletCert, kubeletKey, err := GenerateServerCertAndKey(ca, key, K8S, []string{}, []string{})
	if err != nil {
		t.Errorf("failed to get server ca and key CN = %s err = %v", K8S, err)
		return
	}
	SaveCertAndKey(kubeletCert, kubeletKey, config_path+"/certs", "kubelet")

	kubeletClientCert, KubeletClientKey, err := GenerateClientCertAndKey(ca, key, K8S)
	if err != nil {
		t.Errorf("failed to get client ca and key CN = %s err = %v", K8S, err)
		return
	}
	SaveCertAndKey(kubeletClientCert, KubeletClientKey, config_path+"/certs", "kubelet-client")
	var edge bytes.Buffer
	err = template.Must(template.New("edge").Parse(edge_mode)).Execute(&edge, map[string]interface{}{
		"Token": token,
		"Path":  main_path,
	})
	if err != nil {
		t.Errorf("failed to prase edge_toml err: %v", err)
		return
	}
	SaveFile(config_path+"/edge_mode.toml", edge.String())
	err = InitConf("edge", config_path+"/edge_mode.toml")
	if err != nil {
		t.Errorf("failed to load edge config err = %v", err)
		return
	}
	var cloud bytes.Buffer
	err = template.Must(template.New("cloud").Parse(cloud_toml)).Execute(&cloud, map[string]interface{}{
		"Path": main_path,
	})
	if err != nil {
		t.Errorf("failed to prase cloud_toml err: %v", err)
		return
	}
	SaveFile(config_path+"/cloud_mode.toml", cloud.String())
	SaveFile(config_path+"/token", "default:"+token)
	err = InitConf("cloud", config_path+"/cloud_mode.toml")
	if err != nil {
		t.Errorf("failed to load cloud config err = %v", err)
		return
	}

}

func GenerateCa(CommonName string) (*x509.Certificate, *rsa.PrivateKey, error) {
	klog.Infof("generate ca file with TUNNEL %v", CommonName)
	certSpec := &CertConfig{
		CommonName: CommonName,
	}
	caCert, caKey, err := NewCertificateAuthority(certSpec)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate ca cert and key %+v", err)
	}
	return caCert, caKey, nil
}

func NewCertificateAuthority(config *CertConfig) (*x509.Certificate, *rsa.PrivateKey, error) {
	key, err := util.NewPrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate private key %+v", err)
	}
	cert, err := NewSelfSignedCACert(*config, key)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate cert %+v", err)
	}
	return cert, key, nil
}

func NewSelfSignedCACert(cfg CertConfig, key crypto.Signer) (*x509.Certificate, error) {
	now := time.Now()
	tmpl := x509.Certificate{
		SerialNumber: new(big.Int).SetInt64(0),
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		NotBefore:             now.UTC(),
		NotAfter:              now.Add(duration365d * 10).UTC(),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDERBytes, err := x509.CreateCertificate(cryptorand.Reader, &tmpl, &tmpl, key.Public(), key)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(certDERBytes)
}

func SaveCertAndKey(cert *x509.Certificate, key *rsa.PrivateKey, path, name string) {
	if cert == nil || key == nil {
		klog.Error("unable to write cert and key because cert or key is nil")
		return
	}
	err := certutil.WriteCert(path+"/"+name+".crt", util.EncodeCertPEM(cert))
	if err != nil {
		klog.Errorf("unable to write cert err = %v", err)
		return
	}

	encoded, err := keyutil.MarshalPrivateKeyToPEM(key)
	if err != nil {
		fmt.Printf("unable to write private key err = %v", err)
		return
	}

	if err := keyutil.WriteKey(path+"/"+name+".key", encoded); err != nil {
		klog.Errorf("unable to write private key err = %v", err)
		return
	}
	klog.Info("write cert and key succeddfully")

}

func GenerateServerCertAndKey(caCert *x509.Certificate, caKey *rsa.PrivateKey, serverCN string, ips []string, dns []string) (*x509.Certificate, *rsa.PrivateKey, error) {
	certIps := []net.IP{}
	for _, ip := range ips {
		certIps = append(certIps, net.ParseIP(ip))
	}
	//needs to verify ip first
	config := &certutil.Config{
		CommonName: serverCN,
		AltNames: certutil.AltNames{
			DNSNames: dns,
			IPs:      certIps,
		},
		Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	return generateCertAndKeyConfig(caCert, caKey, config)
}

func generateCertAndKeyConfig(caCert *x509.Certificate, caKey *rsa.PrivateKey, config *certutil.Config) (*x509.Certificate, *rsa.PrivateKey, error) {
	key, err := util.NewPrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate private key %+v", err)
	}
	cert, err := NewSignedCert(config, key, caCert, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate cert %+v", err)
	}
	return cert, key, nil
}

func NewSignedCert(cfg *certutil.Config, key crypto.Signer, caCert *x509.Certificate, caKey crypto.Signer) (*x509.Certificate, error) {
	serial, err := cryptorand.Int(cryptorand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, err
	}
	if len(cfg.CommonName) == 0 {
		return nil, fmt.Errorf("must specify a TUNNEL")
	}
	if len(cfg.Usages) == 0 {
		return nil, fmt.Errorf("must specify at least one ExtKeyUsage")
	}

	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		DNSNames:     cfg.AltNames.DNSNames,
		IPAddresses:  cfg.AltNames.IPs,
		SerialNumber: serial,
		NotBefore:    caCert.NotBefore,
		NotAfter:     time.Now().Add(duration365d).UTC(),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  cfg.Usages,
	}
	certDERBytes, err := x509.CreateCertificate(cryptorand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(certDERBytes)
}

func GenerateClientCertAndKey(caCert *x509.Certificate, caKey *rsa.PrivateKey, clientCN string) (*x509.Certificate, *rsa.PrivateKey, error) {
	klog.Infof("generate client cert and key with CommonName kubernetes-admin")
	clientCertConfig := &certutil.Config{
		CommonName:   clientCN,
		Organization: []string{"system:masters"},
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	return generateCertAndKeyConfig(caCert, caKey, clientCertConfig)
}

func SaveFile(filePath, txt string) {
	f, err := os.Create(filePath)
	if err != nil {
		klog.Errorf("fail to open the file file = %s err = %v", filePath, err)
		return
	}
	_, err = io.WriteString(f, txt)
	if err != nil {
		klog.Errorf("failed to write file file = %s err = %v", filePath, err)
		return
	}
}

func TestInitConf(t *testing.T) {
	err := InitConf("cloud", "./mode.toml")
	fmt.Println(err)
}
