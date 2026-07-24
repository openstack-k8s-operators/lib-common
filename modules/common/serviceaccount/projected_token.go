/*
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

package serviceaccount

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

const (
	// KubeAPIAccessVolumeName is the name of the projected volume that
	// replaces the default automounted service account token.
	KubeAPIAccessVolumeName = "kube-api-access"

	// KubeAPIAccessMountPath is the standard mount path for the projected
	// service account token volume, matching the default automount path.
	KubeAPIAccessMountPath = "/var/run/secrets/kubernetes.io/serviceaccount"

	// DefaultTokenExpirationSeconds is the default expiration for the
	// projected service account token (1 hour).
	DefaultTokenExpirationSeconds int64 = 3600
)

// KubeAPIAccessVolume returns a projected Volume that provides a
// time-limited service account token, the cluster CA certificate, and
// the pod namespace.  Use it together with AutomountServiceAccountToken=false
// on the PodSpec to replace the default long-lived automounted token with a
// short-lived one. An optional expirationSeconds overrides the default
// token lifetime (1 hour).
func KubeAPIAccessVolume(expirationSeconds ...int64) corev1.Volume {
	expiration := DefaultTokenExpirationSeconds
	if len(expirationSeconds) > 0 && expirationSeconds[0] > 0 {
		expiration = expirationSeconds[0]
	}
	return corev1.Volume{
		Name: KubeAPIAccessVolumeName,
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				DefaultMode: ptr.To[int32](0444),
				Sources: []corev1.VolumeProjection{
					{
						ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
							Path:              "token",
							ExpirationSeconds: ptr.To(expiration),
						},
					},
					{
						ConfigMap: &corev1.ConfigMapProjection{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "kube-root-ca.crt",
							},
							Items: []corev1.KeyToPath{
								{
									Key:  "ca.crt",
									Path: "ca.crt",
								},
							},
						},
					},
					{
						DownwardAPI: &corev1.DownwardAPIProjection{
							Items: []corev1.DownwardAPIVolumeFile{
								{
									Path: "namespace",
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.namespace",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// KubeAPIAccessVolumeMount returns a read-only VolumeMount for the
// projected service account token volume at the standard automount path.
func KubeAPIAccessVolumeMount() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      KubeAPIAccessVolumeName,
		MountPath: KubeAPIAccessMountPath,
		ReadOnly:  true,
	}
}
