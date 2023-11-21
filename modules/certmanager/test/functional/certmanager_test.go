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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openstack-k8s-operators/lib-common/modules/certmanager"

	certmgrv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmgrmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
)

var _ = Describe("certmanager module", func() {
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
		Eventually(func(g Gomega) {
			issuer, err := certmanager.GetIssuerByName(
				th.Ctx,
				h,
				names.SelfSignedIssuerName.Name,
				names.SelfSignedIssuerName.Namespace)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(issuer.Spec.SelfSigned).NotTo(BeNil())
			g.Expect(issuer.ObjectMeta.Labels["f"]).To(Equal("l"))
		}, timeout, interval).Should(Succeed())
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
		issuer := th.GetIssuer(names.CAName)
		Expect(issuer.Spec.CA).NotTo(BeNil())
		Expect(issuer.Spec.CA.SecretName).To(Equal("secret"))
		Expect(issuer.Labels["f"]).To(Equal("l"))
	})

	It("deletes issuer", func() {
		i := certmanager.NewIssuer(
			certmanager.CAIssuer(
				names.IssuerName.Name,
				names.IssuerName.Namespace,
				map[string]string{"f": "l"},
				"secret",
			),
			timeout,
		)

		_, err := i.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		issuer := th.GetIssuer(names.IssuerName)
		Expect(issuer).NotTo(BeNil())
		err = i.Delete(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		th.AssertIssuerDoesNotExist(names.IssuerName)
	})

	It("creates certificate", func() {
		c := certmanager.NewCertificate(
			certmanager.Cert(
				names.CertName.Name,
				names.CertName.Namespace,
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
		cert := th.GetCert(names.CertName)
		Expect(cert.Spec.CommonName).To(Equal("keystone-public-openstack.apps-crc.testing"))
		Expect(cert.Spec.SecretName).To(Equal("secret"))
		Expect(cert.Labels["f"]).To(Equal("l"))
	})

	It("deletes certificate", func() {
		c := certmanager.NewCertificate(
			certmanager.Cert(
				names.CertName.Name,
				names.CertName.Namespace,
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
		cert := th.GetCert(names.CertName)
		Expect(cert).NotTo(BeNil())
		err = c.Delete(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		th.AssertIssuerDoesNotExist(names.CertName)
	})
})
