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

package service

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

var (
	svcClusterIP = corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "namespace",
			Labels: map[string]string{
				"foo": "bar",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{},
			Type:  corev1.ServiceTypeClusterIP,
		},
	}
	svcLoadBalancer = corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "namespace",
			Labels: map[string]string{
				"foo": "bar",
			}},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{},
			Type:  corev1.ServiceTypeLoadBalancer,
		},
	}
	portHTTP = []corev1.ServicePort{
		{
			Name:        "foo",
			Protocol:    corev1.ProtocolTCP,
			AppProtocol: nil,
			Port:        int32(80),
			TargetPort:  intstr.FromInt(0),
			NodePort:    0,
		},
	}
	portHTTPS = []corev1.ServicePort{
		{
			Name:        "foo",
			Protocol:    corev1.ProtocolTCP,
			AppProtocol: nil,
			Port:        int32(443),
			TargetPort:  intstr.FromInt(0),
			NodePort:    0,
		},
	}
	portCustom = []corev1.ServicePort{
		{
			Name:        "foo",
			Protocol:    corev1.ProtocolTCP,
			AppProtocol: nil,
			Port:        int32(8080),
			TargetPort:  intstr.FromInt(0),
			NodePort:    0,
		},
	}
	timeout  = time.Duration(5) * time.Second
	override = OverrideSpec{
		Spec: &OverrideServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}
	overrideServiceSpecClusterIP = OverrideServiceSpec{
		Type: corev1.ServiceTypeClusterIP,
	}
	overrideServiceSpecLoadBalancer = OverrideServiceSpec{
		Type: corev1.ServiceTypeLoadBalancer,
	}
)

func TestGenericService(t *testing.T) {
	tests := []struct {
		name    string
		service GenericServiceDetails
		want    corev1.Service
	}{
		{
			name: "Service with port, no labels, selector",
			service: GenericServiceDetails{
				Name:      "foo",
				Namespace: "namespace",
				Labels:    map[string]string{},
				Selector:  map[string]string{},
				Ports: []corev1.ServicePort{
					{
						Name:     "port",
						Port:     int32(80),
						Protocol: corev1.ProtocolTCP,
					},
				},
			},
			want: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "namespace",
					Labels:    map[string]string{},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name:        "port",
							Protocol:    corev1.ProtocolTCP,
							AppProtocol: nil,
							Port:        int32(80),
							TargetPort:  intstr.FromInt(0),
							NodePort:    0,
						},
					},
					Selector:                 map[string]string{},
					Type:                     corev1.ServiceTypeClusterIP,
					PublishNotReadyAddresses: false,
				},
			},
		},
		{
			name: "Service with port, labels, selector",
			service: GenericServiceDetails{
				Name:      "foo",
				Namespace: "namespace",
				Labels: map[string]string{
					"foo": "bar",
				},
				Selector: map[string]string{
					"foo": "bar",
				},
				Ports: []corev1.ServicePort{
					{
						Name:     "port",
						Port:     int32(80),
						Protocol: corev1.ProtocolTCP,
					},
				},
			},
			want: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "namespace",
					Labels: map[string]string{
						"foo": "bar",
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name:        "port",
							Protocol:    corev1.ProtocolTCP,
							AppProtocol: nil,
							Port:        int32(80),
							TargetPort:  intstr.FromInt(0),
							NodePort:    0,
						},
					},
					Selector: map[string]string{
						"foo": "bar",
					},
					Type:                     corev1.ServiceTypeClusterIP,
					PublishNotReadyAddresses: false,
				},
			},
		},
		{
			name: "Headless service with not ready addresses",
			service: GenericServiceDetails{
				Name:      "foo",
				Namespace: "namespace",
				Labels: map[string]string{
					"foo": "bar",
				},
				Selector: map[string]string{
					"foo": "bar",
				},
				Ports: []corev1.ServicePort{
					{
						Name:     "port",
						Port:     int32(80),
						Protocol: corev1.ProtocolTCP,
					},
				},
				ClusterIP:                "None",
				PublishNotReadyAddresses: true,
			},
			want: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "namespace",
					Labels: map[string]string{
						"foo": "bar",
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name:        "port",
							Protocol:    corev1.ProtocolTCP,
							AppProtocol: nil,
							Port:        int32(80),
							TargetPort:  intstr.FromInt(0),
							NodePort:    0,
						},
					},
					Selector: map[string]string{
						"foo": "bar",
					},
					Type:                     corev1.ServiceTypeClusterIP,
					ClusterIP:                "None",
					PublishNotReadyAddresses: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			service := GenericService(&tt.service)

			g.Expect(*service).To(Equal(tt.want))
		})
	}
}

