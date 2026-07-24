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
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestRestrictiveSecurityContext(t *testing.T) {
	var uid int64 = 42457
	sc := RestrictiveSecurityContext(uid)

	if sc.RunAsUser == nil || *sc.RunAsUser != uid {
		t.Errorf("expected RunAsUser %d, got %v", uid, sc.RunAsUser)
	}
	if sc.RunAsGroup == nil || *sc.RunAsGroup != uid {
		t.Errorf("expected RunAsGroup %d, got %v", uid, sc.RunAsGroup)
	}
	if sc.RunAsNonRoot == nil || !*sc.RunAsNonRoot {
		t.Error("expected RunAsNonRoot true")
	}
	if sc.AllowPrivilegeEscalation == nil || *sc.AllowPrivilegeEscalation {
		t.Error("expected AllowPrivilegeEscalation false")
	}
	if sc.Capabilities == nil || len(sc.Capabilities.Drop) != 1 || sc.Capabilities.Drop[0] != "ALL" {
		t.Errorf("expected Capabilities.Drop [ALL], got %v", sc.Capabilities)
	}
	if sc.SeccompProfile == nil || sc.SeccompProfile.Type != corev1.SeccompProfileTypeRuntimeDefault {
		t.Errorf("expected SeccompProfile RuntimeDefault, got %v", sc.SeccompProfile)
	}
	if sc.ReadOnlyRootFilesystem == nil || !*sc.ReadOnlyRootFilesystem {
		t.Error("expected ReadOnlyRootFilesystem true")
	}
}

func TestRestrictiveSecurityContextWithGID(t *testing.T) {
	var uid int64 = 42415
	var gid int64 = 42416
	sc := RestrictiveSecurityContextWithGID(uid, gid)

	if sc.RunAsUser == nil || *sc.RunAsUser != uid {
		t.Errorf("expected RunAsUser %d, got %v", uid, sc.RunAsUser)
	}
	if sc.RunAsGroup == nil || *sc.RunAsGroup != gid {
		t.Errorf("expected RunAsGroup %d, got %v", gid, sc.RunAsGroup)
	}
	if sc.RunAsNonRoot == nil || !*sc.RunAsNonRoot {
		t.Error("expected RunAsNonRoot true")
	}
	if sc.AllowPrivilegeEscalation == nil || *sc.AllowPrivilegeEscalation {
		t.Error("expected AllowPrivilegeEscalation false")
	}
}

func TestRestrictiveSecurityContextNoAddCaps(t *testing.T) {
	sc := RestrictiveSecurityContext(42457)
	if sc.Capabilities.Add != nil {
		t.Errorf("expected no Add capabilities, got %v", sc.Capabilities.Add)
	}
}

func TestRestrictiveSecurityContextWithAddCaps(t *testing.T) {
	sc := RestrictiveSecurityContext(42457, "NET_BIND_SERVICE", "CHOWN")

	if len(sc.Capabilities.Drop) != 1 || sc.Capabilities.Drop[0] != "ALL" {
		t.Errorf("expected Drop [ALL], got %v", sc.Capabilities.Drop)
	}
	if len(sc.Capabilities.Add) != 2 {
		t.Fatalf("expected 2 Add capabilities, got %d", len(sc.Capabilities.Add))
	}
	if sc.Capabilities.Add[0] != "NET_BIND_SERVICE" || sc.Capabilities.Add[1] != "CHOWN" {
		t.Errorf("expected Add [NET_BIND_SERVICE CHOWN], got %v", sc.Capabilities.Add)
	}
}
