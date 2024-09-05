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
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("tls package", func() {
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

	It("validates CA cert secret", func() {
		sname := types.NamespacedName{
			Name:      "ca",
			Namespace: namespace,
		}
		th.CreateEmptySecret(sname)

		// validate bad ca cert secret
		_, err := tls.ValidateCACertSecret(th.Ctx, cClient, sname)
		Expect(err).To(HaveOccurred())

		// update ca cert secret with good data
		th.UpdateSecret(sname, tls.CABundleKey, []byte("foo"))
		hash, err := tls.ValidateCACertSecret(th.Ctx, cClient, sname)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(hash).To(BeIdenticalTo("n56fh645hfbh687hc9h678h87h64bh598h577hch5d6h5c9h5d4h74h84h5f4hfch6dh678h547h9bhbchb6h89h5c4h68dhc9h664h557h595h5c5q"))
	})

	It("validates service cert secret", func() {
		sname := types.NamespacedName{
			Name:      "cert",
			Namespace: namespace,
		}

		// create bad cert secret
		th.CreateEmptySecret(sname)

		// validate bad cert secret
		s := &tls.Service{
			SecretName: sname.Name,
		}
		_, err := s.ValidateCertSecret(th.Ctx, h, namespace)
		Expect(err).To(HaveOccurred())

		// update cert secret with cert, still key missing
		th.UpdateSecret(sname, tls.CertKey, []byte("cert"))
		_, err = s.ValidateCertSecret(th.Ctx, h, namespace)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("field tls.key not found in Secret"))

		// update cert secret with key to be a good cert secret
		th.UpdateSecret(sname, tls.PrivateKey, []byte("key"))

		// validate good cert secret
		hash, err := s.ValidateCertSecret(th.Ctx, h, namespace)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(hash).To(BeIdenticalTo("n547h97h5cfh587h56ch594h79hd4h96h5cfh565h587h569h688h666h685h67ch7fhfbh664h5f9h694h564h9ch645h675h665h78h7h87h566hb6q"))
	})

	It("validates endpoint certs secrets", func() {
		sname := types.NamespacedName{
			Name:      "cert",
			Namespace: namespace,
		}
		// create bad cert secret
		th.CreateSecret(sname, map[string][]byte{
			tls.PrivateKey: []byte("key"),
		})

		endpointCfgs := map[service.Endpoint]tls.Service{}

		// validate empty service map
		_, err := tls.ValidateEndpointCerts(th.Ctx, h, namespace, endpointCfgs)
		Expect(err).ToNot(HaveOccurred())

		endpointCfgs[service.EndpointInternal] = tls.Service{
			SecretName: sname.Name,
		}
		endpointCfgs[service.EndpointPublic] = tls.Service{
			SecretName: sname.Name,
		}

		// validate service map with bad cert secret
		_, err = tls.ValidateEndpointCerts(th.Ctx, h, namespace, endpointCfgs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("field tls.crt not found in Secret"))

		// update cert secret to have missing private key
		th.UpdateSecret(sname, tls.CertKey, []byte("cert"))

		// validate service map with good cert secret
		hash, err := tls.ValidateEndpointCerts(th.Ctx, h, namespace, endpointCfgs)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(hash).To(BeIdenticalTo("n5d7h65dh5d5h569hffh66ch568h95h686h58fhcfh586h5b8hc6hd7h65bh56bh55bh656hfh5f7h84h54bh65dh5c9h8ch64bh64bhdfh8ch589h54bq"))
	})
})
