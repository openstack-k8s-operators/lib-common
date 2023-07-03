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
package functional

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openstack-k8s-operators/lib-common/modules/common/route"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	routev1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getExampleRoute(namespace string) *routev1.Route {
	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-route",
			Namespace: namespace,
			Labels: map[string]string{
				"label":   "a",
				"replace": "a",
			},
			Annotations: map[string]string{
				"anno":    "a",
				"replace": "a",
			},
		},
		Spec: routev1.RouteSpec{
			Host: "some.host.svc",
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromInt(80),
			},
			To: routev1.RouteTargetReference{
				Kind:   "Service",
				Name:   "my-service",
				Weight: pointer.Int32(100),
			},
		},
	}
}

var _ = Describe("route package", func() {
	var namespace string

	BeforeEach(func() {
		// NOTE(gibi): We need to create a unique namespace for each test run
		// as namespaces cannot be deleted in a locally running envtest. See
		// https://book.kubebuilder.io/reference/envtest.html#namespace-usage-limitation
		namespace = uuid.New().String()
		th.CreateNamespace(namespace)
		// We still request the delete of the Namespace to properly cleanup if
		// we run the test in an existing cluster.
		DeferCleanup(th.DeleteNamespace, namespace)

	})

	It("creates route with defaults", func() {
		r := route.NewRoute(
			getExampleRoute(namespace),
			map[string]string{},
			timeout,
			&route.OverrideSpec{},
		)

		_, err := r.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		rv1 := th.AssertRouteExists(types.NamespacedName{Namespace: namespace, Name: "test-route"})
		Expect(rv1.Annotations["anno"]).To(Equal("a"))
		Expect(rv1.Annotations["replace"]).To(Equal("a"))
		Expect(rv1.Labels["label"]).To(Equal("a"))
		Expect(rv1.Labels["replace"]).To(Equal("a"))
		Expect(rv1.Spec.Host).To(Equal("some.host.svc"))
		Expect(rv1.Spec.Port.TargetPort.IntVal).To(Equal(int32(80)))
		Expect(rv1.Spec.To.Name).To(Equal("my-service"))
		Expect(*rv1.Spec.To.Weight).To(Equal(int32(100)))

	})

	It("merges labels to the route", func() {
		r := route.NewRoute(
			getExampleRoute(namespace),
			map[string]string{},
			timeout,
			&route.OverrideSpec{
				EmbeddedLabelsAnnotations: &route.EmbeddedLabelsAnnotations{
					Labels: map[string]string{
						"foo":     "b",
						"replace": "b",
					},
				},
			},
		)

		_, err := r.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		rv1 := th.AssertRouteExists(types.NamespacedName{Namespace: namespace, Name: "test-route"})
		// non overridden label exists
		Expect(rv1.Labels["label"]).To(Equal("a"))
		// adds new label
		Expect(rv1.Labels["foo"]).To(Equal("b"))
		// override replaces existing label
		Expect(rv1.Labels["replace"]).To(Equal("b"))
	})

	It("merges annotations to the route", func() {
		r := route.NewRoute(
			getExampleRoute(namespace),
			map[string]string{},
			timeout,
			&route.OverrideSpec{
				EmbeddedLabelsAnnotations: &route.EmbeddedLabelsAnnotations{
					Annotations: map[string]string{
						"foo":     "b",
						"replace": "b",
					},
				},
			},
		)

		_, err := r.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		rv1 := th.AssertRouteExists(types.NamespacedName{Namespace: namespace, Name: "test-route"})
		// non overridden annotation exists
		Expect(rv1.Annotations["anno"]).To(Equal("a"))
		// adds new annotation
		Expect(rv1.Annotations["foo"]).To(Equal("b"))
		// override replaces existing annotation
		Expect(rv1.Annotations["replace"]).To(Equal("b"))
	})

	It("overrides spec.host if specified", func() {
		r := route.NewRoute(
			getExampleRoute(namespace),
			map[string]string{},
			timeout,
			&route.OverrideSpec{
				Spec: &route.Spec{
					Host: "custom.host.domain",
				},
			},
		)

		_, err := r.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		rv1 := th.AssertRouteExists(types.NamespacedName{Namespace: namespace, Name: "test-route"})
		Expect(rv1.Spec.Host).To(Equal("custom.host.domain"))
	})

	It("overrides spec.subdomain if specified", func() {
		r := route.NewRoute(
			getExampleRoute(namespace),
			map[string]string{},
			timeout,
			&route.OverrideSpec{
				Spec: &route.Spec{
					Subdomain: "subdomain",
				},
			},
		)

		_, err := r.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		rv1 := th.AssertRouteExists(types.NamespacedName{Namespace: namespace, Name: "test-route"})
		Expect(rv1.Spec.Subdomain).To(Equal("subdomain"))
	})

	It("overrides spec.path if specified", func() {
		r := route.NewRoute(
			getExampleRoute(namespace),
			map[string]string{},
			timeout,
			&route.OverrideSpec{
				Spec: &route.Spec{
					Path: "/some/path",
				},
			},
		)

		_, err := r.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		rv1 := th.AssertRouteExists(types.NamespacedName{Namespace: namespace, Name: "test-route"})
		Expect(rv1.Spec.Path).To(Equal("/some/path"))
	})

	It("overrides spec.to if specified", func() {
		r := route.NewRoute(
			getExampleRoute(namespace),
			map[string]string{},
			timeout,
			&route.OverrideSpec{
				Spec: &route.Spec{
					To: route.TargetReference{
						Name:   "my-custom-service",
						Weight: pointer.Int32(10),
					},
				},
			},
		)

		_, err := r.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		rv1 := th.AssertRouteExists(types.NamespacedName{Namespace: namespace, Name: "test-route"})
		Expect(rv1.Spec.To.Kind).To(Equal("Service"))
		Expect(rv1.Spec.To.Name).To(Equal("my-custom-service"))
		Expect(*rv1.Spec.To.Weight).To(Equal(int32(10)))
	})

	It("overrides spec.alternateBackends if specified", func() {
		r := route.NewRoute(
			getExampleRoute(namespace),
			map[string]string{},
			timeout,
			&route.OverrideSpec{
				Spec: &route.Spec{
					AlternateBackends: []route.TargetReference{
						{
							Kind:   "Service",
							Name:   "my-alternate-service",
							Weight: pointer.Int32(200),
						},
					},
				},
			},
		)

		_, err := r.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		rv1 := th.AssertRouteExists(types.NamespacedName{Namespace: namespace, Name: "test-route"})
		Expect(rv1.Spec.AlternateBackends[0].Name).To(Equal("my-alternate-service"))
		Expect(*rv1.Spec.AlternateBackends[0].Weight).To(Equal(int32(200)))
	})

	It("overrides spec.port if specified", func() {
		r := route.NewRoute(
			getExampleRoute(namespace),
			map[string]string{},
			timeout,
			&route.OverrideSpec{
				Spec: &route.Spec{
					Port: &routev1.RoutePort{
						TargetPort: intstr.FromInt(8080),
					},
				},
			},
		)

		_, err := r.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		rv1 := th.AssertRouteExists(types.NamespacedName{Namespace: namespace, Name: "test-route"})
		Expect(rv1.Spec.Port.TargetPort.IntVal).To(Equal(int32(8080)))
	})

	It("overrides spec.tls if specified", func() {
		r := route.NewRoute(
			getExampleRoute(namespace),
			map[string]string{},
			timeout,
			&route.OverrideSpec{
				Spec: &route.Spec{
					TLS: &routev1.TLSConfig{
						Termination:   routev1.TLSTerminationEdge,
						Certificate:   "cert",
						Key:           "key",
						CACertificate: "cacert",
					},
				},
			},
		)

		_, err := r.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		rv1 := th.AssertRouteExists(types.NamespacedName{Namespace: namespace, Name: "test-route"})
		Expect(rv1.Spec.TLS.Termination).To(Equal(routev1.TLSTerminationEdge))
		Expect(rv1.Spec.TLS.Certificate).To(Equal("cert"))
		Expect(rv1.Spec.TLS.Key).To(Equal("key"))
		Expect(rv1.Spec.TLS.CACertificate).To(Equal("cacert"))
	})
})
