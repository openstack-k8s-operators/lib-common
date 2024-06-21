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

package helpers

import (
	"github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// GetService retrieves a Service resource.
//
// Example usage:
//
//	th.GetService(types.NamespacedName{Name: "test-service", Namespace: "test-namespace"})
func (tc *TestHelper) GetService(name types.NamespacedName) *corev1.Service {
	instance := &corev1.Service{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, instance)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	return instance
}

// CreateService creates a new k8s service resource with provided data.
//
// Example usage:
//
//	secret := th.CreateService(types.NamespacedName{Name: "test-secret", Namespace: "test-namespace"}, map[string]string{}, corev1.ServiceSpec{...})
func (tc *TestHelper) CreateService(name types.NamespacedName, labels map[string]string, svcSpec corev1.ServiceSpec) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
			Labels:    labels,
		},
		Spec: svcSpec,
	}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Create(tc.Ctx, svc)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	return svc
}

// AssertServiceExists - asserts the existence of a Service resource in the Kubernetes cluster.
//
// Example usage:
//
//	th.AssertServiceExists(types.NamespacedName{Name: "neutron-public, Namespace: namespace})
func (tc *TestHelper) AssertServiceExists(name types.NamespacedName) *corev1.Service {
	instance := &corev1.Service{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, instance)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
	return instance
}

// DeleteService - deletes a Service resource from the Kubernetes cluster.
//
// Example usage:
//
//	th.DeleteService(types.NamespacedName{Name: "test-service", Namespace: "test-namespace"})
func (tc *TestHelper) DeleteService(name types.NamespacedName) {
	instance := &corev1.Service{}

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

// AssertServiceDoesNotExist ensures the Service resource does not exist in a k8s cluster.
func (tc *TestHelper) AssertServiceDoesNotExist(name types.NamespacedName) {
	instance := &corev1.Service{}
	gomega.Eventually(func(g gomega.Gomega) {
		err := tc.K8sClient.Get(tc.Ctx, name, instance)
		g.Expect(k8s_errors.IsNotFound(err)).To(gomega.BeTrue())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
}

// SimulateLoadBalancerServiceIP retrieves the k8s service and simulates
// that it got a LoadBalancer IP assigned.
//
// example usage:
//
//	th.SimulateLoadBalancerServiceIP(types.NamespacedName{Name: "test-svc", Namespace: "test-namespace"})
func (tc *TestHelper) SimulateLoadBalancerServiceIP(name types.NamespacedName) {
	gomega.Eventually(func(g gomega.Gomega) {
		s := tc.GetService(name)
		g.Expect(s.Spec.Type).To(gomega.Equal(corev1.ServiceTypeLoadBalancer))

		// simulate LoadBalancer assigned IP and updated the k8s service to have a LB IP
		s.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{
			{
				IP: "1.1.1.1",
			},
		}
		g.Expect(tc.K8sClient.Status().Update(tc.Ctx, s)).To(gomega.Succeed())

	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
	tc.Logger.Info("Simulated LoadBalancer IP success", "on", name)
}
