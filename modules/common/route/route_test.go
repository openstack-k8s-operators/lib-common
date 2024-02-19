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

package route

import (
	"testing"
	"time"

	routev1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	. "github.com/onsi/gomega"
)

var (
	route1 = routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "namespace",
			Labels: map[string]string{
				"foo": "bar",
			},
		},
	}
	port1 = routev1.RoutePort{
		TargetPort: intstr.FromInt(80),
	}
	timeout = time.Duration(5) * time.Second
)

func getRouteWithPort(r routev1.Route, port routev1.RoutePort) *routev1.Route {
	r.Spec.Port = &port

	return &r
}

func TestNewRoute(t *testing.T) {
	tests := []struct {
		name     string
		route    *routev1.Route
		override []OverrideSpec
		want     Route
		wantPort string
	}{
		{
			name:     "Route example with no override",
			route:    getRouteWithPort(route1, port1),
			override: nil,
			want: Route{
				route:   getRouteWithPort(route1, port1),
				timeout: timeout,
			},
			wantPort: "80",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			route, err := NewRoute(tt.route, timeout, nil)
			g.Expect(err).ToNot(HaveOccurred())
			// timeout
			g.Expect(route.timeout).To(Equal(timeout))
			// GetLabels
			g.Expect(route.GetLabels()).To(Equal(map[string]string{
				"foo": "bar",
			}))
			// AddAnnotation
			route.AddAnnotation(map[string]string{"foo": "bar"})
			// GetAnnotations
			g.Expect(route.GetAnnotations()).To(Equal(map[string]string{"foo": "bar"}))
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
