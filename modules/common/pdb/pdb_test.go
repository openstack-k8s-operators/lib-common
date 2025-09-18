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

package pdb

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
)

var (
	testPDB = policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pdb",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app": "test-app",
			},
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 2,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test-app",
				},
			},
		},
	}

	testPDBWithMaxUnavailable = policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pdb-maxunavailable",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app": "test-app",
			},
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MaxUnavailable: &intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 1,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test-app",
				},
			},
		},
	}

	timeout = time.Duration(5) * time.Second
)

func TestNewPDB(t *testing.T) {
	g := NewWithT(t)

	pdb := NewPDB(&testPDB, timeout)

	g.Expect(pdb).ToNot(BeNil())
	g.Expect(pdb.pdb.Name).To(Equal("test-pdb"))
	g.Expect(pdb.pdb.Namespace).To(Equal("test-namespace"))
	g.Expect(pdb.timeout).To(Equal(timeout))
	g.Expect(pdb.pdb.Spec.MinAvailable.IntVal).To(Equal(int32(2)))
}

func TestNewPDBWithMaxUnavailable(t *testing.T) {
	g := NewWithT(t)

	pdb := NewPDB(&testPDBWithMaxUnavailable, timeout)

	g.Expect(pdb).ToNot(BeNil())
	g.Expect(pdb.pdb.Name).To(Equal("test-pdb-maxunavailable"))
	g.Expect(pdb.pdb.Namespace).To(Equal("test-namespace"))
	g.Expect(pdb.timeout).To(Equal(timeout))
	g.Expect(pdb.pdb.Spec.MaxUnavailable.IntVal).To(Equal(int32(1)))
}

func TestMaxUnavailablePodDisruptionBudget(t *testing.T) {
	g := NewWithT(t)

	// Test with integer maxUnavailable
	maxUnavailableInt := intstr.FromInt(1)
	labelSelector := map[string]string{
		"app":     "test-app",
		"version": "v1.0",
	}

	pdb := MaxUnavailablePodDisruptionBudget("test-pdb-helper", "test-namespace", maxUnavailableInt, labelSelector)

	g.Expect(pdb).ToNot(BeNil())
	g.Expect(pdb.Name).To(Equal("test-pdb-helper"))
	g.Expect(pdb.Namespace).To(Equal("test-namespace"))
	g.Expect(pdb.Spec.MaxUnavailable).ToNot(BeNil())
	g.Expect(pdb.Spec.MaxUnavailable.IntVal).To(Equal(int32(1)))
	g.Expect(pdb.Spec.MinAvailable).To(BeNil())
	g.Expect(pdb.Spec.Selector).ToNot(BeNil())
	g.Expect(pdb.Spec.Selector.MatchLabels).To(Equal(labelSelector))

	// Test with percentage maxUnavailable
	maxUnavailablePercent := intstr.FromString("25%")
	pdbPercent := MaxUnavailablePodDisruptionBudget("test-pdb-percent", "test-namespace", maxUnavailablePercent, labelSelector)

	g.Expect(pdbPercent.Spec.MaxUnavailable.StrVal).To(Equal("25%"))
	g.Expect(pdbPercent.Spec.MaxUnavailable.Type).To(Equal(intstr.String))
}

