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
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openstack-k8s-operators/lib-common/modules/certmanager"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

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

	It("creates certificates for k8s services with label selector", func() {
		i := certmanager.NewIssuer(
			certmanager.CAIssuer(
				"ca",
				names.Namespace,
				map[string]string{"f": "l"},
				"secret",
			),
			timeout,
		)

		_, err := i.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		issuer := th.GetIssuer(names.CAName)
		Expect(issuer.Spec.CA).NotTo(BeNil())

		svc1Name := types.NamespacedName{Name: "svc1", Namespace: names.Namespace}
		svc2Name := types.NamespacedName{Name: "svc2", Namespace: names.Namespace}
		svc3Name := types.NamespacedName{Name: "svc3", Namespace: names.Namespace}
		th.CreateService(svc1Name, map[string]string{"foo": ""}, corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     svc1Name.Name,
					Port:     int32(1111),
					Protocol: corev1.ProtocolTCP,
				},
			},
		})
		th.CreateService(svc2Name, map[string]string{"foo": ""}, corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     svc2Name.Name,
					Port:     int32(2222),
					Protocol: corev1.ProtocolTCP,
				},
			},
		})
		th.CreateService(svc3Name, map[string]string{}, corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     svc2Name.Name,
					Port:     int32(3333),
					Protocol: corev1.ProtocolTCP,
				},
			},
		})
		// simulate underlying cert secrets exist
		th.CreateCertSecret(types.NamespacedName{Name: "cert-svc1-svc", Namespace: names.Namespace})
		th.CreateCertSecret(types.NamespacedName{Name: "cert-svc2-svc", Namespace: names.Namespace})

		certs, _, err := certmanager.EnsureCertForServicesWithSelector(
			th.Ctx, h, names.Namespace, map[string]string{"foo": ""}, names.CAName.Name)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(certs).To(HaveLen(2))
		Expect(certs).To(HaveKey(fmt.Sprintf("svc1.%s.svc", names.Namespace)))
		Expect(certs).To(HaveKey(fmt.Sprintf("svc2.%s.svc", names.Namespace)))
	})

	It("creates a certificate for a specific k8s service matching label selector", func() {
		i := certmanager.NewIssuer(
			certmanager.CAIssuer(
				"ca",
				names.Namespace,
				map[string]string{"f": "l"},
				"secret",
			),
			timeout,
		)

		_, err := i.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		issuer := th.GetIssuer(names.CAName)
		Expect(issuer.Spec.CA).NotTo(BeNil())

		svc1Name := types.NamespacedName{Name: "svc1", Namespace: names.Namespace}
		svc2Name := types.NamespacedName{Name: "svc2", Namespace: names.Namespace}
		svc3Name := types.NamespacedName{Name: "svc3", Namespace: names.Namespace}
		th.CreateService(svc1Name, map[string]string{"foo": "1"}, corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     svc1Name.Name,
					Port:     int32(1111),
					Protocol: corev1.ProtocolTCP,
				},
			},
		})
		th.CreateService(svc2Name, map[string]string{"foo": "2"}, corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     svc2Name.Name,
					Port:     int32(2222),
					Protocol: corev1.ProtocolTCP,
				},
			},
		})
		th.CreateService(svc3Name, map[string]string{}, corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     svc2Name.Name,
					Port:     int32(3333),
					Protocol: corev1.ProtocolTCP,
				},
			},
		})
		// simulate underlying cert secret exist
		th.CreateCertSecret(types.NamespacedName{Name: "cert-svc2-svc", Namespace: names.Namespace})

		cert, _, err := certmanager.EnsureCertForServiceWithSelector(
			th.Ctx, h, names.Namespace, map[string]string{"foo": "2"}, names.CAName.Name)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(cert).To(Equal("cert-svc2-svc"))

	})

	It("fails to create a certificate for a specific k8s service if the label selector returns not a single service", func() {
		i := certmanager.NewIssuer(
			certmanager.CAIssuer(
				"ca",
				names.Namespace,
				map[string]string{"f": "l"},
				"secret",
			),
			timeout,
		)

		_, err := i.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		issuer := th.GetIssuer(names.CAName)
		Expect(issuer.Spec.CA).NotTo(BeNil())

		svc1Name := types.NamespacedName{Name: "svc1", Namespace: names.Namespace}
		svc2Name := types.NamespacedName{Name: "svc2", Namespace: names.Namespace}
		svc3Name := types.NamespacedName{Name: "svc3", Namespace: names.Namespace}
		th.CreateService(svc1Name, map[string]string{"foo": ""}, corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     svc1Name.Name,
					Port:     int32(1111),
					Protocol: corev1.ProtocolTCP,
				},
			},
		})
		th.CreateService(svc2Name, map[string]string{"foo": ""}, corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     svc2Name.Name,
					Port:     int32(2222),
					Protocol: corev1.ProtocolTCP,
				},
			},
		})
		th.CreateService(svc3Name, map[string]string{}, corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     svc2Name.Name,
					Port:     int32(3333),
					Protocol: corev1.ProtocolTCP,
				},
			},
		})

		_, _, err = certmanager.EnsureCertForServiceWithSelector(
			th.Ctx, h, names.Namespace, map[string]string{"foo": ""}, names.CAName.Name)
		Expect(err).To(HaveOccurred())
	})
})
