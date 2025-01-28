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

package topology

// TopoRef - it models a Topology reference and it can be included in the
// service operators API. It is used to retrieve the referenced Topology
type TopoRef struct {
	// +kubebuilder:validation:Optional
	// Name - The Topology CR name that the Service references
	Name string `json:"name"`

	// +kubebuilder:validation:Optional
	// Namespace - The Namespace to fetch the Topology CR referenced
	// NOTE: Namespace currently points by default to the same namespace where
	// the Service is deployed. Customizing the namespace is not supported and
	// webhooks prevent editing this field to a value different from the
	// current project
	Namespace string `json:"namespace,omitempty"`
}
