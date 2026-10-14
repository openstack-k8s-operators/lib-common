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
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/onsi/gomega" // nolint:revive
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
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
	// portNameMismatch has a port Name that does not match the Service
	// object's Name ("foo") — GetServiceHostnamePort must still find it.
	portNameMismatch = []corev1.ServicePort{
		{
			Name:        "internal",
			Protocol:    corev1.ProtocolTCP,
			AppProtocol: nil,
			Port:        int32(8778),
			TargetPort:  intstr.FromInt(0),
			NodePort:    0,
		},
	}
	// multiPortNamed has the endpoint port (named like the Service object,
	// "foo") second in the list.
	multiPortNamed = []corev1.ServicePort{
		{Name: "metrics", Protocol: corev1.ProtocolTCP, Port: int32(9090), TargetPort: intstr.FromInt(0)},
		{Name: "foo", Protocol: corev1.ProtocolTCP, Port: int32(8080), TargetPort: intstr.FromInt(0)},
	}
	// multiPortUnnamed has no port named like the Service object, so the
	// endpoint port is ambiguous.
	multiPortUnnamed = []corev1.ServicePort{
		{Name: "metrics", Protocol: corev1.ProtocolTCP, Port: int32(9090), TargetPort: intstr.FromInt(0)},
		{Name: "internal", Protocol: corev1.ProtocolTCP, Port: int32(8080), TargetPort: intstr.FromInt(0)},
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
			name:     "ClusterIP service with port Name different from object Name",
			service:  getServiceWithPort(svcClusterIP, portNameMismatch),
			override: OverrideSpec{},
			want: Service{
				service:         getServiceWithPort(svcClusterIP, portNameMismatch),
				timeout:         timeout,
				serviceHostname: "foo.namespace.svc",
			},
			wantPort:                "8778",
			wantOverrideServiceSpec: overrideServiceSpecClusterIP,
		},
		{
			name:     "ClusterIP multi-port service picks the port named like the object",
			service:  getServiceWithPort(svcClusterIP, multiPortNamed),
			override: OverrideSpec{},
			want: Service{
				service:         getServiceWithPort(svcClusterIP, multiPortNamed),
				timeout:         timeout,
				serviceHostname: "foo.namespace.svc",
			},
			wantPort:                "8080",
			wantOverrideServiceSpec: overrideServiceSpecClusterIP,
		},
		{
			name:     "ClusterIP multi-port service with no port named like the object has no endpoint port",
			service:  getServiceWithPort(svcClusterIP, multiPortUnnamed),
			override: OverrideSpec{},
			want: Service{
				service:         getServiceWithPort(svcClusterIP, multiPortUnnamed),
				timeout:         timeout,
				serviceHostname: "foo.namespace.svc",
			},
			wantPort:                "",
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
		wantErr     error
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
			name:        "empty endpointURL uses service URL",
			service:     getServiceWithPort(svcClusterIP, portCustom),
			endpointURL: ptr.To(""),
			proto:       ProtocolHTTP,
			path:        "/path",
			want:        "http://foo.namespace.svc:8080/path",
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
		{
			name:        "Empty EndpointURL falls back to the Service's only port",
			service:     getServiceWithPort(svcClusterIP, portNameMismatch),
			endpointURL: ptr.To(""),
			proto:       ProtocolHTTP,
			path:        "/path",
			want:        "http://foo.namespace.svc:8778/path",
		},
		{
			name: "ExternalName service without ports keeps the hostname-only endpoint",
			service: &corev1.Service{
				ObjectMeta: svcClusterIP.ObjectMeta,
				Spec:       corev1.ServiceSpec{Type: corev1.ServiceTypeExternalName, ExternalName: "ext.example.com"},
			},
			endpointURL: nil,
			proto:       ProtocolHTTPS,
			path:        "/path",
			want:        "https://foo.namespace.svc/path",
		},
		{
			name:        "Multi-port service with no port named like the object, no override",
			service:     getServiceWithPort(svcClusterIP, multiPortUnnamed),
			endpointURL: nil,
			proto:       ProtocolHTTP,
			path:        "/path",
			wantErr:     ErrEndpointPortUnknown,
		},
		{
			name:        "Multi-port service with no port named like the object, empty EndpointURL",
			service:     getServiceWithPort(svcClusterIP, multiPortUnnamed),
			endpointURL: ptr.To(""),
			proto:       ProtocolHTTP,
			path:        "/path",
			wantErr:     ErrEndpointPortUnknown,
		},
		{
			name:        "Multi-port service with no port named like the object, EndpointURL set",
			service:     getServiceWithPort(svcClusterIP, multiPortUnnamed),
			endpointURL: ptr.To("http://this.url"),
			proto:       ProtocolHTTP,
			path:        "/path",
			want:        "http://this.url/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			service, err := NewService(tt.service, timeout, nil)
			g.Expect(err).ToNot(HaveOccurred())
			url, err := service.GetAPIEndpoint(tt.endpointURL, ptr.To(tt.proto), tt.path)
			if tt.wantErr != nil {
				g.Expect(err).To(MatchError(tt.wantErr))
				return
			}
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(url).To(Equal(tt.want))
		})
	}
}

