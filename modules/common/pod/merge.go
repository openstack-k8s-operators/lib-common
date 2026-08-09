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

package pod

import (
	corev1 "k8s.io/api/core/v1"
)

// MergeContainersByName merges desired container specs into existing containers
// matched by name. It starts from the desired container and preserves only the
// server-defaulted fields (TerminationMessagePath, TerminationMessagePolicy,
// ImagePullPolicy) from the existing container. All other fields come from the
// desired spec, which ensures that new fields added in future Kubernetes
// versions are not silently dropped.
//
// When container counts differ or a desired container name is not found in
// existing, the existing slice is replaced with the desired containers.
func MergeContainersByName(existing *[]corev1.Container, desired []corev1.Container) {
	if len(*existing) != len(desired) {
		*existing = desired
		return
	}

	existingByName := make(map[string]int, len(*existing))
	for i := range *existing {
		existingByName[(*existing)[i].Name] = i
	}

	for _, d := range desired {
		idx, ok := existingByName[d.Name]
		if !ok {
			*existing = desired
			return
		}
		// Preserve server-defaulted fields from the existing container
		// only when the desired spec doesn't explicitly set them.
		if d.ImagePullPolicy == "" {
			d.ImagePullPolicy = (*existing)[idx].ImagePullPolicy
		}
		if d.TerminationMessagePath == "" {
			d.TerminationMessagePath = (*existing)[idx].TerminationMessagePath
		}
		if d.TerminationMessagePolicy == "" {
			d.TerminationMessagePolicy = (*existing)[idx].TerminationMessagePolicy
		}
		(*existing)[idx] = d
	}
}
