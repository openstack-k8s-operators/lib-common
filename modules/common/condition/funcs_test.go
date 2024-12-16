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
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	unknownReady = UnknownCondition(ReadyCondition, RequestedReason, ReadyInitMessage)
	trueReady    = TrueCondition(ReadyCondition, ReadyMessage)

	unknownA     = UnknownCondition("a", "reason unknownA", "message unknownA")
	falseA       = FalseCondition("a", "reason falseA", SeverityInfo, "message falseA")
	trueA        = TrueCondition("a", "message trueA")
	unknownB     = UnknownCondition("b", "reason unknownB", "message unknownB")
	falseB       = FalseCondition("b", "reason falseB", SeverityInfo, "message falseB")
	falseBError  = FalseCondition("b", "reason falseBError", SeverityError, "message falseBError")
	trueB        = TrueCondition("b", "message trueB")
	falseInfo    = FalseCondition("falseInfo", "reason falseInfo", SeverityInfo, "message falseInfo")
	falseWarning = FalseCondition("falseWarning", "reason falseWarning", SeverityWarning, "message falseWarning")
	falseError   = FalseCondition("falseError", "reason falseError", SeverityError, "message falseError")
)

func TestInit(t *testing.T) {
	tests := []struct {
		name       string
		conditions Conditions
		want       Conditions
	}{
		{
			name:       "Init conditions without optional condition",
			conditions: CreateList(nil),
			want:       CreateList(unknownReady),
		},
		{
			name:       "Init conditions with an optional condition",
			conditions: CreateList(unknownA),
			want:       CreateList(unknownReady, unknownA),
		},
		{
			name:       "Init conditions with optional conditions",
			conditions: CreateList(unknownA, unknownB),
			want:       CreateList(unknownReady, unknownA, unknownB),
		},
		{
			name:       "Init conditions with optional conditions",
			conditions: CreateList(unknownB, unknownA),
			want:       CreateList(unknownReady, unknownA, unknownB),
		},
		{
			name:       "Init conditions with duplicate optional condition",
			conditions: CreateList(unknownA, unknownA),
			want:       CreateList(unknownReady, unknownA),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			someCondition := TrueCondition("foo", "to be removed on Init()")
			conditions := Conditions{
				*someCondition,
			}

			conditions.Init(&tt.conditions)
			g.Expect(conditions).To(haveSameConditionsOf(tt.want))
		})
	}
}

func TestSet(t *testing.T) {
	conditions := Conditions{}

	time1 := metav1.NewTime(time.Date(2022, time.August, 9, 10, 0, 0, 0, time.UTC))
	time2 := metav1.NewTime(time.Date(2022, time.August, 10, 10, 0, 0, 0, time.UTC))
	falseBTime1 := falseB.DeepCopy()
	falseBTime1.LastTransitionTime = time1

	falseBTime2 := falseB.DeepCopy()
	falseBTime2.LastTransitionTime = time2

	tests := []struct {
		name      string
		condition *Condition
		want      Conditions
	}{
		{
			name:      "Add nil condition",
			condition: nil,
			want:      CreateList(unknownReady),
		},
		{
			name:      "Add unknownB condition",
			condition: unknownB,
			want:      CreateList(unknownReady, unknownB),
		},
		{
			name:      "Add another condition unknownA, gets sorted",
			condition: unknownA,
			want:      CreateList(unknownReady, unknownA, unknownB),
		},
		{
			name:      "Add same condition unknownA, won't duplicate",
			condition: unknownA,
			want:      CreateList(unknownReady, unknownA, unknownB),
		},
		{
			name:      "Change condition unknownA, to be Status=False",
			condition: falseA,
			want:      CreateList(unknownReady, falseA, unknownB),
		},
		{
			name:      "Change ready condition to True",
			condition: trueReady,
			want:      CreateList(trueReady, falseA, unknownB),
		},
	}

	g := NewWithT(t)

	conditions.Init(nil)
	g.Expect(conditions).To(haveSameConditionsOf(CreateList(unknownReady)))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			conditions.Set(tt.condition)

			g.Expect(conditions).To(haveSameConditionsOf(tt.want))
		})
	}

	// test time LastTransitionTime won't change if same status get set
	// a) set conditions with time1
	conditions.Set(falseBTime1)
	c1 := conditions.Get(falseB.Type)
	g.Expect(c1.LastTransitionTime).To(BeIdenticalTo(time1))

	// b) set condition with same state, but new time. Time must be still time1
	conditions.Set(falseBTime2)
	c2 := conditions.Get(falseB.Type)
	g.Expect(c2.LastTransitionTime).To(BeIdenticalTo(time1))
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name       string
		conditions Conditions
		cType      Type
		expected   Conditions
	}{
		{
			name: "present",
			conditions: CreateList(
				unknownReady,
				unknownA,
				unknownB,
			),
			cType: unknownA.Type,
			expected: CreateList(
				unknownReady,
				unknownB,
			),
		},
		{
			name: "unknownReady-remove",
			conditions: CreateList(
				unknownReady,
				unknownA,
				unknownB,
			),
			cType: unknownReady.Type,
			expected: CreateList(
				unknownA,
				unknownB,
			),
		},
		{
			name: "absent",
			conditions: CreateList(
				unknownReady,
				unknownA,
			),
			cType: unknownB.Type,
			expected: CreateList(
				unknownReady,
				unknownA,
			),
		},
		{
			name:       "empty",
			conditions: Conditions{},
			cType:      unknownA.Type,
			expected:   Conditions{},
		},
	}

	g := NewWithT(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.conditions.Remove(tt.cType)
			g.Expect(tt.expected).To(haveSameConditionsOf(tt.conditions))
		})
	}
}

