/*
Copyright 2025 Red Hat

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

package backup

// Restore order constants (gaps of 10 allow insertion)
const (
	// RestoreOrder00 is PVCs - storage foundation
	RestoreOrder00 = "00"
	// RestoreOrder10 is NADs, Secrets, ConfigMaps - foundation resources
	RestoreOrder10 = "10"
	// RestoreOrder20 is OpenStackVersion - restored before ControlPlane
	RestoreOrder20 = "20"
	// RestoreOrder30 is OpenStackControlPlane - restored after Version
	RestoreOrder30 = "30"
	// RestoreOrder40 is backup config and user resources
	RestoreOrder40 = "40"
	// RestoreOrder50 is manual steps - database/RabbitMQ restore, resume deployment
	RestoreOrder50 = "50"
	// RestoreOrder60 is OpenStackDataPlaneNodeSet - restored after ControlPlane
	RestoreOrder60 = "60"
)

// Category constants for backup/restore scope
const (
	// CategoryControlPlane identifies control plane resources
	CategoryControlPlane = "controlplane"
	// CategoryDataPlane identifies data plane resources
	CategoryDataPlane = "dataplane"
)
