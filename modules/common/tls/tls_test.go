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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	. "github.com/onsi/gomega"
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"github.com/openstack-k8s-operators/lib-common/modules/storage"
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
		want    storage.Volume
	}{
		{
			name:    "No Secrets",
			service: &Service{},
			want:    storage.Volume{},
		},
		{
			name:    "Only TLS Secret",
			service: &Service{SecretName: "cert-secret"},
			id:      "foo",
			want: storage.Volume{
				Name: "foo-tls-certs",
				VolumeSource: storage.VolumeSource{
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
			want: storage.Volume{
				Name: "default-tls-certs",
				VolumeSource: storage.VolumeSource{
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
		want storage.Volume
	}{
		{
			name: "Empty Ca",
			ca:   &Ca{},
			want: storage.Volume{},
		},
		{
			name: "Set CaBundleSecretName",
			ca: &Ca{
				CaBundleSecretName: "ca-secret",
			},
			want: storage.Volume{
				Name: "combined-ca-bundle",
				VolumeSource: storage.VolumeSource{
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
