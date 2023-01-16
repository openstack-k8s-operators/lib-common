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
	"github.com/google/uuid"
	t "github.com/onsi/gomega"
	keystonev1 "github.com/openstack-k8s-operators/keystone-operator/api/v1beta1"
	"github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"
)

// CreateKeystoneAPI -
func (tc *TestHelper) CreateKeystoneAPI(namespace string) types.NamespacedName {
	keystone := &keystonev1.KeystoneAPI{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "keystone.openstack.org/v1beta1",
			Kind:       "KeystoneAPI",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "keystone-" + uuid.New().String(),
			Namespace: namespace,
		},
		Spec: keystonev1.KeystoneAPISpec{},
	}

	t.Expect(tc.k8sClient.Create(tc.ctx, keystone.DeepCopy())).Should(t.Succeed())
	name := types.NamespacedName{Namespace: namespace, Name: keystone.Name}

	// the Status field needs to be written via a separate client
	keystone = tc.GetKeystoneAPI(name)
	keystone.Status = keystonev1.KeystoneAPIStatus{
		APIEndpoints: map[string]string{"public": "http://keystone-public-openstack.testing"},
	}
	t.Expect(tc.k8sClient.Status().Update(tc.ctx, keystone.DeepCopy())).Should(t.Succeed())

	tc.logger.Info("KeystoneAPI created", "KeystoneAPI", name)
	return name
}

// DeleteKeystoneAPI -
func (tc *TestHelper) DeleteKeystoneAPI(name types.NamespacedName) {
	t.Eventually(func(g t.Gomega) {
		keystone := &keystonev1.KeystoneAPI{}
		err := tc.k8sClient.Get(tc.ctx, name, keystone)
		// if it is already gone that is OK
		if k8s_errors.IsNotFound(err) {
			return
		}
		g.Expect(err).Should(t.BeNil())

		g.Expect(tc.k8sClient.Delete(tc.ctx, keystone)).Should(t.Succeed())

		err = tc.k8sClient.Get(tc.ctx, name, keystone)
		g.Expect(k8s_errors.IsNotFound(err)).To(t.BeTrue())
	}, tc.timeout, tc.interval).Should(t.Succeed())
}

// GetKeystoneAPI -
func (tc *TestHelper) GetKeystoneAPI(name types.NamespacedName) *keystonev1.KeystoneAPI {
	instance := &keystonev1.KeystoneAPI{}
	t.Eventually(func(g t.Gomega) {
		g.Expect(tc.k8sClient.Get(tc.ctx, name, instance)).Should(t.Succeed())
	}, tc.timeout, tc.interval).Should(t.Succeed())
	return instance
}

// GetKeystoneService -
func (tc *TestHelper) GetKeystoneService(name types.NamespacedName) *keystonev1.KeystoneService {
	instance := &keystonev1.KeystoneService{}
	t.Eventually(func(g t.Gomega) {
		g.Expect(tc.k8sClient.Get(tc.ctx, name, instance)).Should(t.Succeed())
	}, tc.timeout, tc.interval).Should(t.Succeed())
	return instance
}

// SimulateKeystoneServiceReady -
func (tc *TestHelper) SimulateKeystoneServiceReady(name types.NamespacedName) {
	t.Eventually(func(g t.Gomega) {
		service := tc.GetKeystoneService(name)
		service.Status.Conditions.MarkTrue(condition.ReadyCondition, "Ready")
		g.Expect(tc.k8sClient.Status().Update(tc.ctx, service)).To(t.Succeed())
	}, tc.timeout, tc.interval).Should(t.Succeed())
	tc.logger.Info("Simulated KeystoneService ready", "on", name)
}

// AssertServiceExists -
func (tc *TestHelper) AssertServiceExists(name types.NamespacedName) *corev1.Service {
	instance := &corev1.Service{}
	t.Eventually(func(g t.Gomega) {
		g.Expect(tc.k8sClient.Get(tc.ctx, name, instance)).Should(t.Succeed())
	}, tc.timeout, tc.interval).Should(t.Succeed())
	return instance
}

// GetKeystoneEndpoint -
func (tc *TestHelper) GetKeystoneEndpoint(name types.NamespacedName) *keystonev1.KeystoneEndpoint {
	instance := &keystonev1.KeystoneEndpoint{}
	t.Eventually(func(g t.Gomega) {
		g.Expect(tc.k8sClient.Get(tc.ctx, name, instance)).Should(t.Succeed())
	}, tc.timeout, tc.interval).Should(t.Succeed())
	return instance
}

// SimulateKeystoneEndpointReady -
func (tc *TestHelper) SimulateKeystoneEndpointReady(name types.NamespacedName) {
	t.Eventually(func(g t.Gomega) {
		endpoint := tc.GetKeystoneEndpoint(name)
		endpoint.Status.Conditions.MarkTrue(condition.ReadyCondition, "Ready")
		g.Expect(tc.k8sClient.Status().Update(tc.ctx, endpoint)).To(t.Succeed())
	}, tc.timeout, tc.interval).Should(t.Succeed())
	tc.logger.Info("Simulated KeystoneEndpoint ready", "on", name)
}
