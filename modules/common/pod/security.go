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
	"k8s.io/utils/ptr"
)

// RestrictiveSecurityContext returns a hardened container SecurityContext
// suitable for unprivileged workloads. It sets RunAsNonRoot, drops all
// capabilities, disables privilege escalation, and applies the RuntimeDefault
// seccomp profile. The provided uid is used for both RunAsUser and RunAsGroup.
// Optional addCapabilities are added back after dropping ALL.
func RestrictiveSecurityContext(uid int64, addCapabilities ...corev1.Capability) *corev1.SecurityContext {
	return RestrictiveSecurityContextWithGID(uid, uid, addCapabilities...)
}

// RestrictiveSecurityContextWithGID is like RestrictiveSecurityContext but
// allows specifying a different GID.
func RestrictiveSecurityContextWithGID(uid, gid int64, addCapabilities ...corev1.Capability) *corev1.SecurityContext {
	caps := &corev1.Capabilities{
		Drop: []corev1.Capability{"ALL"},
	}
	if len(addCapabilities) > 0 {
		caps.Add = addCapabilities
	}
	return &corev1.SecurityContext{
		RunAsUser:                ptr.To(uid),
		RunAsGroup:               ptr.To(gid),
		RunAsNonRoot:             ptr.To(true),
		ReadOnlyRootFilesystem:   ptr.To(true),
		AllowPrivilegeEscalation: ptr.To(false),
		Capabilities:             caps,
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
	}
}
