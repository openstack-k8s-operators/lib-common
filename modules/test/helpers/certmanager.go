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
	certmgrv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"

	"github.com/onsi/gomega"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// GetIssuer waits for and retrieves a Issuer resource from the Kubernetes cluster
//
// Example:
//
//	issuer := th.GetIssuer(types.NamespacedName{Name: "my-issuer", Namespace: "my-namespace"})
func (tc *TestHelper) GetIssuer(name types.NamespacedName) *certmgrv1.Issuer {
	instance := &certmgrv1.Issuer{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, instance)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
	return instance
}

// AssertIssuerDoesNotExist ensures the Issuer resource does not exist in a k8s cluster.
func (tc *TestHelper) AssertIssuerDoesNotExist(name types.NamespacedName) {
	instance := &certmgrv1.Issuer{}
	gomega.Eventually(func(g gomega.Gomega) {
		err := tc.K8sClient.Get(tc.Ctx, name, instance)
		g.Expect(k8s_errors.IsNotFound(err)).To(gomega.BeTrue())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
}

// GetCert waits for and retrieves a Certificate resource from the Kubernetes cluster
//
// Example:
//
//	cert := th.GetCert(types.NamespacedName{Name: "my-issuer", Namespace: "my-namespace"})
func (tc *TestHelper) GetCert(name types.NamespacedName) *certmgrv1.Certificate {
	instance := &certmgrv1.Certificate{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, instance)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
	return instance
}

// AssertCertDoesNotExist ensures the Certificate resource does not exist in a k8s cluster.
func (tc *TestHelper) AssertCertDoesNotExist(name types.NamespacedName) {
	instance := &certmgrv1.Certificate{}
	gomega.Eventually(func(g gomega.Gomega) {
		err := tc.K8sClient.Get(tc.Ctx, name, instance)
		g.Expect(k8s_errors.IsNotFound(err)).To(gomega.BeTrue())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
}