func TestReset(t *testing.T) {
	tests := []struct {
		name       string
		conditions Conditions
		expected   Conditions
	}{
		{
			name:       "empty",
			conditions: Conditions{},
			expected:   Conditions{},
		},
		{
			name: "present",
			conditions: CreateList(
				unknownReady,
				unknownA,
				unknownB,
			),
			expected: Conditions{},
		},
	}

	g := NewWithT(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.conditions.Reset()
			g.Expect(tt.expected).To(haveSameConditionsOf(tt.conditions))
		})
	}
}

func TestHasSameState(t *testing.T) {
	g := NewWithT(t)

	// same condition
	falseInfo2 := falseInfo.DeepCopy()
	g.Expect(HasSameState(falseInfo, falseInfo2)).To(BeTrue())

	// different LastTransitionTime does not impact state
	falseInfo2 = falseInfo.DeepCopy()
	falseInfo2.LastTransitionTime = metav1.NewTime(time.Date(1900, time.November, 10, 23, 0, 0, 0, time.UTC))
	g.Expect(HasSameState(falseInfo, falseInfo2)).To(BeTrue())

	// different Type, Status, Reason, Severity and Message determine different state
	falseInfo2 = falseInfo.DeepCopy()
	falseInfo2.Type = "another type"
	g.Expect(HasSameState(falseInfo, falseInfo2)).To(BeFalse())

	falseInfo2 = falseInfo.DeepCopy()
	falseInfo2.Status = corev1.ConditionTrue
	g.Expect(HasSameState(falseInfo, falseInfo2)).To(BeFalse())

	falseInfo2 = falseInfo.DeepCopy()
	falseInfo2.Severity = SeverityWarning
	g.Expect(HasSameState(falseInfo, falseInfo2)).To(BeFalse())

	falseInfo2 = falseInfo.DeepCopy()
	falseInfo2.Reason = "another reason"
	g.Expect(HasSameState(falseInfo, falseInfo2)).To(BeFalse())

	falseInfo2 = falseInfo.DeepCopy()
	falseInfo2.Message = "another message"
	g.Expect(HasSameState(falseInfo, falseInfo2)).To(BeFalse())
}

func TestLess(t *testing.T) {
	g := NewWithT(t)

	// alphabetical order of Type is respected
	g.Expect(less(trueA, trueB)).To(BeTrue())
	g.Expect(less(trueB, trueA)).To(BeFalse())

	// Ready condition is always expected to be first
	g.Expect(less(trueReady, trueA)).To(BeTrue())
	g.Expect(less(trueA, trueReady)).To(BeFalse())

}

