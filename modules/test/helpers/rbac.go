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
	"k8s.io/apimachinery/pkg/types"
)

// GetServiceAccount -
func (tc *TestHelper) GetServiceAccount(name types.NamespacedName) *corev1.ServiceAccount {
	instance := &corev1.ServiceAccount{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.k8sClient.Get(tc.ctx, name, instance)).Should(gomega.Succeed())
	}, tc.timeout, tc.interval).Should(gomega.Succeed())

	return instance
}

// GetRole -
func (tc *TestHelper) GetRole(name types.NamespacedName) *rbacv1.Role {
	instance := &rbacv1.Role{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.k8sClient.Get(tc.ctx, name, instance)).Should(gomega.Succeed())
	}, tc.timeout, tc.interval).Should(gomega.Succeed())

	return instance
}

// GetRoleBinding -
func (tc *TestHelper) GetRoleBinding(name types.NamespacedName) *rbacv1.RoleBinding {
	instance := &rbacv1.RoleBinding{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.k8sClient.Get(tc.ctx, name, instance)).Should(gomega.Succeed())
	}, tc.timeout, tc.interval).Should(gomega.Succeed())

	return instance
}
