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
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Init - init new condition list with the overall ReadyCondition set to:
// Type: ReadyCondition, Status: Unknown, Reason, Severity and Message.
//
// Optional conditions list can be passed as parameter which allows to initialize
// additional conditions at the beginning.
func (conditions *Conditions) Init(cl *Conditions) {
	conditions.Set(UnknownCondition(ReadyCondition, RequestedReason, ReadyInitMessage))

	// add all optional conditions if no not nil
	if cl != nil {
		for _, c := range *cl {
			conditions.Set(&c)
		}
	}
}

// Set - sets new condition on the conditions list.
//
// If a condition already exists, the LastTransitionTime is only updated when there is a change
// in any of the fields: Status, Reason, Severity and Message.
//
// The conditons list get sorted so that the Ready condition always goes first, followed by
// all the other Member conditions sorted by Type. This makes it easy to identify the overall
// state of the service
func (conditions *Conditions) Set(c *Condition) {
	// if it is an empty condition, just return and don't set it
	if c == nil {
		return
	}

	// set the transition time only if not already set
	if c.LastTransitionTime.IsZero() {
		c.LastTransitionTime = metav1.NewTime(time.Now().UTC().Truncate(time.Second))
	}

	// Check for the existence of a particular condition type in a list of conditions
	// and change it only if there is a status change
	exists := false
	for i, existingCondition := range *conditions {
		if existingCondition.Type == c.Type {
			exists = true
			if !hasSameState(&existingCondition, c) {
				(*conditions)[i] = *c

				break
			}
			break
		}
	}

	// If the condition does not exist, add it
	if !exists {
		*conditions = append(*conditions, *c)
	}

	// Sort conditions list
	conditions.Sort()
}

// Get returns the condition with the given type, if the condition does not exists,
// it returns nil.
func (conditions *Conditions) Get(t Type) *Condition {
	for _, condition := range *conditions {
		if condition.Type == t {
			return &condition
		}
	}
	return nil
}

// Has returns true if a condition with the given type exists.
func (conditions *Conditions) Has(t Type) bool {
	return conditions.Get(t) != nil
}

// MarkTrue sets Status=True for the condition with the given type.
func (conditions *Conditions) MarkTrue(t Type, messageFormat string, messageArgs ...interface{}) {
	conditions.Set(TrueCondition(t, messageFormat, messageArgs...))
}

// MarkFalse sets Status=False for the condition with the given type.
func (conditions *Conditions) MarkFalse(t Type, reason Reason, severity Severity, messageFormat string, messageArgs ...interface{}) {
	conditions.Set(FalseCondition(t, reason, severity, messageFormat, messageArgs...))
}

// MarkUnknown sets Status=Unknown for the condition with the given type.
func (conditions *Conditions) MarkUnknown(t Type, reason Reason, messageFormat string, messageArgs ...interface{}) {
	conditions.Set(UnknownCondition(t, reason, messageFormat, messageArgs...))
}

// IsTrue is true if the condition with the given type is True, otherwise it return false
// if the condition is not True or if the condition does not exist (is nil).
func (conditions *Conditions) IsTrue(t Type) bool {
	if c := conditions.Get(t); c != nil {
		return c.Status == corev1.ConditionTrue
	}
	return false
}

// IsFalse is true if the condition with the given type is False, otherwise it return false
// if the condition is not False or if the condition does not exist (is nil).
func (conditions *Conditions) IsFalse(t Type) bool {
	if c := conditions.Get(t); c != nil {
		return c.Status == corev1.ConditionFalse
	}
	return false
}

// IsUnknown is true if the condition with the given type is Unknown or if the condition
// does not exist (is nil).
func (conditions *Conditions) IsUnknown(t Type) bool {
	if c := conditions.Get(t); c != nil {
		return c.Status == corev1.ConditionUnknown
	}
	return true
}

// Sort - Sorts the list so that the Ready condition always goes first, followed by all the other
// conditions sorted by Type. This makes it easy to identify the overall state of
// the service
func (conditions *Conditions) Sort() {
	// Sorts conditions for convenience of the consumer, i.e. cli client.
	// According to this the Ready condition always goes first, followed by all the other
	// conditions sorted by Type. This makes it easy to identify the overall state of
	// the service
	sort.Slice(*conditions, func(i, j int) bool {
		return less(&(*conditions)[i], &(*conditions)[j])
	})
}

// SortByLastTransitionTime - Sorts a list of conditions by the LastTransitionTime
func (conditions *Conditions) SortByLastTransitionTime() {
	sort.Slice(*conditions, func(i, j int) bool {
		return lessLastTransitionTime(&(*conditions)[i], &(*conditions)[j])
	})
}

// less returns true if a condition is less than another with regards to the
// order of conditions designed for better consumption i.e. cli client.
// According to this the Ready condition always goes first, followed by all the other
// conditions sorted by Type. This makes it easy to identify the overall state of
// the service
func less(i, j *Condition) bool {
	return (i.Type == ReadyCondition || i.Type < j.Type) && j.Type != ReadyCondition
}

