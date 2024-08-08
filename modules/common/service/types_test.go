/*
Copyright 2024 Red Hat

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
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestEndpointValidate(t *testing.T) {
	tests := []struct {
		name string
		e    Endpoint
		want error
	}{
		{
			name: "Valid endpoint",
			e:    EndpointInternal,
			want: nil,
		},
		{
			name: "Wrong endpoint",
			e:    Endpoint("wrooong"),
			want: fmt.Errorf("invalid endpoint type: wrooong"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			if tt.want == nil {
				g.Expect(tt.e.Validate()).To(Succeed())
			} else {
				g.Expect(tt.e.Validate()).To(Equal(tt.want))
			}
		})
	}
}

func TestValidateRoutedOverrides(t *testing.T) {
	//basePath := field.NewPath("spec")

	tests := []struct {
		name      string
		basePath  *field.Path
		overrides map[Endpoint]RoutedOverrideSpec
		want      field.ErrorList
	}{
		{
			name:     "Valid override config",
			basePath: field.NewPath("spec").Child("override").Child("service"),
			overrides: map[Endpoint]RoutedOverrideSpec{
				EndpointInternal: {},
			},
			want: field.ErrorList{},
		},
		{
			name:     "Wrong override endpoint",
			basePath: field.NewPath("spec").Child("override").Child("service"),
			overrides: map[Endpoint]RoutedOverrideSpec{
				Endpoint("wrooong"): {},
			},
			want: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.override.service[wrooong]",
					BadValue: "wrooong",
					Detail:   "invalid endpoint type: wrooong",
				},
			},
		},
		{
			name:     "Both good and wrong override endpoint configs",
			basePath: field.NewPath("spec").Child("foo").Child("bar"),
			overrides: map[Endpoint]RoutedOverrideSpec{
				EndpointInternal:    {},
				Endpoint("wrooong"): {},
			},
			want: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.foo.bar[wrooong]",
					BadValue: "wrooong",
					Detail:   "invalid endpoint type: wrooong",
				},
			},
		},
		{
			name:     "Multiple wrong override endpoints",
			basePath: field.NewPath("spec").Child("foo"),
			overrides: map[Endpoint]RoutedOverrideSpec{
				EndpointInternal:       {},
				Endpoint("wrooong"):    {},
				Endpoint("wroooooong"): {},
			},
			want: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.foo[wrooong]",
					BadValue: "wrooong",
					Detail:   "invalid endpoint type: wrooong",
				},
				&field.Error{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.foo[wroooooong]",
					BadValue: "wroooooong",
					Detail:   "invalid endpoint type: wroooooong",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(ValidateRoutedOverrides(tt.basePath, tt.overrides)).To(ContainElements(tt.want))
		})
	}
}
