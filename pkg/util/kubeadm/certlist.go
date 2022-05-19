/*
Copyright 2019 The Kubernetes Authors.
Copyright 2020 Authors of SuperEdge - file modified.

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

package kubeadm

import (
	"crypto"
	"crypto/x509"

	"github.com/pkg/errors"

	certutil "k8s.io/client-go/util/cert"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
)

type configMutatorsFunc func(*kubeadmapi.InitConfiguration, *pkiutil.CertConfig) error

// KubeadmCert represents a certificate that Kubeadm will create to function properly.
type KubeadmCert struct {
	Name     string
	LongName string
	BaseName string
	CAName   string
	// Some attributes will depend on the InitConfiguration, only known at runtime.
	// These functions will be run in series, passed both the InitConfiguration and a cert Config.
	configMutators []configMutatorsFunc
	config         pkiutil.CertConfig
}

const (
	IngressClientCertAndKeyBaseName = "tunnel-anp-client"
)

// GetConfig returns the definition for the given cert given the provided InitConfiguration
func (k *KubeadmCert) GetConfig(ic *kubeadmapi.InitConfiguration) (*pkiutil.CertConfig, error) {
	for _, f := range k.configMutators {
		if err := f(ic, &k.config); err != nil {
			return nil, err
		}
	}

	k.config.PublicKeyAlgorithm = ic.ClusterConfiguration.PublicKeyAlgorithm()
	return &k.config, nil
}

// CreateFromCA makes and writes a certificate using the given CA cert and key.
func (k *KubeadmCert) CreateFromCA(ic *kubeadmapi.InitConfiguration, caCert *x509.Certificate, caKey crypto.Signer) error {
	cfg, err := k.GetConfig(ic)
	if err != nil {
		return errors.Wrapf(err, "couldn't create %q certificate", k.Name)
	}
	cert, key, err := pkiutil.NewCertAndKey(caCert, caKey, cfg)
	if err != nil {
		return err
	}
	err = writeCertificateFilesIfNotExist(
		ic.CertificatesDir,
		k.BaseName,
		caCert,
		cert,
		key,
		cfg,
	)

	if err != nil {
		return errors.Wrapf(err, "failed to write or validate certificate %q", k.Name)
	}

	return nil
}

// CreateAsCA creates a certificate authority, writing the files to disk and also returning the created CA so it can be used to sign child certs.
func (k *KubeadmCert) CreateAsCA(ic *kubeadmapi.InitConfiguration) (*x509.Certificate, crypto.Signer, error) {
	cfg, err := k.GetConfig(ic)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "couldn't get configuration for %q CA certificate", k.Name)
	}
	caCert, caKey, err := pkiutil.NewCertificateAuthority(cfg)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "couldn't generate %q CA certificate", k.Name)
	}

	err = writeCertificateAuthorityFilesIfNotExist(
		ic.CertificatesDir,
		k.BaseName,
		caCert,
		caKey,
	)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "couldn't write out %q CA certificate", k.Name)
	}

	return caCert, caKey, nil
}

// CertificateTree is represents a one-level-deep tree, mapping a CA to the certs that depend on it.
type CertificateTree map[*KubeadmCert]Certificates

// CreateTree creates the CAs, certs signed by the CAs, and writes them all to disk.
func (t CertificateTree) CreateTree(ic *kubeadmapi.InitConfiguration) error {
	for ca, leaves := range t {
		cfg, err := ca.GetConfig(ic)
		if err != nil {
			return err
		}

		var caKey crypto.Signer

		caCert, err := pkiutil.TryLoadCertFromDisk(ic.CertificatesDir, ca.BaseName)
		if err == nil {
			// Cert exists already, make sure it's valid
			if !caCert.IsCA {
				return errors.Errorf("certificate %q is not a CA", ca.Name)
			}
			// Try and load a CA Key
			caKey, err = pkiutil.TryLoadKeyFromDisk(ic.CertificatesDir, ca.BaseName)
			if err != nil {
				// If there's no CA key, make sure every certificate exists.
				for _, leaf := range leaves {
					cl := certKeyLocation{
						pkiDir:   ic.CertificatesDir,
						baseName: leaf.BaseName,
						uxName:   leaf.Name,
					}
					if err := validateSignedCertWithCA(cl, caCert); err != nil {
						return errors.Wrapf(err, "could not load expected certificate %q or validate the existence of key %q for it", leaf.Name, ca.Name)
					}
				}
				continue
			}
			// CA key exists; just use that to create new certificates.
		} else {
			// CACert doesn't already exist, create a new cert and key.
			caCert, caKey, err = pkiutil.NewCertificateAuthority(cfg)
			if err != nil {
				return err
			}

			err = writeCertificateAuthorityFilesIfNotExist(
				ic.CertificatesDir,
				ca.BaseName,
				caCert,
				caKey,
			)
			if err != nil {
				return err
			}
		}

		for _, leaf := range leaves {
			if err := leaf.CreateFromCA(ic, caCert, caKey); err != nil {
				return err
			}
		}
	}
	return nil
}

// CertificateMap is a flat map of certificates, keyed by Name.
type CertificateMap map[string]*KubeadmCert

// CertTree returns a one-level-deep tree, mapping a CA cert to an array of certificates that should be signed by it.
func (m CertificateMap) CertTree() (CertificateTree, error) {
	caMap := make(CertificateTree)

	for _, cert := range m {
		if cert.CAName == "" {
			if _, ok := caMap[cert]; !ok {
				caMap[cert] = []*KubeadmCert{}
			}
		} else {
			ca, ok := m[cert.CAName]
			if !ok {
				return nil, errors.Errorf("certificate %q references unknown CA %q", cert.Name, cert.CAName)
			}
			caMap[ca] = append(caMap[ca], cert)
		}
	}

	return caMap, nil
}

// Certificates is a list of Certificates that Kubeadm should create.
type Certificates []*KubeadmCert

// AsMap returns the list of certificates as a map, keyed by name.
func (c Certificates) AsMap() CertificateMap {
	certMap := make(map[string]*KubeadmCert)
	for _, cert := range c {
		certMap[cert.Name] = cert
	}

	return certMap
}

// GetDefaultCertList returns  all of the certificates kubeadm requires to function.
func GetEdgeCertList() Certificates {
	return Certificates{
		&KubeadmCertRootCA,
		&KubeadmIngress,
	}
}

var (
	// KubeadmCertRootCA is the definition of the Kubernetes Root CA for the API Server and kubelet.
	KubeadmCertRootCA = KubeadmCert{
		Name:     "ca",
		LongName: "self-signed Kubernetes CA to provision identities for other Kubernetes components",
		BaseName: kubeadmconstants.CACertAndKeyBaseName,
		config: pkiutil.CertConfig{
			Config: certutil.Config{
				CommonName: "kubernetes",
			},
		},
	}

	KubeadmIngress = KubeadmCert{
		Name:     "ingress-client",
		LongName: "self-signed CA to provision identities for etcd",
		BaseName: IngressClientCertAndKeyBaseName,
		CAName:   "ca",
		config: pkiutil.CertConfig{
			Config: certutil.Config{
				CommonName: "ingress-client",
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			},
		},
	}
)

func makeAltNamesMutator(f func(*kubeadmapi.InitConfiguration) (*certutil.AltNames, error)) configMutatorsFunc {
	return func(mc *kubeadmapi.InitConfiguration, cc *pkiutil.CertConfig) error {
		altNames, err := f(mc)
		if err != nil {
			return err
		}
		cc.AltNames = *altNames
		return nil
	}
}

func setCommonNameToNodeName() configMutatorsFunc {
	return func(mc *kubeadmapi.InitConfiguration, cc *pkiutil.CertConfig) error {
		cc.CommonName = mc.NodeRegistration.Name
		return nil
	}
}
