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

// GetConfigMap -
func (tc *TestHelper) GetConfigMap(name types.NamespacedName) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, cm)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	return cm
}

// ListConfigMaps -
func (tc *TestHelper) ListConfigMaps(namespace string) *corev1.ConfigMapList {
	cms := &corev1.ConfigMapList{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.List(tc.Ctx, cms, client.InNamespace(namespace))).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	return cms
}

// CreateEmptyConfigMap -
func (tc *TestHelper) CreateEmptyConfigMap(name types.NamespacedName) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Data: map[string]string{},
	}
	gomega.Expect(tc.K8sClient.Create(tc.Ctx, cm)).Should(gomega.Succeed())

	return cm
}

// DeleteConfigMap -
func (tc *TestHelper) DeleteConfigMap(name types.NamespacedName) {
	gomega.Eventually(func(g gomega.Gomega) {
		configMap := &corev1.ConfigMap{}
		err := tc.K8sClient.Get(tc.Ctx, name, configMap)
		// if it is already gone that is OK
		if k8s_errors.IsNotFound(err) {
			return
		}
		g.Expect(err).ShouldNot(gomega.HaveOccurred())

		g.Expect(tc.K8sClient.Delete(tc.Ctx, configMap)).Should(gomega.Succeed())

		err = tc.K8sClient.Get(tc.Ctx, name, configMap)
		g.Expect(k8s_errors.IsNotFound(err)).To(gomega.BeTrue())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
}