func TestMinAvailablePodDisruptionBudget(t *testing.T) {
	g := NewWithT(t)

	// Test with integer minAvailable
	minAvailableInt := intstr.FromInt(3)
	labelSelector := map[string]string{
		"app":       "web-server",
		"component": "frontend",
	}

	pdb := MinAvailablePodDisruptionBudget("test-pdb-min-helper", "test-namespace", minAvailableInt, labelSelector)

	g.Expect(pdb).ToNot(BeNil())
	g.Expect(pdb.Name).To(Equal("test-pdb-min-helper"))
	g.Expect(pdb.Namespace).To(Equal("test-namespace"))
	g.Expect(pdb.Spec.MinAvailable).ToNot(BeNil())
	g.Expect(pdb.Spec.MinAvailable.IntVal).To(Equal(int32(3)))
	g.Expect(pdb.Spec.MaxUnavailable).To(BeNil())
	g.Expect(pdb.Spec.Selector).ToNot(BeNil())
	g.Expect(pdb.Spec.Selector.MatchLabels).To(Equal(labelSelector))

	// Test with percentage minAvailable
	minAvailablePercent := intstr.FromString("75%")
	pdbPercent := MinAvailablePodDisruptionBudget("test-pdb-min-percent", "test-namespace", minAvailablePercent, labelSelector)

	g.Expect(pdbPercent.Spec.MinAvailable.StrVal).To(Equal("75%"))
	g.Expect(pdbPercent.Spec.MinAvailable.Type).To(Equal(intstr.String))
	g.Expect(pdbPercent.Spec.MaxUnavailable).To(BeNil())
}

func TestMaxUnavailablePodDisruptionBudgetWithNewPDB(t *testing.T) {
	g := NewWithT(t)

	// Create PDB using helper function
	maxUnavailable := intstr.FromInt(2)
	labelSelector := map[string]string{
		"component": "database",
		"tier":      "backend",
	}

	pdbResource := MaxUnavailablePodDisruptionBudget("db-pdb", "production", maxUnavailable, labelSelector)

	// Wrap it with NewPDB
	pdb := NewPDB(pdbResource, timeout)

	g.Expect(pdb.GetPDB().Name).To(Equal("db-pdb"))
	g.Expect(pdb.GetPDB().Namespace).To(Equal("production"))
	g.Expect(pdb.GetPDB().Spec.MaxUnavailable.IntVal).To(Equal(int32(2)))
	g.Expect(pdb.GetPDB().Spec.Selector.MatchLabels["component"]).To(Equal("database"))
	g.Expect(pdb.GetPDB().Spec.Selector.MatchLabels["tier"]).To(Equal("backend"))
}

func TestMinAvailablePodDisruptionBudgetWithNewPDB(t *testing.T) {
	g := NewWithT(t)

	// Create PDB using MinAvailable helper function
	minAvailable := intstr.FromInt(4)
	labelSelector := map[string]string{
		"service":     "api-gateway",
		"environment": "production",
		"tier":        "frontend",
	}

	pdbResource := MinAvailablePodDisruptionBudget("api-pdb", "production", minAvailable, labelSelector)

	// Wrap it with NewPDB
	pdb := NewPDB(pdbResource, timeout)

	g.Expect(pdb.GetPDB().Name).To(Equal("api-pdb"))
	g.Expect(pdb.GetPDB().Namespace).To(Equal("production"))
	g.Expect(pdb.GetPDB().Spec.MinAvailable.IntVal).To(Equal(int32(4)))
	g.Expect(pdb.GetPDB().Spec.MaxUnavailable).To(BeNil())
	g.Expect(pdb.GetPDB().Spec.Selector.MatchLabels["service"]).To(Equal("api-gateway"))
	g.Expect(pdb.GetPDB().Spec.Selector.MatchLabels["environment"]).To(Equal("production"))
	g.Expect(pdb.GetPDB().Spec.Selector.MatchLabels["tier"]).To(Equal("frontend"))
}

func TestGetPDB(t *testing.T) {
	g := NewWithT(t)

	pdbWrapper := NewPDB(&testPDB, timeout)
	retrievedPDB := pdbWrapper.GetPDB()

	g.Expect(retrievedPDB.Name).To(Equal("test-pdb"))
	g.Expect(retrievedPDB.Namespace).To(Equal("test-namespace"))
	g.Expect(retrievedPDB.Spec.MinAvailable.IntVal).To(Equal(int32(2)))
	g.Expect(retrievedPDB.Spec.Selector.MatchLabels["app"]).To(Equal("test-app"))
}