func getServiceWithPort(svc corev1.Service, ports []corev1.ServicePort) *corev1.Service {
	svc.Spec.Ports = ports

	return &svc
}

func TestNewService(t *testing.T) {
	tests := []struct {
		name                    string
		service                 *corev1.Service
		override                OverrideSpec
		want                    Service
		wantPort                string
		wantOverrideServiceSpec OverrideServiceSpec
	}{
		{
			name:     "HTTP ClusterIP service no override",
			service:  getServiceWithPort(svcClusterIP, portHTTP),
			override: OverrideSpec{},
			want: Service{
				service:         getServiceWithPort(svcClusterIP, portHTTP),
				timeout:         timeout,
				serviceHostname: "foo.namespace.svc",
			},
			wantPort:                "80",
			wantOverrideServiceSpec: overrideServiceSpecClusterIP,
		},
		{
			name:     "HTTPS ClusterIP service no override",
			service:  getServiceWithPort(svcClusterIP, portHTTPS),
			override: OverrideSpec{},
			want: Service{
				service:         getServiceWithPort(svcClusterIP, portHTTPS),
				timeout:         timeout,
				serviceHostname: "foo.namespace.svc",
			},
			wantPort:                "443",
			wantOverrideServiceSpec: overrideServiceSpecClusterIP,
		},
		{
			name:     "None ClusterIP service no override",
			service:  getServiceWithPort(svcClusterIP, portCustom),
			override: OverrideSpec{},
			want: Service{
				service:         getServiceWithPort(svcClusterIP, portCustom),
				timeout:         timeout,
				serviceHostname: "foo.namespace.svc",
			},
			wantPort:                "8080",
			wantOverrideServiceSpec: overrideServiceSpecClusterIP,
		},
		{
			name:     "HTTP ClusterIP service override service Type to LoadBalancer",
			service:  getServiceWithPort(svcClusterIP, portHTTP),
			override: override,
			want: Service{
				service:         getServiceWithPort(svcLoadBalancer, portHTTP),
				timeout:         timeout,
				serviceHostname: "foo.namespace.svc",
			},
			wantPort:                "80",
			wantOverrideServiceSpec: overrideServiceSpecLoadBalancer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			service, err := NewService(tt.service, timeout, &tt.override)
			g.Expect(err).ToNot(HaveOccurred())
			// timeout
			g.Expect(service.timeout).To(Equal(timeout))
			// GetServiceType
			g.Expect(service.GetServiceType()).To(Equal(tt.want.service.Spec.Type))
			// GetLabels
			g.Expect(service.GetLabels()).To(Equal(map[string]string{
				"foo": "bar",
			}))
			// AddAnnotation
			service.AddAnnotation(map[string]string{"foo": "bar"})
			// GetAnnotations
			g.Expect(service.GetAnnotations()).To(Equal(map[string]string{"foo": "bar"}))
			// GetServiceHostname
			g.Expect(service.GetServiceHostname()).To(Equal(tt.want.serviceHostname))
			// GetServiceHostnamePort
			hostname, port := service.GetServiceHostnamePort()
			g.Expect(hostname).To(Equal(tt.want.serviceHostname))
			g.Expect(port).To(Equal(tt.wantPort))
			// GetSpec
			g.Expect(*service.GetSpec()).To(Equal(tt.want.service.Spec))
			// ToOverrideServiceSpec
			dd, err := service.ToOverrideServiceSpec()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(dd).ToNot(BeNil())
			g.Expect(*dd).To(Equal(tt.wantOverrideServiceSpec))
		})
	}
}

