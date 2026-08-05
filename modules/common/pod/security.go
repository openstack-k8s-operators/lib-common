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
// suitable for unprivileged workloads. It sets RunAsUser to uid, RunAsGroup
// to gid, RunAsNonRoot, drops all capabilities, disables privilege escalation,
// and applies the RuntimeDefault seccomp profile.
// Optional addCapabilities are added back after dropping ALL.
//
// Does not set ReadOnlyRootFilesystem -- a future RestrictiveReadOnlySecurityContext
// is the intended way for an individual service to opt into that later, without
// changing this function's behavior for every existing caller.
func RestrictiveSecurityContext(uid, gid int64, addCapabilities ...corev1.Capability) *corev1.SecurityContext {
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
		AllowPrivilegeEscalation: ptr.To(false),
		Capabilities:             caps,
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
	}
}

// RestrictivePodSecurityContext returns a hardened PodSecurityContext for
// unprivileged workloads. It sets RunAsUser to uid, RunAsGroup and FSGroup
// to gid, RunAsNonRoot, and applies the RuntimeDefault seccomp profile.
// FSGroup ensures that volumes mounted from Secrets/ConfigMaps are
// group-readable by the service process without needing chown.
//
// Optional supplementalGroups grant additional GIDs to the pod — use this
// when the workload needs to read files not covered by FSGroup, e.g.
// RPM-shipped configs baked into the container image with restrictive
// group ownership rather than mounted from a Secret/ConfigMap.
func RestrictivePodSecurityContext(uid, gid int64, supplementalGroups ...int64) *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{
		RunAsUser:          ptr.To(uid),
		RunAsGroup:         ptr.To(gid),
		RunAsNonRoot:       ptr.To(true),
		FSGroup:            ptr.To(gid),
		SupplementalGroups: supplementalGroups,
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
	}
}

// RestrictiveSecurityContextWithGID is an alias for RestrictiveSecurityContext.
// Deprecated: use RestrictiveSecurityContext directly.
func RestrictiveSecurityContextWithGID(uid, gid int64, addCapabilities ...corev1.Capability) *corev1.SecurityContext {
	return RestrictiveSecurityContext(uid, gid, addCapabilities...)
}

// PrivilegedSecurityContext returns a SecurityContext for a workload that
// needs full Privileged access to the host (e.g. LVM/iSCSI/multipath device
// management via nsenter'd host binaries) and therefore cannot use
// RestrictiveSecurityContext — Privileged is incompatible with
// ReadOnlyRootFilesystem and capability dropping. RunAsUser is set to uid,
// RunAsGroup to gid.
func PrivilegedSecurityContext(uid, gid int64) *corev1.SecurityContext {
	return &corev1.SecurityContext{
		RunAsUser:    ptr.To(uid),
		RunAsGroup:   ptr.To(gid),
		RunAsNonRoot: ptr.To(true),
		Privileged:   ptr.To(true),
	}
}
