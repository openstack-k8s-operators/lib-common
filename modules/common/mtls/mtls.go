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

package mtls

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	// CertKey - key of the secret entry holding the cert
	CertKey = "tls.crt"
	// PrivateKey - key of the secret entry holding the cert private key
	PrivateKey = "tls.key"
	// CAKey - key of the secret entry holding the CA
	CAKey = "ca.crt"

	// CaPath - path to the ca certificate
	CaPath = "/etc/pki/tls/certs/mtls-ca.crt"
	// CertPath - path to the client certificate
	CertPath = "/etc/pki/tls/certs/mtls.crt"
	// KeyPath - path to the key
	KeyPath = "/etc/pki/tls/private/mtls.key"
)

// CreateMTLSVolumeMounts - add volume mount for MTLS certificates and CA certificate
func CreateMTLSVolumeMounts(SecretName string) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{}
	if SecretName != "" {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      SecretName,
			MountPath: CertPath,
			SubPath:   CertKey,
			ReadOnly:  true,
		}, corev1.VolumeMount{
			Name:      SecretName,
			MountPath: KeyPath,
			SubPath:   PrivateKey,
			ReadOnly:  true,
		}, corev1.VolumeMount{
			Name:      SecretName,
			MountPath: CaPath,
			SubPath:   CAKey,
			ReadOnly:  true,
		})
	}

	return volumeMounts
}

// CreateMTLSVolume - add volume for MTLS certificates and CA certificate for the service
func CreateMTLSVolume(SecretName string) corev1.Volume {
	volume := corev1.Volume{}
	if SecretName != "" {
		volume = corev1.Volume{
			Name: SecretName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: SecretName,
					//DefaultMode: ptr.To[int32](0400),
				},
			},
		}
	}

	return volume
}

// CaMountPath - returns path to the ca certificate
func CaMountPath() string {
	return CaPath
}

// CertMountPath - returns path to the certificate
func CertMountPath() string {
	return CertPath
}

// KeyMountPath - returns path to the key
func KeyMountPath() string {
	return KeyPath
}
