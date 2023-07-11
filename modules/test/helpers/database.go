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
	mariadbv1 "github.com/openstack-k8s-operators/mariadb-operator/api/v1beta1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"
)

// CreateDBService creates a k8s Service object that matches with the
// Expectations of lib-common database module as a Service for the MariaDB
func (tc *TestHelper) CreateDBService(namespace string, mariadbCRName string, spec corev1.ServiceSpec) types.NamespacedName {
	// The Name is used as the hostname to access the service. So
	// we generate something unique for the MariaDB CR it represents
	// so we can assert that the correct Service is selected.
	serviceName := "hostname-for-" + mariadbCRName
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
			// NOTE(gibi): The lib-common databvase module looks up the
			// Service exposed by MariaDB via these labels.
			Labels: map[string]string{
				"app": "mariadb",
				"cr":  "mariadb-" + mariadbCRName,
			},
		},
		Spec: spec,
	}
	t.Expect(tc.K8sClient.Create(tc.Ctx, service)).Should(t.Succeed())

	return types.NamespacedName{Name: serviceName, Namespace: namespace}
}

// DeleteDBService The function deletes the Service if exists and wait for it to disappear from the API.
// If the Service does not exists then it is assumed to be successfully deleted.
// Example:
//
//	th.DeleteDBService(types.NamespacedName{Name: "my-service", Namespace: "my-namespace"})
//
// or:
//
//	DeferCleanup(th.DeleteDBService, th.CreateDBService(cell0.MariaDBDatabaseName.Namespace, cell0.MariaDBDatabaseName.Name, serviceSpec))
func (tc *TestHelper) DeleteDBService(name types.NamespacedName) {
	t.Eventually(func(g t.Gomega) {
		service := &corev1.Service{}
		err := tc.K8sClient.Get(tc.Ctx, name, service)
		// if it is already gone that is OK
		if k8s_errors.IsNotFound(err) {
			return
		}
		g.Expect(err).NotTo(t.HaveOccurred())

		g.Expect(tc.K8sClient.Delete(tc.Ctx, service)).Should(t.Succeed())

		err = tc.K8sClient.Get(tc.Ctx, name, service)
		g.Expect(k8s_errors.IsNotFound(err)).To(t.BeTrue())
	}, tc.Timeout, tc.Interval).Should(t.Succeed())
}

// GetMariaDBDatabase waits for and retrieves a MariaDBDatabase resource from the Kubernetes cluster
//
// Example:
//
//	mariadbDatabase := th.GetMariaDBDatabase(types.NamespacedName{Name: "my-mariadb-database", Namespace: "my-namespace"})
func (tc *TestHelper) GetMariaDBDatabase(name types.NamespacedName) *mariadbv1.MariaDBDatabase {
	instance := &mariadbv1.MariaDBDatabase{}
	t.Eventually(func(g t.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, instance)).Should(t.Succeed())
	}, tc.Timeout, tc.Interval).Should(t.Succeed())
	return instance
}

// SimulateMariaDBDatabaseCompleted simulates a completed state for a MariaDBDatabase resource in a Kubernetes cluster.
//
// Example:
//
//	th.SimulateMariaDBDatabaseCompleted(types.NamespacedName{Name: "my-mariadb-database", Namespace: "my-namespace"})
//
// or
//
//	DeferCleanup(th.SimulateMariaDBDatabaseCompleted, types.NamespacedName{Name: "my-mariadb-database", Namespace: "my-namespace"})
func (tc *TestHelper) SimulateMariaDBDatabaseCompleted(name types.NamespacedName) {
	t.Eventually(func(g t.Gomega) {
		db := tc.GetMariaDBDatabase(name)
		db.Status.Completed = true
		// This can return conflict so we have the t.Eventually block to retry
		g.Expect(tc.K8sClient.Status().Update(tc.Ctx, db)).To(t.Succeed())

	}, tc.Timeout, tc.Interval).Should(t.Succeed())

	tc.Logger.Info("Simulated DB completed", "on", name)
}
