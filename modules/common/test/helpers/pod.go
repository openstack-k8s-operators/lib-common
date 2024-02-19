/*
Copyright 2024 Red Hat
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
)

// GetPod - retrieves a Pod resource.
//
// example usage:
//
//	th.Pod(types.NamespacedName{Name: "test-pod", Namespace: "test-namespace"})
func (tc *TestHelper) GetPod(name types.NamespacedName) *corev1.Pod {
	pod := &corev1.Pod{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, pod)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
	return pod
}

// SimulatePodPhaseRunning retrieves the Pod and simulates
// a running phase for the Pod in a Kubernetes cluster.
//
// example usage:
//
//	th.SimulatePodPhaseRunning(types.NamespacedName{Name: "test-pod", Namespace: "test-namespace"})
func (tc *TestHelper) SimulatePodPhaseRunning(name types.NamespacedName) {
	gomega.Eventually(func(g gomega.Gomega) {
		pod := tc.GetPod(name)
		pod.Status.Phase = corev1.PodRunning
		g.Expect(tc.K8sClient.Status().Update(tc.Ctx, pod)).To(gomega.Succeed())

	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
	tc.Logger.Info("Simulated pod running phase", "on", name)
}
