/*
Copyright 2026 Red Hat

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

package webhook

import (
	"testing"

	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"k8s.io/apimachinery/pkg/util/validation/field"

	. "github.com/onsi/gomega" // nolint:revive
)

func TestValidateServicePortNames(t *testing.T) {
	tests := []struct {
		name       string
		ports      []service.OverrideServicePort
		validNames []string
		wantErrs   int
		wantDetail string
	}{
		{
			name:       "no ports",
			ports:      nil,
			validNames: []string{"placement-public"},
			wantErrs:   0,
		},
		{
			name:       "matching single valid name",
			ports:      []service.OverrideServicePort{{Name: "placement-public", Port: 13778}},
			validNames: []string{"placement-public"},
			wantErrs:   0,
		},
		{
			name:       "mismatched single valid name",
			ports:      []service.OverrideServicePort{{Name: "wrong", Port: 13778}},
			validNames: []string{"placement-public"},
			wantErrs:   1,
			wantDetail: `must be "placement-public"`,
		},
		{
			name:  "matching one of several valid names",
			ports: []service.OverrideServicePort{{Name: "glance-default-public", Port: 13292}},
			validNames: []string{
				"glance-default-public",
				"glance-other-public",
			},
			wantErrs: 0,
		},
		{
			name:  "mismatched with several valid names",
			ports: []service.OverrideServicePort{{Name: "wrong", Port: 13292}},
			validNames: []string{
				"glance-default-public",
				"glance-other-public",
			},
			wantErrs:   1,
			wantDetail: `must be one of ["glance-default-public" "glance-other-public"]`,
		},
		{
			name:       "no valid names configured",
			ports:      []service.OverrideServicePort{{Name: "placement-public", Port: 13778}},
			validNames: nil,
			wantErrs:   1,
			wantDetail: "no override port names are valid here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			errs := ValidateServicePortNames(field.NewPath("spec", "ports"), tt.ports, tt.validNames...)
			g.Expect(errs).To(HaveLen(tt.wantErrs))
			if tt.wantErrs > 0 {
				g.Expect(errs[0].Detail).To(Equal(tt.wantDetail))
				g.Expect(errs[0].Field).To(Equal("spec.ports[0].name"))
			}
		})
	}
}

func TestDefaultServicePortNames(t *testing.T) {
	g := NewWithT(t)
	g.Expect(DefaultServicePortNames()).To(Equal(map[service.Endpoint][]string{
		service.EndpointPublic:   {"public"},
		service.EndpointInternal: {"internal"},
	}))
}

func overrideWithPorts(ports ...service.OverrideServicePort) service.RoutedOverrideSpec {
	return service.RoutedOverrideSpec{
		OverrideSpec: service.OverrideSpec{
			Spec: &service.OverrideServiceSpec{Ports: ports},
		},
	}
}

func TestValidatePortNameOverrides(t *testing.T) {
	validNames := map[service.Endpoint][]string{
		service.EndpointPublic:   {"placement-public"},
		service.EndpointInternal: {"placement-internal"},
	}

	tests := []struct {
		name       string
		overrides  map[service.Endpoint]service.RoutedOverrideSpec
		validNames map[service.Endpoint][]string
		wantFields []string
	}{
		{
			name:       "no overrides",
			overrides:  nil,
			validNames: validNames,
			wantFields: nil,
		},
		{
			name: "no spec set for the endpoint",
			overrides: map[service.Endpoint]service.RoutedOverrideSpec{
				service.EndpointPublic: {},
			},
			validNames: validNames,
			wantFields: nil,
		},
		{
			name: "endpoint not present in validNames is not checked",
			overrides: map[service.Endpoint]service.RoutedOverrideSpec{
				service.EndpointPublic: overrideWithPorts(service.OverrideServicePort{Name: "wrong", Port: 13778}),
			},
			validNames: map[service.Endpoint][]string{service.EndpointInternal: {"placement-internal"}},
			wantFields: nil,
		},
		{
			name: "matching name on one endpoint",
			overrides: map[service.Endpoint]service.RoutedOverrideSpec{
				service.EndpointPublic: overrideWithPorts(service.OverrideServicePort{Name: "placement-public", Port: 13778}),
			},
			validNames: validNames,
			wantFields: nil,
		},
		{
			name: "mismatched name on one endpoint",
			overrides: map[service.Endpoint]service.RoutedOverrideSpec{
				service.EndpointPublic: overrideWithPorts(service.OverrideServicePort{Name: "wrong", Port: 13778}),
			},
			validNames: validNames,
			wantFields: []string{"spec.override.service[public].spec.ports[0].name"},
		},
		{
			name: "DefaultServicePortNames composes with matching port names",
			overrides: map[service.Endpoint]service.RoutedOverrideSpec{
				service.EndpointPublic:   overrideWithPorts(service.OverrideServicePort{Name: "public", Port: 443}),
				service.EndpointInternal: overrideWithPorts(service.OverrideServicePort{Name: "internal", Port: 80}),
			},
			validNames: DefaultServicePortNames(),
			wantFields: nil,
		},
		{
			name: "DefaultServicePortNames composes with mismatched port name",
			overrides: map[service.Endpoint]service.RoutedOverrideSpec{
				service.EndpointPublic: overrideWithPorts(service.OverrideServicePort{Name: "wrong", Port: 443}),
			},
			validNames: DefaultServicePortNames(),
			wantFields: []string{"spec.override.service[public].spec.ports[0].name"},
		},
		{
			name: "mismatched name on both endpoints is reported in sorted (internal, public) order",
			overrides: map[service.Endpoint]service.RoutedOverrideSpec{
				service.EndpointPublic:   overrideWithPorts(service.OverrideServicePort{Name: "wrong-public", Port: 13778}),
				service.EndpointInternal: overrideWithPorts(service.OverrideServicePort{Name: "wrong-internal", Port: 8778}),
			},
			validNames: validNames,
			wantFields: []string{
				"spec.override.service[internal].spec.ports[0].name",
				"spec.override.service[public].spec.ports[0].name",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			errs := ValidatePortNameOverrides(
				field.NewPath("spec", "override", "service"), tt.overrides, tt.validNames)
			var gotFields []string
			for _, e := range errs {
				gotFields = append(gotFields, e.Field)
			}
			g.Expect(gotFields).To(Equal(tt.wantFields))
		})
	}
}
