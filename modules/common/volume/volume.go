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
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	// RunHttpdVolumeName is the standard volume name for the httpd PID file directory.
	RunHttpdVolumeName = "run-httpd"
	// RunHttpdMountPath is the canonical mount path for the httpd PID directory.
	RunHttpdMountPath = "/run/httpd"
	// VarLogHttpdVolumeName is the standard volume name for the httpd log directory.
	VarLogHttpdVolumeName = "var-log-httpd"
	// VarLogHttpdMountPath is the mount path for the httpd log directory.
	VarLogHttpdMountPath = "/var/log/httpd"
	// TmpVolumeName is the standard volume name for /tmp.
	TmpVolumeName = "tmp"
	// TmpMountPath is the mount path for /tmp.
	TmpMountPath = "/tmp"
	// HomeDirCacheSubdir is the ".cache" subdirectory under a service's
	// home directory. RHEL's python3-setuptools downstream patch caches
	// iter_entry_points() scans under $HOME/.cache/python-entrypoints/
	// on every process start. Needs a writable emptyDir mount when
	// ReadOnlyRootFilesystem is enabled.
	HomeDirCacheSubdir = ".cache"
)

var configSecretMode int32 = 0440

// WritableDirVolume returns an emptyDir Volume. Used for any path that needs
// to be writable by a non-root service user: /run/httpd (PID file),
// /var/log/httpd, /var/log/<service>, /tmp, service home-dir subdirs, etc.
// Pass an optional sizeLimit to cap the volume's ephemeral storage.
func WritableDirVolume(name string, sizeLimit ...*resource.Quantity) corev1.Volume {
	var limit *resource.Quantity
	if len(sizeLimit) > 0 {
		limit = sizeLimit[0]
	}
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{SizeLimit: limit},
		},
	}
}

// WritableDirVolumeMount returns a VolumeMount for a writable emptyDir.
func WritableDirVolumeMount(name, mountPath string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      name,
		MountPath: mountPath,
	}
}

// WritableDirSubPathMounts returns SubPath VolumeMounts onto the emptyDir
// named volumeName for the given subdirectories of baseDir. Mounted via
// SubPath rather than shadowing baseDir itself, since the image may bake
// real content there (e.g. shell dotfiles under a service home directory).
func WritableDirSubPathMounts(volumeName, baseDir string, subdirs ...string) []corev1.VolumeMount {
	mounts := make([]corev1.VolumeMount, 0, len(subdirs))
	for _, subdir := range subdirs {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: baseDir + "/" + subdir,
			SubPath:   subdir,
		})
	}
	return mounts
}

// ConfigSecretVolumes returns Volumes and VolumeMounts for a list of Secret
// names, each mounted read-only at /var/lib/config-data/secret-{idx} with
// DefaultMode 0440.
func ConfigSecretVolumes(secretNames []string) ([]corev1.Volume, []corev1.VolumeMount) {
	volumes := make([]corev1.Volume, 0, len(secretNames))
	mounts := make([]corev1.VolumeMount, 0, len(secretNames))

	for idx, secretName := range secretNames {
		volumes = append(volumes, corev1.Volume{
			Name: secretName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  secretName,
					DefaultMode: &configSecretMode,
				},
			},
		})
		mounts = append(mounts, corev1.VolumeMount{
			Name:      secretName,
			MountPath: "/var/lib/config-data/secret-" + strconv.Itoa(idx),
			ReadOnly:  true,
		})
	}

	return volumes, mounts
}