func TestGetAPIEndpoint(t *testing.T) {
	tests := []struct {
		name        string
		service     *corev1.Service
		endpointURL *string
		proto       Protocol
		port        string
		path        string
		want        string
	}{
		{
			name:        "HTTP ClusterIP service default port 80, no override",
			service:     getServiceWithPort(svcClusterIP, portHTTP),
			endpointURL: nil,
			proto:       ProtocolHTTP,
			path:        "",
			want:        "http://foo.namespace.svc",
		},
		{
			name:        "HTTP ClusterIP service non default 8080 port, no override",
			service:     getServiceWithPort(svcClusterIP, portCustom),
			endpointURL: nil,
			proto:       ProtocolHTTP,
			path:        "/path",
			want:        "http://foo.namespace.svc:8080/path",
		},
		{
			name:        "HTTPS ClusterIP service default 443 port, no override",
			service:     getServiceWithPort(svcClusterIP, portHTTPS),
			endpointURL: nil,
			proto:       ProtocolHTTPS,
			path:        "/path",
			want:        "https://foo.namespace.svc/path",
		},
		{
			name:        "HTTPS ClusterIP service non default 8080 port, no override",
			service:     getServiceWithPort(svcClusterIP, portCustom),
			endpointURL: nil,
			proto:       ProtocolHTTPS,
			path:        "/path",
			want:        "https://foo.namespace.svc:8080/path",
		},
		{
			name:        "None ClusterIP service port 80 no override",
			service:     getServiceWithPort(svcClusterIP, portHTTP),
			endpointURL: nil,
			proto:       ProtocolNone,
			path:        "/path",
			want:        "foo.namespace.svc:80/path",
		},
		{
			name:        "None ClusterIP service port 8080 override",
			service:     getServiceWithPort(svcClusterIP, portCustom),
			endpointURL: nil,
			proto:       ProtocolNone,
			path:        "/path",
			want:        "foo.namespace.svc:8080/path",
		},
		{
			name:        "Override EndpointURL with path",
			service:     getServiceWithPort(svcClusterIP, portCustom),
			endpointURL: ptr.To("http://this.url"),
			proto:       ProtocolNone,
			path:        "/path",
			want:        "http://this.url/path",
		},
		{
			name:        "Override EndpointURL no path",
			service:     getServiceWithPort(svcClusterIP, portCustom),
			endpointURL: ptr.To("http://this.url"),
			proto:       ProtocolNone,
			path:        "",
			want:        "http://this.url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			service, err := NewService(tt.service, timeout, nil)
			g.Expect(err).ToNot(HaveOccurred())
			url, err := service.GetAPIEndpoint(tt.endpointURL, ptr.To(tt.proto), tt.path)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(url).To(Equal(tt.want))
		})
	}
}

func TestToOverrideServiceSpec(t *testing.T) {
	tests := []struct {
		name     string
		service  *corev1.Service
		override OverrideSpec
		want     OverrideServiceSpec
	}{
		{
			name:     "No override",
			service:  getServiceWithPort(svcClusterIP, portHTTP),
			override: OverrideSpec{},
			want:     overrideServiceSpecClusterIP,
		},
		{
			name:     "HTTP ClusterIP service override service Type to LoadBalancer",
			service:  getServiceWithPort(svcClusterIP, portHTTP),
			override: override,
			want:     overrideServiceSpecLoadBalancer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			service, err := NewService(tt.service, timeout, &tt.override)
			g.Expect(err).ToNot(HaveOccurred())
			// ToOverrideServiceSpec
			ovrrdServiceSpec, err := service.ToOverrideServiceSpec()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(ovrrdServiceSpec).ToNot(BeNil())
			g.Expect(*ovrrdServiceSpec).To(Equal(tt.want))
		})
	}
}

