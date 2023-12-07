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
	"github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
)

// AssertVolumeExists - asserts the existence of a named volume in a []corev1.Volume.
//
// Example usage:
//
//	th.AssertVolumeExists("name", []corev1.Volume{...})
func (tc *TestHelper) AssertVolumeExists(name string, volumes []corev1.Volume) {
	gomega.Expect(volumes).To(gomega.ContainElement(gomega.HaveField("Name", name)))
}

// AssertVolumeMountExists - asserts the existence of a named volumeMount with a subPath (if provided) in a []corev1.VolumeMount.
//
// Example usage:
//
//	th.AssertVolumeMountExists("name", "subPath", []corev1.VolumeMount{...})
func (tc *TestHelper) AssertVolumeMountExists(name string, subPath string, volumeMounts []corev1.VolumeMount) bool {
	exist := false
	for _, v := range volumeMounts {
		if v.Name == name {
			if subPath != "" {
				if v.SubPath == subPath {
					exist = true
				}
			} else {
				exist = true
			}
		}
	}
	gomega.Expect(exist).To(gomega.BeTrue())

	return exist
}
