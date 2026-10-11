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

// +kubebuilder:object:generate:=true

// Package tls provides utilities for managing TLS certificates and configurations
package tls

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/openstack-k8s-operators/lib-common/modules/common/env"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/secret"
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// CABundleSecret -
	CABundleSecret = "combined-ca-bundle"
	// CABundleLabel added to the CA bundle secret for the namespace
	CABundleLabel = "combined-ca-bundle"
	// CABundleKey - key in CaBundleSecret holding a full CA bundle
	CABundleKey = "tls-ca-bundle.pem"
	// InternalCABundleKey - key in CABundleSecret only holding the internal CA
	InternalCABundleKey = "internal-ca-bundle.pem"

	// DefaultCAPrefix -
	DefaultCAPrefix = "rootca-"
	// DownstreamTLSCABundlePath -
	DownstreamTLSCABundlePath = "/etc/pki/ca-trust/extracted/pem/" + CABundleKey
	// UpstreamTLSCABundlePath -
	UpstreamTLSCABundlePath = "/etc/ssl/certs/ca-certificates.crt"

	// CertKey - key of the secret entry holding the cert
	CertKey = "tls.crt"
	// PrivateKey - key of the secret entry holding the cert private key
	PrivateKey = "tls.key"
	// CAKey - key of the secret entry holding the CA
	CAKey = "ca.crt"
	// DefaultCertMountDir - updated default path to mount cert files inside container
	DefaultCertMountDir = "/var/lib/config-data/tls/certs"
	// DefaultKeyMountDir - updated default path to mount cert keys inside container
	DefaultKeyMountDir = "/var/lib/config-data/tls/private"

	// TLSHashName - Name of the hash of hashes of all cert resources used to identify a change
	TLSHashName = "certs"

	// AdditionalSubjectNamesKey - Comma separated list of additionalSubjectNames
	// that should be passed to the CertificateRequest
	AdditionalSubjectNamesKey = "additionalSubjectNames"

	// DefaultClusterInternalDomain - cluster internal dns domain
	DefaultClusterInternalDomain = "cluster.local"
)

// SimpleService defines the observed state of TLS for a single service
type SimpleService struct {
	// +kubebuilder:validation:optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// Secret containing certificates for the service
	GenericService `json:",inline"`

	// +kubebuilder:validation:optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// Secret containing CA bundle
	Ca `json:",inline"`
}

// API defines the observed state of TLS with API only
type API struct {
	// +kubebuilder:validation:optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// API tls type which encapsulates for API services
	API APIService `json:"api,omitempty"`

	// +kubebuilder:validation:optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// Secret containing CA bundle
	Ca `json:",inline"`
}

// APIService - API tls type which encapsulates for API services
type APIService struct {
	// +kubebuilder:validation:optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// Public GenericService - holds the secret for the public endpoint
	Public GenericService `json:"public,omitempty"`

	// +kubebuilder:validation:optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// Internal GenericService - holds the secret for the internal endpoint
	Internal GenericService `json:"internal,omitempty"`
}

// GenericService contains server-specific TLS secret or issuer
type GenericService struct {
	// +kubebuilder:validation:Optional
	// SecretName - holding the cert, key for the service
	SecretName *string `json:"secretName,omitempty"`
}

// Ca contains CA-specific settings, which could be used both by services (to define their own CA certificates)
// and by clients (to verify the server's certificate)
type Ca struct {
	// CaBundleSecretName - holding the CA certs in a pre-created bundle file
	CaBundleSecretName string `json:"caBundleSecretName,omitempty"`
}

// Service contains server-specific TLS secret
// +kubebuilder:object:generate:=false
type Service struct {
	// SecretName - holding the cert, key for the service
	SecretName string `json:"secretName"`

	// CertMount - dst location to mount the service tls.crt cert. Can be used to override the default location which is /etc/tls/certs/<service id>.crt
	CertMount *string `json:"certMount,omitempty"`

	// KeyMount - dst location to mount the service tls.key  key. Can be used to override the default location which is /etc/tls/private/<service id>.key
	KeyMount *string `json:"keyMount,omitempty"`

	// CaMount - dst location to mount the CA cert ca.crt to. Can be used if the service CA cert should be mounted specifically, e.g. to be set in a service config for validation, instead of the env wide bundle.
	CaMount *string `json:"caMount,omitempty"`
}

