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

// AssertVolumeMountPathExists - Returns true if mountPath and subPath exist in
// the volumeMounts array.
// Assumptions:
//   - an empty []corev1.VolumeMount results in a failure
//   - an empty "name" results in a failure
//   - when no mountPath and no subPath are passed (empty string) it results in a
//     failure
//   - if no mountPath is provided it is skipped during the evaluation and
//   - if no subPath is provided it is skipped during the evaluation
//   - when both mountPath and subPath are passed, it returns true only when
//     both are found in the volumeMount array
func (tc *TestHelper) AssertVolumeMountPathExists(
	name string,
	mountPath string,
	subPath string,
	volumeMounts []corev1.VolumeMount,
) bool {
	// Early return for an empty volumeMounts list
	if len(volumeMounts) == 0 {
		gomega.Expect(false).To(gomega.BeTrue(), "Volume mounts list is empty")
		return false
	}

	// Early return for an empty VolumeMount name
	if name == "" {
		gomega.Expect(false).To(gomega.BeTrue(), "Volume name cannot be empty")
		return false
	}

	// Early return if both mountPath and subPath are empty strings
	if mountPath == "" && subPath == "" {
		gomega.Expect(false).To(gomega.BeTrue(), "Both mountPath and subPath are empty")
		return false
	}

	// init two bool variables to represent the status of mountPath and
	// subPath check: when the function reaches this point we have at least an
	// input that should be checked
	mountPathExists := mountPath == ""
	subPathExists := subPath == ""

	for _, v := range volumeMounts {
		// only process the current volumeMount if the name
		// mathes
		if v.Name == name {
			// mountPath is passed as input and it has not
			// been evaluated
			if v.MountPath == mountPath && !mountPathExists {
				mountPathExists = true
			}
			// subPath is passed as input and it has not
			// been evaluated
			if v.SubPath == subPath && !subPathExists {
				subPathExists = true
			}
		}
		// check if we can break early and optimize the
		// iterations
		if mountPathExists && subPathExists {
			break
		}
	}
	gomega.Expect(mountPathExists && subPathExists).To(gomega.BeTrue())
	return (mountPathExists && subPathExists)
}
