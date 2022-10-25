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

// GetConfigMap -
func (tc *TestHelper) GetConfigMap(name types.NamespacedName) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.k8sClient.Get(tc.ctx, name, cm)).Should(gomega.Succeed())
	}, tc.timeout, tc.interval).Should(gomega.Succeed())

	return cm
}

// ListConfigMaps -
func (tc *TestHelper) ListConfigMaps(namespace string) *corev1.ConfigMapList {
	cms := &corev1.ConfigMapList{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.k8sClient.List(tc.ctx, cms, client.InNamespace(namespace))).Should(gomega.Succeed())
	}, tc.timeout, tc.interval).Should(gomega.Succeed())

	return cms
}
