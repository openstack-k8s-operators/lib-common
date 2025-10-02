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
	. "github.com/onsi/ginkgo/v2" // nolint:revive
	. "github.com/onsi/gomega"    // nolint:revive
	"github.com/openstack-k8s-operators/lib-common/modules/common/object"

	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("object package", func() {
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

	It("now new owner gets added when adding same ownerref", func() {
		cmName := types.NamespacedName{
			Namespace: namespace,
			Name:      "test-cm",
		}

		cm := th.CreateConfigMap(cmName, map[string]any{})

		err := object.EnsureOwnerRef(th.Ctx, h, cm, cm)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(object.CheckOwnerRefExist(cm.GetUID(), cm.GetOwnerReferences())).To(BeTrue())
		Expect(cm.GetOwnerReferences()).To(HaveLen(1))
	})

	It("adds an additional owner to the ownerref list", func() {
		// create owner obj
		owner := types.NamespacedName{
			Namespace: namespace,
			Name:      "test-owner",
		}
		ownerCM := th.CreateConfigMap(owner, map[string]any{})

		// create target obj we add the owner ref to
		cmName := types.NamespacedName{
			Namespace: namespace,
			Name:      "test-cm",
		}
		cm := th.CreateConfigMap(cmName, map[string]any{})

		err := object.EnsureOwnerRef(th.Ctx, h, ownerCM, cm)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(object.CheckOwnerRefExist(ownerCM.GetUID(), cm.GetOwnerReferences())).To(BeTrue())
	})
})