// Enabled - returns true if TLS is configured for the service
func (s *GenericService) Enabled() bool {
	return s.SecretName != nil && *s.SecretName != ""
}

// Enabled - returns true if TLS is configured for the public and internal
func (a *APIService) Enabled(endpt service.Endpoint) bool {
	switch endpt {
	case service.EndpointPublic:
		return a.Public.Enabled()
	case service.EndpointInternal:
		return a.Internal.Enabled()
	}

	return false
}

// ValidateCertSecrets - validates the content of the cert secrets to make sure "tls-ca-bundle.pem" key exists
func (a *APIService) ValidateCertSecrets(
	ctx context.Context,
	h *helper.Helper,
	namespace string,
) (string, error) {
	var svc GenericService
	certHashes := map[string]env.Setter{}
	for _, endpt := range []service.Endpoint{service.EndpointInternal, service.EndpointPublic} {
		switch endpt {
		case service.EndpointPublic:
			if !a.Enabled(service.EndpointPublic) {
				continue
			}

			svc = a.Public

		case service.EndpointInternal:
			if !a.Enabled(service.EndpointInternal) {
				continue
			}

			svc = a.Internal
		}

		hash, err := svc.ValidateCertSecret(ctx, h, namespace)
		if err != nil {
			return "", err
		}
		certHashes["cert-"+endpt.String()] = env.SetValue(hash)
	}

	certsHash, err := util.HashOfInputHashes(certHashes)
	if err != nil {
		return "", err
	}
	return certsHash, nil
}

// ToService - convert tls.APIService to tls.Service
func (s *GenericService) ToService() (*Service, error) {
	toS := &Service{}

	sBytes, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("error marshalling api service: %w", err)
	}

	err = json.Unmarshal(sBytes, toS)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling tls service: %w", err)
	}

	return toS, nil
}

// ValidateCertSecret - validates the content of the cert secrets to make sure "tls-ca-bundle.pem" key exists
func (s *GenericService) ValidateCertSecret(
	ctx context.Context,
	h *helper.Helper,
	namespace string,
) (string, error) {
	hash := ""

	endptTLSCfg, err := s.ToService()
	if err != nil {
		return "", err
	}

	if endptTLSCfg.SecretName != "" {
		// validate the cert secret has the expected keys
		hash, err = endptTLSCfg.ValidateCertSecret(ctx, h, namespace)
		if err != nil {
			return "", err
		}
	}

	return hash, nil
}

// ValidateCACertSecret - validates the content of the cert secret to make sure "tls-ca-bundle.pem" key exists
func ValidateCACertSecret(
	ctx context.Context,
	c client.Client,
	caSecret types.NamespacedName,
) (string, error) {
	hash, ctrlResult, err := secret.VerifySecret(
		ctx,
		caSecret,
		[]string{CABundleKey},
		c,
		5*time.Second)
	if err != nil {
		return "", err
	} else if (ctrlResult != ctrl.Result{}) {
		return "", k8s_errors.NewNotFound(
			appsv1.Resource("Secret"),
			fmt.Sprintf("%s in namespace %s", caSecret.Name, caSecret.Namespace),
		)
	}

	return hash, nil
}

// ValidateCertSecret - validates the content of the cert secret to make sure "tls.key", "tls.crt" and optional "ca.crt" keys exist
func (s *Service) ValidateCertSecret(ctx context.Context, h *helper.Helper, namespace string) (string, error) {
	// define keys to expect in cert secret
	keys := []string{PrivateKey, CertKey}
	if s.CaMount != nil {
		keys = append(keys, CAKey)
	}

	hash, ctrlResult, err := secret.VerifySecret(
		ctx,
		types.NamespacedName{Name: s.SecretName, Namespace: namespace},
		keys,
		h.GetClient(),
		5*time.Second)
	if err != nil {
		return "", err
	} else if (ctrlResult != ctrl.Result{}) {
		return "", k8s_errors.NewNotFound(
			corev1.Resource(corev1.ResourceSecrets.String()),
			fmt.Sprintf("%s in namespace %s", s.SecretName, namespace),
		)
	}

	return hash, nil
}

