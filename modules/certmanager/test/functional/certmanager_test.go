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
package functional

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openstack-k8s-operators/lib-common/modules/certmanager"

	certmgrv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmgrmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("certmanager module", func() {
	var namespace string

	BeforeEach(func() {
		// NOTE(gibi): We need to create a unique namespace for each test run
		// as namespaces cannot be deleted in a locally running envtest. See
		// https://book.kubebuilder.io/reference/envtest.html#namespace-usage-limitation
		namespace = uuid.New().String()
		th.CreateNamespace(namespace)
		// We still request the delete of the Namespace to properly cleanup if
		// we run the test in an existing cluster.
		DeferCleanup(th.DeleteNamespace, namespace)

	})

	It("creates selfsigned issuer", func() {
		i := certmanager.NewIssuer(
			certmanager.SelfSignedIssuer(
				"selfsigned",
				namespace,
				map[string]string{"f": "l"},
			),
			timeout,
		)

		_, err := i.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		issuer := th.GetIssuer(types.NamespacedName{Namespace: namespace, Name: "selfsigned"})
		Expect(issuer.Spec.SelfSigned).NotTo(BeNil())
		Expect(issuer.Labels["f"]).To(Equal("l"))

	})

	It("creates CA issuer", func() {
		i := certmanager.NewIssuer(
			certmanager.CAIssuer(
				"ca",
				namespace,
				map[string]string{"f": "l"},
				"secret",
			),
			timeout,
		)

		_, err := i.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		issuer := th.GetIssuer(types.NamespacedName{Namespace: namespace, Name: "ca"})
		Expect(issuer.Spec.CA).NotTo(BeNil())
		Expect(issuer.Spec.CA.SecretName).To(Equal("secret"))
		Expect(issuer.Labels["f"]).To(Equal("l"))
	})

	It("deletes issuer", func() {
		issuerName := types.NamespacedName{Namespace: namespace, Name: "issuer"}
		i := certmanager.NewIssuer(
			certmanager.CAIssuer(
				issuerName.Name,
				issuerName.Namespace,
				map[string]string{"f": "l"},
				"secret",
			),
			timeout,
		)

		_, err := i.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		issuer := th.GetIssuer(issuerName)
		Expect(issuer).NotTo(BeNil())
		err = i.Delete(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		th.AssertIssuerDoesNotExist(issuerName)
	})

	It("creates certificate", func() {
		c := certmanager.NewCertificate(
			certmanager.Cert(
				"cert",
				namespace,
				map[string]string{"f": "l"},
				certmgrv1.CertificateSpec{
					CommonName: "keystone-public-openstack.apps-crc.testing",
					DNSNames: []string{
						"keystone-public-openstack",
						"keystone-public-openstack.apps-crc.testing",
					},
					IssuerRef: certmgrmetav1.ObjectReference{
						Kind: "Issuer",
						Name: "issuerName",
					},
					SecretName: "secret",
				},
			),
			timeout,
		)

		_, err := c.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		cert := th.GetCert(types.NamespacedName{Namespace: namespace, Name: "cert"})
		Expect(cert.Spec.CommonName).To(Equal("keystone-public-openstack.apps-crc.testing"))
		Expect(cert.Spec.SecretName).To(Equal("secret"))
		Expect(cert.Labels["f"]).To(Equal("l"))
	})

	It("deletes certificate", func() {
		certName := types.NamespacedName{Namespace: namespace, Name: "cert"}
		c := certmanager.NewCertificate(
			certmanager.Cert(
				certName.Name,
				certName.Namespace,
				map[string]string{"f": "l"},
				certmgrv1.CertificateSpec{
					CommonName: "keystone-public-openstack.apps-crc.testing",
					DNSNames: []string{
						"keystone-public-openstack",
						"keystone-public-openstack.apps-crc.testing",
					},
					IssuerRef: certmgrmetav1.ObjectReference{
						Kind: "Issuer",
						Name: "issuerName",
					},
					SecretName: "secret",
				},
			),
			timeout,
		)

		_, err := c.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		cert := th.GetCert(certName)
		Expect(cert).NotTo(BeNil())
		err = c.Delete(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		th.AssertIssuerDoesNotExist(certName)
	})
})
