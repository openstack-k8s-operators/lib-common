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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("object package - observedGeneration tests", func() {
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

	When("checking if owner service is ready with observedGeneration", func() {
		It("returns false when controller owner is ready but observedGeneration does not match generation", func() {
			// Create a Deployment as the owner
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "owner-deployment-not-reconciled",
					Namespace: namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: func(i int32) *int32 { return &i }(1),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "test"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "test:latest",
								},
							},
						},
					},
				},
			}
			Expect(th.K8sClient.Create(th.Ctx, deployment)).Should(Succeed())

			// Update deployment to increment generation
			Eventually(func(g Gomega) {
				dep := th.GetDeployment(types.NamespacedName{
					Namespace: namespace,
					Name:      "owner-deployment-not-reconciled",
				})
				dep.Spec.Replicas = func(i int32) *int32 { return &i }(2)
				g.Expect(th.K8sClient.Update(th.Ctx, dep)).Should(Succeed())
			}, th.Timeout, th.Interval).Should(Succeed())

			// Simulate deployment with Ready condition but observedGeneration < generation
			Eventually(func(g Gomega) {
				dep := th.GetDeployment(types.NamespacedName{
					Namespace: namespace,
					Name:      "owner-deployment-not-reconciled",
				})

				// Set status with Ready-like condition (Available) and observedGeneration behind
				dep.Status.ObservedGeneration = dep.Generation - 1
				dep.Status.Conditions = []appsv1.DeploymentCondition{
					{
						Type:   appsv1.DeploymentAvailable,
						Status: corev1.ConditionTrue,
					},
				}
				// Also add a Ready condition for completeness
				dep.Status.Conditions = append(dep.Status.Conditions, appsv1.DeploymentCondition{
					Type:   "Ready",
					Status: corev1.ConditionTrue,
				})

				g.Expect(th.K8sClient.Status().Update(th.Ctx, dep)).Should(Succeed())
			}, th.Timeout, th.Interval).Should(Succeed())

			// Get the deployment to get its UID
			dep := th.GetDeployment(types.NamespacedName{
				Namespace: namespace,
				Name:      "owner-deployment-not-reconciled",
			})

			// Create a child ConfigMap owned by the Deployment
			rawCM := map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "child-cm-not-reconciled",
					"namespace": namespace,
					"ownerReferences": []interface{}{
						map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "Deployment",
							"name":       dep.Name,
							"uid":        string(dep.GetUID()),
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

		It("returns true when controller owner is ready and observedGeneration matches generation", func() {
			// Create a Deployment as the owner
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "owner-deployment-reconciled",
					Namespace: namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: func(i int32) *int32 { return &i }(1),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "test"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "test:latest",
								},
							},
						},
					},
				},
			}
			Expect(th.K8sClient.Create(th.Ctx, deployment)).Should(Succeed())

			// Simulate deployment with Ready condition and observedGeneration = generation
			Eventually(func(g Gomega) {
				dep := th.GetDeployment(types.NamespacedName{
					Namespace: namespace,
					Name:      "owner-deployment-reconciled",
				})

				// Set status with Ready condition and observedGeneration matching generation
				dep.Status.ObservedGeneration = dep.Generation
				dep.Status.Conditions = []appsv1.DeploymentCondition{
					{
						Type:   "Ready",
						Status: corev1.ConditionTrue,
					},
				}

				g.Expect(th.K8sClient.Status().Update(th.Ctx, dep)).Should(Succeed())
			}, th.Timeout, th.Interval).Should(Succeed())

			// Get the deployment to get its UID
			dep := th.GetDeployment(types.NamespacedName{
				Namespace: namespace,
				Name:      "owner-deployment-reconciled",
			})

			// Create a child ConfigMap owned by the Deployment
			rawCM := map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "child-cm-reconciled",
					"namespace": namespace,
					"ownerReferences": []interface{}{
						map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "Deployment",
							"name":       dep.Name,
							"uid":        string(dep.GetUID()),
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

		It("returns false when controller owner is ready but has no observedGeneration field", func() {
			// Create a Deployment as the owner
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "owner-deployment-no-observed",
					Namespace: namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: func(i int32) *int32 { return &i }(1),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "test"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "test:latest",
								},
							},
						},
					},
				},
			}
			Expect(th.K8sClient.Create(th.Ctx, deployment)).Should(Succeed())

			// Set status with Ready condition but explicitly set observedGeneration to 0 (unset)
			Eventually(func(g Gomega) {
				dep := th.GetDeployment(types.NamespacedName{
					Namespace: namespace,
					Name:      "owner-deployment-no-observed",
				})

				// Set status with Ready condition but no observedGeneration (leave it at 0)
				dep.Status.ObservedGeneration = 0
				dep.Status.Conditions = []appsv1.DeploymentCondition{
					{
						Type:   "Ready",
						Status: corev1.ConditionTrue,
					},
				}

				g.Expect(th.K8sClient.Status().Update(th.Ctx, dep)).Should(Succeed())
			}, th.Timeout, th.Interval).Should(Succeed())

			// Get the deployment to get its UID
			dep := th.GetDeployment(types.NamespacedName{
				Namespace: namespace,
				Name:      "owner-deployment-no-observed",
			})

			// Create a child ConfigMap owned by the Deployment
			rawCM := map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "child-cm-no-observed",
					"namespace": namespace,
					"ownerReferences": []interface{}{
						map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "Deployment",
							"name":       dep.Name,
							"uid":        string(dep.GetUID()),
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
