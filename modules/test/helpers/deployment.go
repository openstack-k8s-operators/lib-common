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
	"encoding/json"
	networkv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	appsv1 "k8s.io/api/apps/v1"
)

// GetDeployment - retrieves a Deployment resource from cluster.
// The function uses the Gomega library's Eventually function to
// repeatedly attempt to get the Deployment until it is successful or
// the test's timeout is reached.
//
// The function returns a pointer to the retrieved Deployment.
// If the function cannot find the Deployment within the timeout,
// it will cause the test to fail.
//
// Example usage:
//
//	  deployment := th.GetDeployment(
//					types.NamespacedName{
//						Namespace: neutronAPIName.Namespace,
//						Name:      "neutron",
//					},
//				)
func (tc *TestHelper) GetDeployment(name types.NamespacedName) *appsv1.Deployment {
	deployment := &appsv1.Deployment{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, deployment)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	return deployment
}

// SimulateDeploymentReplicaReady function retrieves the Deployment resource and
// simulate that replicas are ready
// Example usage:
//
//	th.SimulateDeploymentReplicaReady(ironicNames.INAName)
func (tc *TestHelper) SimulateDeploymentReplicaReady(name types.NamespacedName) {
	gomega.Eventually(func(g gomega.Gomega) {
		deployment := tc.GetDeployment(name)

		deployment.Status.Replicas = 1
		deployment.Status.ReadyReplicas = 1
		g.Expect(tc.K8sClient.Status().Update(tc.Ctx, deployment)).To(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	tc.Logger.Info("Simulated Deployment success", "on", name)
}

// SimulateDeploymentReadyWithPods simulates a Deployment with ready replicas
// by creating and updating the corresponding Pods.
//
// Example:
//
//	    th.SimulateDeploymentReadyWithPods(
//					manilaTest.Instance,
//					map[string][]string{manilaName.Namespace + "/internalapi": {"10.0.0.1"}},
//				)
func (tc *TestHelper) SimulateDeploymentReadyWithPods(name types.NamespacedName, networkIPs map[string][]string) {
	ss := tc.GetDeployment(name)
	for i := 0; i < int(*ss.Spec.Replicas); i++ {
		pod := &corev1.Pod{
			ObjectMeta: ss.Spec.Template.ObjectMeta,
			Spec:       ss.Spec.Template.Spec,
		}
		pod.ObjectMeta.Namespace = name.Namespace
		pod.ObjectMeta.GenerateName = name.Name

		var netStatus []networkv1.NetworkStatus
		for network, IPs := range networkIPs {
			netStatus = append(
				netStatus,
				networkv1.NetworkStatus{
					Name: network,
					IPs:  IPs,
				},
			)
		}
		netStatusAnnotation, err := json.Marshal(netStatus)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		pod.Annotations[networkv1.NetworkStatusAnnot] = string(netStatusAnnotation)

		gomega.Expect(tc.K8sClient.Create(tc.Ctx, pod)).Should(gomega.Succeed())
	}

	gomega.Eventually(func(g gomega.Gomega) {
		ss := tc.GetDeployment(name)
		ss.Status.Replicas = 1
		ss.Status.ReadyReplicas = 1
		g.Expect(tc.K8sClient.Status().Update(tc.Ctx, ss)).To(gomega.Succeed())

	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	tc.Logger.Info("Simulated deployment success", "on", name)
}
