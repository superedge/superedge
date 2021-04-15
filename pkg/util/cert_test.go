package util

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"gotest.tools/assert"
	"math/big"
	"testing"
	"time"
)

func TestCertHasExpired(t *testing.T) {
	now := time.Now()

	oldCert := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "superedge",
			Organization: []string{"superedge"},
		},
		NotBefore: time.Unix(1000, 0),
		NotAfter:  time.Unix(10000, 0),
		KeyUsage:  x509.KeyUsageCertSign,
	}
	assert.Check(t, CertHasExpired(oldCert))

	newCert := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "superedge",
			Organization: []string{"superedge"},
		},
		NotBefore: now.Add(time.Hour),
		NotAfter:  now.Add(2 * time.Hour),
		KeyUsage:  x509.KeyUsageCertSign,
	}
	assert.Check(t, CertHasExpired(newCert))

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "superedge",
			Organization: []string{"superedge"},
		},
		NotBefore: time.Unix(1000, 0),
		NotAfter:  now.Add(2 * time.Hour),
		KeyUsage:  x509.KeyUsageCertSign,
	}
	assert.Check(t, !CertHasExpired(cert))
}