func TestGetAndHas(t *testing.T) {
	g := NewWithT(t)

	conditions := Conditions{}
	conditions.Init(nil)
	g.Expect(conditions).To(haveSameConditionsOf(CreateList(unknownReady)))
	g.Expect(conditions.Has(ReadyCondition)).To(BeTrue())
	g.Expect(conditions.Get(ReadyCondition)).To(haveSameStateOf(unknownReady))
	g.Expect(conditions.Get("notExistingCond")).To(BeNil())
	g.Expect(conditions.Get("notExistingCond")).To(BeNil())

	conditions.Set(unknownA)
	g.Expect(conditions.Has(ReadyCondition)).To(BeTrue())
	g.Expect(conditions.Has("a")).To(BeTrue())
	g.Expect(conditions.Get("a")).To(haveSameStateOf(unknownA))
}

func TestIsMethods(t *testing.T) {
	g := NewWithT(t)

	conditions := Conditions{}
	cl := CreateList(trueA, falseInfo, unknownB)
	conditions.Init(&cl)
	g.Expect(conditions).To(haveSameConditionsOf(CreateList(unknownReady, trueA, unknownB, falseInfo)))

	// test isTrue
	g.Expect(conditions.IsTrue("a")).To(BeTrue())
	g.Expect(conditions.IsTrue("falseInfo")).To(BeFalse())
	g.Expect(conditions.IsTrue("unknownB")).To(BeFalse())

	// test isFalse
	g.Expect(conditions.IsFalse("a")).To(BeFalse())
	g.Expect(conditions.IsFalse("falseInfo")).To(BeTrue())
	g.Expect(conditions.IsFalse("unknownB")).To(BeFalse())

	// test isUnknown
	g.Expect(conditions.IsUnknown("a")).To(BeFalse())
	g.Expect(conditions.IsUnknown("falseInfo")).To(BeFalse())
	g.Expect(conditions.IsUnknown("unknownB")).To(BeTrue())
}

func TestAllSubConditionIsTrue(t *testing.T) {
	conditions := Conditions{}

	time1 := metav1.NewTime(time.Date(2022, time.August, 9, 10, 0, 0, 0, time.UTC))
	time2 := metav1.NewTime(time.Date(2022, time.August, 10, 10, 0, 0, 0, time.UTC))
	falseBTime1 := falseB.DeepCopy()
	falseBTime1.LastTransitionTime = time1

	falseBTime2 := falseB.DeepCopy()
	falseBTime2.LastTransitionTime = time2

	tests := []struct {
		name      string
		condition *Condition
		want      bool
	}{
		{
			name:      "Add nil condition",
			condition: nil,
			want:      true,
		},
		{
			name:      "Add unknownB condition",
			condition: unknownB,
			want:      false,
		},
		{
			name:      "Add another condition unknownA",
			condition: unknownA,
			want:      false,
		},
		{
			name:      "Change condition unknownA, to be Status=True",
			condition: trueA,
			want:      false,
		},
		{
			name:      "Change condition unknownB, to be Status=True",
			condition: trueB,
			want:      true,
		},
	}

	g := NewWithT(t)

	conditions.Init(nil)
	g.Expect(conditions).To(haveSameConditionsOf(CreateList(unknownReady)))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			conditions.Set(tt.condition)
			g.Expect(conditions.AllSubConditionIsTrue()).To(BeIdenticalTo(tt.want))
		})
	}
}

func TestMarkMethods(t *testing.T) {
	g := NewWithT(t)

	conditions := Conditions{}
	conditions.Init(nil)
	g.Expect(conditions).To(haveSameConditionsOf(CreateList(unknownReady)))
	g.Expect(conditions.Get(ReadyCondition).Severity).To(BeEmpty())

	// test MarkTrue
	conditions.MarkTrue(ReadyCondition, ReadyMessage)
	g.Expect(conditions.Get(ReadyCondition)).To(haveSameStateOf(trueReady))
	g.Expect(conditions.Get(ReadyCondition).Severity).To(BeEmpty())

	// test MarkFalse
	conditions.MarkFalse("falseError", "reason falseError", SeverityError, "message falseError")
	g.Expect(conditions.Get("falseError")).To(haveSameStateOf(falseError))

	// test MarkTrue of previous false condition
	conditions.MarkTrue("falseError", "now True")
	g.Expect(conditions.Get("falseError").Severity).To(BeEmpty())

	// test MarkUnknown
	conditions.MarkUnknown("a", "reason unknownA", "message unknownA")
	g.Expect(conditions.Get("a")).To(haveSameStateOf(unknownA))
}

