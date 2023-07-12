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
	"encoding/json"
	"fmt"

	networkv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	t "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// GetStatefulSet - retrieves a StatefulSet resource.
//
// example usage:
//
//	th.GetStatefulSet(types.NamespacedName{Name: "test-statefulset", Namespace: "test-namespace"})
func (tc *TestHelper) GetStatefulSet(name types.NamespacedName) *appsv1.StatefulSet {
	ss := &appsv1.StatefulSet{}
	t.Eventually(func(g t.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, ss)).Should(t.Succeed())
	}, tc.Timeout, tc.Interval).Should(t.Succeed())
	return ss
}

// SimulateStatefulSetReplicaReady retrieves the StatefulSet  and simulates
// a ready state for a StatefulSet's replica in a Kubernetes cluster.
//
// example usage:
//
//	th.SimulateStatefulSetReplicaReady(types.NamespacedName{Name: "test-statefulset", Namespace: "test-namespace"})
func (tc *TestHelper) SimulateStatefulSetReplicaReady(name types.NamespacedName) {
	t.Eventually(func(g t.Gomega) {
		ss := tc.GetStatefulSet(name)
		ss.Status.Replicas = 1
		ss.Status.ReadyReplicas = 1
		g.Expect(tc.K8sClient.Status().Update(tc.Ctx, ss)).To(t.Succeed())

	}, tc.Timeout, tc.Interval).Should(t.Succeed())
	tc.Logger.Info("Simulated statefulset success", "on", name)
}

// SimulateStatefulSetReplicaReadyWithPods simulates a StatefulSet with ready replicas
// by creating and updating the corresponding Pods.
//
// example usage:
//
//		th.SimulateStatefulSetReplicaReadyWithPods(
//	 	cell0.ConductorStatefulSetName,
//	 	map[string][]string{cell0.CellName.Namespace + "/internalapi": {"10.0.0.1"}},
//	 )
func (tc *TestHelper) SimulateStatefulSetReplicaReadyWithPods(name types.NamespacedName, networkIPs map[string][]string) {
	ss := tc.GetStatefulSet(name)
	for i := 0; i < int(*ss.Spec.Replicas); i++ {
		pod := &corev1.Pod{
			ObjectMeta: ss.Spec.Template.ObjectMeta,
			Spec:       ss.Spec.Template.Spec,
		}
		pod.ObjectMeta.Namespace = name.Namespace
		pod.ObjectMeta.Name = fmt.Sprintf("%s-%d", name.Name, i)

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
		t.Expect(err).NotTo(t.HaveOccurred())
		pod.Annotations[networkv1.NetworkStatusAnnot] = string(netStatusAnnotation)

		t.Expect(tc.K8sClient.Create(tc.Ctx, pod)).Should(t.Succeed())
	}

	t.Eventually(func(g t.Gomega) {
		ss := tc.GetStatefulSet(name)
		ss.Status.Replicas = 1
		ss.Status.ReadyReplicas = 1
		g.Expect(tc.K8sClient.Status().Update(tc.Ctx, ss)).To(t.Succeed())

	}, tc.Timeout, tc.Interval).Should(t.Succeed())

	tc.Logger.Info("Simulated statefulset success", "on", name)
}
