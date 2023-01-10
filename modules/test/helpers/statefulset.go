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
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
)

// GetStatefulSet -
func (tc *TestHelper) GetStatefulSet(name types.NamespacedName) *appsv1.StatefulSet {
	ss := &appsv1.StatefulSet{}
	t.Eventually(func(g t.Gomega) {
		g.Expect(tc.k8sClient.Get(tc.ctx, name, ss)).Should(t.Succeed())
	}, tc.timeout, tc.interval).Should(t.Succeed())
	return ss
}

// ListStatefulSets -
func (tc *TestHelper) ListStatefulSets(namespace string) *appsv1.StatefulSetList {
	sss := &appsv1.StatefulSetList{}
	t.Expect(tc.k8sClient.List(tc.ctx, sss, client.InNamespace(namespace))).Should(t.Succeed())
	return sss

}

// SimulateStatefulSetReplicaReady -
func (tc *TestHelper) SimulateStatefulSetReplicaReady(name types.NamespacedName) {
	t.Eventually(func(g t.Gomega) {
		ss := tc.GetStatefulSet(name)
		ss.Status.Replicas = 1
		ss.Status.ReadyReplicas = 1
		g.Expect(tc.k8sClient.Status().Update(tc.ctx, ss)).To(t.Succeed())

	}, tc.timeout, tc.interval).Should(t.Succeed())
	tc.logger.Info("Simulated statefulset success", "on", name)
}
