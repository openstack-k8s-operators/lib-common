/*
Copyright 2025 Red Hat

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

package helpers

import (
	"github.com/onsi/gomega"

	policyv1 "k8s.io/api/policy/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// GetPodDisruptionBudget retrieves a PodDisruptionBudget resource.
//
// Example usage:
//
//	th.GetPodDisruptionBudget(types.NamespacedName{Name: "test-pdb", Namespace: "test-namespace"})
func (tc *TestHelper) GetPodDisruptionBudget(name types.NamespacedName) *policyv1.PodDisruptionBudget {
	instance := &policyv1.PodDisruptionBudget{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, instance)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	return instance
}

// CreatePodDisruptionBudget creates a new k8s PodDisruptionBudget resource with provided data.
//
// Example usage:
//
//	pdb := th.CreatePodDisruptionBudget(types.NamespacedName{Name: "test-pdb", Namespace: "test-namespace"}, map[string]string{}, policyv1.PodDisruptionBudgetSpec{...})
func (tc *TestHelper) CreatePodDisruptionBudget(name types.NamespacedName, labels map[string]string, pdbSpec policyv1.PodDisruptionBudgetSpec) *policyv1.PodDisruptionBudget {
	pdb := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
			Labels:    labels,
		},
		Spec: pdbSpec,
	}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Create(tc.Ctx, pdb)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	return pdb
}

// AssertPodDisruptionBudgetExists - asserts the existence of a PodDisruptionBudget resource in the Kubernetes cluster.
//
// Example usage:
//
//	th.AssertPodDisruptionBudgetExists(types.NamespacedName{Name: "app-pdb", Namespace: namespace})
func (tc *TestHelper) AssertPodDisruptionBudgetExists(name types.NamespacedName) *policyv1.PodDisruptionBudget {
	instance := &policyv1.PodDisruptionBudget{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, instance)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
	return instance
}

// DeletePodDisruptionBudget - deletes a PodDisruptionBudget resource from the Kubernetes cluster.
//
// Example usage:
//
//	th.DeletePodDisruptionBudget(types.NamespacedName{Name: "test-pdb", Namespace: "test-namespace"})
func (tc *TestHelper) DeletePodDisruptionBudget(name types.NamespacedName) {
	instance := &policyv1.PodDisruptionBudget{}

	gomega.Eventually(func(g gomega.Gomega) {
		name := types.NamespacedName{Name: name.Name, Namespace: name.Namespace}
		err := tc.K8sClient.Get(tc.Ctx, name, instance)
		// if it is already gone that is OK
		if k8s_errors.IsNotFound(err) {
			return
		}
		g.Expect(err).ShouldNot(gomega.HaveOccurred())

		g.Expect(tc.K8sClient.Delete(tc.Ctx, instance)).Should(gomega.Succeed())

		err = tc.K8sClient.Get(tc.Ctx, name, instance)
		g.Expect(k8s_errors.IsNotFound(err)).To(gomega.BeTrue())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

}

// AssertPodDisruptionBudgetDoesNotExist ensures the PodDisruptionBudget resource does not exist in a k8s cluster.
func (tc *TestHelper) AssertPodDisruptionBudgetDoesNotExist(name types.NamespacedName) {
	instance := &policyv1.PodDisruptionBudget{}
	gomega.Consistently(func(g gomega.Gomega) {
		err := tc.K8sClient.Get(tc.Ctx, name, instance)
		g.Expect(k8s_errors.IsNotFound(err)).To(gomega.BeTrue())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
}

// SimulatePodDisruptionBudgetReady simulates a PodDisruptionBudget being ready by updating its status.
//
// Example usage:
//
//	th.SimulatePodDisruptionBudgetReady(types.NamespacedName{Name: "test-pdb", Namespace: "test-namespace"})
func (tc *TestHelper) SimulatePodDisruptionBudgetReady(name types.NamespacedName) {
	gomega.Eventually(func(g gomega.Gomega) {
		pdb := tc.GetPodDisruptionBudget(name)
		pdb.Status.ObservedGeneration = pdb.Generation
		pdb.Status.ExpectedPods = 3
		pdb.Status.CurrentHealthy = 2
		pdb.Status.DesiredHealthy = 2
		pdb.Status.DisruptionsAllowed = 1
		g.Expect(tc.K8sClient.Status().Update(tc.Ctx, pdb)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
}
