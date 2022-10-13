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
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetSecret -
func GetSecret(name types.NamespacedName) corev1.Secret {
	secret := &corev1.Secret{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(k8sClient.Get(ctx, name, secret)).Should(gomega.Succeed())
	}, timeout, interval).Should(gomega.Succeed())

	return *secret
}

// ListSecrets -
func ListSecrets(namespace string) corev1.SecretList {
	secrets := &corev1.SecretList{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(k8sClient.List(ctx, secrets, client.InNamespace(namespace))).Should(gomega.Succeed())
	}, timeout, interval).Should(gomega.Succeed())

	return *secrets
}
