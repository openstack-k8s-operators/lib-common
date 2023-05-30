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
)

// CreateKeystoneAPI creates a new KeystoneAPI instance with the specified namespace in the Kubernetes cluster.
//
// Example usage:
//
//	keystoneAPI := th.CreateKeystoneAPI(namespace)
//	DeferCleanup(th.DeleteKeystoneAPI, keystoneAPI)
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

	t.Expect(tc.K8sClient.Create(tc.Ctx, keystone.DeepCopy())).Should(t.Succeed())
	name := types.NamespacedName{Namespace: namespace, Name: keystone.Name}

	// the Status field needs to be written via a separate client
	keystone = tc.GetKeystoneAPI(name)
	keystone.Status = keystonev1.KeystoneAPIStatus{
		APIEndpoints: map[string]string{
			"public":   "http://keystone-public-openstack.testing",
			"internal": "http://keystone-internal.openstack.svc:5000",
		},
	}
	t.Expect(tc.K8sClient.Status().Update(tc.Ctx, keystone.DeepCopy())).Should(t.Succeed())

	tc.Logger.Info("KeystoneAPI created", "KeystoneAPI", name)
	return name
}

// DeleteKeystoneAPI deletes a KeystoneAPI resource from the Kubernetes cluster.
//
// # After the deletion, the function checks again if the KeystoneAPI is successfully deleted
//
// Example usage:
//
//	keystoneAPI := th.CreateKeystoneAPI(namespace)
//	DeferCleanup(th.DeleteKeystoneAPI, keystoneAPI)
func (tc *TestHelper) DeleteKeystoneAPI(name types.NamespacedName) {
	t.Eventually(func(g t.Gomega) {
		keystone := &keystonev1.KeystoneAPI{}
		err := tc.K8sClient.Get(tc.Ctx, name, keystone)
		// if it is already gone that is OK
		if k8s_errors.IsNotFound(err) {
			return
		}
		g.Expect(err).NotTo(t.HaveOccurred())

		g.Expect(tc.K8sClient.Delete(tc.Ctx, keystone)).Should(t.Succeed())

		err = tc.K8sClient.Get(tc.Ctx, name, keystone)
		g.Expect(k8s_errors.IsNotFound(err)).To(t.BeTrue())
	}, tc.Timeout, tc.Interval).Should(t.Succeed())
}

// GetKeystoneAPI retrieves a KeystoneAPI resource.
//
// The function returns a pointer to the retrieved KeystoneAPI resource.
// example usage:
//
//	  keystoneAPIName := th.CreateKeystoneAPI(novaNames.NovaName.Namespace)
//		 DeferCleanup(th.DeleteKeystoneAPI, keystoneAPIName)
//		 keystoneAPI := th.GetKeystoneAPI(keystoneAPIName)
func (tc *TestHelper) GetKeystoneAPI(name types.NamespacedName) *keystonev1.KeystoneAPI {
	instance := &keystonev1.KeystoneAPI{}
	t.Eventually(func(g t.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, instance)).Should(t.Succeed())
	}, tc.Timeout, tc.Interval).Should(t.Succeed())
	return instance
}

// GetKeystoneService function retrieves and returns the KeystoneService resource
//
// Example usage:
//
//	keystoneServiceName := th.CreateKeystoneService(namespace)
func (tc *TestHelper) GetKeystoneService(name types.NamespacedName) *keystonev1.KeystoneService {
	instance := &keystonev1.KeystoneService{}
	t.Eventually(func(g t.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, instance)).Should(t.Succeed())
	}, tc.Timeout, tc.Interval).Should(t.Succeed())
	return instance
}

// SimulateKeystoneServiceReady simulates the readiness of a KeystoneService
// resource by seting the Ready condition of the KeystoneService to true
//
// Example usage:
// keystoneServiceName := th.CreateKeystoneService(namespace)
func (tc *TestHelper) SimulateKeystoneServiceReady(name types.NamespacedName) {
	t.Eventually(func(g t.Gomega) {
		service := tc.GetKeystoneService(name)
		service.Status.Conditions.MarkTrue(condition.ReadyCondition, "Ready")
		g.Expect(tc.K8sClient.Status().Update(tc.Ctx, service)).To(t.Succeed())
	}, tc.Timeout, tc.Interval).Should(t.Succeed())
	tc.Logger.Info("Simulated KeystoneService ready", "on", name)
}

// GetKeystoneEndpoint retrieves a KeystoneEndpoint resource from the Kubernetes cluster.
//
// Example usage:
//
//	keystoneEndpointName := th.CreateKeystoneEndpoint(namespace)
func (tc *TestHelper) GetKeystoneEndpoint(name types.NamespacedName) *keystonev1.KeystoneEndpoint {
	instance := &keystonev1.KeystoneEndpoint{}
	t.Eventually(func(g t.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, instance)).Should(t.Succeed())
	}, tc.Timeout, tc.Interval).Should(t.Succeed())
	return instance
}

// SimulateKeystoneEndpointReady function retrieves the KeystoneEndpoint resource and
// simulates a KeystoneEndpoint resource being marked as ready.
//
// Example usage:
//
//	keystoneEndpointName := th.CreateKeystoneEndpoint(namespace)
//	th.SimulateKeystoneEndpointReady(keystoneEndpointName)
func (tc *TestHelper) SimulateKeystoneEndpointReady(name types.NamespacedName) {
	t.Eventually(func(g t.Gomega) {
		endpoint := tc.GetKeystoneEndpoint(name)
		endpoint.Status.Conditions.MarkTrue(condition.ReadyCondition, "Ready")
		g.Expect(tc.K8sClient.Status().Update(tc.Ctx, endpoint)).To(t.Succeed())
	}, tc.Timeout, tc.Interval).Should(t.Succeed())
	tc.Logger.Info("Simulated KeystoneEndpoint ready", "on", name)
}
