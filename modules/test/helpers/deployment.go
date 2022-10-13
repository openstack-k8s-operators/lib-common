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
	"time"

	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
)

// GetDeployment -
func GetDeployment(name types.NamespacedName, timeout time.Duration, interval time.Duration) *appsv1.Deployment {
	deployment := &appsv1.Deployment{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(k8sClient.Get(ctx, name, deployment)).Should(gomega.Succeed())
	}, timeout, interval).Should(gomega.Succeed())

	return deployment
}

// ListDeployments -
func ListDeployments(namespace string) *appsv1.DeploymentList {
	deployments := &appsv1.DeploymentList{}
	gomega.Expect(k8sClient.List(ctx, deployments, client.InNamespace(namespace))).Should(gomega.Succeed())

	return deployments
}

// SimulateDeploymentReplicaReady -
func SimulateDeploymentReplicaReady(name types.NamespacedName, timeout time.Duration, interval time.Duration) {
	deployment := GetDeployment(name, timeout, interval)
	// NOTE(gibi): We don't need to do this when run against a real
	// env as there the deployment could reach the ready state automatically.
	// But for that we would need another set of test setup, i.e. deploying
	// the mariadb-operator.

	deployment.Status.Replicas = 1
	deployment.Status.ReadyReplicas = 1
	gomega.Expect(k8sClient.Status().Update(ctx, deployment)).To(gomega.Succeed())
}
