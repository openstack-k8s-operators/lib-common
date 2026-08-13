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

package volume

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
)

func TestWritableDirVolumeNoSizeLimit(t *testing.T) {
	v := WritableDirVolume("run-httpd")
	if v.Name != "run-httpd" {
		t.Errorf("expected Name %q, got %q", "run-httpd", v.Name)
	}
	if v.EmptyDir == nil {
		t.Fatal("expected EmptyDir to be set")
	}
	if v.EmptyDir.SizeLimit != nil {
		t.Errorf("expected no SizeLimit, got %v", v.EmptyDir.SizeLimit)
	}
}

func TestWritableDirVolumeWithSizeLimit(t *testing.T) {
	limit := resource.MustParse("64Mi")
	v := WritableDirVolume("home", &limit)
	if v.EmptyDir == nil || v.EmptyDir.SizeLimit == nil {
		t.Fatal("expected SizeLimit to be set")
	}
	if v.EmptyDir.SizeLimit.String() != "64Mi" {
		t.Errorf("expected SizeLimit 64Mi, got %v", v.EmptyDir.SizeLimit)
	}
}

func TestWritableDirVolumeMount(t *testing.T) {
	m := WritableDirVolumeMount("run-httpd", "/run/httpd")
	if m.Name != "run-httpd" {
		t.Errorf("expected Name %q, got %q", "run-httpd", m.Name)
	}
	if m.MountPath != "/run/httpd" {
		t.Errorf("expected MountPath %q, got %q", "/run/httpd", m.MountPath)
	}
}

func TestWritableDirSubPathMounts(t *testing.T) {
	mounts := WritableDirSubPathMounts("home", "/var/lib/service", "tmp", ".cache")
	if len(mounts) != 2 {
		t.Fatalf("expected 2 mounts, got %d", len(mounts))
	}
	if mounts[0].MountPath != "/var/lib/service/tmp" || mounts[0].SubPath != "tmp" {
		t.Errorf("unexpected first mount: %+v", mounts[0])
	}
	if mounts[1].MountPath != "/var/lib/service/.cache" || mounts[1].SubPath != ".cache" {
		t.Errorf("unexpected second mount: %+v", mounts[1])
	}
}

func TestWritableDirSubPathMountsNoSubdirs(t *testing.T) {
	mounts := WritableDirSubPathMounts("home", "/var/lib/service")
	if len(mounts) != 0 {
		t.Errorf("expected no mounts, got %v", mounts)
	}
}

func TestConfigSecretVolumes(t *testing.T) {
	vols, mounts := ConfigSecretVolumes([]string{"secret-a", "secret-b"})
	if len(vols) != 2 || len(mounts) != 2 {
		t.Fatalf("expected 2 volumes and 2 mounts, got %d and %d", len(vols), len(mounts))
	}
	if vols[0].Secret == nil || vols[0].Secret.SecretName != "secret-a" {
		t.Errorf("expected SecretName %q", "secret-a")
	}
	if vols[0].Secret.DefaultMode == nil || *vols[0].Secret.DefaultMode != 0440 {
		t.Errorf("expected DefaultMode 0440, got %v", vols[0].Secret.DefaultMode)
	}
	if mounts[0].MountPath != "/var/lib/config-data/secret-0" {
		t.Errorf("expected MountPath %q, got %q", "/var/lib/config-data/secret-0", mounts[0].MountPath)
	}
	if !mounts[0].ReadOnly {
		t.Error("expected ReadOnly to be true")
	}
}

func TestConfigSecretVolumesEmpty(t *testing.T) {
	vols, mounts := ConfigSecretVolumes(nil)
	if len(vols) != 0 || len(mounts) != 0 {
		t.Errorf("expected empty results, got %d volumes, %d mounts", len(vols), len(mounts))
	}
}

func TestConstants(t *testing.T) {
	tests := []struct {
		name, volume, path string
	}{
		{"RunHttpd", RunHttpdVolumeName, RunHttpdMountPath},
		{"VarLogHttpd", VarLogHttpdVolumeName, VarLogHttpdMountPath},
		{"Tmp", TmpVolumeName, TmpMountPath},
	}
	for _, tt := range tests {
		v := WritableDirVolume(tt.volume)
		m := WritableDirVolumeMount(tt.volume, tt.path)
		if v.Name != tt.volume {
			t.Errorf("%s: volume Name = %q, want %q", tt.name, v.Name, tt.volume)
		}
		if m.MountPath != tt.path {
			t.Errorf("%s: mount Path = %q, want %q", tt.name, m.MountPath, tt.path)
		}
	}
}
