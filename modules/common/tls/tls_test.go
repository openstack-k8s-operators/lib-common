/*
Copyright 2023 Red Hat

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

package tls

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	. "github.com/onsi/gomega" // nolint:revive
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	helpers "github.com/openstack-k8s-operators/lib-common/modules/common/test/helpers"
)

func TestAPIEnabled(t *testing.T) {
	tests := []struct {
		name  string
		endpt service.Endpoint
		api   *APIService
		want  bool
	}{
		{
			name:  "empty API",
			endpt: service.EndpointInternal,
			api:   &APIService{},
			want:  false,
		},
		{
			name:  "Internal SecretName nil",
			endpt: service.EndpointInternal,
			api: &APIService{
				Internal: GenericService{SecretName: nil},
				Public:   GenericService{SecretName: nil},
			},
			want: false,
		},
		{
			name:  "Internal SecretName defined",
			endpt: service.EndpointInternal,
			api: &APIService{
				Internal: GenericService{SecretName: ptr.To("foo")},
				Public:   GenericService{SecretName: nil},
			},
			want: true,
		},
		{
			name:  "Public SecretName nil",
			endpt: service.EndpointPublic,
			api: &APIService{
				Internal: GenericService{SecretName: nil},
				Public:   GenericService{SecretName: nil},
			},
			want: false,
		},
		{
			name:  "Public SecretName defined",
			endpt: service.EndpointPublic,
			api: &APIService{
				Internal: GenericService{SecretName: nil},
				Public:   GenericService{SecretName: ptr.To("foo")},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(tt.api.Enabled(tt.endpt)).To(BeEquivalentTo(tt.want))
		})
	}
}

func TestGenericServiceToService(t *testing.T) {
	tests := []struct {
		name    string
		service *GenericService
		want    Service
	}{
		{
			name:    "empty APIService",
			service: &GenericService{},
			want:    Service{},
		},
		{
			name: "APIService SecretName specified",
			service: &GenericService{
				SecretName: ptr.To("foo"),
			},
			want: Service{
				SecretName: "foo",
			},
		},
		{
			name: "APIService SecretName nil",
			service: &GenericService{
				SecretName: nil,
			},
			want: Service{
				SecretName: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			s, err := tt.service.ToService()
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(s).NotTo(BeNil())
		})
	}
}

func TestServiceCreateVolumeMounts(t *testing.T) {
	tests := []struct {
		name    string
		service *Service
		id      string
		want    []corev1.VolumeMount
	}{
		{
			name:    "No TLS Secret",
			service: &Service{},
			id:      "foo",
			want:    []corev1.VolumeMount{},
		},
		{
			name:    "Only TLS Secret",
			service: &Service{SecretName: "cert-secret"},
			id:      "foo",
			want: []corev1.VolumeMount{
				{
					MountPath: "/var/lib/config-data/tls/certs/foo.crt",
					Name:      "foo-tls-certs",
					ReadOnly:  true,
					SubPath:   "tls.crt",
				},
				{
					MountPath: "/var/lib/config-data/tls/private/foo.key",
					Name:      "foo-tls-certs",
					ReadOnly:  true,
					SubPath:   "tls.key",
				},
			},
		},
		{
			name: "TLS and CA Secrets",
			service: &Service{
				SecretName: "cert-secret",
				CaMount:    ptr.To("/var/lib/config-data/ca-bundle/ca.crt"),
			},
			id: "foo",
			want: []corev1.VolumeMount{
				{
					MountPath: "/var/lib/config-data/tls/certs/foo.crt",
					Name:      "foo-tls-certs",
					ReadOnly:  true,
					SubPath:   "tls.crt",
				},
				{
					MountPath: "/var/lib/config-data/tls/private/foo.key",
					Name:      "foo-tls-certs",
					ReadOnly:  true,
					SubPath:   "tls.key",
				},
				{
					MountPath: "/var/lib/config-data/ca-bundle/ca.crt",
					Name:      "foo-tls-certs",
					ReadOnly:  true,
					SubPath:   "ca.crt",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			mounts := tt.service.CreateVolumeMounts(tt.id)
			g.Expect(mounts).To(HaveLen(len(tt.want)))
			g.Expect(mounts).To(Equal(tt.want))
		})
	}
}

func TestServiceCreateVolume(t *testing.T) {
	tests := []struct {
		name    string
		service *Service
		id      string
		want    corev1.Volume
	}{
		{
			name:    "No Secrets",
			service: &Service{},
			want:    corev1.Volume{},
		},
		{
			name:    "Only TLS Secret",
			service: &Service{SecretName: "cert-secret"},
			id:      "foo",
			want: corev1.Volume{
				Name: "foo-tls-certs",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  "cert-secret",
						DefaultMode: ptr.To[int32](0400),
					},
				},
			},
		},
		{
			name:    "Only TLS Secret no serviceID",
			service: &Service{SecretName: "cert-secret"},
			want: corev1.Volume{
				Name: "default-tls-certs",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  "cert-secret",
						DefaultMode: ptr.To[int32](0400),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			volume := tt.service.CreateVolume(tt.id)
			g.Expect(volume).To(Equal(tt.want))
		})
	}
}

func TestCACreateVolumeMounts(t *testing.T) {
	tests := []struct {
		name          string
		ca            *Ca
		caBundleMount *string
		want          []corev1.VolumeMount
	}{
		{
			name: "Empty Ca",
			ca:   &Ca{},
			want: []corev1.VolumeMount{},
		},
		{
			name: "Only CaBundleSecretName no caBundleMount",
			ca: &Ca{
				CaBundleSecretName: "ca-secret",
			},
			want: []corev1.VolumeMount{
				{
					MountPath: "/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem",
					Name:      "combined-ca-bundle",
					ReadOnly:  true,
					SubPath:   "tls-ca-bundle.pem",
				},
			},
		},
		{
			name: "CaBundleSecretName and caBundleMount",
			ca: &Ca{
				CaBundleSecretName: "ca-secret",
			},
			caBundleMount: ptr.To("/mount/my/ca.crt"),
			want: []corev1.VolumeMount{
				{
					MountPath: "/mount/my/ca.crt",
					Name:      "combined-ca-bundle",
					ReadOnly:  true,
					SubPath:   "tls-ca-bundle.pem",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			mounts := tt.ca.CreateVolumeMounts(tt.caBundleMount)
			g.Expect(mounts).To(HaveLen(len(tt.want)))
			g.Expect(mounts).To(Equal(tt.want))
		})
	}
}

func TestCaCreateVolume(t *testing.T) {
	tests := []struct {
		name string
		ca   *Ca
		want corev1.Volume
	}{
		{
			name: "Empty Ca",
			ca:   &Ca{},
			want: corev1.Volume{},
		},
		{
			name: "Set CaBundleSecretName",
			ca: &Ca{
				CaBundleSecretName: "ca-secret",
			},
			want: corev1.Volume{
				Name: "combined-ca-bundle",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  "ca-secret",
						DefaultMode: ptr.To[int32](0444),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			volume := tt.ca.CreateVolume()
			g.Expect(volume).To(Equal(tt.want))
		})
	}
}

func TestAnalyzeCertificate(t *testing.T) {
	tests := []struct {
		name               string
		certGenerator      func() ([]byte, error)
		wantTLS13          bool
		wantPQCSafe        bool
		wantKeyAlgorithm   string
		wantKeySize        int
		wantMinCipherCount int
	}{
		{
			name: "RSA 2048-bit certificate (TLS 1.3 compatible, not PQC-safe)",
			certGenerator: func() ([]byte, error) {
				cfg := &helpers.CertConfig{
					KeyType:      "rsa",
					KeySize:      2048,
					CommonName:   "test-rsa-2048.example.com",
					DNSNames:     []string{"test-rsa-2048.example.com"},
					Organization: "Test Org",
					NotBefore:    ptr.To(time.Now()).Add(-1 * time.Hour),
					NotAfter:     ptr.To(time.Now()).Add(24 * time.Hour),
				}
				cert, err := helpers.GenerateCertificate(cfg)
				if err != nil {
					return nil, err
				}
				return cert.CertPEM, nil
			},
			wantTLS13:          true,
			wantPQCSafe:        false,
			wantKeyAlgorithm:   "RSA",
			wantKeySize:        2048,
			wantMinCipherCount: 1,
		},
		{
			name: "RSA 3072-bit certificate (TLS 1.3 compatible, PQC-safe)",
			certGenerator: func() ([]byte, error) {
				cfg := &helpers.CertConfig{
					KeyType:      "rsa",
					KeySize:      3072,
					CommonName:   "test-rsa-3072.example.com",
					DNSNames:     []string{"test-rsa-3072.example.com"},
					Organization: "Test Org",
					NotBefore:    ptr.To(time.Now()).Add(-1 * time.Hour),
					NotAfter:     ptr.To(time.Now()).Add(24 * time.Hour),
				}
				cert, err := helpers.GenerateCertificate(cfg)
				if err != nil {
					return nil, err
				}
				return cert.CertPEM, nil
			},
			wantTLS13:          true,
			wantPQCSafe:        true,
			wantKeyAlgorithm:   "RSA",
			wantKeySize:        3072,
			wantMinCipherCount: 1,
		},
		{
			name: "RSA 4096-bit certificate (TLS 1.3 compatible, PQC-safe)",
			certGenerator: func() ([]byte, error) {
				cfg := &helpers.CertConfig{
					KeyType:      "rsa",
					KeySize:      4096,
					CommonName:   "test-rsa-4096.example.com",
					DNSNames:     []string{"test-rsa-4096.example.com"},
					Organization: "Test Org",
					NotBefore:    ptr.To(time.Now()).Add(-1 * time.Hour),
					NotAfter:     ptr.To(time.Now()).Add(24 * time.Hour),
				}
				cert, err := helpers.GenerateCertificate(cfg)
				if err != nil {
					return nil, err
				}
				return cert.CertPEM, nil
			},
			wantTLS13:          true,
			wantPQCSafe:        true,
			wantKeyAlgorithm:   "RSA",
			wantKeySize:        4096,
			wantMinCipherCount: 1,
		},
		{
			name: "ECDSA P-256 certificate (TLS 1.3 compatible, not PQC-safe)",
			certGenerator: func() ([]byte, error) {
				cfg := &helpers.CertConfig{
					KeyType:      "ecdsa",
					KeySize:      256,
					CommonName:   "test-ecdsa-p256.example.com",
					DNSNames:     []string{"test-ecdsa-p256.example.com"},
					Organization: "Test Org",
					NotBefore:    ptr.To(time.Now()).Add(-1 * time.Hour),
					NotAfter:     ptr.To(time.Now()).Add(24 * time.Hour),
				}
				cert, err := helpers.GenerateCertificate(cfg)
				if err != nil {
					return nil, err
				}
				return cert.CertPEM, nil
			},
			wantTLS13:          true,
			wantPQCSafe:        false,
			wantKeyAlgorithm:   "ECDSA",
			wantKeySize:        256,
			wantMinCipherCount: 1,
		},
		{
			name: "ECDSA P-384 certificate (TLS 1.3 compatible, PQC-safe)",
			certGenerator: func() ([]byte, error) {
				cfg := &helpers.CertConfig{
					KeyType:      "ecdsa",
					KeySize:      384,
					CommonName:   "test-ecdsa-p384.example.com",
					DNSNames:     []string{"test-ecdsa-p384.example.com"},
					Organization: "Test Org",
					NotBefore:    ptr.To(time.Now()).Add(-1 * time.Hour),
					NotAfter:     ptr.To(time.Now()).Add(24 * time.Hour),
				}
				cert, err := helpers.GenerateCertificate(cfg)
				if err != nil {
					return nil, err
				}
				return cert.CertPEM, nil
			},
			wantTLS13:          true,
			wantPQCSafe:        true,
			wantKeyAlgorithm:   "ECDSA",
			wantKeySize:        384,
			wantMinCipherCount: 1,
		},
		{
			name: "ECDSA P-521 certificate (TLS 1.3 compatible, PQC-safe)",
			certGenerator: func() ([]byte, error) {
				cfg := &helpers.CertConfig{
					KeyType:      "ecdsa",
					KeySize:      521,
					CommonName:   "test-ecdsa-p521.example.com",
					DNSNames:     []string{"test-ecdsa-p521.example.com"},
					Organization: "Test Org",
					NotBefore:    ptr.To(time.Now()).Add(-1 * time.Hour),
					NotAfter:     ptr.To(time.Now()).Add(24 * time.Hour),
				}
				cert, err := helpers.GenerateCertificate(cfg)
				if err != nil {
					return nil, err
				}
				return cert.CertPEM, nil
			},
			wantTLS13:          true,
			wantPQCSafe:        true,
			wantKeyAlgorithm:   "ECDSA",
			wantKeySize:        521,
			wantMinCipherCount: 1,
		},
		{
			name: "Ed25519 certificate (TLS 1.3 compatible, not PQC-safe)",
			certGenerator: func() ([]byte, error) {
				cfg := &helpers.CertConfig{
					KeyType:      "ed25519",
					KeySize:      0,
					CommonName:   "test-ed25519.example.com",
					DNSNames:     []string{"test-ed25519.example.com"},
					Organization: "Test Org",
					NotBefore:    ptr.To(time.Now()).Add(-1 * time.Hour),
					NotAfter:     ptr.To(time.Now()).Add(24 * time.Hour),
				}
				cert, err := helpers.GenerateCertificate(cfg)
				if err != nil {
					return nil, err
				}
				return cert.CertPEM, nil
			},
			wantTLS13:          true,
			wantPQCSafe:        false,
			wantKeyAlgorithm:   "Ed25519",
			wantKeySize:        256,
			wantMinCipherCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			certPEM, err := tt.certGenerator()
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(certPEM).NotTo(BeEmpty())

			analysis, err := AnalyzeCertificate(certPEM)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(analysis).NotTo(BeNil())

			g.Expect(analysis.SupportsTLS13).To(Equal(tt.wantTLS13),
				"TLS 1.3 support should be %v", tt.wantTLS13)
			g.Expect(analysis.IsPQCSafe).To(Equal(tt.wantPQCSafe),
				"PQC safety should be %v", tt.wantPQCSafe)
			g.Expect(analysis.KeyAlgorithm).To(Equal(tt.wantKeyAlgorithm),
				"Key algorithm should be %s", tt.wantKeyAlgorithm)
			g.Expect(analysis.KeySize).To(Equal(tt.wantKeySize),
				"Key size should be %d", tt.wantKeySize)
			g.Expect(analysis.SignatureAlgorithm).NotTo(BeEmpty(),
				"Signature algorithm should not be empty")
			g.Expect(len(analysis.CipherSuites)).To(BeNumerically(">=", tt.wantMinCipherCount),
				"Should have at least %d cipher suites", tt.wantMinCipherCount)
		})
	}
}

func TestAnalyzeCertificateErrors(t *testing.T) {
	tests := []struct {
		name      string
		certPEM   []byte
		wantError bool
	}{
		{
			name:      "Empty PEM data",
			certPEM:   []byte{},
			wantError: true,
		},
		{
			name:      "Invalid PEM data",
			certPEM:   []byte("not a valid PEM"),
			wantError: true,
		},
		{
			name: "Valid PEM but invalid certificate",
			certPEM: []byte(`-----BEGIN CERTIFICATE-----
invalid certificate data
-----END CERTIFICATE-----`),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			analysis, err := AnalyzeCertificate(tt.certPEM)

			if tt.wantError {
				g.Expect(err).To(HaveOccurred())
				g.Expect(analysis).To(BeNil())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(analysis).NotTo(BeNil())
			}
		})
	}
}

func TestPQCSafeAlgorithms(t *testing.T) {
	tests := []struct {
		name        string
		keyType     string
		keySize     int
		wantPQCSafe bool
	}{
		{"RSA 1024 not PQC-safe", "rsa", 1024, false},
		{"RSA 2048 not PQC-safe", "rsa", 2048, false},
		{"RSA 3072 PQC-safe", "rsa", 3072, true},
		{"RSA 4096 PQC-safe", "rsa", 4096, true},
		{"ECDSA P-256 not PQC-safe", "ecdsa", 256, false},
		{"ECDSA P-384 PQC-safe", "ecdsa", 384, true},
		{"ECDSA P-521 PQC-safe", "ecdsa", 521, true},
		{"Ed25519 not PQC-safe", "ed25519", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			cfg := &helpers.CertConfig{
				KeyType:      tt.keyType,
				KeySize:      tt.keySize,
				CommonName:   "pqc-test.example.com",
				DNSNames:     []string{"pqc-test.example.com"},
				Organization: "Test Org",
				NotBefore:    ptr.To(time.Now()).Add(-1 * time.Hour),
				NotAfter:     ptr.To(time.Now()).Add(24 * time.Hour),
			}

			cert, err := helpers.GenerateCertificate(cfg)
			g.Expect(err).NotTo(HaveOccurred())

			analysis, err := AnalyzeCertificate(cert.CertPEM)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(analysis.IsPQCSafe).To(Equal(tt.wantPQCSafe),
				"%s with key size %d should have PQC-safe=%v", tt.keyType, tt.keySize, tt.wantPQCSafe)
		})
	}
}

func TestTLS13Compatibility(t *testing.T) {
	tests := []struct {
		name          string
		keyType       string
		keySize       int
		wantTLS13Compat bool
	}{
		{"RSA 1024 not TLS 1.3 compatible", "rsa", 1024, false},
		{"RSA 2048 TLS 1.3 compatible", "rsa", 2048, true},
		{"RSA 3072 TLS 1.3 compatible", "rsa", 3072, true},
		{"RSA 4096 TLS 1.3 compatible", "rsa", 4096, true},
		{"ECDSA P-256 TLS 1.3 compatible", "ecdsa", 256, true},
		{"ECDSA P-384 TLS 1.3 compatible", "ecdsa", 384, true},
		{"ECDSA P-521 TLS 1.3 compatible", "ecdsa", 521, true},
		{"Ed25519 TLS 1.3 compatible", "ed25519", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			cfg := &helpers.CertConfig{
				KeyType:      tt.keyType,
				KeySize:      tt.keySize,
				CommonName:   "tls13-test.example.com",
				DNSNames:     []string{"tls13-test.example.com"},
				Organization: "Test Org",
				NotBefore:    ptr.To(time.Now()).Add(-1 * time.Hour),
				NotAfter:     ptr.To(time.Now()).Add(24 * time.Hour),
			}

			cert, err := helpers.GenerateCertificate(cfg)
			g.Expect(err).NotTo(HaveOccurred())

			analysis, err := AnalyzeCertificate(cert.CertPEM)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(analysis.SupportsTLS13).To(Equal(tt.wantTLS13Compat),
				"%s with key size %d should have TLS 1.3 compatibility=%v",
				tt.keyType, tt.keySize, tt.wantTLS13Compat)
		})
	}
}

func TestCipherSuiteRecommendations(t *testing.T) {
	tests := []struct {
		name           string
		keyType        string
		keySize        int
		wantPQCSafe    bool
		expectStronger bool
	}{
		{
			name:           "RSA 2048 should get standard cipher suites",
			keyType:        "rsa",
			keySize:        2048,
			wantPQCSafe:    false,
			expectStronger: false,
		},
		{
			name:           "RSA 3072 should get stronger cipher suites",
			keyType:        "rsa",
			keySize:        3072,
			wantPQCSafe:    true,
			expectStronger: true,
		},
		{
			name:           "ECDSA P-384 should get stronger cipher suites",
			keyType:        "ecdsa",
			keySize:        384,
			wantPQCSafe:    true,
			expectStronger: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			cfg := &helpers.CertConfig{
				KeyType:      tt.keyType,
				KeySize:      tt.keySize,
				CommonName:   "cipher-test.example.com",
				DNSNames:     []string{"cipher-test.example.com"},
				Organization: "Test Org",
				NotBefore:    ptr.To(time.Now()).Add(-1 * time.Hour),
				NotAfter:     ptr.To(time.Now()).Add(24 * time.Hour),
			}

			cert, err := helpers.GenerateCertificate(cfg)
			g.Expect(err).NotTo(HaveOccurred())

			analysis, err := AnalyzeCertificate(cert.CertPEM)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(analysis.IsPQCSafe).To(Equal(tt.wantPQCSafe))
			g.Expect(len(analysis.CipherSuites)).To(BeNumerically(">", 0))

			if tt.expectStronger {
				// PQC-safe configs should prefer TLS_AES_256_GCM_SHA384
				g.Expect(analysis.CipherSuites[0]).To(Equal("TLS_AES_256_GCM_SHA384"))
			}
		})
	}
}