// ValidateEndpointCerts - validates all services from an endpointCfgs and
// returns the hash of hashes for all the certificates
func ValidateEndpointCerts(
	ctx context.Context,
	h *helper.Helper,
	namespace string,
	endpointCfgs map[service.Endpoint]Service,
) (string, error) {
	certHashes := map[string]env.Setter{}
	for endpt, endpointTLSCfg := range endpointCfgs {
		if endpointTLSCfg.SecretName != "" {
			// validate the cert secret has the expected keys
			hash, err := endpointTLSCfg.ValidateCertSecret(ctx, h, namespace)
			if err != nil {
				return "", err
			}

			certHashes["cert-"+endpt.String()] = env.SetValue(hash)
		}
	}

	certsHash, err := util.HashOfInputHashes(certHashes)
	if err != nil {
		return "", err
	}
	return certsHash, nil
}

// getCertMountPath - return certificate mount path
func (s *Service) getCertMountPath(serviceID string) string {
	if serviceID == "" {
		serviceID = "default"
	}

	certMountPath := fmt.Sprintf("%s/%s.crt", DefaultCertMountDir, serviceID)
	if s.CertMount != nil {
		certMountPath = *s.CertMount
	}

	return certMountPath
}

// getKeyMountPath - return key mount path
func (s *Service) getKeyMountPath(serviceID string) string {
	if serviceID == "" {
		serviceID = "default"
	}

	keyMountPath := fmt.Sprintf("%s/%s.key", DefaultKeyMountDir, serviceID)
	if s.KeyMount != nil {
		keyMountPath = *s.KeyMount
	}

	return keyMountPath
}

// CreateVolumeMounts - add volume mount for TLS certificates and CA certificate for the service
func (s *Service) CreateVolumeMounts(serviceID string) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{}
	if serviceID == "" {
		serviceID = "default"
	}
	if s.SecretName != "" {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      serviceID + "-tls-certs",
			MountPath: s.getCertMountPath(serviceID),
			SubPath:   CertKey,
			ReadOnly:  true,
		}, corev1.VolumeMount{
			Name:      serviceID + "-tls-certs",
			MountPath: s.getKeyMountPath(serviceID),
			SubPath:   PrivateKey,
			ReadOnly:  true,
		})

		if s.CaMount != nil {
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      serviceID + "-tls-certs",
				MountPath: *s.CaMount,
				SubPath:   CAKey,
				ReadOnly:  true,
			})
		}
	}

	return volumeMounts
}

// CreateVolume - add volume for TLS certificates and CA certificate for the service
func (s *Service) CreateVolume(serviceID string) corev1.Volume {
	volume := corev1.Volume{}
	if serviceID == "" {
		serviceID = "default"
	}
	if s.SecretName != "" {
		volume = corev1.Volume{
			Name: serviceID + "-tls-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  s.SecretName,
					DefaultMode: ptr.To[int32](0400),
				},
			},
		}
	}

	return volume
}

// CreateVolumeMounts creates volume mounts for CA bundle file
func (c *Ca) CreateVolumeMounts(caBundleMount *string) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{}

	if caBundleMount == nil {
		caBundleMount = ptr.To(DownstreamTLSCABundlePath)
	}

	if c.CaBundleSecretName != "" {
		volumeMounts = []corev1.VolumeMount{
			{
				Name:      CABundleLabel,
				MountPath: *caBundleMount,
				SubPath:   CABundleKey,
				ReadOnly:  true,
			},
		}
	}

	return volumeMounts
}

// CreateVolume creates volumes for CA bundle file
func (c *Ca) CreateVolume() corev1.Volume {
	volume := corev1.Volume{}

	if c.CaBundleSecretName != "" {
		volume = corev1.Volume{
			Name: CABundleLabel,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  c.CaBundleSecretName,
					DefaultMode: ptr.To[int32](0444),
				},
			},
		}
	}

	return volume
}

