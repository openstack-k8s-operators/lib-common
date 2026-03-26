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
package functional

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openstack-k8s-operators/lib-common/modules/common/secret"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Secret helpers", func() {
	var namespace string

	BeforeEach(func() {
		namespace = uuid.New().String()
		th.CreateNamespace(namespace)
	})

	When("CreateOrPatchSecretPreserve is called", func() {
		It("creates a secret with initial data", func() {
			owner := th.CreateNamespace("secret-owner")
			s := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: namespace,
				},
				Type: corev1.SecretTypeOpaque,
				Data: map[string][]byte{
					"password": []byte("initial-password"),
				},
			}

			hash, op, err := secret.CreateOrPatchSecretPreserve(ctx, h, owner, s, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(op).To(Equal(controllerutil.OperationResultCreated))
			Expect(hash).NotTo(BeEmpty())

			got := &corev1.Secret{}
			Expect(cClient.Get(ctx, types.NamespacedName{
				Name:      "test-secret",
				Namespace: namespace,
			}, got)).To(Succeed())
			Expect(got.Data).To(HaveKeyWithValue("password", []byte("initial-password")))
		})

		It("preserves existing data on subsequent calls", func() {
			owner := th.CreateNamespace("secret-preserve-owner")

			// First call: create with initial data
			s1 := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "preserve-secret",
					Namespace: namespace,
				},
				Type: corev1.SecretTypeOpaque,
				Data: map[string][]byte{
					"password": []byte("original-password"),
				},
			}
			_, _, err := secret.CreateOrPatchSecretPreserve(ctx, h, owner, s1, false)
			Expect(err).NotTo(HaveOccurred())

			// Second call: try to update with different data
			s2 := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "preserve-secret",
					Namespace: namespace,
				},
				Type: corev1.SecretTypeOpaque,
				Data: map[string][]byte{
					"password": []byte("new-password-should-be-ignored"),
				},
			}
			_, op, err := secret.CreateOrPatchSecretPreserve(ctx, h, owner, s2, false)
			Expect(err).NotTo(HaveOccurred())
			// No change since data was preserved
			Expect(op).To(Equal(controllerutil.OperationResultNone))

			// Verify the original data is preserved
			got := &corev1.Secret{}
			Expect(cClient.Get(ctx, types.NamespacedName{
				Name:      "preserve-secret",
				Namespace: namespace,
			}, got)).To(Succeed())
			Expect(got.Data).To(HaveKeyWithValue("password", []byte("original-password")))
		})

		It("adds new labels while preserving existing ones and data", func() {
			owner := th.CreateNamespace("secret-labels-owner")

			s1 := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "labels-secret",
					Namespace: namespace,
					Labels:    map[string]string{"version": "v1"},
				},
				Data: map[string][]byte{"key": []byte("value")},
			}
			_, _, err := secret.CreateOrPatchSecretPreserve(ctx, h, owner, s1, false)
			Expect(err).NotTo(HaveOccurred())

			s2 := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "labels-secret",
					Namespace: namespace,
					Labels:    map[string]string{"new-label": "yes"},
				},
				Data: map[string][]byte{"key": []byte("new-value")},
			}
			_, _, err = secret.CreateOrPatchSecretPreserve(ctx, h, owner, s2, false)
			Expect(err).NotTo(HaveOccurred())

			got := &corev1.Secret{}
			Expect(cClient.Get(ctx, types.NamespacedName{
				Name:      "labels-secret",
				Namespace: namespace,
			}, got)).To(Succeed())

			// Existing labels preserved, new labels added (MergeStringMaps behavior)
			Expect(got.Labels).To(HaveKeyWithValue("version", "v1"))
			Expect(got.Labels).To(HaveKeyWithValue("new-label", "yes"))
			// Data should be preserved (original value)
			Expect(got.Data).To(HaveKeyWithValue("key", []byte("value")))
		})
	})
})
