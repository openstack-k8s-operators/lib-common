/*
Copyright 2022 Red Hat
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

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetSecret -
func (tc *TestHelper) GetSecret(name types.NamespacedName) corev1.Secret {
	secret := &corev1.Secret{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.k8sClient.Get(tc.ctx, name, secret)).Should(gomega.Succeed())
	}, tc.timeout, tc.interval).Should(gomega.Succeed())

	return *secret
}

// ListSecrets -
func (tc *TestHelper) ListSecrets(namespace string) corev1.SecretList {
	secrets := &corev1.SecretList{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.k8sClient.List(tc.ctx, secrets, client.InNamespace(namespace))).Should(gomega.Succeed())
	}, tc.timeout, tc.interval).Should(gomega.Succeed())

	return *secrets
}

// CreateSecret -
func (tc *TestHelper) CreateSecret(name types.NamespacedName, data map[string][]byte) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Data: data,
	}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.k8sClient.Create(tc.ctx, secret)).Should(gomega.Succeed())
	}, tc.timeout, tc.interval).Should(gomega.Succeed())

	return secret
}

// CreateEmptySecret -
func (tc *TestHelper) CreateEmptySecret(name types.NamespacedName) *corev1.Secret {
	return tc.CreateSecret(name, map[string][]byte{})
}

// DeleteSecret -
func (tc *TestHelper) DeleteSecret(name types.NamespacedName) {
	gomega.Eventually(func(g gomega.Gomega) {
		secret := &corev1.Secret{}
		err := tc.k8sClient.Get(tc.ctx, name, secret)
		// if it is already gone that is OK
		if k8s_errors.IsNotFound(err) {
			return
		}
		g.Expect(err).ShouldNot(gomega.HaveOccurred())

		g.Expect(tc.k8sClient.Delete(tc.ctx, secret)).Should(gomega.Succeed())

		err = tc.k8sClient.Get(tc.ctx, name, secret)
		g.Expect(k8s_errors.IsNotFound(err)).To(gomega.BeTrue())
	}, tc.timeout, tc.interval).Should(gomega.Succeed())
}