type TLSAnalysis struct {
	// SupportsTLS13 indicates if TLS 1.3 is supported
	SupportsTLS13 bool
	// MinTLSVersion is the minimum TLS version supported
	MinTLSVersion uint16
	// MaxTLSVersion is the maximum TLS version supported
	MaxTLSVersion uint16
	// IsPQCSafe indicates whether the certificate uses post-quantum safe algorithms
	IsPQCSafe bool
	// SignatureAlgorithm is the signature algorithm used by the certificate
	SignatureAlgorithm string
	// KeyAlgorithm is the public key algorithm used
	KeyAlgorithm string
	// KeySize is the size of the public key in bits
	KeySize int
	// CipherSuites lists the cipher suites that would be used
	CipherSuites []string
}

var pqcSafeAlgorithms = map[x509.SignatureAlgorithm]bool{
	x509.SHA256WithRSA:    false, // Depends on key size
	x509.SHA384WithRSA:    false, // Depends on key size
	x509.SHA512WithRSA:    false, // Depends on key size
	x509.ECDSAWithSHA256:  false, // Depends on key size
	x509.ECDSAWithSHA384:  true,  // P-384 is transitionally safe
	x509.ECDSAWithSHA512:  true,  // P521 is transitionally safe
	x509.SHA256WithRSAPSS: false, // Depends on key size
	x509.SHA384WithRSAPSS: false, // Depends on key size
	x509.SHA512WithRSAPSS: false, // Depends on key size
	x509.PureEd25519:      false, // Ed25519 is not PQC-safe
}

// Minimum key sizes required for transitional PQC safety (see NIST SP 800-57)
const (
	minPQCSafeRSAKeySize   = 3072
	minPQCSafeECDSAKeySize = 384 // P-384 curve
	minTLS13RSAKeySize     = 2048
	minTLS13ECDSAKeySize   = 256
)

// AnalyzeCertificate analyzes a certificate for TLS 1.3 enablement and PQC-safe algorithm usage
func AnalyzeCertificate(certPEM []byte) (*TLSAnalysis, error) {
	// Decode the PEM block
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	analysis := &TLSAnalysis{
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
	}

	// Determine if we're PQC-safe based on algorithm and key size
	analysis.IsPQCSafe = isPQCSafe(cert)

	// Determine algorithm and key size
	switch pubKey := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		analysis.KeyAlgorithm = "RSA"
		analysis.KeySize = pubKey.N.BitLen()
	case *ecdsa.PublicKey:
		analysis.KeyAlgorithm = "ECDSA"
		analysis.KeySize = pubKey.Curve.Params().BitSize
	case ed25519.PublicKey:
		analysis.KeyAlgorithm = "Ed25519"
		analysis.KeySize = ed25519.PublicKeySize * 8 // PublicKeySize is in bytes
	default:
		analysis.KeyAlgorithm = "Unknown"
	}

	// Check TLS 1.3 support
	// The Certificate itself doesn't dictate TLS version, but we can see if it's compatible with
	// TLS 1.3 requirements.
	analysis.SupportsTLS13 = isTLS13Compatible(cert)

	// Set version info (these typically come from server config, not the certificate)
	analysis.MinTLSVersion = tls.VersionTLS12
	analysis.MaxTLSVersion = tls.VersionTLS13

	// Get the recommended cipher suites
	analysis.CipherSuites = getRecommendedCipherSuites(analysis.IsPQCSafe)

	return analysis, nil
}

// isPQCSafe determines if a certificate is PQC-safe through using a PQC-safe algorithm or a
// large enough key size
func isPQCSafe(cert *x509.Certificate) bool {
	// Check the signature algorithm
	baseSafe, exists := pqcSafeAlgorithms[cert.SignatureAlgorithm]

	// For algorithms where PQC-safety depends on key length
	if exists && !baseSafe {
		switch pubKey := cert.PublicKey.(type) {
		case *rsa.PublicKey:
			// RSA keys need to be >= 3072 bits for PQC transitional safety
			return pubKey.N.BitLen() >= minPQCSafeRSAKeySize
		case *ecdsa.PublicKey:
			// ECDSA needs a P-384 or P-521 curve
			return pubKey.Curve.Params().BitSize >= minPQCSafeECDSAKeySize
		}
	}

	return baseSafe
}

