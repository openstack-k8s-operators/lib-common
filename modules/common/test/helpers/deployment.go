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
	"fmt"

	networkv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
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

// SimulateDeploymentAnyNumberReplicaReady function retrieves the Deployment resource and
// simulate that replicas are ready
// Example usage:
//
//	th.SimulateDeploymentAnyNumberReplicaReady(ironicNames.INAName, 0)
func (tc *TestHelper) SimulateDeploymentAnyNumberReplicaReady(name types.NamespacedName, replica int32) {
	gomega.Eventually(func(g gomega.Gomega) {
		deployment := tc.GetDeployment(name)

		deployment.Status.AvailableReplicas = replica
		deployment.Status.Replicas = replica
		deployment.Status.ReadyReplicas = replica
		deployment.Status.UpdatedReplicas = replica
		deployment.Status.ObservedGeneration = deployment.Generation
		g.Expect(tc.K8sClient.Status().Update(tc.Ctx, deployment)).To(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	tc.Logger.Info("Simulated Deployment success", "on", name)
}

// SimulateDeploymentReplicaReady function retrieves the Deployment resource and
// simulate that replicas are ready
// Example usage:
//
//	th.SimulateDeploymentReplicaReady(ironicNames.INAName)
func (tc *TestHelper) SimulateDeploymentReplicaReady(name types.NamespacedName) {
	gomega.Eventually(func(g gomega.Gomega) {
		deployment := tc.GetDeployment(name)

		deployment.Status.AvailableReplicas = *deployment.Spec.Replicas
		deployment.Status.Replicas = *deployment.Spec.Replicas
		deployment.Status.ReadyReplicas = *deployment.Spec.Replicas
		deployment.Status.UpdatedReplicas = *deployment.Spec.Replicas
		deployment.Status.ObservedGeneration = deployment.Generation

		tc.Logger.Info("Simulated Deployment success", "ObservedGeneration", deployment.Status.ObservedGeneration)
		tc.Logger.Info("Simulated Deployment success", "Generation", deployment.Generation)

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
	depl := tc.GetDeployment(name)
	for i := 0; i < int(*depl.Spec.Replicas); i++ {
		pod := &corev1.Pod{
			ObjectMeta: depl.Spec.Template.ObjectMeta,
			Spec:       depl.Spec.Template.Spec,
		}
		pod.ObjectMeta.Namespace = name.Namespace
		pod.ObjectMeta.GenerateName = name.Name
		// NOTE(gibi): If there is a mount that refers to a volume created via
		// persistent volume claim then that mount won't have a corresponding
		// volume created in EnvTest as we are not simulating the k8s volume
		// claim logic here at the moment. Therefore the Pod create would fail
		// with a missing volume. So to avoid that we remove every mount and
		// volume from the pod we create here.
		pod.Spec.Volumes = []corev1.Volume{}
		for i := range pod.Spec.Containers {
			pod.Spec.Containers[i].VolumeMounts = []corev1.VolumeMount{}
		}
		for i := range pod.Spec.InitContainers {
			pod.Spec.InitContainers[i].VolumeMounts = []corev1.VolumeMount{}
		}
		// Skip adding network annotations if networkIPs is empty
		if len(networkIPs) > 0 {
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
		}

		gomega.Expect(tc.K8sClient.Create(tc.Ctx, pod)).Should(gomega.Succeed())
	}

	gomega.Eventually(func(g gomega.Gomega) {
		depl := tc.GetDeployment(name)
		depl.Status.AvailableReplicas = *depl.Spec.Replicas
		depl.Status.Replicas = *depl.Spec.Replicas
		depl.Status.ReadyReplicas = *depl.Spec.Replicas
		depl.Status.UpdatedReplicas = *depl.Spec.Replicas
		depl.Status.ObservedGeneration = depl.Generation
		g.Expect(tc.K8sClient.Status().Update(tc.Ctx, depl)).To(gomega.Succeed())

	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	tc.Logger.Info("Simulated deployment success", "on", name)
}

// AssertDeploymentDoesNotExist ensures the Deployment resource does not exist in a k8s cluster.
func (tc *TestHelper) AssertDeploymentDoesNotExist(name types.NamespacedName) {
	instance := &appsv1.Deployment{}
	gomega.Eventually(func(g gomega.Gomega) {
		err := tc.K8sClient.Get(tc.Ctx, name, instance)
		g.Expect(k8s_errors.IsNotFound(err)).To(gomega.BeTrue())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
}

// SimulateDeploymentProgressing function retrieves the Deployment resource and
// simulate that replicas are progressing
// Example usage:
//
//	th.SimulateDeploymentProgressing(ironicNames.INAName)
func (tc *TestHelper) SimulateDeploymentProgressing(name types.NamespacedName) {
	gomega.Eventually(func(g gomega.Gomega) {
		deployment := tc.GetDeployment(name)

		deployment.Status.AvailableReplicas = *deployment.Spec.Replicas
		deployment.Status.Replicas = *deployment.Spec.Replicas + 1
		deployment.Status.ReadyReplicas = *deployment.Spec.Replicas
		deployment.Status.UpdatedReplicas = *deployment.Spec.Replicas
		deployment.Status.ObservedGeneration = deployment.Generation

		/*
			conditions:
			- lastTransitionTime: "2025-03-11T10:53:00Z"
			  lastUpdateTime: "2025-03-11T10:53:00Z"
			  message: Deployment has minimum availability.
			  reason: MinimumReplicasAvailable
			  status: "True"
			  type: Available
			- lastTransitionTime: "2025-03-11T16:17:41Z"
			  lastUpdateTime: "2025-03-11T16:27:31Z"
			  message: ReplicaSet "keystone-869cb5d44c" is progressing.
			  reason: ReplicaSetUpdated
			  status: "True"
			  type: Progressing
		*/

		deployment.Status.Conditions = []appsv1.DeploymentCondition{
			{
				Message: "Deployment has minimum availability",
				Reason:  "MinimumReplicasAvailable",
				Status:  corev1.ConditionTrue,
				Type:    appsv1.DeploymentAvailable,
			},
			{
				Message: fmt.Sprintf("ReplicaSet \"%s-869cb5d44c\" is progressing.", deployment.Name),
				Reason:  "ReplicaSetUpdated",
				Status:  corev1.ConditionTrue,
				Type:    appsv1.DeploymentProgressing,
			}}

		tc.Logger.Info("Simulated Deployment progressing", "ObservedGeneration", deployment.Status.ObservedGeneration)
		tc.Logger.Info("Simulated Deployment progressing", "Generation", deployment.Generation)

		g.Expect(tc.K8sClient.Status().Update(tc.Ctx, deployment)).To(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	tc.Logger.Info("Simulated Deployment progressing", "on", name)
}

// SimulateDeploymentProgressDeadlineExceeded function retrieves the Deployment resource and
// simulate that it hit ProgressDeadlineExceeded
// Example usage:
//
//	th.SimulateDeploymentProgressDeadlineExceeded(ironicNames.INAName)
func (tc *TestHelper) SimulateDeploymentProgressDeadlineExceeded(name types.NamespacedName) {
	gomega.Eventually(func(g gomega.Gomega) {
		deployment := tc.GetDeployment(name)

		deployment.Status.AvailableReplicas = *deployment.Spec.Replicas
		deployment.Status.Replicas = *deployment.Spec.Replicas + 1
		deployment.Status.ReadyReplicas = *deployment.Spec.Replicas
		deployment.Status.UpdatedReplicas = *deployment.Spec.Replicas
		deployment.Status.ObservedGeneration = deployment.Generation
		deployment.Status.UnavailableReplicas = 1

		/*
			conditions:
			- lastTransitionTime: "2025-03-14T11:09:42Z"
			  lastUpdateTime: "2025-03-14T11:09:42Z"
			  message: Deployment has minimum availability.
			  reason: MinimumReplicasAvailable
			  status: "True"
			  type: Available
			- lastTransitionTime: "2025-03-18T13:49:18Z"
			  lastUpdateTime: "2025-03-18T13:49:18Z"
			  message: ReplicaSet "keystone-5d9c965546" has timed out progressing.
			  reason: ProgressDeadlineExceeded
			  status: "False"
			  type: Progressing
		*/

		deployment.Status.Conditions = []appsv1.DeploymentCondition{
			{
				Message: "Deployment has minimum availability",
				Reason:  "MinimumReplicasAvailable",
				Status:  corev1.ConditionTrue,
				Type:    appsv1.DeploymentAvailable,
			},
			{
				Message: fmt.Sprintf("ReplicaSet \"%s-869cb5d44c\" has timed out progressing.", deployment.Name),
				Reason:  "ProgressDeadlineExceeded",
				Status:  corev1.ConditionFalse,
				Type:    appsv1.DeploymentProgressing,
			}}

		tc.Logger.Info("Simulated Deployment ProgressDeadlineExceeded", "ObservedGeneration", deployment.Status.ObservedGeneration)
		tc.Logger.Info("Simulated Deployment ProgressDeadlineExceeded", "Generation", deployment.Generation)

		g.Expect(tc.K8sClient.Status().Update(tc.Ctx, deployment)).To(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	tc.Logger.Info("Simulated Deployment ProgressDeadlineExceeded", "on", name)
}