func TestSortByLastTransitionTime(t *testing.T) {
	g := NewWithT(t)

	time1 := metav1.NewTime(time.Date(2020, time.August, 9, 10, 0, 0, 0, time.UTC))
	time2 := metav1.NewTime(time.Date(2020, time.August, 10, 10, 0, 0, 0, time.UTC))
	time3 := metav1.NewTime(time.Date(2020, time.August, 11, 10, 0, 0, 0, time.UTC))

	falseA.LastTransitionTime = time1
	falseB.LastTransitionTime = time2
	falseError.LastTransitionTime = time3

	g.Expect(lessLastTransitionTime(falseA, falseB)).To(BeFalse())
	g.Expect(lessLastTransitionTime(falseB, falseA)).To(BeTrue())

	conditions := Conditions{}
	cl := CreateList(falseB, falseError, falseA)
	conditions.Init(&cl)
	conditions.SortByLastTransitionTime()

	// unknownReady has the current time stamp, so is first
	g.Expect(conditions).To(haveSameConditionsOf(CreateList(unknownReady, falseError, falseB, falseA)))
}

func TestMirror(t *testing.T) {
	g := NewWithT(t)

	time1 := metav1.NewTime(time.Date(2020, time.August, 9, 10, 0, 0, 0, time.UTC))
	time2 := metav1.NewTime(time.Date(2020, time.August, 10, 10, 0, 0, 0, time.UTC))

	trueA.LastTransitionTime = time1
	falseB.LastTransitionTime = time2

	conditions := Conditions{}
	conditions.Init(nil)
	g.Expect(conditions).To(haveSameConditionsOf(CreateList(unknownReady)))
	targetCondition := conditions.Mirror("targetConditon")
	g.Expect(targetCondition.Status).To(BeIdenticalTo(unknownReady.Status))
	g.Expect(targetCondition.Severity).To(BeIdenticalTo(unknownReady.Severity))
	g.Expect(targetCondition.Reason).To(BeIdenticalTo(unknownReady.Reason))
	g.Expect(targetCondition.Message).To(BeIdenticalTo(unknownReady.Message))

	conditions.Set(trueA)
	g.Expect(conditions).To(haveSameConditionsOf(CreateList(unknownReady, trueA)))
	targetCondition = conditions.Mirror("targetConditon")
	// expect to be mirrored unknownReady
	g.Expect(targetCondition.Status).To(BeIdenticalTo(unknownReady.Status))
	g.Expect(targetCondition.Severity).To(BeIdenticalTo(unknownReady.Severity))
	g.Expect(targetCondition.Reason).To(BeIdenticalTo(unknownReady.Reason))
	g.Expect(targetCondition.Message).To(BeIdenticalTo(unknownReady.Message))

	conditions.Set(falseB)
	g.Expect(conditions).To(haveSameConditionsOf(CreateList(unknownReady, trueA, falseB)))
	targetCondition = conditions.Mirror("targetConditon")
	// expect to be mirrored falseB
	g.Expect(targetCondition.Status).To(BeIdenticalTo(falseB.Status))
	g.Expect(targetCondition.Severity).To(BeIdenticalTo(falseB.Severity))
	g.Expect(targetCondition.Reason).To(BeIdenticalTo(falseB.Reason))
	g.Expect(targetCondition.Message).To(BeIdenticalTo(falseB.Message))

	conditions.Set(falseBError)
	g.Expect(conditions).To(haveSameConditionsOf(CreateList(unknownReady, trueA, falseBError)))
	targetCondition = conditions.Mirror("targetConditon")
	// expect to be mirrored falseBError
	g.Expect(targetCondition.Status).To(BeIdenticalTo(falseBError.Status))
	g.Expect(targetCondition.Severity).To(BeIdenticalTo(falseBError.Severity))
	g.Expect(targetCondition.Reason).To(BeIdenticalTo(falseBError.Reason))
	g.Expect(targetCondition.Message).To(BeIdenticalTo(falseBError.Message))

	// mark ReadyCondition to true
	// We expect that ReadyCondition True means the overall status of the
	// resource is ready that condition is mirrored even if there are other
	// conditions in non True state.
	conditions.MarkTrue(ReadyCondition, ReadyMessage)
	conditions.Set(unknownA)
	g.Expect(conditions).To(haveSameConditionsOf(CreateList(trueReady, unknownA, falseBError)))
	targetCondition = conditions.Mirror("targetConditon")
	// expect to be mirrored trueReady
	g.Expect(targetCondition.Status).To(BeIdenticalTo(trueReady.Status))
	g.Expect(targetCondition.Severity).To(BeIdenticalTo(trueReady.Severity))
	g.Expect(targetCondition.Reason).To(BeIdenticalTo(trueReady.Reason))
	g.Expect(targetCondition.Message).To(BeIdenticalTo(trueReady.Message))
}