func TestGetVHostServerName(t *testing.T) {
	tests := []struct {
		name                 string
		endpointURL          *string
		defaultName          string
		extraAliases         []string
		wantServer           string
		wantAliases          []string
		wantErr              bool
		wantIPErr            bool
		wantDefaultNameIPErr bool
	}{
		{
			name:        "no endpointURL uses default name, FQDN alias",
			endpointURL: nil,
			defaultName: "placement-public.openstack.svc",
			wantServer:  "placement-public.openstack.svc",
			wantAliases: []string{"placement-public.openstack.svc.cluster.local"},
		},
		{
			name:         "no endpointURL includes FQDN and extraAliases",
			endpointURL:  nil,
			defaultName:  "placement-public.openstack.svc",
			extraAliases: []string{"placement.extra.example"},
			wantServer:   "placement-public.openstack.svc",
			wantAliases:  []string{"placement-public.openstack.svc.cluster.local", "placement.extra.example"},
		},
		{
			name:                 "IP default name without endpointURL returns error",
			endpointURL:          nil,
			defaultName:          "192.0.2.1",
			wantErr:              true,
			wantDefaultNameIPErr: true,
		},
		{
			name:                 "IP default name with DNS endpointURL returns error",
			endpointURL:          ptr.To("https://placement.foo.bar:8778"),
			defaultName:          "192.0.2.1",
			wantErr:              true,
			wantDefaultNameIPErr: true,
		},
		{
			name:        "endpointURL host becomes ServerName, default and FQDN become aliases",
			endpointURL: ptr.To("https://placement.foo.bar:8778"),
			defaultName: "placement-public.openstack.svc",
			wantServer:  "placement.foo.bar",
			wantAliases: []string{"placement-public.openstack.svc", "placement-public.openstack.svc.cluster.local"},
		},
		{
			name:         "extraAliases are appended and deduplicated against ServerName",
			endpointURL:  ptr.To("https://placement.foo.bar:8778"),
			defaultName:  "placement-public.openstack.svc",
			extraAliases: []string{"placement-public.openstack.svc.cluster.local", "placement.foo.bar"},
			wantServer:   "placement.foo.bar",
			wantAliases:  []string{"placement-public.openstack.svc", "placement-public.openstack.svc.cluster.local"},
		},
		{
			name:        "endpointURL host same as default name yields no duplicate alias, FQDN still present",
			endpointURL: ptr.To("https://placement-public.openstack.svc:8778"),
			defaultName: "placement-public.openstack.svc",
			wantServer:  "placement-public.openstack.svc",
			wantAliases: []string{"placement-public.openstack.svc.cluster.local"},
		},
		{
			name:        "endpointURL host same as FQDN yields no duplicate alias, short name still present",
			endpointURL: ptr.To("https://placement-public.openstack.svc.cluster.local:8778"),
			defaultName: "placement-public.openstack.svc",
			wantServer:  "placement-public.openstack.svc.cluster.local",
			wantAliases: []string{"placement-public.openstack.svc"},
		},
		{
			name:        "malformed endpointURL returns error",
			endpointURL: ptr.To("http://foo bar.com:8778"),
			defaultName: "placement-public.openstack.svc",
			wantErr:     true,
		},
		{
			name:        "endpointURL with no host returns error",
			endpointURL: ptr.To("/just/a/path"),
			defaultName: "placement-public.openstack.svc",
			wantErr:     true,
		},
		{
			name:        "IPv4 endpointURL host returns error",
			endpointURL: ptr.To("https://192.0.2.1:8778"),
			defaultName: "placement-public.openstack.svc",
			wantErr:     true,
			wantIPErr:   true,
		},
		{
			name:        "IPv4 endpointURL host with trailing root dot returns error",
			endpointURL: ptr.To("https://127.0.0.1.:8778"),
			defaultName: "placement-public.openstack.svc",
			wantErr:     true,
			wantIPErr:   true,
		},
		{
			name:        "IPv6 endpointURL host returns error",
			endpointURL: ptr.To("https://[2001:db8::1]:8778"),
			defaultName: "placement-public.openstack.svc",
			wantErr:     true,
			wantIPErr:   true,
		},
		{
			name:        "zone-scoped IPv6 endpointURL host returns error",
			endpointURL: ptr.To("https://[fe80::1%25eth0]:8778"),
			defaultName: "placement-public.openstack.svc",
			wantErr:     true,
			wantIPErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			override := RoutedOverrideSpec{EndpointURL: tt.endpointURL}
			serverName, aliases, err := GetVHostServerName(override, tt.defaultName, tt.extraAliases...)

			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				if tt.wantIPErr {
					g.Expect(errors.Is(err, ErrEndpointURLIP)).To(BeTrue())
				}
				if tt.wantDefaultNameIPErr {
					g.Expect(errors.Is(err, ErrDefaultNameIP)).To(BeTrue())
				}
				return
			}
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(serverName).To(Equal(tt.wantServer))
			g.Expect(aliases).To(Equal(tt.wantAliases))
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

func TestNewServicePortsOverride(t *testing.T) {
	basePorts := []corev1.ServicePort{
		{Name: "amqp", Protocol: corev1.ProtocolTCP, Port: 5672, TargetPort: intstr.FromInt(5672)},
		{Name: "amqps", Protocol: corev1.ProtocolTCP, Port: 5671, TargetPort: intstr.FromInt(5671)},
	}

	tests := []struct {
		name     string
		override *OverrideSpec
		want     []corev1.ServicePort
	}{
		{
			name:     "no port override keeps base ports",
			override: &OverrideSpec{Spec: &OverrideServiceSpec{Type: corev1.ServiceTypeLoadBalancer}},
			want:     basePorts,
		},
		{
			name: "change matching port keeps other fields and other ports",
			override: &OverrideSpec{Spec: &OverrideServiceSpec{
				Ports: []OverrideServicePort{{Name: "amqp", Port: 5673}},
			}},
			// amqp Port changes to 5673, TargetPort/Protocol are preserved from
			// base; amqps is untouched.
			want: []corev1.ServicePort{
				{Name: "amqp", Protocol: corev1.ProtocolTCP, Port: 5673, TargetPort: intstr.FromInt(5672)},
				{Name: "amqps", Protocol: corev1.ProtocolTCP, Port: 5671, TargetPort: intstr.FromInt(5671)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			base := svcClusterIP.DeepCopy()
			base.Spec.Ports = append([]corev1.ServicePort{}, basePorts...)

			svc, err := NewService(base, timeout, tt.override)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(svc.GetSpec().Ports).To(Equal(tt.want))
		})
	}
}

// TestNewServicePortsOverrideTargetPort verifies that overriding only the Port
// of a base port that omits TargetPort preserves backend routing: TargetPort is
// pinned to the original base Port instead of following the new Port via
// Kubernetes' TargetPort=Port defaulting.
func TestNewServicePortsOverrideTargetPort(t *testing.T) {
	g := NewWithT(t)

	base := svcClusterIP.DeepCopy()
	// single port, no TargetPort set -> would default to Port (80)
	base.Spec.Ports = []corev1.ServicePort{
		{Name: "foo", Protocol: corev1.ProtocolTCP, Port: 80},
	}

	override := &OverrideSpec{Spec: &OverrideServiceSpec{
		Ports: []OverrideServicePort{{Name: "foo", Port: 8080}},
	}}

	svc, err := NewService(base, timeout, override)
	g.Expect(err).ToNot(HaveOccurred())
	// Port changes to 8080 but backend routing stays at the original 80.
	g.Expect(svc.GetSpec().Ports).To(Equal([]corev1.ServicePort{
		{Name: "foo", Protocol: corev1.ProtocolTCP, Port: 8080, TargetPort: intstr.FromInt32(80)},
	}))
}

func TestNewServicePortsOverrideInvalid(t *testing.T) {
	basePorts := []corev1.ServicePort{
		{Name: "amqp", Protocol: corev1.ProtocolTCP, Port: 5672, TargetPort: intstr.FromInt(5672)},
		{Name: "amqps", Protocol: corev1.ProtocolTCP, Port: 5671, TargetPort: intstr.FromInt(5671)},
	}

	tests := []struct {
		name     string
		override *OverrideSpec
		errMatch string
	}{
		{
			name: "changing a port to collide with another is rejected",
			override: &OverrideSpec{Spec: &OverrideServiceSpec{
				Ports: []OverrideServicePort{{Name: "amqp", Port: 5671}},
			}},
			errMatch: "duplicate service port after applying ports override: 5671/TCP",
		},
		{
			name: "override referencing an unknown port name is rejected",
			override: &OverrideSpec{Spec: &OverrideServiceSpec{
				Ports: []OverrideServicePort{{Name: "does-not-exist", Port: 9999}},
			}},
			errMatch: `ports override references unknown service port name: "does-not-exist"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			base := svcClusterIP.DeepCopy()
			base.Spec.Ports = append([]corev1.ServicePort{}, basePorts...)

			_, err := NewService(base, timeout, tt.override)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring(tt.errMatch))
		})
	}
}

func TestRemoveServiceAnnotation(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	owner := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "owner", Namespace: "namespace", UID: "owner-uid"}}
	current := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo", Namespace: "namespace",
			Annotations: map[string]string{"managed": "old", "other": "preserved"},
		},
		Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP},
	}
	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(current).Build()
	h, err := helper.NewHelper(owner, client, nil, scheme.Scheme, ctrl.Log)
	g.Expect(err).NotTo(HaveOccurred())
	svc, err := NewService(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "namespace"},
		Spec:       corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP},
	}, 0, nil)
	g.Expect(err).NotTo(HaveOccurred())
	svc.RemoveAnnotation("managed")
	_, err = svc.CreateOrPatch(ctx, h)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(client.Get(ctx, types.NamespacedName{Name: "foo", Namespace: "namespace"}, current)).To(Succeed())
	g.Expect(current.Annotations).NotTo(HaveKey("managed"))
	g.Expect(current.Annotations).To(HaveKeyWithValue("other", "preserved"))
}

func TestRemoveThenAddServiceAnnotation(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	owner := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "owner", Namespace: "namespace", UID: "owner-uid"}}
	current := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo", Namespace: "namespace",
			Annotations: map[string]string{"managed": "old"},
		},
		Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP},
	}
	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(current).Build()
	h, err := helper.NewHelper(owner, client, nil, scheme.Scheme, ctrl.Log)
	g.Expect(err).NotTo(HaveOccurred())
	svc, err := NewService(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "namespace"},
		Spec:       corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP},
	}, 0, nil)
	g.Expect(err).NotTo(HaveOccurred())
	svc.RemoveAnnotation("managed")
	svc.AddAnnotation(map[string]string{"managed": "new"})
	_, err = svc.CreateOrPatch(ctx, h)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(client.Get(ctx, types.NamespacedName{Name: "foo", Namespace: "namespace"}, current)).To(Succeed())
	g.Expect(current.Annotations).To(HaveKeyWithValue("managed", "new"))
}
