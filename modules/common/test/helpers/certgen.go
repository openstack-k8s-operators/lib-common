/*
Copyright 2026 Red hat
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
	http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for specific language governing permissions and
limitations under the License.
*/

package helpers

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// CertConfig defines the configuration used to generate a test certificate
type CertConfig struct {
	// KeyType is the type of key to generate: "rsa", "ecdsa", or "ed25519"
	KeyType string
	// KeySize is the size of the key in bits (for RSA: 2048, 3072, 4096; for ECDSA: 256, 384, 521)
	KeySize int
	// CommonName for the certificate
	CommonName string
	// DNSNames for Subject Alternative Names
	DNSNames []string
	// Organization name
	Organization string
	// NotBefore time (defaults to now)
	NotBefore time.Time
	// NotAfter time (defaults to now + 1 year)
	NotAfter time.Time
}

// GeneratedCert contains the generated certificate and key in PEM format
type GeneratedCert struct {
	CertPEM []byte
	KeyPEM  []byte
	CAPEM   []byte
}

// DefaultCertConfig returns a default certificate configuration with RSA-2048 and some sane
// defaults
func DefaultCertConfig() *CertConfig {
	return &CertConfig{
		KeyType:      "rsa",
		KeySize:      2048,
		CommonName:   "test.example.com",
		DNSNames:     []string{"test.example.com", "localhost"},
		Organization: "Test Org",
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
	}
}

// RSA2048CertConfig returns a configuration for an RSA 2048-bit certificate
// This certificate is TLS 1.3 compatible but is not PQC-safe
func RSA2048CertConfig() *CertConfig {
	cfg := DefaultCertConfig()
	cfg.KeyType = "rsa"
	cfg.KeySize = 2048
	return cfg
}

// RSA3072CertConfig returns a configuration for an RSA 3072-bit certificate
// This certificate is TLS 1.3 compatible and is quantum-resistant
func RSA3072CertConfig() *CertConfig {
	cfg := DefaultCertConfig()
	cfg.KeyType = "rsa"
	cfg.KeySize = 3072
	return cfg
}

// RSA4096CertConfig returns a configuration for an RSA 4096-bit certificate
// This certificate is TLS 1.3 compatible and is quantum-resistant
func RSA4096CertConfig() *CertConfig {
	cfg := DefaultCertConfig()
	cfg.KeyType = "rsa"
	cfg.KeySize = 4096
	return cfg
}

// ECDSAP256CertConfig returns a configuration for an ECDSA P-256 certificate
// This certificate is TLS 1.3 compatible but is not PQC-safe
func ECDSAP256CertConfig() *CertConfig {
	cfg := DefaultCertConfig()
	cfg.KeyType = "ecdsa"
	cfg.KeySize = 256
	return cfg
}

// ECDSAP384CertConfig returns a configuration for an ECDSA P-384 certificate
// This certificate is TLS 1.3 compatible and is quantum-resistant
func ECDSAP384CertConfig() *CertConfig {
	cfg := DefaultCertConfig()
	cfg.KeyType = "ecdsa"
	cfg.KeySize = 384
	return cfg
}

// ECDSAP521CertConfig returns a configuration for an ECDSA P-521 certificate
// This certificate is TLS 1.3 compatible and is quantum-resistant
func ECDSAP521CertConfig() *CertConfig {
	cfg := DefaultCertConfig()
	cfg.KeyType = "ecdsa"
	cfg.KeySize = 521
	return cfg
}

// Ed25519CertConfig returns a configuration for an Ed25519 certificate
// This certificate is TLS 1.3 compatible but is not PQC-safe
func Ed25519CertConfig() *CertConfig {
	cfg := DefaultCertConfig()
	cfg.KeyType = "ed25519"
	cfg.KeySize = 0 // Ed25519 has a fixed key size
	return cfg
}

