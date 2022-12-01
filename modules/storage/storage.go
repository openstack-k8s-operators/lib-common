/*
Copyright 2022 Red Hat

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

package storage

import (
	corev1 "k8s.io/api/core/v1"
)

// ExtraVolType represents a "label" that can be optionally added to the VolMounts
// instance
type ExtraVolType string

// PropagationType identifies the Service, Group or instance (e.g. the backend) that
// receives an Extra Volume that can potentially be mounted
type PropagationType string

const (
	// PropagationEverywhere is used to define a propagation policy that allows
	// to get the volumes mounted to all the OpenStack services
	PropagationEverywhere PropagationType = "All"
	// DBSync represents a common ServiceType defined by the OpenStack operators
	// that keeps track of the DBSync pod
	DBSync PropagationType = "DBSync"
	// Compute represents a common ServiceType that can be translated into an
	// external-data-plane related propagation policy
	Compute PropagationType = "Compute"
)

// VolMounts is the data structure used to expose Volumes and Mounts that can
// be added to a pod according to the defined Propagation policy
type VolMounts struct {
	// +kubebuilder:validation:type={PropagationEverywhere}
	// Propagation defines which pod should mount the volume
	Propagation []PropagationType `json:"propagation,omitempty"`
	// Label associated to a given extraMount
	// +kubebuilder:validation:Optional
	ExtraVolType ExtraVolType `json:"extraVolType,omitempty"`
	// +kubebuilder:validation:Required
	Volumes []corev1.Volume `json:"volumes"`
	// +kubebuilder:validation:Required
	Mounts []corev1.VolumeMount `json:"mounts"`
}

// Propagate allows services to filter and mount extra volumes according to
// the specified policy
func (v *VolMounts) Propagate(svc []PropagationType) []VolMounts {

	var vl []VolMounts

	// if propagation is not specified the defined volumes are mounted
	// by default: this allows operators that don't require propagation
	// to rely on this feature without adopting the propagation part
	if len(v.Propagation) == 0 {
		vl = append(vl, VolMounts{
			Volumes: v.Volumes,
			Mounts:  v.Mounts,
		})
	}

	for _, p := range v.Propagation {
		if canPropagate(p, svc) {
			vl = append(vl, VolMounts{
				Volumes: v.Volumes,
				Mounts:  v.Mounts,
			})
		}
	}

	return vl
}
