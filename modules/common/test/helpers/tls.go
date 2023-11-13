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

package helpers

// ServiceConfig contains server-specific TLS secret
type ServiceConfig struct {
	SecretName             string
	CertMount              *string
	KeyMount               *string
	CaMount                *string
	DisableNonTLSListeners bool
}

// CaConfig contains CA-specific settings
type CaConfig struct {
	CaBundleSecretName string
	CaBundleMount      *string
}

// InitializeServiceConfig creates a custom CaConfig structure for testing with provided data
//
// Example usage:
//
// serviceConfig := th.InitializeServiceConfig("test-tls-secret", "/etc/pki/tls/certs/tls.crt", "/etc/pki/tls/private/tls.key", "/etc/pki/tls/ca.crt", true)
func (tc *TestHelper) InitializeServiceConfig(secretName, certMount, keyMount, caMount string, disableNonTLSListeners bool) *ServiceConfig {
	serviceConfig := &ServiceConfig{
		SecretName:             secretName,
		CertMount:              &certMount,
		KeyMount:               &keyMount,
		CaMount:                &caMount,
		DisableNonTLSListeners: disableNonTLSListeners,
	}

	return serviceConfig
}

// InitializeCaConfig creates a custom CaConfig structure for testing with provided data
//
// Example usage:
//
// caConfig := th.InitializeCaConfig("test-ca-secret", "/etc/pki/ca-trust/extracted/pem/ca-bundle.pem")
func (tc *TestHelper) InitializeCaConfig(caBundleSecretName, caBundleMount string) *CaConfig {
	caConfig := &CaConfig{
		CaBundleSecretName: caBundleSecretName,
		CaBundleMount:      &caBundleMount,
	}

	return caConfig
}
