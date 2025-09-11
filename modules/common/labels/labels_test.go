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

package labels

import (
	"testing"

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetGroupLabel(t *testing.T) {
	t.Run("Get group label", func(t *testing.T) {
		g := NewWithT(t)

		gl := GetGroupLabel("foo")

		g.Expect(gl).To(BeIdenticalTo("foo.openstack.org"))
	})
}

func TestGetOwnerUIDLabelSelector(t *testing.T) {
	t.Run("Get owner uid label selector", func(t *testing.T) {
		g := NewWithT(t)

		gl := GetGroupLabel("foo")
		g.Expect(gl).To(BeIdenticalTo("foo.openstack.org"))

		ls := GetOwnerUIDLabelSelector(gl)
		g.Expect(ls).To(BeIdenticalTo("foo.openstack.org/uid"))
	})
}

func TestGetOwnerNameSpaceLabelSelector(t *testing.T) {
	t.Run("Get owner namespace label selector", func(t *testing.T) {
		g := NewWithT(t)

		gl := GetGroupLabel("foo")
		g.Expect(gl).To(BeIdenticalTo("foo.openstack.org"))

		ls := GetOwnerNameSpaceLabelSelector(gl)
		g.Expect(ls).To(BeIdenticalTo("foo.openstack.org/namespace"))
	})
}

func TestGetOwnerNameLabelSelector(t *testing.T) {
	t.Run("Get owner name label selector", func(t *testing.T) {
		g := NewWithT(t)

		gl := GetGroupLabel("foo")
		g.Expect(gl).To(BeIdenticalTo("foo.openstack.org"))

		ls := GetOwnerNameLabelSelector(gl)
		g.Expect(ls).To(BeIdenticalTo("foo.openstack.org/name"))
	})
}

func TestGetLabels(t *testing.T) {

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "podname",
			Namespace: "podnamespace",
			UID:       "11111111-1111-1111-1111-111111111111",
		},
	}

	tests := []struct {
		name   string
		labels map[string]string
		want   map[string]string
	}{
		{
			name:   "Get default labels",
			labels: map[string]string{},
			want: map[string]string{
				"foo.openstack.org/uid":       "11111111-1111-1111-1111-111111111111",
				"foo.openstack.org/namespace": "podnamespace",
				"foo.openstack.org/name":      "podname",
			},
		},
		{
			name: "Get default + additional custom labels",
			labels: map[string]string{
				"customlabel": "value",
			},
			want: map[string]string{
				"customlabel":                 "value",
				"foo.openstack.org/uid":       "11111111-1111-1111-1111-111111111111",
				"foo.openstack.org/namespace": "podnamespace",
				"foo.openstack.org/name":      "podname",
			},
		},
	}

	gl := GetGroupLabel("foo")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(gl).To(BeIdenticalTo("foo.openstack.org"))

			l := GetLabels(pod, gl, tt.labels)

			g.Expect(l).To(HaveLen(len(tt.want)))
			g.Expect(l).To(BeEquivalentTo(tt.want))
		})
	}
}

// Given a map[string]string, get the corresponding labelSelectors and compare
// them via the EqualLabelSelectors utility
func TestEqualLabelSelectors(t *testing.T) {
	t.Run("Compare labelSelectors", func(t *testing.T) {
		g := NewWithT(t)

		l0 := GetLabelSelector(map[string]string{})
		l1 := GetLabelSelector(map[string]string{"app": "foo", "version": "v1", "property": "bar"})
		l2 := l1
		l3 := GetLabelSelector(map[string]string{"app": "api", "version": "v1"})

		g.Expect(EqualLabelSelectors(l1, l0)).To(BeFalse())
		g.Expect(EqualLabelSelectors(l1, l2)).To(BeTrue())
		g.Expect(EqualLabelSelectors(l1, l3)).To(BeFalse())
	})
}