func TestMirrorInvalidStatus(t *testing.T) {
	g := NewWithT(t)

	conditions := Conditions{}
	invalidStatusA := Condition{
		Type:     "a",
		Status:   "FooBar",
		Reason:   "",
		Severity: SeverityNone,
		Message:  "",
	}
	conditions.Init(&Conditions{invalidStatusA})
	// NOTE(gibi): we always have a ReadyCondition with Unknown status added
	// automatically so it is picked up by Mirror before reaching to the
	// condition with FooBar status as FooBar status handled as the lowest prio
	// as it is not matching of any expectes Status value. So to trigger the
	// error case we need to remove the ReadyCondition explicitly.
	conditions.Remove(ReadyCondition)

	g.Expect(func() { conditions.Mirror("targetConditon") }).To(
		PanicWith(MatchRegexp(`Condition \{a FooBar .*\} has invalid status value 'FooBar'.`)))
}

func TestIsError(t *testing.T) {
	g := NewWithT(t)

	g.Expect(IsError(nil)).To(BeFalse())
	// wrong reason
	g.Expect(IsError(falseBError)).To(BeFalse())
	g.Expect(IsError(falseB)).To(BeFalse())
	g.Expect(IsError(trueB)).To(BeFalse())
	g.Expect(IsError(FalseCondition("errorReason", ErrorReason, SeverityError, "message Error"))).To(BeTrue())
}

func TestGetHigherPrioCondition(t *testing.T) {
	g := NewWithT(t)

	g.Expect(GetHigherPrioCondition(nil, nil)).To(BeNil())

	c := GetHigherPrioCondition(unknownA, nil)
	g.Expect(HasSameState(c, unknownA)).To(BeTrue())

	c = GetHigherPrioCondition(nil, unknownA)
	g.Expect(HasSameState(c, unknownA)).To(BeTrue())

	// Status Unknown has higher prio then Status True
	c = GetHigherPrioCondition(unknownA, trueA)
	g.Expect(HasSameState(c, unknownA)).To(BeTrue())

	// Status False has higher prio then Status Unknown
	c = GetHigherPrioCondition(falseA, unknownA)
	g.Expect(HasSameState(c, falseA)).To(BeTrue())

	// Status False has higher prio then Status True
	c = GetHigherPrioCondition(falseA, trueA)
	g.Expect(HasSameState(c, falseA)).To(BeTrue())

	// When both Status=False, Severity Error has higher prio then Info
	c = GetHigherPrioCondition(falseInfo, falseError)
	g.Expect(HasSameState(c, falseError)).To(BeTrue())

	// When both Status=False, Severity Error has higher prio then Warning
	c = GetHigherPrioCondition(falseWarning, falseError)
	g.Expect(HasSameState(c, falseError)).To(BeTrue())

	// When both Status=False, Severity Warning has higher prio then Info
	c = GetHigherPrioCondition(falseWarning, falseInfo)
	g.Expect(HasSameState(c, falseWarning)).To(BeTrue())

	// When both Status=False, and same Severity return the one with later
	// LastTransitionTime.
	warning1 := falseWarning.DeepCopy()
	warning2 := falseWarning.DeepCopy()

	time1 := metav1.NewTime(time.Date(2020, time.August, 9, 10, 0, 0, 0, time.UTC))
	time2 := metav1.NewTime(time.Date(2020, time.August, 10, 10, 0, 0, 0, time.UTC))
	warning1.LastTransitionTime = time1
	warning1.Message = "warning1"
	warning2.LastTransitionTime = time2
	warning2.Message = "warning2"

	c = GetHigherPrioCondition(warning1, warning2)
	g.Expect(HasSameState(c, warning2)).To(BeTrue())
}

