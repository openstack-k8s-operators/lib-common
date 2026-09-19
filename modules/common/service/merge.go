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

package service

import (
	corev1 "k8s.io/api/core/v1"
)

// MergeServicePorts merges desired service port specs into existing ports
// matched by name. It starts from the desired port and preserves only the
// server-defaulted fields (Protocol, TargetPort) from the existing port when
// the desired spec doesn't explicitly set them. All other fields come from the
// desired spec.
//
// When port counts differ or a desired port name is not found in existing, the
// existing slice is replaced with the desired ports.
func MergeServicePorts(existing *[]corev1.ServicePort, desired []corev1.ServicePort) {
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
		// Preserve server-defaulted fields from the existing port
		// only when the desired spec doesn't explicitly set them.
		if d.Protocol == "" {
			d.Protocol = (*existing)[idx].Protocol
		}
		if d.TargetPort.IntValue() == 0 && d.TargetPort.StrVal == "" {
			d.TargetPort = (*existing)[idx].TargetPort
		}
		(*existing)[idx] = d
	}
}
