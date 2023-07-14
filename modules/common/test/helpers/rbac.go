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
	rbacv1 "k8s.io/api/rbac/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// GetServiceAccount fetches a ServiceAccount resource
//
// Example usage:
//
//	th.GetServiceAccount(types.NamespacedName{Name: "test-service-account", Namespace: "test-namespace"})
func (tc *TestHelper) GetServiceAccount(name types.NamespacedName) *corev1.ServiceAccount {
	instance := &corev1.ServiceAccount{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, instance)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	return instance
}

// GetRole fetches a Role resource.
//
// Example usage:
//
//	th.GetRole(types.NamespacedName{Name: "test-role", Namespace: "test-namespace"})
func (tc *TestHelper) GetRole(name types.NamespacedName) *rbacv1.Role {
	instance := &rbacv1.Role{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, instance)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	return instance
}

// GetRoleBinding - fetches a RoleBinding resource
//
// Example usage:
//
//	th.GetRoleBinding(types.NamespacedName{Name: "test-rolebinding", Namespace: "test-namespace"})
func (tc *TestHelper) GetRoleBinding(name types.NamespacedName) *rbacv1.RoleBinding {
	instance := &rbacv1.RoleBinding{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, instance)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	return instance
}

// AssertRoleBindingDoesNotExist ensures the RoleBinding resource does not exist in a k8s cluster.
func (tc *TestHelper) AssertRoleBindingDoesNotExist(name types.NamespacedName) {
	instance := &rbacv1.RoleBinding{}
	gomega.Eventually(func(g gomega.Gomega) {
		err := tc.K8sClient.Get(tc.Ctx, name, instance)
		g.Expect(k8s_errors.IsNotFound(err)).To(gomega.BeTrue())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
}
