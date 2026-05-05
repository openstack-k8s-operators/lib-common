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

package annotations

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// PausedAnnotation is set on a resource to pause reconciliation.
	// The annotation key presence is what matters; the value is ignored.
	PausedAnnotation = "openstack.org/paused"

	// ReconcileTriggerAnnotation can be set on a resource to trigger
	// a reconciliation. Changing its value forces a new reconcile event.
	// Used in multiple operators, therefore defined as a shared constant.
	ReconcileTriggerAnnotation = "openstack.org/reconcile-trigger"
)

// IsPaused returns true if the PausedAnnotation key is present on the object.
func IsPaused(obj metav1.Object) bool {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		return false
	}
	_, exists := annotations[PausedAnnotation]
	return exists
}
