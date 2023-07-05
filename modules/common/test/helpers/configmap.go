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
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetConfigMap retrieves a ConfigMap resource from a k8s cluster.
//
// Example usage:
// cm := th.GetConfigMap(types.NamespacedName{Namespace: "default", Name: "example-configmap"})
func (tc *TestHelper) GetConfigMap(name types.NamespacedName) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, cm)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	return cm
}

// ListConfigMaps retrieves a list of ConfigMap resources from a specific namespace
//
// Example usage:
//
//	cms := th.ListConfigMaps(novaNames.MetadataName.Name)
func (tc *TestHelper) ListConfigMaps(namespace string) *corev1.ConfigMapList {
	cms := &corev1.ConfigMapList{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.List(tc.Ctx, cms, client.InNamespace(namespace))).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	return cms
}

// DeleteConfigMap deletes a ConfigMap resource from a Kubernetes cluster.
//
// Example usage:
//
//	th.DeleteConfigMap(types.NamespacedName{Namespace: "default", Name: "example-configmap"})
//
// or
//
//	DeferCleanup(th.DeleteConfigMap, inventoryName)
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

// CreateConfigMap creates a new ConfigMap resource with the provided data.
//
// Example usage:
//
//	data := map[string]interface{}{"key": "value"}
//	cm := th.CreateConfigMap(types.NamespacedName{Namespace: "default", Name: "example-configmap"}, data)
func (tc *TestHelper) CreateConfigMap(name types.NamespacedName, data map[string]interface{}) client.Object {
	raw := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name":      name.Name,
			"namespace": name.Namespace,
		},
		"data": data,
	}

	return tc.CreateUnstructured(raw)
}

// AssertConfigMapDoesNotExist ensures the ConfigMap resource does not exist in a k8s cluster.
func (tc *TestHelper) AssertConfigMapDoesNotExist(name types.NamespacedName) {
	instance := &corev1.ConfigMap{}
	gomega.Eventually(func(g gomega.Gomega) {
		err := tc.K8sClient.Get(tc.Ctx, name, instance)
		g.Expect(k8s_errors.IsNotFound(err)).To(gomega.BeTrue())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
}
