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
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// List - A list of conditions
type List []Condition

// Condition - A particular overall condition of a certain resource
type Condition struct {
	Type               Type                   `json:"type"`
	Status             corev1.ConditionStatus `json:"status"`
	Reason             Reason                 `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
	LastHeartbeatTime  metav1.Time            `json:"lastHearbeatTime,omitempty"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty"`
}

// Type - A summarizing name for a given condition
type Type string

// Reason - Why a particular condition is true, false or unknown
type Reason string

// NewCondition - Create a new condition object
func NewCondition(
	conditionType Type,
	status corev1.ConditionStatus,
	reason Reason,
	message string,
) Condition {
	now := metav1.NewTime(time.Now().UTC().Truncate(time.Second))
	condition := Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastHeartbeatTime:  now,
		LastTransitionTime: now,
	}
	return condition
}

// Set - Set a particular condition in a given condition list
func (conditions *List) set(c *Condition) {
	// Check for the existence of a particular condition type in a list of conditions
	// and change it only if there is a status change
	exists := false
	for i := range *conditions {
		existingCondition := (*conditions)[i]
		if existingCondition.Type == c.Type {
			exists = true
			if !hasSameState(&existingCondition, c) {
				(*conditions)[i] = *c
				break
			}
			c.LastTransitionTime = existingCondition.LastTransitionTime
			break
		}
	}

	// If the condition does not exist, add it, setting the transition time only if not already set
	if !exists {
		*conditions = append(*conditions, *c)
	}
}

// hasSameState returns true if a condition has the same state of another
func hasSameState(i, j *Condition) bool {
	return i.Type == j.Type &&
		i.Status == j.Status &&
		i.Reason == j.Reason &&
		i.Message == j.Message
}

// GetCurrentCondition - Get current condition with status == corev1.ConditionTrue
func (conditions List) getCurrentCondition() *Condition {
	for i, cond := range conditions {
		if cond.Status == corev1.ConditionTrue {
			return &conditions[i]
		}
	}

	return nil
}

// UpdateCurrentCondition - sets new condition, and sets previous condition to corev1.ConditionFalse
func (conditions *List) UpdateCurrentCondition(c Condition) {
	// if it is an empty condition, just return and don't set it
	if c == (Condition{}) {
		return
	}

	//
	// get current condition and update to corev1.ConditionFalse
	//
	currentCondition := conditions.getCurrentCondition()
	if currentCondition != nil {
		currentCondition.Status = corev1.ConditionFalse
		conditions.set(currentCondition)
	}

	//
	// set new condition
	//
	conditions.set(&c)
}
