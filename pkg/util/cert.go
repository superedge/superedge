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

import (
	"bytes"
	"crypto"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"time"

	"k8s.io/client-go/util/cert"
	k8scert "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
)

const (
	rsaKeySize  = 2048
	duration10y = time.Hour * 24 * 365 * 10
)

const (
	CertificateBlockType = "CERTIFICATE"

	// ECPrivateKeyBlockType is a possible value for pem.Block.Type.
	ECPrivateKeyBlockType = "EC PRIVATE KEY"

	// RSAPrivateKeyBlockType is a possible value for pem.Block.Type.
	RSAPrivateKeyBlockType = "RSA PRIVATE KEY"

	// PrivateKeyBlockType is a possible value for pem.Block.Type.
	PrivateKeyBlockType = "PRIVATE KEY"

	// PublicKeyBlockType is a possible value for pem.Block.Type.
	PublicKeyBlockType = "PUBLIC KEY"
)

// EncodeCertPEM returns PEM-endcoded certificate data
func EncodeCertPEM(cert *x509.Certificate) []byte {
	block := pem.Block{
		Type:  CertificateBlockType, //tobe replaced
		Bytes: cert.Raw,
	}
	return pem.EncodeToMemory(&block)
}

// ParseCertsPEM returns the x509.Certificates contained in the given PEM-encoded byte array
// Returns an error if a certificate could not be parsed, or if the data does not contain any certificates
func ParseCertsPEM(pemCerts []byte) ([]*x509.Certificate, error) {
	ok := false
	certs := []*x509.Certificate{}
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}
		// Only use PEM "CERTIFICATE" blocks without extra headers
		if block.Type != CertificateBlockType || len(block.Headers) != 0 {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return certs, err
		}

		certs = append(certs, cert)
		ok = true
	}

	if !ok {
		return certs, fmt.Errorf("data does not contain any valid RSA or ECDSA certificates")
	}
	return certs, nil
}

// ParsePrivateKeyPEM returns a private key parsed from a PEM block in the supplied data.
// Recognizes PEM blocks for "EC PRIVATE KEY", "RSA PRIVATE KEY", or "PRIVATE KEY"
func parsePrivateKeyPEM(keyData []byte) (interface{}, error) {
	var privateKeyPemBlock *pem.Block
	for {
		privateKeyPemBlock, keyData = pem.Decode(keyData)
		if privateKeyPemBlock == nil {
			break
		}

		switch privateKeyPemBlock.Type {
		case ECPrivateKeyBlockType:
			// ECDSA Private Key in ASN.1 format
			if key, err := x509.ParseECPrivateKey(privateKeyPemBlock.Bytes); err == nil {
				return key, nil
			}
		case RSAPrivateKeyBlockType:
			// RSA Private Key in PKCS#1 format
			if key, err := x509.ParsePKCS1PrivateKey(privateKeyPemBlock.Bytes); err == nil {
				return key, nil
			}
		case PrivateKeyBlockType:
			// RSA or ECDSA Private Key in unencrypted PKCS#8 format
			if key, err := x509.ParsePKCS8PrivateKey(privateKeyPemBlock.Bytes); err == nil {
				return key, nil
			}
		}
		// tolerate non-key PEM blocks for compatibility with things like "EC PARAMETERS" blocks
		// originally, only the first PEM block was parsed and expected to be a key block
	}
	// we read all the PEM blocks and didn't recognize one
	return nil, fmt.Errorf("data does not contain a valid RSA or ECDSA private key")
}

func ParsePrivateKeyPEMRSA(keyData []byte) (*rsa.PrivateKey, error) {
	privKey, err := parsePrivateKeyPEM(keyData)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse the private key %+v", err)
	}
	// Allow RSA format only
	var key *rsa.PrivateKey
	switch k := privKey.(type) {
	case *rsa.PrivateKey:
		key = k
	default:
		return nil, fmt.Errorf("the private key isn't in RSA format")
	}
	return key, nil
}

// NewPrivateKey creates an RSA private key change fa
func NewPrivateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(cryptorand.Reader, rsaKeySize)
}

//Tobe optimized
func GenerateCertAndKeyConfig(caCert *x509.Certificate, caKey *rsa.PrivateKey,
	config *cert.Config) (*x509.Certificate, *rsa.PrivateKey, error) {
	return generateCertAndKeyConfig(caCert, caKey, config)
}

func GetServiceCertByRootca(commonName, organization string, dns []string, rootCA *x509.Certificate, key *rsa.PrivateKey) ([]byte, []byte, error) {
	svCert, svKey, err := GenerateCertAndKeyConfig(rootCA, key, &k8scert.Config{
		CommonName:   commonName,
		Organization: []string{organization},
		AltNames: k8scert.AltNames{
			DNSNames: dns,
		},
		Usages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth},
	})

	svKeyBytes, err := keyutil.MarshalPrivateKeyToPEM(svKey)
	if err != nil {
		return nil, nil, err
	}
	return EncodeCertPEM(svCert), svKeyBytes, nil
}

func GenerateCA(organization string) (caBundle string, caCert *x509.Certificate, caPrivKey *rsa.PrivateKey, err error) {
	serial, err := cryptorand.Int(cryptorand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return
	}
	ca := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{organization},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err = rsa.GenerateKey(cryptorand.Reader, 2048)
	if err != nil {
		return
	}
	caBytes, err := x509.CreateCertificate(cryptorand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return
	}
	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  CertificateBlockType,
		Bytes: caBytes,
	})
	caBundle = base64.StdEncoding.EncodeToString(caPEM.Bytes())
	caCerts, err := ParseCertsPEM(caPEM.Bytes())
	if err != nil {
		return
	}
	caCert = caCerts[0]
	return
}

func generateCertAndKeyConfig(caCert *x509.Certificate, caKey *rsa.PrivateKey,
	config *cert.Config) (*x509.Certificate, *rsa.PrivateKey, error) {
	key, err := NewPrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate private key %+v", err)
	}
	cert, err := NewSignedCert(config, key, caCert, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate cert %+v", err)
	}
	return cert, key, nil
}

func GenerateClientCertAndKey(caCert *x509.Certificate, caKey *rsa.PrivateKey,
	clientCN string) (*x509.Certificate, *rsa.PrivateKey, error) {
	clientCertConfig := &cert.Config{
		CommonName:   clientCN,
		Organization: []string{"system:masters"},
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	return generateCertAndKeyConfig(caCert, caKey, clientCertConfig)
}

// NewSignedCert creates a signed certificate using the given CA certificate and key
func NewSignedCert(cfg *cert.Config, key crypto.Signer,
	caCert *x509.Certificate, caKey crypto.Signer) (*x509.Certificate, error) {
	serial, err := cryptorand.Int(cryptorand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, err
	}
	if len(cfg.CommonName) == 0 {
		return nil, fmt.Errorf("must specify a CommonName")
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
		NotAfter:     time.Now().Add(duration10y).UTC(),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  cfg.Usages,
	}
	certDERBytes, err := x509.CreateCertificate(cryptorand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(certDERBytes)
}

func GetRandToken(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func CertHasExpired(cert *x509.Certificate) bool {
	now := time.Now()
	return now.Before(cert.NotBefore) || now.After(cert.NotAfter)
}
