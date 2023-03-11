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

package ceph

// Defaults is used as type to reference defaults
type Defaults string

// Ceph defaults
const (
	DefaultUser             Defaults = "openstack"
	DefaultCinderPool       Defaults = "volumes"
	DefaultCinderBackupPool Defaults = "backups"
	DefaultNovaPool         Defaults = "vms"
	DefaultGlancePool       Defaults = "images"
	CError                  Defaults = ""
)

// Backend defines the Ceph client parameters
type Backend struct {
	// +kubebuilder:validation:Required
	// ClusterFSID defines the fsid
	ClusterFSID string `json:"cephFsid"`
	// +kubebuilder:validation:Required
	// ClusterMons defines the commma separated mon list
	ClusterMonHosts string `json:"cephMons"`
	// +kubebuilder:validation:Required
	// ClientKey set the Ceph cluster key
	ClientKey string `json:"cephClientKey"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="CephUser"
	// User set the Ceph cluster pool
	User string `json:"cephUser"`
	// +kubebuilder:validation:Optional
	// Pools - Map of chosen names to spec definitions for the Ceph cluster
	// pools
	Pools map[string]PoolSpec `json:"cephPools,omitempty"`
}

// PoolSpec defines the Ceph pool Spec parameters
type PoolSpec struct {
	// +kubebuilder:validation:Required
	// PoolName defines the name of the pool
	PoolName string `json:"name"`
}