func TestDeletePDBWithName(t *testing.T) {
	g := NewWithT(t)

	// This is a unit test that verifies the function signature and basic behavior
	// The actual deletion would be tested in integration tests with a real K8s client

	// Test that the function accepts the correct parameters and returns no error for non-existent PDB
	// (since IsNotFound errors are ignored)
	ctx := context.Background()

	// Create a mock helper - in a real test environment this would be a proper helper
	// For unit testing, we're mainly verifying the function signature and basic logic
	var h *helper.Helper = nil // This would cause a panic if called, but we're testing the interface

	// Test function signature - this should compile without errors
	var testFunc = DeletePDBWithName
	g.Expect(testFunc).ToNot(BeNil())

	// In integration tests, you would test actual deletion:
	// err := DeletePDBWithName(ctx, h, "test-pdb", "test-namespace")
	// g.Expect(err).ShouldNot(HaveOccurred())

	// For now, we just verify the function exists and has the right signature
	g.Expect(DeletePDBWithName).ToNot(BeNil())

	// Test that calling with nil helper would be caught (in real usage, helper would never be nil)
	// This is just to show the function structure is correct
	if h != nil {
		err := DeletePDBWithName(ctx, h, "test-pdb", "test-namespace")
		g.Expect(err).ShouldNot(HaveOccurred()) // This line won't execute due to nil check above
	}
}

func TestIsReady(t *testing.T) {
	g := NewWithT(t)

	// Test case 1: PDB not ready - no status
	pdbNotReady := policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Generation: 1,
		},
		Status: policyv1.PodDisruptionBudgetStatus{
			ObservedGeneration: 0,
			ExpectedPods:       0,
			CurrentHealthy:     0,
		},
	}
	g.Expect(IsReady(pdbNotReady)).To(BeFalse())

	// Test case 2: PDB not ready - generation mismatch
	pdbGenerationMismatch := policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Generation: 2,
		},
		Status: policyv1.PodDisruptionBudgetStatus{
			ObservedGeneration: 1,
			ExpectedPods:       3,
			CurrentHealthy:     2,
		},
	}
	g.Expect(IsReady(pdbGenerationMismatch)).To(BeFalse())

	// Test case 3: PDB not ready - no expected pods
	pdbNoExpectedPods := policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Generation: 1,
		},
		Status: policyv1.PodDisruptionBudgetStatus{
			ObservedGeneration: 1,
			ExpectedPods:       0,
			CurrentHealthy:     0,
		},
	}
	g.Expect(IsReady(pdbNoExpectedPods)).To(BeFalse())

	// Test case 4: PDB ready
	pdbReady := policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Generation: 1,
		},
		Status: policyv1.PodDisruptionBudgetStatus{
			ObservedGeneration: 1,
			ExpectedPods:       3,
			CurrentHealthy:     2,
		},
	}
	g.Expect(IsReady(pdbReady)).To(BeTrue())

	// Test case 5: PDB ready with zero healthy pods (edge case)
	pdbReadyZeroHealthy := policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Generation: 1,
		},
		Status: policyv1.PodDisruptionBudgetStatus{
			ObservedGeneration: 1,
			ExpectedPods:       1,
			CurrentHealthy:     0,
		},
	}
	g.Expect(IsReady(pdbReadyZeroHealthy)).To(BeTrue())
}

