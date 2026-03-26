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

package statefulset

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestMergeContainersByName(t *testing.T) {
	tests := []struct {
		name     string
		existing []corev1.Container
		desired  []corev1.Container
		verify   func(t *testing.T, containers []corev1.Container)
	}{
		{
			name: "successful merge preserves server defaults",
			existing: []corev1.Container{
				{
					Name:                     "app",
					Image:                    "old-image:v1",
					TerminationMessagePath:   "/dev/termination-log",
					TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					ImagePullPolicy:          corev1.PullIfNotPresent,
				},
			},
			desired: []corev1.Container{
				{
					Name:  "app",
					Image: "new-image:v2",
					Env: []corev1.EnvVar{
						{Name: "FOO", Value: "bar"},
					},
				},
			},
			verify: func(t *testing.T, containers []corev1.Container) {
				c := containers[0]
				if c.Image != "new-image:v2" {
					t.Errorf("Image = %q, want %q", c.Image, "new-image:v2")
				}
				if len(c.Env) != 1 || c.Env[0].Name != "FOO" {
					t.Errorf("Env not merged correctly")
				}
				// Server defaults should be preserved
				if c.TerminationMessagePath != "/dev/termination-log" {
					t.Errorf("TerminationMessagePath lost: %q", c.TerminationMessagePath)
				}
				if c.TerminationMessagePolicy != corev1.TerminationMessageReadFile {
					t.Errorf("TerminationMessagePolicy lost: %v", c.TerminationMessagePolicy)
				}
				if c.ImagePullPolicy != corev1.PullIfNotPresent {
					t.Errorf("ImagePullPolicy lost: %v", c.ImagePullPolicy)
				}
			},
		},
		{
			name: "multi-container merge by name not order",
			existing: []corev1.Container{
				{Name: "sidecar", Image: "sidecar:v1", ImagePullPolicy: corev1.PullAlways},
				{Name: "main", Image: "main:v1", ImagePullPolicy: corev1.PullIfNotPresent},
			},
			desired: []corev1.Container{
				{Name: "main", Image: "main:v2"},
				{Name: "sidecar", Image: "sidecar:v2"},
			},
			verify: func(t *testing.T, containers []corev1.Container) {
				// Order should be preserved (existing order)
				if containers[0].Name != "sidecar" || containers[0].Image != "sidecar:v2" {
					t.Errorf("sidecar not merged: %+v", containers[0])
				}
				if containers[1].Name != "main" || containers[1].Image != "main:v2" {
					t.Errorf("main not merged: %+v", containers[1])
				}
				// ImagePullPolicy preserved
				if containers[0].ImagePullPolicy != corev1.PullAlways {
					t.Errorf("sidecar ImagePullPolicy lost")
				}
				if containers[1].ImagePullPolicy != corev1.PullIfNotPresent {
					t.Errorf("main ImagePullPolicy lost")
				}
			},
		},
		{
			name: "merges all operator-controlled fields",
			existing: []corev1.Container{
				{
					Name:            "app",
					Image:           "old:v1",
					ImagePullPolicy: corev1.PullAlways,
				},
			},
			desired: []corev1.Container{
				{
					Name:    "app",
					Image:   "new:v2",
					Command: []string{"/bin/sh"},
					Args:    []string{"-c", "echo"},
					Env:     []corev1.EnvVar{{Name: "K", Value: "V"}},
					Ports:   []corev1.ContainerPort{{ContainerPort: 8080}},
					VolumeMounts: []corev1.VolumeMount{
						{Name: "data", MountPath: "/data"},
					},
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("100m"),
						},
					},
					LivenessProbe:   &corev1.Probe{InitialDelaySeconds: 5},
					ReadinessProbe:  &corev1.Probe{InitialDelaySeconds: 10},
					Lifecycle:       &corev1.Lifecycle{},
					SecurityContext: &corev1.SecurityContext{},
				},
			},
			verify: func(t *testing.T, containers []corev1.Container) {
				c := containers[0]
				if c.Image != "new:v2" {
					t.Errorf("Image not merged")
				}
				if len(c.Command) != 1 || c.Command[0] != "/bin/sh" {
					t.Errorf("Command not merged")
				}
				if len(c.Args) != 2 {
					t.Errorf("Args not merged")
				}
				if len(c.Env) != 1 {
					t.Errorf("Env not merged")
				}
				if len(c.Ports) != 1 {
					t.Errorf("Ports not merged")
				}
				if len(c.VolumeMounts) != 1 {
					t.Errorf("VolumeMounts not merged")
				}
				if c.Resources.Limits == nil {
					t.Errorf("Resources not merged")
				}
				if c.LivenessProbe == nil {
					t.Errorf("LivenessProbe not merged")
				}
				if c.ReadinessProbe == nil {
					t.Errorf("ReadinessProbe not merged")
				}
				if c.Lifecycle == nil {
					t.Errorf("Lifecycle not merged")
				}
				if c.SecurityContext == nil {
					t.Errorf("SecurityContext not merged")
				}
				// Server default preserved
				if c.ImagePullPolicy != corev1.PullAlways {
					t.Errorf("ImagePullPolicy lost")
				}
			},
		},
		{
			name: "desired fields override existing for StartupProbe, WorkingDir, EnvFrom",
			existing: []corev1.Container{
				{
					Name:            "app",
					Image:           "old:v1",
					ImagePullPolicy: corev1.PullAlways,
					StartupProbe:    &corev1.Probe{InitialDelaySeconds: 15},
					WorkingDir:      "/old/dir",
					EnvFrom: []corev1.EnvFromSource{
						{Prefix: "OLD_"},
					},
				},
			},
			desired: []corev1.Container{
				{
					Name:         "app",
					Image:        "new:v2",
					StartupProbe: &corev1.Probe{InitialDelaySeconds: 30},
					WorkingDir:   "/new/dir",
					EnvFrom: []corev1.EnvFromSource{
						{Prefix: "NEW_"},
					},
				},
			},
			verify: func(t *testing.T, containers []corev1.Container) {
				c := containers[0]
				if c.Image != "new:v2" {
					t.Errorf("Image not merged")
				}
				if c.StartupProbe == nil || c.StartupProbe.InitialDelaySeconds != 30 {
					t.Errorf("StartupProbe should come from desired, got %v", c.StartupProbe)
				}
				if c.WorkingDir != "/new/dir" {
					t.Errorf("WorkingDir should come from desired, got %q", c.WorkingDir)
				}
				if len(c.EnvFrom) != 1 || c.EnvFrom[0].Prefix != "NEW_" {
					t.Errorf("EnvFrom should come from desired, got %v", c.EnvFrom)
				}
				if c.ImagePullPolicy != corev1.PullAlways {
					t.Errorf("ImagePullPolicy should be preserved from existing")
				}
			},
		},
		{
			name: "desired without optional fields clears them from existing",
			existing: []corev1.Container{
				{
					Name:            "app",
					Image:           "old:v1",
					ImagePullPolicy: corev1.PullAlways,
					StartupProbe:    &corev1.Probe{InitialDelaySeconds: 15},
					WorkingDir:      "/old/dir",
					EnvFrom:         []corev1.EnvFromSource{{Prefix: "OLD_"}},
					VolumeDevices:   []corev1.VolumeDevice{{Name: "dev", DevicePath: "/dev/xvda"}},
				},
			},
			desired: []corev1.Container{
				{
					Name:  "app",
					Image: "new:v2",
				},
			},
			verify: func(t *testing.T, containers []corev1.Container) {
				c := containers[0]
				if c.StartupProbe != nil {
					t.Errorf("StartupProbe should be nil when not in desired")
				}
				if c.WorkingDir != "" {
					t.Errorf("WorkingDir should be empty when not in desired")
				}
				if c.EnvFrom != nil {
					t.Errorf("EnvFrom should be nil when not in desired")
				}
				if c.VolumeDevices != nil {
					t.Errorf("VolumeDevices should be nil when not in desired")
				}
				if c.ImagePullPolicy != corev1.PullAlways {
					t.Errorf("ImagePullPolicy should be preserved from existing")
				}
			},
		},
		{
			name: "count mismatch falls back to replacement",
			existing: []corev1.Container{
				{Name: "app", Image: "old:v1"},
			},
			desired: []corev1.Container{
				{Name: "app", Image: "new:v2"},
				{Name: "sidecar", Image: "sidecar:v1"},
			},
			verify: func(t *testing.T, containers []corev1.Container) {
				if len(containers) != 2 {
					t.Errorf("expected 2 containers after replacement, got %d", len(containers))
				}
				if containers[0].Name != "app" || containers[0].Image != "new:v2" {
					t.Errorf("first container not replaced correctly: %+v", containers[0])
				}
				if containers[1].Name != "sidecar" || containers[1].Image != "sidecar:v1" {
					t.Errorf("second container not replaced correctly: %+v", containers[1])
				}
			},
		},
		{
			name: "name mismatch falls back to replacement",
			existing: []corev1.Container{
				{Name: "app", Image: "old:v1"},
			},
			desired: []corev1.Container{
				{Name: "different", Image: "new:v2"},
			},
			verify: func(t *testing.T, containers []corev1.Container) {
				if len(containers) != 1 || containers[0].Name != "different" {
					t.Errorf("expected replacement with desired, got %+v", containers)
				}
			},
		},
		{
			name: "desired explicit server-default fields are honored",
			existing: []corev1.Container{
				{
					Name:                     "app",
					Image:                    "old:v1",
					ImagePullPolicy:          corev1.PullIfNotPresent,
					TerminationMessagePath:   "/dev/termination-log",
					TerminationMessagePolicy: corev1.TerminationMessageReadFile,
				},
			},
			desired: []corev1.Container{
				{
					Name:                     "app",
					Image:                    "new:v2",
					ImagePullPolicy:          corev1.PullAlways,
					TerminationMessagePath:   "/custom/path",
					TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
				},
			},
			verify: func(t *testing.T, containers []corev1.Container) {
				c := containers[0]
				if c.ImagePullPolicy != corev1.PullAlways {
					t.Errorf("ImagePullPolicy = %v, want PullAlways (from desired)", c.ImagePullPolicy)
				}
				if c.TerminationMessagePath != "/custom/path" {
					t.Errorf("TerminationMessagePath = %q, want /custom/path (from desired)", c.TerminationMessagePath)
				}
				if c.TerminationMessagePolicy != corev1.TerminationMessageFallbackToLogsOnError {
					t.Errorf("TerminationMessagePolicy = %v, want FallbackToLogsOnError (from desired)", c.TerminationMessagePolicy)
				}
			},
		},
		{
			name:     "empty slices succeed",
			existing: []corev1.Container{},
			desired:  []corev1.Container{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MergeContainersByName(&tt.existing, tt.desired)
			if tt.verify != nil {
				tt.verify(t, tt.existing)
			}
		})
	}
}
