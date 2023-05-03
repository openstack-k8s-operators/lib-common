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
	t "github.com/onsi/gomega"
	rabbitmqv1 "github.com/openstack-k8s-operators/infra-operator/apis/rabbitmq/v1beta1"
	"k8s.io/apimachinery/pkg/types"
)

// GetTransportURL -
func (tc *TestHelper) GetTransportURL(name types.NamespacedName) *rabbitmqv1.TransportURL {
	instance := &rabbitmqv1.TransportURL{}
	t.Eventually(func(g t.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, instance)).Should(t.Succeed())
	}, tc.Timeout, tc.Interval).Should(t.Succeed())
	return instance
}

// SimulateTransportURLReady -
func (tc *TestHelper) SimulateTransportURLReady(name types.NamespacedName) {
	t.Eventually(func(g t.Gomega) {
		transport := tc.GetTransportURL(name)
		transport.Status.SecretName = transport.Spec.RabbitmqClusterName + "-secret"
		transport.Status.Conditions.MarkTrue("TransportURLReady", "Ready")
		g.Expect(tc.K8sClient.Status().Update(tc.Ctx, transport)).To(t.Succeed())

	}, tc.Timeout, tc.Interval).Should(t.Succeed())
	tc.Logger.Info("Simulated TransportURL ready", "on", name)
}
