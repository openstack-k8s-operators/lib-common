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

package condition

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Severity expresses the severity of a Condition Type failing.
type Severity string

const (
	// SeverityError specifies that a condition with `Status=False` is an error.
	SeverityError Severity = "Error"

	// SeverityWarning specifies that a condition with `Status=False` is a warning.
	SeverityWarning Severity = "Warning"

	// SeverityInfo specifies that a condition with `Status=False` is informative.
	SeverityInfo Severity = "Info"

	// SeverityNone should apply only to conditions with `Status=True`.
	SeverityNone Severity = ""
)

// Type - A summarizing name for a given condition
// ---
// Many .condition.type values are consistent across resources like Available, but because arbitrary conditions can be
// useful (see .node.status.conditions), the ability to deconflict is important.
// The regex it matches is (dns1123SubdomainFmt/)?(qualifiedNameFmt)
// +kubebuilder:validation:Pattern=`^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$`
// +kubebuilder:validation:MaxLength=316
type Type string

// Reason - Why a particular condition is true, false or unknown
// reason contains a programmatic identifier indicating the reason for the condition's last transition.
// Producers of specific condition types may define expected values and meanings for this field,
// and whether the values are considered a guaranteed API.
// The value should be a CamelCase string.
// This field may not be empty.
// +kubebuilder:validation:MaxLength=1024
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:Pattern=`^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z0-9_])?$`
type Reason string

// Condition contains details for one aspect of the current state of this API Resource.
// ---
// This struct is intended for direct use as an array at the field path .status.conditions.  For example,
//
//	type FooStatus struct{
//	    // Represents the observations of a foo's current state.
//	    // Known .status.conditions.type are: "Available", "Progressing", and "Degraded"
//	    // +patchMergeKey=type
//	    // +patchStrategy=merge
//	    // +listType=map
//	    // +listMapKey=type
//	    Conditions []condition.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
//
//	    // other fields
//	}
type Condition struct {
	// Type of condition in CamelCase.
	// +required
	// +kubebuilder:validation:Required
	Type Type `json:"type"`

	// status of the condition, one of True, False, Unknown.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=True;False;Unknown
	Status corev1.ConditionStatus `json:"status"`

	// Severity provides a classification of Reason code, so the current situation is immediately
	// understandable and could act accordingly.
	// It is meant for situations where Status=False and it should be indicated if it is just
	// informational, warning (next reconciliation might fix it) or an error (e.g. DB create issue
	// and no actions to automatically resolve the issue can/should be done).
	// For conditions where Status=Unknown or Status=True the Severity should be SeverityNone.
	Severity Severity `json:"severity,omitempty"`

	// lastTransitionTime is the last time the condition transitioned from one status to another.
	// This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=date-time
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`

	// The reason for the condition's last transition in CamelCase.
	// +optional
	Reason Reason `json:"reason,omitempty"`

	// message is a human readable message indicating details about the transition.
	// This may be an empty string.
	// +optional
	// +kubebuilder:validation:MaxLength=32768
	Message string `json:"message,omitempty"`
}

// Conditions provide observations of the operational state of a API resource.
// +patchMergeKey=type
// +patchStrategy=merge
// +listType=map
// +listMapKey=type
type Conditions []Condition

// conditionGroup defines a group of conditions with the same status and severity,
type conditionGroup struct {
	status     corev1.ConditionStatus
	severity   Severity
	conditions Conditions
}