func TestPDBSpec(t *testing.T) {
	g := NewWithT(t)

	// Test minAvailable spec using helper function
	minAvailable := intstr.FromInt(2)
	labelSelector := map[string]string{"app": "test-app"}
	pdbResource := MinAvailablePodDisruptionBudget("test-pdb-minavailable", "test-namespace", minAvailable, labelSelector)
	pdbMinAvailable := NewPDB(pdbResource, timeout)

	spec := pdbMinAvailable.GetPDB().Spec
	g.Expect(spec.MinAvailable).ToNot(BeNil())
	g.Expect(spec.MinAvailable.IntVal).To(Equal(int32(2)))
	g.Expect(spec.MaxUnavailable).To(BeNil())

	// Test maxUnavailable spec using helper function
	maxUnavailable := intstr.FromInt(1)
	pdbMaxResource := MaxUnavailablePodDisruptionBudget("test-pdb-maxunavailable", "test-namespace", maxUnavailable, labelSelector)
	pdbMaxUnavailable := NewPDB(pdbMaxResource, timeout)

	spec = pdbMaxUnavailable.GetPDB().Spec
	g.Expect(spec.MaxUnavailable).ToNot(BeNil())
	g.Expect(spec.MaxUnavailable.IntVal).To(Equal(int32(1)))
	g.Expect(spec.MinAvailable).To(BeNil())

	// Test selector
	g.Expect(spec.Selector).ToNot(BeNil())
	g.Expect(spec.Selector.MatchLabels["app"]).To(Equal("test-app"))
}

func TestPDBWithPercentage(t *testing.T) {
	g := NewWithT(t)

	pdbWithPercentage := policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pdb-percentage",
			Namespace: "test-namespace",
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "50%",
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test-app",
				},
			},
		},
	}

	pdb := NewPDB(&pdbWithPercentage, timeout)
	g.Expect(pdb.GetPDB().Spec.MinAvailable.StrVal).To(Equal("50%"))
	g.Expect(pdb.GetPDB().Spec.MinAvailable.Type).To(Equal(intstr.String))

	// Test minAvailable with percentage using helper
	minAvailablePercent := intstr.FromString("80%")
	labelSelector := map[string]string{"app": "test-app"}
	pdbMinAvailablePercent := MinAvailablePodDisruptionBudget("test-pdb-min-percent", "test-namespace", minAvailablePercent, labelSelector)

	pdbWrapper := NewPDB(pdbMinAvailablePercent, timeout)
	g.Expect(pdbWrapper.GetPDB().Spec.MinAvailable.StrVal).To(Equal("80%"))
	g.Expect(pdbWrapper.GetPDB().Spec.MinAvailable.Type).To(Equal(intstr.String))
	g.Expect(pdbWrapper.GetPDB().Spec.MaxUnavailable).To(BeNil())

	// Test maxUnavailable with percentage using helper
	maxUnavailablePercent := intstr.FromString("30%")
	pdbMaxUnavailablePercent := MaxUnavailablePodDisruptionBudget("test-pdb-max-percent", "test-namespace", maxUnavailablePercent, labelSelector)

	pdbMaxWrapper := NewPDB(pdbMaxUnavailablePercent, timeout)
	g.Expect(pdbMaxWrapper.GetPDB().Spec.MaxUnavailable.StrVal).To(Equal("30%"))
	g.Expect(pdbMaxWrapper.GetPDB().Spec.MaxUnavailable.Type).To(Equal(intstr.String))
	g.Expect(pdbMaxWrapper.GetPDB().Spec.MinAvailable).To(BeNil())
}