func TestRestoreLastTransitionTimes(t *testing.T) {
	time1 := metav1.NewTime(time.Date(2022, time.August, 9, 10, 0, 0, 0, time.UTC))
	time2 := metav1.NewTime(time.Date(2022, time.August, 10, 10, 0, 0, 0, time.UTC))

	tests := []struct {
		name  string
		patch func(condition *Condition)
		want  metav1.Time
	}{
		// If the patch function modifies any field that causes HasSameCondition()
		// to fail, then testCond should retain its LastTransitionTime (time1).
		{
			name: "Different condition type",
			patch: func(condition *Condition) {
				condition.Type = "X"
			},
			want: time1,
		},
		{
			name: "Different condition status",
			patch: func(condition *Condition) {
				condition.Status = corev1.ConditionUnknown
			},
			want: time1,
		},
		{
			name: "Different condition reason",
			patch: func(condition *Condition) {
				condition.Reason = "reason X"
			},
			want: time1,
		},
		{
			name: "Different condition severity",
			patch: func(condition *Condition) {
				condition.Severity = SeverityWarning
			},
			want: time1,
		},
		{
			name: "Different condition message",
			patch: func(condition *Condition) {
				condition.Message = "message X"
			},
			want: time1,
		},
		// LastTransitionTime should change to time2 when HasSameCondition() passes.
		{
			name:  "Same condition state",
			patch: func(condition *Condition) {},
			want:  time2,
		},
	}

	g := NewWithT(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCond := falseA.DeepCopy()
			testCond.LastTransitionTime = time1
			conditions := CreateList(testCond)

			// Patch a copy of testCond and set a different LastTransitionTime
			savedCond := testCond.DeepCopy()
			tt.patch(savedCond)
			savedCond.LastTransitionTime = time2
			savedConditions := CreateList(savedCond)

			RestoreLastTransitionTimes(&conditions, savedConditions)

			g.Expect(conditions.Get(testCond.Type).LastTransitionTime).To(BeIdenticalTo(tt.want))
		})
	}
}

// haveSameConditionsOf matches a conditions list to be the same as another.
func haveSameConditionsOf(expected Conditions) types.GomegaMatcher {
	return &conditionsMatcher{
		Expected: expected,
	}
}

type conditionsMatcher struct {
	Expected Conditions
}

func (matcher *conditionsMatcher) Match(actual interface{}) (success bool, err error) {
	actualConditions, ok := actual.(Conditions)
	if !ok {
		return false, errors.New("Value should be a conditions list")
	}

	if len(actualConditions) != len(matcher.Expected) {
		return false, nil
	}

	for i := range actualConditions {
		if !HasSameState(&actualConditions[i], &matcher.Expected[i]) {
			return false, nil
		}
	}
	return true, nil
}

func (matcher *conditionsMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to have the same conditions of", matcher.Expected)
}

func (matcher *conditionsMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to have the same conditions of", matcher.Expected)
}

// haveSameStateOf matches a condition to have the same state of another.
func haveSameStateOf(expected *Condition) types.GomegaMatcher {
	return &conditionMatcher{
		Expected: expected,
	}
}

type conditionMatcher struct {
	Expected *Condition
}

func (matcher *conditionMatcher) Match(actual interface{}) (success bool, err error) {
	actualCondition, ok := actual.(*Condition)
	if !ok {
		return false, errors.New("value should be a condition")
	}

	return HasSameState(actualCondition, matcher.Expected), nil
}

func (matcher *conditionMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to have the same state of", matcher.Expected)
}

func (matcher *conditionMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to have the same state of", matcher.Expected)
}
