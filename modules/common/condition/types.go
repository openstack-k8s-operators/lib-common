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
type Type string

// Reason - Why a particular condition is true, false or unknown
type Reason string

// Condition defines an observation of a API resource operational state.
type Condition struct {
	// Type of condition in CamelCase.
	Type Type `json:"type"`

	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`

	// Severity provides a classification of Reason code, so the current situation is immediately
	// understandable and could act accordingly.
	// It is meant for situations where Status=False and it should be indicated if it is just
	// informational, warning (next reconciliation might fix it) or an error (e.g. DB create issue
	// and no actions to automatically resolve the issue can/should be done).
	// For conditions where Status=Unknown or Status=True the Severity should be SeverityNone.
	Severity Severity `json:"severity,omitempty"`

	// Last time the condition transitioned from one status to another.
	// This should be when the underlying condition changed. If that is not known, then using the time when
	// the API field changed is acceptable.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`

	// The reason for the condition's last transition in CamelCase.
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// Conditions provide observations of the operational state of a API resource.
type Conditions []Condition