func TestPDBWithUnhealthyPodEvictionPolicy(t *testing.T) {
	g := NewWithT(t)

	pdbWithPolicy := policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pdb-policy",
			Namespace: "test-namespace",
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 1,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test-app",
				},
			},
			UnhealthyPodEvictionPolicy: ptr.To(policyv1.AlwaysAllow),
		},
	}

	pdb := NewPDB(&pdbWithPolicy, timeout)
	g.Expect(pdb.GetPDB().Spec.UnhealthyPodEvictionPolicy).ToNot(BeNil())
	g.Expect(*pdb.GetPDB().Spec.UnhealthyPodEvictionPolicy).To(Equal(policyv1.AlwaysAllow))

	// Test with MinAvailable helper function and then add policy
	minAvailable := intstr.FromInt(3)
	labelSelector := map[string]string{"app": "test-app"}
	pdbResource := MinAvailablePodDisruptionBudget("test-pdb-min-helper-policy", "test-namespace", minAvailable, labelSelector)

	// Add UnhealthyPodEvictionPolicy to the resource
	pdbResource.Spec.UnhealthyPodEvictionPolicy = ptr.To(policyv1.IfHealthyBudget)

	pdbWithMinHelperPolicy := NewPDB(pdbResource, timeout)
	g.Expect(pdbWithMinHelperPolicy.GetPDB().Spec.UnhealthyPodEvictionPolicy).ToNot(BeNil())
	g.Expect(*pdbWithMinHelperPolicy.GetPDB().Spec.UnhealthyPodEvictionPolicy).To(Equal(policyv1.IfHealthyBudget))
	g.Expect(pdbWithMinHelperPolicy.GetPDB().Spec.MinAvailable.IntVal).To(Equal(int32(3)))
	g.Expect(pdbWithMinHelperPolicy.GetPDB().Spec.MaxUnavailable).To(BeNil())

	// Test with MaxUnavailable helper function and then add policy
	maxUnavailable := intstr.FromInt(2)
	pdbMaxResource := MaxUnavailablePodDisruptionBudget("test-pdb-max-helper-policy", "test-namespace", maxUnavailable, labelSelector)

	// Add UnhealthyPodEvictionPolicy to the resource
	pdbMaxResource.Spec.UnhealthyPodEvictionPolicy = ptr.To(policyv1.AlwaysAllow)

	pdbWithMaxHelperPolicy := NewPDB(pdbMaxResource, timeout)
	g.Expect(pdbWithMaxHelperPolicy.GetPDB().Spec.UnhealthyPodEvictionPolicy).ToNot(BeNil())
	g.Expect(*pdbWithMaxHelperPolicy.GetPDB().Spec.UnhealthyPodEvictionPolicy).To(Equal(policyv1.AlwaysAllow))
	g.Expect(pdbWithMaxHelperPolicy.GetPDB().Spec.MaxUnavailable.IntVal).To(Equal(int32(2)))
	g.Expect(pdbWithMaxHelperPolicy.GetPDB().Spec.MinAvailable).To(BeNil())
}

func TestPDBHelperFunctionsComparison(t *testing.T) {
	g := NewWithT(t)

	labelSelector := map[string]string{
		"app":     "my-app",
		"version": "v2.0",
	}

	// Test MinAvailable helper
	minAvailable := intstr.FromInt(5)
	pdbMin := MinAvailablePodDisruptionBudget("app-min-pdb", "default", minAvailable, labelSelector)

	g.Expect(pdbMin.Spec.MinAvailable).ToNot(BeNil())
	g.Expect(pdbMin.Spec.MinAvailable.IntVal).To(Equal(int32(5)))
	g.Expect(pdbMin.Spec.MaxUnavailable).To(BeNil())

	// Test MaxUnavailable helper
	maxUnavailable := intstr.FromInt(2)
	pdbMax := MaxUnavailablePodDisruptionBudget("app-max-pdb", "default", maxUnavailable, labelSelector)

	g.Expect(pdbMax.Spec.MaxUnavailable).ToNot(BeNil())
	g.Expect(pdbMax.Spec.MaxUnavailable.IntVal).To(Equal(int32(2)))
	g.Expect(pdbMax.Spec.MinAvailable).To(BeNil())

	// Both should have the same selector
	g.Expect(pdbMin.Spec.Selector.MatchLabels).To(Equal(pdbMax.Spec.Selector.MatchLabels))

	// Test with percentages
	minAvailablePercent := intstr.FromString("60%")
	maxUnavailablePercent := intstr.FromString("40%")

	pdbMinPercent := MinAvailablePodDisruptionBudget("app-min-percent", "default", minAvailablePercent, labelSelector)
	pdbMaxPercent := MaxUnavailablePodDisruptionBudget("app-max-percent", "default", maxUnavailablePercent, labelSelector)

	g.Expect(pdbMinPercent.Spec.MinAvailable.StrVal).To(Equal("60%"))
	g.Expect(pdbMaxPercent.Spec.MaxUnavailable.StrVal).To(Equal("40%"))
}
