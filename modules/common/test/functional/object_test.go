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
	. "github.com/onsi/ginkgo/v2" // nolint:revive
	. "github.com/onsi/gomega"    // nolint:revive
	"github.com/openstack-k8s-operators/lib-common/modules/common/object"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("object package", func() {
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

	It("now new owner gets added when adding same ownerref", func() {
		cmName := types.NamespacedName{
			Namespace: namespace,
			Name:      "test-cm",
		}

		cm := th.CreateConfigMap(cmName, map[string]interface{}{})

		err := object.EnsureOwnerRef(th.Ctx, h, cm, cm)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(object.CheckOwnerRefExist(cm.GetUID(), cm.GetOwnerReferences())).To(BeTrue())
		Expect(cm.GetOwnerReferences()).To(HaveLen(1))
	})

	It("adds an additional owner to the ownerref list", func() {
		// create owner obj
		owner := types.NamespacedName{
			Namespace: namespace,
			Name:      "test-owner",
		}
		ownerCM := th.CreateConfigMap(owner, map[string]interface{}{})

		// create target obj we add the owner ref to
		cmName := types.NamespacedName{
			Namespace: namespace,
			Name:      "test-cm",
		}
		cm := th.CreateConfigMap(cmName, map[string]interface{}{})

		err := object.EnsureOwnerRef(th.Ctx, h, ownerCM, cm)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(object.CheckOwnerRefExist(ownerCM.GetUID(), cm.GetOwnerReferences())).To(BeTrue())
	})

	When("checking if owner service is ready", func() {
		It("returns true when object has no controller owner", func() {
			cmName := types.NamespacedName{
				Namespace: namespace,
				Name:      "test-cm-no-owner",
			}
			cm := th.CreateConfigMap(cmName, map[string]interface{}{})

			ready, err := object.IsOwnerServiceReady(th.Ctx, h, cm)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ready).To(BeTrue())
		})

		It("returns true when controller owner is deleted", func() {
			// Create a ConfigMap with an owner reference to a non-existent ConfigMap
			rawCM := map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "test-cm-deleted-owner",
					"namespace": namespace,
					"ownerReferences": []interface{}{
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"name":       "non-existent-configmap",
							"uid":        "11111111-1111-1111-1111-111111111111",
							"controller": true,
						},
					},
				},
			}
			cm := th.CreateUnstructured(rawCM)

			ready, err := object.IsOwnerServiceReady(th.Ctx, h, cm)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ready).To(BeTrue())
		})

		It("returns false when controller owner exists but has no status", func() {
			// Create owner ConfigMap (no status.conditions)
			ownerCM := th.CreateConfigMap(types.NamespacedName{
				Namespace: namespace,
				Name:      "owner-cm-no-status",
			}, map[string]interface{}{})

			// Create child ConfigMap with owner reference
			rawCM := map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "child-cm-no-status",
					"namespace": namespace,
					"ownerReferences": []interface{}{
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"name":       ownerCM.GetName(),
							"uid":        string(ownerCM.GetUID()),
							"controller": true,
						},
					},
				},
			}
			cm := th.CreateUnstructured(rawCM)

			ready, err := object.IsOwnerServiceReady(th.Ctx, h, cm)
			Expect(err).ShouldNot(HaveOccurred())
			// ConfigMaps don't have status.conditions, so should return false
			Expect(ready).To(BeFalse())
		})

		It("returns false when controller owner has Ready condition with status True but no observedGeneration (Pod case)", func() {
			// Create a Pod as the owner (Pods do not have observedGeneration, so the check will return false)
			th.CreatePod(types.NamespacedName{
				Namespace: namespace,
				Name:      "owner-pod-ready",
			}, map[string]string{}, map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":  "test",
						"image": "test:latest",
					},
				},
			})

			// Update the Pod status to include a Ready condition
			Eventually(func(g Gomega) {
				pod := th.GetPod(types.NamespacedName{
					Namespace: namespace,
					Name:      "owner-pod-ready",
				})
				// Manually set status with Ready condition
				pod.Status.Conditions = []corev1.PodCondition{
					{
						Type:   "Ready",
						Status: corev1.ConditionTrue,
					},
				}
				err := th.K8sClient.Status().Update(th.Ctx, pod)
				g.Expect(err).ShouldNot(HaveOccurred())
			}, th.Timeout, th.Interval).Should(Succeed())

			// Get the updated pod to get its UID
			pod := th.GetPod(types.NamespacedName{
				Namespace: namespace,
				Name:      "owner-pod-ready",
			})

			// Create a ConfigMap owned by the Pod
			rawCM := map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "child-cm-ready-owner",
					"namespace": namespace,
					"ownerReferences": []interface{}{
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Pod",
							"name":       pod.Name,
							"uid":        string(pod.GetUID()),
							"controller": true,
						},
					},
				},
			}
			cm := th.CreateUnstructured(rawCM)

			ready, err := object.IsOwnerServiceReady(th.Ctx, h, cm)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ready).To(BeFalse())
		})

		It("returns false when controller owner has Ready condition with status False", func() {
			// Create a Pod as the owner
			th.CreatePod(types.NamespacedName{
				Namespace: namespace,
				Name:      "owner-pod-not-ready",
			}, map[string]string{}, map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":  "test",
						"image": "test:latest",
					},
				},
			})

			// Update the Pod status to include a Ready=False condition
			Eventually(func(g Gomega) {
				pod := th.GetPod(types.NamespacedName{
					Namespace: namespace,
					Name:      "owner-pod-not-ready",
				})
				pod.Status.Conditions = []corev1.PodCondition{
					{
						Type:   "Ready",
						Status: corev1.ConditionFalse,
					},
				}
				err := th.K8sClient.Status().Update(th.Ctx, pod)
				g.Expect(err).ShouldNot(HaveOccurred())
			}, th.Timeout, th.Interval).Should(Succeed())

			// Get the updated pod
			pod := th.GetPod(types.NamespacedName{
				Namespace: namespace,
				Name:      "owner-pod-not-ready",
			})

			// Create a ConfigMap owned by the Pod
			rawCM := map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "child-cm-not-ready-owner",
					"namespace": namespace,
					"ownerReferences": []interface{}{
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Pod",
							"name":       pod.Name,
							"uid":        string(pod.GetUID()),
							"controller": true,
						},
					},
				},
			}
			cm := th.CreateUnstructured(rawCM)

			ready, err := object.IsOwnerServiceReady(th.Ctx, h, cm)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ready).To(BeFalse())
		})
	})
})
