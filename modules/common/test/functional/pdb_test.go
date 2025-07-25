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
package functional

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openstack-k8s-operators/lib-common/modules/common/pdb"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getExamplePDBWithMinAvailable(namespace string, minAvailable intstr.IntOrString) *policyv1.PodDisruptionBudget {
	return &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pdb",
			Namespace: namespace,
			Labels: map[string]string{
				"app":     "test-app",
				"replace": "old-value",
			},
			Annotations: map[string]string{
				"description": "test pdb",
				"replace":     "old-value",
			},
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &minAvailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test-app",
				},
			},
		},
	}
}

func getExamplePDBWithMaxUnavailable(namespace string, maxUnavailable intstr.IntOrString) *policyv1.PodDisruptionBudget {
	return &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pdb-max",
			Namespace: namespace,
			Labels: map[string]string{
				"app":  "test-app",
				"type": "maxunavailable",
			},
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MaxUnavailable: &maxUnavailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test-app",
				},
			},
		},
	}
}

var _ = Describe("pdb package", func() {
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

	It("creates PDB with minAvailable", func() {
		minAvailable := intstr.FromInt(2)
		p := pdb.NewPDB(
			getExamplePDBWithMinAvailable(namespace, minAvailable),
			timeout,
		)

		_, err := p.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())

		pdbResource := th.AssertPodDisruptionBudgetExists(types.NamespacedName{Namespace: namespace, Name: "test-pdb"})
		Expect(pdbResource.Spec.MinAvailable).ToNot(BeNil())
		Expect(pdbResource.Spec.MinAvailable.IntVal).To(Equal(int32(2)))
		Expect(pdbResource.Spec.MaxUnavailable).To(BeNil())
		Expect(pdbResource.Spec.Selector.MatchLabels["app"]).To(Equal("test-app"))
		Expect(pdbResource.Labels["app"]).To(Equal("test-app"))
		Expect(pdbResource.Labels["replace"]).To(Equal("old-value"))
		Expect(pdbResource.Annotations["description"]).To(Equal("test pdb"))
		Expect(pdbResource.Annotations["replace"]).To(Equal("old-value"))

		// Test Getters
		retrievedPDB := p.GetPDB()
		Expect(retrievedPDB.Name).To(Equal("test-pdb"))
		Expect(retrievedPDB.Spec.MinAvailable.IntVal).To(Equal(int32(2)))

		// GetPDBWithName()
		pdbFromGet, err := pdb.GetPDBWithName(th.Ctx, h, "test-pdb", namespace)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(pdbFromGet.Name).To(Equal("test-pdb"))

		// Test IsReady() function - PDB is not ready initially
		Expect(pdb.IsReady(*pdbResource)).To(BeFalse())

		// Simulate PDB becoming ready
		th.SimulatePodDisruptionBudgetReady(types.NamespacedName{Namespace: namespace, Name: "test-pdb"})

		// Verify PDB is now ready
		pdbResource = th.AssertPodDisruptionBudgetExists(types.NamespacedName{Namespace: namespace, Name: "test-pdb"})
		Expect(pdb.IsReady(*pdbResource)).To(BeTrue())

		// Delete method
		err = p.Delete(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		th.AssertPodDisruptionBudgetDoesNotExist(types.NamespacedName{Namespace: namespace, Name: "test-pdb"})
	})

	It("creates PDB with maxUnavailable", func() {
		maxUnavailable := intstr.FromInt(1)
		p := pdb.NewPDB(
			getExamplePDBWithMaxUnavailable(namespace, maxUnavailable),
			timeout,
		)

		_, err := p.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())

		pdbResource := th.AssertPodDisruptionBudgetExists(types.NamespacedName{Namespace: namespace, Name: "test-pdb-max"})
		Expect(pdbResource.Spec.MaxUnavailable).ToNot(BeNil())
		Expect(pdbResource.Spec.MaxUnavailable.IntVal).To(Equal(int32(1)))
		Expect(pdbResource.Spec.MinAvailable).To(BeNil())
		Expect(pdbResource.Spec.Selector.MatchLabels["app"]).To(Equal("test-app"))
		Expect(pdbResource.Labels["type"]).To(Equal("maxunavailable"))
	})

	It("creates PDB with percentage values", func() {
		minAvailablePercent := intstr.FromString("75%")
		pdbResource := pdb.MinAvailablePodDisruptionBudget("pdb-percent", namespace, minAvailablePercent, map[string]string{
			"app": "percentage-test",
		})

		p := pdb.NewPDB(pdbResource, timeout)

		_, err := p.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())

		createdPDB := th.AssertPodDisruptionBudgetExists(types.NamespacedName{Namespace: namespace, Name: "pdb-percent"})
		Expect(createdPDB.Spec.MinAvailable).ToNot(BeNil())
		Expect(createdPDB.Spec.MinAvailable.StrVal).To(Equal("75%"))
		Expect(createdPDB.Spec.MinAvailable.Type).To(Equal(intstr.String))
		Expect(createdPDB.Spec.MaxUnavailable).To(BeNil())
		Expect(createdPDB.Spec.Selector.MatchLabels["app"]).To(Equal("percentage-test"))
	})

	It("tests MaxUnavailablePodDisruptionBudget helper function", func() {
		maxUnavailable := intstr.FromInt(2)
		labelSelector := map[string]string{
			"service": "web-frontend",
			"tier":    "frontend",
		}

		pdbResource := pdb.MaxUnavailablePodDisruptionBudget("frontend-pdb", namespace, maxUnavailable, labelSelector)
		p := pdb.NewPDB(pdbResource, timeout)

		_, err := p.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())

		createdPDB := th.AssertPodDisruptionBudgetExists(types.NamespacedName{Namespace: namespace, Name: "frontend-pdb"})
		Expect(createdPDB.Spec.MaxUnavailable).ToNot(BeNil())
		Expect(createdPDB.Spec.MaxUnavailable.IntVal).To(Equal(int32(2)))
		Expect(createdPDB.Spec.MinAvailable).To(BeNil())
		Expect(createdPDB.Spec.Selector.MatchLabels["service"]).To(Equal("web-frontend"))
		Expect(createdPDB.Spec.Selector.MatchLabels["tier"]).To(Equal("frontend"))
	})

	It("tests MinAvailablePodDisruptionBudget helper function", func() {
		minAvailable := intstr.FromString("60%")
		labelSelector := map[string]string{
			"service": "database",
			"tier":    "backend",
		}

		pdbResource := pdb.MinAvailablePodDisruptionBudget("backend-pdb", namespace, minAvailable, labelSelector)
		p := pdb.NewPDB(pdbResource, timeout)

		_, err := p.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())

		createdPDB := th.AssertPodDisruptionBudgetExists(types.NamespacedName{Namespace: namespace, Name: "backend-pdb"})
		Expect(createdPDB.Spec.MinAvailable).ToNot(BeNil())
		Expect(createdPDB.Spec.MinAvailable.StrVal).To(Equal("60%"))
		Expect(createdPDB.Spec.MinAvailable.Type).To(Equal(intstr.String))
		Expect(createdPDB.Spec.MaxUnavailable).To(BeNil())
		Expect(createdPDB.Spec.Selector.MatchLabels["service"]).To(Equal("database"))
		Expect(createdPDB.Spec.Selector.MatchLabels["tier"]).To(Equal("backend"))
	})

	It("creates PDB with UnhealthyPodEvictionPolicy", func() {
		minAvailable := intstr.FromInt(3)
		pdbSpec := policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &minAvailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "policy-test",
				},
			},
			UnhealthyPodEvictionPolicy: ptr.To(policyv1.AlwaysAllow),
		}

		pdbResource := &policyv1.PodDisruptionBudget{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "policy-pdb",
				Namespace: namespace,
			},
			Spec: pdbSpec,
		}

		p := pdb.NewPDB(pdbResource, timeout)

		_, err := p.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())

		createdPDB := th.AssertPodDisruptionBudgetExists(types.NamespacedName{Namespace: namespace, Name: "policy-pdb"})
		Expect(createdPDB.Spec.UnhealthyPodEvictionPolicy).ToNot(BeNil())
		Expect(*createdPDB.Spec.UnhealthyPodEvictionPolicy).To(Equal(policyv1.AlwaysAllow))
		Expect(createdPDB.Spec.MinAvailable.IntVal).To(Equal(int32(3)))
		Expect(createdPDB.Spec.Selector.MatchLabels["app"]).To(Equal("policy-test"))
	})

	It("updates existing PDB on subsequent CreateOrPatch calls", func() {
		minAvailable := intstr.FromInt(2)
		initialPDB := getExamplePDBWithMinAvailable(namespace, minAvailable)
		p := pdb.NewPDB(initialPDB, timeout)

		// Create the PDB
		_, err := p.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())

		// Verify initial creation
		pdbResource := th.AssertPodDisruptionBudgetExists(types.NamespacedName{Namespace: namespace, Name: "test-pdb"})
		Expect(pdbResource.Spec.MinAvailable.IntVal).To(Equal(int32(2)))

		// Update the PDB spec
		newMinAvailable := intstr.FromInt(4)
		updatedPDB := getExamplePDBWithMinAvailable(namespace, newMinAvailable)
		updatedPDB.Labels["new-label"] = "added"
		updatedPDB.Annotations["new-annotation"] = "added"

		p2 := pdb.NewPDB(updatedPDB, timeout)

		// Apply the update
		_, err = p2.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())

		// Verify the update
		updatedResource := th.AssertPodDisruptionBudgetExists(types.NamespacedName{Namespace: namespace, Name: "test-pdb"})
		Expect(updatedResource.Spec.MinAvailable.IntVal).To(Equal(int32(4)))
		Expect(updatedResource.Labels["new-label"]).To(Equal("added"))
		Expect(updatedResource.Annotations["new-annotation"]).To(Equal("added"))
		// Original labels and annotations should still be present
		Expect(updatedResource.Labels["app"]).To(Equal("test-app"))
		Expect(updatedResource.Annotations["description"]).To(Equal("test pdb"))
	})

	It("handles PDB creation with complex selectors", func() {
		labelSelector := map[string]string{
			"app":         "complex-app",
			"version":     "v2.0",
			"environment": "production",
			"tier":        "frontend",
		}

		maxUnavailable := intstr.FromString("25%")
		pdbResource := pdb.MaxUnavailablePodDisruptionBudget("complex-pdb", namespace, maxUnavailable, labelSelector)

		p := pdb.NewPDB(pdbResource, timeout)

		_, err := p.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())

		createdPDB := th.AssertPodDisruptionBudgetExists(types.NamespacedName{Namespace: namespace, Name: "complex-pdb"})
		Expect(createdPDB.Spec.MaxUnavailable.StrVal).To(Equal("25%"))
		Expect(createdPDB.Spec.Selector.MatchLabels).To(Equal(labelSelector))

		// Verify all label selectors are preserved
		for key, value := range labelSelector {
			Expect(createdPDB.Spec.Selector.MatchLabels[key]).To(Equal(value))
		}
	})

	It("tests DeletePDBWithName function", func() {
		// First create a PDB using helper function
		minAvailable := intstr.FromInt(3)
		labelSelector := map[string]string{
			"app": "delete-test",
		}

		pdbResource := pdb.MinAvailablePodDisruptionBudget("delete-test-pdb", namespace, minAvailable, labelSelector)
		p := pdb.NewPDB(pdbResource, timeout)

		// Create the PDB
		_, err := p.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())

		// Verify it exists
		createdPDB := th.AssertPodDisruptionBudgetExists(types.NamespacedName{Namespace: namespace, Name: "delete-test-pdb"})
		Expect(createdPDB.Spec.MinAvailable.IntVal).To(Equal(int32(3)))
		Expect(createdPDB.Spec.Selector.MatchLabels["app"]).To(Equal("delete-test"))

		// Test DeletePDBWithName function
		err = pdb.DeletePDBWithName(ctx, h, "delete-test-pdb", namespace)
		Expect(err).ShouldNot(HaveOccurred())

		// Verify the PDB is deleted
		th.AssertPodDisruptionBudgetDoesNotExist(types.NamespacedName{Namespace: namespace, Name: "delete-test-pdb"})

		// Test deleting non-existent PDB (should not return error)
		err = pdb.DeletePDBWithName(ctx, h, "non-existent-pdb", namespace)
		Expect(err).ShouldNot(HaveOccurred())
	})
})
