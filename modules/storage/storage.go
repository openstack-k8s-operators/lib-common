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
// +kubebuilder:object:generate:=true

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

// VolumeSource our slimmed down version of the VolumeSource struct with deprecated and "removed" fields removed to save space
type VolumeSource struct {
	HostPath *corev1.HostPathVolumeSource `json:"hostPath,omitempty" protobuf:"bytes,1,opt,name=hostPath"`
	// emptyDir represents a temporary directory that shares a pod's lifetime.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes#emptydir
	// +optional
	EmptyDir *corev1.EmptyDirVolumeSource `json:"emptyDir,omitempty" protobuf:"bytes,2,opt,name=emptyDir"`
	// secret represents a secret that should populate this volume.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes#secret
	// +optional
	Secret *corev1.SecretVolumeSource `json:"secret,omitempty" protobuf:"bytes,6,opt,name=secret"`
	// nfs represents an NFS mount on the host that shares a pod's lifetime
	// More info: https://kubernetes.io/docs/concepts/storage/volumes#nfs
	// +optional
	NFS *corev1.NFSVolumeSource `json:"nfs,omitempty" protobuf:"bytes,7,opt,name=nfs"`
	// iscsi represents an ISCSI Disk resource that is attached to a
	// kubelet's host machine and then exposed to the pod.
	// More info: https://examples.k8s.io/volumes/iscsi/README.md
	// +optional
	ISCSI *corev1.ISCSIVolumeSource `json:"iscsi,omitempty" protobuf:"bytes,8,opt,name=iscsi"`
	// persistentVolumeClaimVolumeSource represents a reference to a
	// PersistentVolumeClaim in the same namespace.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims
	// +optional
	PersistentVolumeClaim *corev1.PersistentVolumeClaimVolumeSource `json:"persistentVolumeClaim,omitempty" protobuf:"bytes,10,opt,name=persistentVolumeClaim"`
	// cephFS represents a Ceph FS mount on the host that shares a pod's lifetime
	// +optional
	CephFS *corev1.CephFSVolumeSource `json:"cephfs,omitempty" protobuf:"bytes,14,opt,name=cephfs"`
	// downwardAPI represents downward API about the pod that should populate this volume
	// +optional
	DownwardAPI *corev1.DownwardAPIVolumeSource `json:"downwardAPI,omitempty" protobuf:"bytes,16,opt,name=downwardAPI"`
	// fc represents a Fibre Channel resource that is attached to a kubelet's host machine and then exposed to the pod.
	// +optional
	FC *corev1.FCVolumeSource `json:"fc,omitempty" protobuf:"bytes,17,opt,name=fc"`
	// configMap represents a configMap that should populate this volume
	// +optional
	ConfigMap *corev1.ConfigMapVolumeSource `json:"configMap,omitempty" protobuf:"bytes,19,opt,name=configMap"`
	// photonPersistentDisk represents a PhotonController persistent disk attached and mounted on kubelets host machine
	PhotonPersistentDisk *corev1.PhotonPersistentDiskVolumeSource `json:"photonPersistentDisk,omitempty" protobuf:"bytes,23,opt,name=photonPersistentDisk"`
	// projected items for all in one resources secrets, configmaps, and downward API
	Projected *corev1.ProjectedVolumeSource `json:"projected,omitempty" protobuf:"bytes,26,opt,name=projected"`
	// scaleIO represents a ScaleIO persistent volume attached and mounted on Kubernetes nodes.
	// +optional
	ScaleIO *corev1.ScaleIOVolumeSource `json:"scaleIO,omitempty" protobuf:"bytes,25,opt,name=scaleIO"`
	// storageOS represents a StorageOS volume attached and mounted on Kubernetes nodes.
	// +optional
	StorageOS *corev1.StorageOSVolumeSource `json:"storageos,omitempty" protobuf:"bytes,27,opt,name=storageos"`
	// csi (Container Storage Interface) represents ephemeral storage that is handled by certain external CSI drivers (Beta feature).
	// +optional
	CSI *corev1.CSIVolumeSource `json:"csi,omitempty" protobuf:"bytes,28,opt,name=csi"`
	// ephemeral represents a volume that is handled by a cluster storage driver.
	// The volume's lifecycle is tied to the pod that defines it - it will be created before the pod starts,
	// and deleted when the pod is removed.
	//
	// +optional
	Ephemeral *corev1.EphemeralVolumeSource `json:"ephemeral,omitempty" protobuf:"bytes,29,opt,name=ephemeral"`
}

// Volume our slimmed down version of Volume
type Volume struct {
	// +kubebuilder:validation:Required
	// Name of the volume
	Name string `json:"name"`
	// +kubebuilder:validation:Required
	// VolumeSource defines the source of a volume to be mounted
	VolumeSource VolumeSource `json:"volumeSource"`
}

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
	Volumes []Volume `json:"volumes"`
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

// ConvertVolumeSource function to convert from a VolumeSource to a corev1.VolumeSource
func ConvertVolumeSource(v *VolumeSource) corev1.VolumeSource {
	return corev1.VolumeSource{
		HostPath:              v.HostPath,
		EmptyDir:              v.EmptyDir,
		Secret:                v.Secret,
		NFS:                   v.NFS,
		ISCSI:                 v.ISCSI,
		PersistentVolumeClaim: v.PersistentVolumeClaim,
		CephFS:                v.CephFS,
		DownwardAPI:           v.DownwardAPI,
		FC:                    v.FC,
		ConfigMap:             v.ConfigMap,
		PhotonPersistentDisk:  v.PhotonPersistentDisk,
		Projected:             v.Projected,
		ScaleIO:               v.ScaleIO,
		StorageOS:             v.StorageOS,
		CSI:                   v.CSI,
		Ephemeral:             v.Ephemeral,
	}
}
