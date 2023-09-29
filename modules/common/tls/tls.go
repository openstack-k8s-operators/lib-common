/*
Copyright 2022 Red Hat

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

package tls

// Service contains server-specific TLS secret
type Service struct {
	// Server-specific settings
	SecretName             string `json:"secretName,omitempty"`
	DisableNonTLSListeners bool   `json:"disableNonTLSListeners,omitempty"`
}

// Ca contains CA-specific settings, which could be used both by services (to define their own CA certificates)
// and by clients (to verify the server's certificate)
type Ca struct {
	// CA-specific settings
	CaSecretName string `json:"caSecretName,omitempty"`
}

// TLS - a generic type, which encapsulates both the service and CA configurations
type TLS struct {
	Service *Service `json:"service"`
	Ca      *Ca      `json:"ca"`
}
