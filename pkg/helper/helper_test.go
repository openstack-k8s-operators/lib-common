/*
Copyright 2020 Red Hat

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

package helper

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	keystonev1 "github.com/openstack-k8s-operators/keystone-operator/api/v1beta1"
)

func TestToUnstructured(t *testing.T) {
	t.Run("with a typed object", func(t *testing.T) {
		g := NewWithT(t)
		// Test with a typed object.
		obj := &keystonev1.KeystoneAPI{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "keystone",
				Namespace: "openstack",
			},
			Spec: keystonev1.KeystoneAPISpec{
				DatabaseHostname: "dbhost",
			},
		}
		newObj, err := ToUnstructured(obj)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(newObj.GetName()).To(Equal(obj.Name))
		g.Expect(newObj.GetNamespace()).To(Equal(obj.Namespace))

		// Change a spec field and validate that it stays the same in the incoming object.
		g.Expect(unstructured.SetNestedField(newObj.Object, "dbhost1", "spec", "databaseHostname")).To(Succeed())
		g.Expect(obj.Spec.DatabaseHostname).To(Equal("dbhost"))
	})

	t.Run("with an unstructured object", func(t *testing.T) {
		g := NewWithT(t)

		obj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "test.x.y.z/v1",
				"metadata": map[string]interface{}{
					"name":      "keystone",
					"namespace": "openstack",
				},
				"spec": map[string]interface{}{
					"databaseHostname": "dbhost",
				},
			},
		}

		newObj, err := ToUnstructured(obj)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(newObj.GetName()).To(Equal(obj.GetName()))
		g.Expect(newObj.GetNamespace()).To(Equal(obj.GetNamespace()))

		// Validate that the maps point to different addresses.
		g.Expect(obj.Object).ToNot(BeIdenticalTo(newObj.Object))

		// Change a spec field and validate that it stays the same in the incoming object.
		g.Expect(unstructured.SetNestedField(newObj.Object, "dbhost1", "spec", "databaseHostname")).To(Succeed())
		dbHostValue, _, err := unstructured.NestedString(obj.Object, "spec", "databaseHostname")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(dbHostValue).To(Equal("dbhost"))

		// Change the name of the new object and make sure it doesn't change it the old one.
		newObj.SetName("keystone-1")
		g.Expect(obj.GetName()).To(Equal("keystone"))
	})
}
