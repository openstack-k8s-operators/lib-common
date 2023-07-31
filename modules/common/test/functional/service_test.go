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
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getExampleService(namespace string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-svc",
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
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "test-port",
					Port:       int32(80),
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(8080),
				},
			},
			Selector: map[string]string{
				"internal": "true",
				"service":  "foo",
			},
		},
	}
}

var _ = Describe("service package", func() {
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

	It("creates service with defaults", func() {
		s, err := service.NewService(
			getExampleService(namespace),
			timeout,
			&service.OverrideSpec{},
		)
		Expect(err).ShouldNot(HaveOccurred())

		_, err = s.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		svc := th.AssertServiceExists(types.NamespacedName{Namespace: namespace, Name: "test-svc"})
		Expect(svc.Annotations["anno"]).To(Equal("a"))
		Expect(svc.Annotations["replace"]).To(Equal("a"))
		Expect(svc.Labels["label"]).To(Equal("a"))
		Expect(svc.Labels["replace"]).To(Equal("a"))
		Expect(svc.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
		Expect(svc.Spec.Ports[0].Name).To(Equal("test-port"))
		Expect(svc.Spec.Ports[0].Port).To(Equal(int32(80)))
		Expect(svc.Spec.Ports[0].Protocol).To(Equal(corev1.ProtocolTCP))
		Expect(svc.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(8080)))
	})

	It("merges labels to the service", func() {
		s, err := service.NewService(
			getExampleService(namespace),
			timeout,
			&service.OverrideSpec{
				EmbeddedLabelsAnnotations: &service.EmbeddedLabelsAnnotations{
					Labels: map[string]string{
						"foo":     "b",
						"replace": "b",
					},
				},
			},
		)
		Expect(err).ShouldNot(HaveOccurred())

		_, err = s.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		rv1 := th.AssertServiceExists(types.NamespacedName{Namespace: namespace, Name: "test-svc"})
		// non overridden label exists
		Expect(rv1.Labels["label"]).To(Equal("a"))
		// adds new label
		Expect(rv1.Labels["foo"]).To(Equal("b"))
		// override replaces existing label
		Expect(rv1.Labels["replace"]).To(Equal("b"))
	})

	It("merges annotations to the service", func() {
		s, err := service.NewService(
			getExampleService(namespace),
			timeout,
			&service.OverrideSpec{
				EmbeddedLabelsAnnotations: &service.EmbeddedLabelsAnnotations{
					Annotations: map[string]string{
						"foo":     "b",
						"replace": "b",
					},
				},
			},
		)
		Expect(err).ShouldNot(HaveOccurred())

		_, err = s.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		rv1 := th.AssertServiceExists(types.NamespacedName{Namespace: namespace, Name: "test-svc"})
		// non overridden annotation exists
		Expect(rv1.Annotations["anno"]).To(Equal("a"))
		// adds new annotation
		Expect(rv1.Annotations["foo"]).To(Equal("b"))
		// override replaces existing annotation
		Expect(rv1.Annotations["replace"]).To(Equal("b"))
	})

	It("overrides spec.Type to LoadBalancer", func() {
		s, err := service.NewService(
			getExampleService(namespace),
			timeout,
			&service.OverrideSpec{
				Spec: &service.OverrideServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
		)
		Expect(err).ShouldNot(HaveOccurred())

		_, err = s.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		svc := th.AssertServiceExists(types.NamespacedName{Namespace: namespace, Name: "test-svc"})
		Expect(svc.Spec.Type).To(Equal(corev1.ServiceTypeLoadBalancer))
	})
})
