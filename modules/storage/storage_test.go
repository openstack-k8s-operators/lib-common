/*
Copyright 2026 Red Hat

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

package storage

import (
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func TestPropagate(t *testing.T) {
	sampleVolMounts := VolMounts{
		Propagation: []PropagationType{"Neutron"},
		Volumes: []Volume{
			{Name: "extra-vol"},
		},
		Mounts: []corev1.VolumeMount{
			{Name: "extra-vol", MountPath: "/extra"},
		},
	}

	t.Run("Empty propagation returns one copy by default", func(t *testing.T) {
		g := NewWithT(t)
		vm := sampleVolMounts
		vm.Propagation = nil
		vl := vm.Propagate([]PropagationType{"Nova"})
		g.Expect(vl).To(HaveLen(1))
		g.Expect(vl[0].Volumes).To(Equal(vm.Volumes))
		g.Expect(vl[0].Mounts).To(Equal(vm.Mounts))
	})

	t.Run("Single matching selector returns one copy", func(t *testing.T) {
		g := NewWithT(t)
		vl := sampleVolMounts.Propagate([]PropagationType{"Neutron"})
		g.Expect(vl).To(HaveLen(1))
		g.Expect(vl[0].Volumes).To(Equal(sampleVolMounts.Volumes))
		g.Expect(vl[0].Mounts).To(Equal(sampleVolMounts.Mounts))
	})

	t.Run("Multiple selectors with one match return one copy", func(t *testing.T) {
		g := NewWithT(t)
		vm := sampleVolMounts
		vm.Propagation = []PropagationType{"Nova", "Neutron"}
		vl := vm.Propagate([]PropagationType{"Neutron"})
		g.Expect(vl).To(HaveLen(1))
	})

	t.Run("Multiple matching selectors return one copy", func(t *testing.T) {
		g := NewWithT(t)
		vm := sampleVolMounts
		vm.Propagation = []PropagationType{PropagationEverywhere, "Neutron"}
		vl := vm.Propagate([]PropagationType{"Neutron"})
		g.Expect(vl).To(HaveLen(1))
	})

	t.Run("Multiple matching selectors against multiple services return one copy", func(t *testing.T) {
		g := NewWithT(t)
		vm := sampleVolMounts
		vm.Propagation = []PropagationType{"Neutron", "Ironic"}
		vl := vm.Propagate([]PropagationType{"Ironic", "Neutron"})
		g.Expect(vl).To(HaveLen(1))
	})

	t.Run("No matching selector returns no copy", func(t *testing.T) {
		g := NewWithT(t)
		vl := sampleVolMounts.Propagate([]PropagationType{"Nova"})
		g.Expect(vl).To(BeEmpty())
	})
}
