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

package tls

import (
	"context"
	"fmt"
	"strings"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/secret"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
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
)

// Service contains server-specific TLS secret
type Service struct {
	// +kubebuilder:validation:Optional
	SecretName string `json:"secretName,omitempty"`
	// +kubebuilder:validation:Optional
	DisableNonTLSListeners bool `json:"disableNonTLSListeners,omitempty"`
}

// Ca contains CA-specific settings, which could be used both by services (to define their own CA certificates)
// and by clients (to verify the server's certificate)
type Ca struct {
	// +kubebuilder:validation:Optional
	CaSecretName string `json:"caSecretName,omitempty"`
}

// TLS - a generic type, which encapsulates both the service and CA configurations
type TLS struct {
	Service *Service `json:"service"`
	Ca      *Ca      `json:"ca"`
}

// NewTLS - initialize and return a TLS struct
func NewTLS(ctx context.Context, h *helper.Helper, namespace string, service *Service, ca *Ca) (*TLS, error) {

	// Ensure service SecretName exists or return an error
	if service != nil && service.SecretName != "" {
		secretData, _, err := secret.GetSecret(ctx, h, service.SecretName, namespace)
		if err != nil {
			return nil, fmt.Errorf("error ensuring secret %s exists: %w", service.SecretName, err)
		}

		_, keyOk := secretData.Data["tls.key"]
		_, certOk := secretData.Data["tls.crt"]
		if !keyOk || !certOk {
			return nil, fmt.Errorf("secret %s does not contain both tls.key and tls.crt", service.SecretName)
		}
	}

	return &TLS{
		Service: service,
		Ca:      ca,
	}, nil
}

// CreateVolumeMounts - add volume mount for TLS certificate and CA certificates, this counts on openstack-operator providing CA certs with unique names
func (t *TLS) CreateVolumeMounts() []corev1.VolumeMount {
	var volumeMounts []corev1.VolumeMount

	if t.Service != nil && t.Service.SecretName != "" {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "tls-crt",
			MountPath: "/etc/pki/tls/certs/tls.crt",
			SubPath:   "tls.crt",
			ReadOnly:  true,
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "tls-key",
			MountPath: "/etc/pki/tls/certs/tls.key",
			SubPath:   "tls.key",
			ReadOnly:  true,
		})
	}

	if t.Ca != nil && t.Ca.CaSecretName != "" {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "ca-certs",
			MountPath: "/etc/pki/ca-trust/extracted/pem",
			ReadOnly:  true,
		})
	}

	return volumeMounts
}

// CreateVolumes - add volume for TLS certificate and CA certificates
func (t *TLS) CreateVolumes() []corev1.Volume {
	var volumes []corev1.Volume

	if t.Service != nil && t.Service.SecretName != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "tls-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  t.Service.SecretName,
					DefaultMode: ptr.To[int32](0440),
				},
			},
		})
	}

	if t.Ca != nil && t.Ca.CaSecretName != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "ca-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  t.Ca.CaSecretName,
					DefaultMode: ptr.To[int32](0444),
				},
			},
		})
	}

	return volumes
}

// CreateDatabaseClientConfig - connection flags for the MySQL client
// Configures TLS connections for clients that use TLS certificates
// returns a string of mysql config statements
func (t *TLS) CreateDatabaseClientConfig() string {
	conn := []string{}
	// This assumes certificates are always injected in
	// a common directory for all services
	if t.Service.SecretName != "" {
		conn = append(conn,
			"ssl-cert=/etc/pki/tls/certs/tls.crt",
			"ssl-key=/etc/pki/tls/private/tls.key")
	}
	// Client uses a CA certificate that gets merged
	// into the pod's CA bundle by kolla_start
	if t.Ca.CaSecretName != "" {
		conn = append(conn,
			"ssl-ca=/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem")
	}
	if len(conn) > 0 {
		conn = append([]string{"ssl=1"}, conn...)
	}
	return strings.Join(conn, "\n")
}