// isTLS13Compatible checks if a certificate is compatible with TLS 1.3
func isTLS13Compatible(cert *x509.Certificate) bool {
	// TLS 1.3 removed support for RSA-PSS and requires specific signature algorithms.
	// Generally, certificates with RSA >= 2048, ECDSA with curves P-256+,
	// or Ed25519 are compatible.
	// NOTE: TLS 1.3 compatibility does NOT necessarily mean that it's PQC-safe!
	switch pubKey := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		return pubKey.N.BitLen() >= minTLS13RSAKeySize
	case *ecdsa.PublicKey:
		return pubKey.Curve.Params().BitSize >= minTLS13ECDSAKeySize
	case ed25519.PublicKey:
		return true
	default:
		return false
	}
}

// getRecommendedCipherSuites returns the recommended cipher suites depending on whether we want
// PQC ciphers or not.
func getRecommendedCipherSuites(pqcSafe bool) []string {
	// TLS 1.3 cipher suites (these are always used for TLS 1.3)
	tls13Suites := []string{
		"TLS_AES_128_GCM_SHA256",
		"TLS_AES_256_GCM_SHA384",
		"TLS_CHACHA20_POLY1305_SHA256",
	}

	if pqcSafe {
		// For PQC-safe configs, prefer stronger ciphers
		return append([]string{
			"TLS_AES_256_GCM_SHA384",
		}, tls13Suites...)
	}

	return tls13Suites
}

// AnalyzeCertSecret analyzes a certificate stored in a Kubernetes secret
func AnalyzeCertSecret(
	ctx context.Context,
	c client.Client,
	secretName types.NamespacedName,
) (*TLSAnalysis, error) {
	// Get the secret
	certSecret := &corev1.Secret{}
	err := c.Get(ctx, secretName, certSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate secret: %w", err)
	}

	// Get the certificate's data
	certData, exists := certSecret.Data[CertKey]
	if !exists {
		return nil, fmt.Errorf("certificate data not found in secret")
	}

	// Analyze the certificate
	return AnalyzeCertificate(certData)
}

// IsTLS13Enabled checks if TLS 1.3 is enabled for a service
func (s *Service) IsTLS13Enabled(
	ctx context.Context,
	h *helper.Helper,
	namespace string,
) (bool, error) {
	if s.SecretName == "" {
		return false, fmt.Errorf("no certificate configured")
	}

	analysis, err := AnalyzeCertSecret(
		ctx,
		h.GetClient(),
		types.NamespacedName{Name: s.SecretName, Namespace: namespace},
	)
	if err != nil {
		return false, err
	}

	return analysis.SupportsTLS13, nil
}

// IsPQCSafe checks if the certificate uses PQC-safe algorithms/key lengths
func (s *Service) IsPQCSafe(ctx context.Context, h *helper.Helper, namespace string) (bool, error) {
	if s.SecretName == "" {
		return false, fmt.Errorf("no certificate configured")
	}

	analysis, err := AnalyzeCertSecret(
		ctx,
		h.GetClient(),
		types.NamespacedName{Name: s.SecretName, Namespace: namespace},
	)
	if err != nil {
		return false, err
	}

	return analysis.IsPQCSafe, nil
}

// GetTLSAnalysis returns a comprehensive TLS analysis for a service
func (s *Service) GetTLSAnalysis(
	ctx context.Context,
	h *helper.Helper,
	namespace string,
) (*TLSAnalysis, error) {
	if s.SecretName == "" {
		return nil, fmt.Errorf("no certificate configured")
	}

	return AnalyzeCertSecret(
		ctx,
		h.GetClient(),
		types.NamespacedName{Name: s.SecretName, Namespace: namespace},
	)
}
