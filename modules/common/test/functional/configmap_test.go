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
	"github.com/openstack-k8s-operators/lib-common/modules/common/configmap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("ConfigMap helpers", func() {
	var namespace string

	BeforeEach(func() {
		namespace = uuid.New().String()
		th.CreateNamespace(namespace)
	})

	When("CreateOrPatchRawConfigMap is called", func() {
		It("creates a ConfigMap from raw data", func() {
			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cm",
					Namespace: namespace,
					Labels: map[string]string{
						"app": "test",
					},
				},
				Data: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			}
			hash, op, err := configmap.CreateOrPatchRawConfigMap(
				ctx, h, th.CreateNamespace("cm-owner"), cm, false,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(op).To(Equal(controllerutil.OperationResultCreated))
			Expect(hash).NotTo(BeEmpty())

			got := &corev1.ConfigMap{}
			Expect(cClient.Get(ctx, types.NamespacedName{
				Name:      "test-cm",
				Namespace: namespace,
			}, got)).To(Succeed())

			Expect(got.Data).To(HaveKeyWithValue("key1", "value1"))
			Expect(got.Data).To(HaveKeyWithValue("key2", "value2"))
			Expect(got.Labels).To(HaveKeyWithValue("app", "test"))
		})

		It("patches an existing ConfigMap with new data", func() {
			owner := th.CreateNamespace("cm-patch-owner")

			cm1 := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "patch-cm",
					Namespace: namespace,
				},
				Data: map[string]string{"key1": "old"},
			}
			_, _, err := configmap.CreateOrPatchRawConfigMap(
				ctx, h, owner, cm1, false,
			)
			Expect(err).NotTo(HaveOccurred())

			cm2 := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "patch-cm",
					Namespace: namespace,
				},
				Data: map[string]string{"key1": "new", "key2": "added"},
			}
			hash2, op, err := configmap.CreateOrPatchRawConfigMap(
				ctx, h, owner, cm2, false,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(op).To(Equal(controllerutil.OperationResultUpdated))
			Expect(hash2).NotTo(BeEmpty())

			got := &corev1.ConfigMap{}
			Expect(cClient.Get(ctx, types.NamespacedName{
				Name:      "patch-cm",
				Namespace: namespace,
			}, got)).To(Succeed())

			Expect(got.Data).To(HaveKeyWithValue("key1", "new"))
			Expect(got.Data).To(HaveKeyWithValue("key2", "added"))
		})

		It("returns consistent hash for same data", func() {
			owner := th.CreateNamespace("cm-hash-owner")
			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hash-cm",
					Namespace: namespace,
				},
				Data: map[string]string{"k": "v"},
			}

			hash1, _, err := configmap.CreateOrPatchRawConfigMap(
				ctx, h, owner, cm, false,
			)
			Expect(err).NotTo(HaveOccurred())

			hash2, _, err := configmap.CreateOrPatchRawConfigMap(
				ctx, h, owner, cm, false,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(hash2).To(Equal(hash1))
		})
	})
})