// GenerateCertificate generates a self-signed certificate from the provided configuration
func GenerateCertificate(config *CertConfig) (*GeneratedCert, error) {
	// Generate the private key based on type
	var privateKey interface{}
	var err error

	switch config.KeyType {
	case "rsa":
		privateKey, err = rsa.GenerateKey(rand.Reader, config.KeySize)
	case "ecdsa":
		var curve elliptic.Curve
		switch config.KeySize {
		case 256:
			curve = elliptic.P256()
		case 384:
			curve = elliptic.P384()
		case 521:
			curve = elliptic.P521()
		default:
			curve = elliptic.P256()
		}
		privateKey, err = ecdsa.GenerateKey(curve, rand.Reader)
	case "ed25519":
		_, privateKey, err = ed25519.GenerateKey(rand.Reader)
	default:
		privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	}

	if err != nil {
		return nil, err
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   config.CommonName,
			Organization: []string{config.Organization},
		},
		DNSNames:              config.DNSNames,
		NotBefore:             config.NotBefore,
		NotAfter:              config.NotAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create the certificate
	var publicKey interface{}
	switch key := privateKey.(type) {
	case *rsa.PrivateKey:
		publicKey = &key.PublicKey
	case *ecdsa.PrivateKey:
		publicKey = &key.PublicKey
	case ed25519.PrivateKey:
		publicKey = key.Public()
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey, privateKey)
	if err != nil {
		return nil, err
	}

	// Encode the certificate with PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode private key with PEM
	var keyPEM []byte
	switch key := privateKey.(type) {
	case *rsa.PrivateKey:
		keyPEM = pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		})
	case *ecdsa.PrivateKey:
		keyBytes, err := x509.MarshalECPrivateKey(key)
		if err != nil {
			return nil, err
		}
		keyPEM = pem.EncodeToMemory(&pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: keyBytes,
		})
	case ed25519.PrivateKey:
		keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			return nil, err
		}
		keyPEM = pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: keyBytes,
		})
	}

	return &GeneratedCert{
		CertPEM: certPEM,
		KeyPEM:  keyPEM,
		CAPEM:   certPEM, // Self-signed, so the CA is the same as the certificate itself
	}, nil
}

// CreateCertSecretWithConfig creates a Kubernetes secret with a generated, self-signed certificate
func (tc *TestHelper) CreateCertSecretWithConfig(
	name types.NamespacedName,
	config *CertConfig,
) (*corev1.Secret, *GeneratedCert, error) {
	cert, err := GenerateCertificate(config)
	if err != nil {
		return nil, nil, err
	}

	data := map[string][]byte{
		"tls.crt": cert.CertPEM,
		"tls.key": cert.KeyPEM,
		"ca.crt":  cert.CAPEM,
	}

	secret := tc.CreateSecret(name, data)
	return secret, cert, nil
}

// CreateRSA2048CertSecret creates a Kubernetes secret with an RSA 2048-bit certificate
func (tc *TestHelper) CreateRSA2048CertSecret(name types.NamespacedName) (*corev1.Secret, *GeneratedCert, error) {
	return tc.CreateCertSecretWithConfig(name, RSA2048CertConfig())
}

// CreateRSA3072CertSecret creates a Kubernetes secret with an RSA 3072-bit certificate
// (quantum-resistant)
func (tc *TestHelper) CreateRSA3072CertSecret(name types.NamespacedName) (*corev1.Secret,
	*GeneratedCert, error) {
	return tc.CreateCertSecretWithConfig(name, RSA3072CertConfig())
}

// CreateRSA4096CertSecret creates a Kubernetes secret with an RSA 4096-bit certificate
// (quantum-resistant)
func (tc *TestHelper) CreateRSA4096CertSecret(name types.NamespacedName) (*corev1.Secret,
	*GeneratedCert, error) {
	return tc.CreateCertSecretWithConfig(name, RSA4096CertConfig())
}

// CreateECDSAP256CertSecret creates a Kubernetes secret with an ECDSA P-256 certificate
func (tc *TestHelper) CreateECDSAP256CertSecret(name types.NamespacedName) (*corev1.Secret,
	*GeneratedCert, error) {
	return tc.CreateCertSecretWithConfig(name, ECDSAP256CertConfig())
}

// CreateECDSAP384CertSecret creates a Kubernetes secret with an ECDSA P-384 certificate
// (quantum-resistant)
func (tc *TestHelper) CreateECDSAP384CertSecret(name types.NamespacedName) (*corev1.Secret,
	*GeneratedCert, error) {
	return tc.CreateCertSecretWithConfig(name, ECDSAP384CertConfig())
}

// CreateECDSAP521CertSecret creates a Kubernetes secret with an ECDSA P-521 certificate
// (quantum-resistant)
func (tc *TestHelper) CreateECDSAP521CertSecret(name types.NamespacedName) (*corev1.Secret,
	*GeneratedCert, error) {
	return tc.CreateCertSecretWithConfig(name, ECDSAP521CertConfig())
}

// CreateEd25519CertSecret creates a Kubernetes secret with an Ed25519 certificate
func (tc *TestHelper) CreateEd25519CertSecret(name types.NamespacedName) (*corev1.Secret, *GeneratedCert, error) {
	return tc.CreateCertSecretWithConfig(name, Ed25519CertConfig())
}