func TestOverrideSpecAddAnnotation(t *testing.T) {
	tests := []struct {
		name       string
		override   OverrideSpec
		annotation map[string]string
		want       OverrideSpec
	}{
		{
			name:       "No override, no custom annotation",
			override:   OverrideSpec{},
			annotation: map[string]string{},
			want: OverrideSpec{
				EmbeddedLabelsAnnotations: &EmbeddedLabelsAnnotations{
					Annotations: map[string]string{},
				},
			},
		},
		{
			name: "override, no custom annotation",
			override: OverrideSpec{EmbeddedLabelsAnnotations: &EmbeddedLabelsAnnotations{
				Annotations: map[string]string{"key": "val"},
			}},
			annotation: map[string]string{},
			want: OverrideSpec{
				EmbeddedLabelsAnnotations: &EmbeddedLabelsAnnotations{
					Annotations: map[string]string{"key": "val"},
				},
			},
		},
		{
			name: "override, additional custom annotation",
			override: OverrideSpec{EmbeddedLabelsAnnotations: &EmbeddedLabelsAnnotations{
				Annotations: map[string]string{"key": "val"},
			}},
			annotation: map[string]string{"custom": "val"},
			want: OverrideSpec{
				EmbeddedLabelsAnnotations: &EmbeddedLabelsAnnotations{
					Annotations: map[string]string{
						"key":    "val",
						"custom": "val",
					},
				},
			},
		},
		{
			name: "override, additional custom same annotation",
			override: OverrideSpec{EmbeddedLabelsAnnotations: &EmbeddedLabelsAnnotations{
				Annotations: map[string]string{"key": "val"},
			}},
			annotation: map[string]string{"key": "custom"},
			want: OverrideSpec{
				EmbeddedLabelsAnnotations: &EmbeddedLabelsAnnotations{
					Annotations: map[string]string{
						"key": "val",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			tt.override.AddAnnotation(tt.annotation)
			g.Expect(tt.override.Annotations).To(Equal(tt.want.Annotations))
		})
	}
}

func TestRoutedOverrideSpecAddAnnotation(t *testing.T) {
	tests := []struct {
		name       string
		override   RoutedOverrideSpec
		annotation map[string]string
		want       RoutedOverrideSpec
	}{
		{
			name:       "No override, no custom annotation",
			override:   RoutedOverrideSpec{},
			annotation: map[string]string{},
			want: RoutedOverrideSpec{
				OverrideSpec: OverrideSpec{
					EmbeddedLabelsAnnotations: &EmbeddedLabelsAnnotations{
						Annotations: map[string]string{},
					},
				},
			},
		},
		{
			name: "override, no custom annotation",
			override: RoutedOverrideSpec{
				OverrideSpec: OverrideSpec{
					EmbeddedLabelsAnnotations: &EmbeddedLabelsAnnotations{
						Annotations: map[string]string{"key": "val"},
					},
				}},
			annotation: map[string]string{},
			want: RoutedOverrideSpec{
				OverrideSpec: OverrideSpec{

					EmbeddedLabelsAnnotations: &EmbeddedLabelsAnnotations{
						Annotations: map[string]string{"key": "val"},
					},
				},
			},
		},
		{
			name: "override, additional custom annotation",
			override: RoutedOverrideSpec{
				OverrideSpec: OverrideSpec{
					EmbeddedLabelsAnnotations: &EmbeddedLabelsAnnotations{
						Annotations: map[string]string{"key": "val"},
					},
				}},
			annotation: map[string]string{"custom": "val"},
			want: RoutedOverrideSpec{
				OverrideSpec: OverrideSpec{
					EmbeddedLabelsAnnotations: &EmbeddedLabelsAnnotations{
						Annotations: map[string]string{
							"key":    "val",
							"custom": "val",
						},
					},
				},
			},
		},
		{
			name: "override, additional custom same annotation",
			override: RoutedOverrideSpec{
				OverrideSpec: OverrideSpec{
					EmbeddedLabelsAnnotations: &EmbeddedLabelsAnnotations{
						Annotations: map[string]string{"key": "val"},
					},
				}},
			annotation: map[string]string{"key": "custom"},
			want: RoutedOverrideSpec{
				OverrideSpec: OverrideSpec{
					EmbeddedLabelsAnnotations: &EmbeddedLabelsAnnotations{
						Annotations: map[string]string{
							"key": "val",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			tt.override.AddAnnotation(tt.annotation)
			g.Expect(tt.override.Annotations).To(Equal(tt.want.Annotations))
		})
	}
}