// lessLastTransitionTime returns true if a conditions LastTransitionTime is not before
// another ones
func lessLastTransitionTime(i, j *Condition) bool {
	return !i.LastTransitionTime.Before(&j.LastTransitionTime)
}

// hasSameState returns true if a condition has the same state of another
func hasSameState(i, j *Condition) bool {
	return i.Type == j.Type &&
		i.Status == j.Status &&
		i.Reason == j.Reason &&
		i.Severity == j.Severity &&
		i.Message == j.Message
}

// TrueCondition returns a condition with Status=True and the given type.
func TrueCondition(t Type, messageFormat string, messageArgs ...interface{}) *Condition {
	return &Condition{
		Type:     t,
		Status:   corev1.ConditionTrue,
		Reason:   ReadyReason,
		Severity: SeverityNone,
		Message:  fmt.Sprintf(messageFormat, messageArgs...),
	}
}

// FalseCondition returns a condition with Status=False and the given type.
func FalseCondition(t Type, reason Reason, severity Severity, messageFormat string, messageArgs ...interface{}) *Condition {
	return &Condition{
		Type:     t,
		Status:   corev1.ConditionFalse,
		Reason:   reason,
		Severity: severity,
		Message:  fmt.Sprintf(messageFormat, messageArgs...),
	}
}

// UnknownCondition returns a condition with Status=Unknown and the given type.
func UnknownCondition(t Type, reason Reason, messageFormat string, messageArgs ...interface{}) *Condition {
	return &Condition{
		Type:     t,
		Status:   corev1.ConditionUnknown,
		Reason:   reason,
		Severity: SeverityNone,
		Message:  fmt.Sprintf(messageFormat, messageArgs...),
	}
}

// CreateList returns a conditions from a parameter list of several conditions.
func CreateList(conditions ...*Condition) Conditions {
	cs := Conditions{}
	for _, x := range conditions {
		if x != nil {
			cs = append(cs, *x)
		}
	}
	return cs
}

// Mirror - mirrors Status, Message, Reason and Severity from the latest condition
// of a sorted conditionGroup list into a target condition of type t.
// The conditionGroup entries are split by Status with the order False, True, Unknown.
// If Status=False its again split into Severity with the order Error, Warning, Info.
func (conditions *Conditions) Mirror(t Type) *Condition {

	if conditions == nil || len(*conditions) == 0 {
		return nil
	}

	g := conditions.getConditionGroups()
	if len(g) == 0 {
		return nil
	}

	// Get the ConditionTrue group and validate if the overall ReadyCondition is true.
	// If the overall ReadyConditon is true, it is expected that this
	// is the actual state and no other groups need to be checked
	cg := g[groupOrder(*TrueCondition(ReadyCondition, "foo"))]
	if len(cg.conditions) > 0 && cg.conditions.IsTrue(ReadyCondition) {
		c := cg.conditions.Get(ReadyCondition)
		mirrorCondition := TrueCondition(t, c.Message)
		mirrorCondition.LastTransitionTime = c.LastTransitionTime

		return mirrorCondition
	}

	mirrorCondition := &Condition{}
	for _, cg := range g {
		if len(cg.conditions) == 0 {
			continue
		}

		cl := &cg.conditions
		// get the first conditon of the group which is the one with the latest LastTransitionTime
		cl.SortByLastTransitionTime()
		c := (*cl)[0]

		if c.Status == corev1.ConditionTrue {
			mirrorCondition = TrueCondition(t, c.Message)
			mirrorCondition.LastTransitionTime = c.LastTransitionTime
			break
		}

		if c.Status == corev1.ConditionFalse {
			mirrorCondition = FalseCondition(t, c.Reason, c.Severity, c.Message)
			mirrorCondition.LastTransitionTime = c.LastTransitionTime
			break
		}

		mirrorCondition = UnknownCondition(t, c.Reason, c.Message)
		mirrorCondition.LastTransitionTime = c.LastTransitionTime
	}

	return mirrorCondition
}

// getConditionGroups groups a list of conditions according to status, severity values.
// The groups are sorted by Status and Severity.
func (conditions *Conditions) getConditionGroups() []conditionGroup {
	// 6 make possible number of groups 3xConditionFalse, 1xConditionTrue
	// 1xConditionUnknown and a catch all one which should never happen
	groups := make([]conditionGroup, 6)

	for _, condition := range *conditions {
		added := false
		for i := range groups {
			if groups[i].status == condition.Status && groups[i].severity == condition.Severity {
				groups[i].conditions = append(groups[i].conditions, condition)
				added = true
				break
			}
		}
		if !added {
			index := groupOrder(condition)

			groups[index] = conditionGroup{
				conditions: []Condition{condition},
				status:     condition.Status,
				severity:   condition.Severity,
			}
		}
	}

	return groups
}

func groupOrder(c Condition) int {
	switch c.Status {
	case corev1.ConditionFalse:
		switch c.Severity {
		case SeverityError:
			return 0
		case SeverityWarning:
			return 1
		case SeverityInfo:
			return 2
		}
	case corev1.ConditionTrue:
		return 3
	case corev1.ConditionUnknown:
		return 4
	}

	// this hopefully never happens
	return 5
}
