/*
Copyright 2024 Red Hat
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
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	networkv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
)

// GetNAD - retrieves a NetworkAttachmentDefinition resource.
//
// Example usage:
//
//	th.GetNAD(types.NamespacedName{Name: "test-nad", Namespace: "test-namespace"})
func (tc *TestHelper) GetNAD(name types.NamespacedName) *networkv1.NetworkAttachmentDefinition {
	nad := &networkv1.NetworkAttachmentDefinition{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, nad)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
	return nad
}

// CreateNAD creates a new NetworkAttachmentDefinition resource with the provided spec.
//
// Example usage:
//
//	spec := map[string]interface{}{"key": "value"}
//	p := th.CreateNAD(types.NamespacedName{Namespace: "default", Name: "example"}, spec)
func (tc *TestHelper) CreateNAD(name types.NamespacedName, spec map[string]interface{}) client.Object {
	raw := map[string]interface{}{
		"apiVersion": "k8s.cni.cncf.io/v1",
		"kind":       "NetworkAttachmentDefinition",
		"metadata": map[string]interface{}{
			"name":      name.Name,
			"namespace": name.Namespace,
		},
		"spec": spec,
	}
	return tc.CreateUnstructured(raw)
}
