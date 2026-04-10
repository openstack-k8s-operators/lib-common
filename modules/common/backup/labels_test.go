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

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestGetBackupLabels(t *testing.T) {
	tests := []struct {
		name     string
		category string
		want     map[string]string
	}{
		{
			name:     "PVC with controlplane category",
			category: CategoryControlPlane,
			want: map[string]string{
				BackupLabel:         "true",
				BackupCategoryLabel: "controlplane",
			},
		},
		{
			name:     "PVC with dataplane category",
			category: CategoryDataPlane,
			want: map[string]string{
				BackupLabel:         "true",
				BackupCategoryLabel: "dataplane",
			},
		},
		{
			name:     "PVC without category",
			category: "",
			want: map[string]string{
				BackupLabel: "true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBackupLabels(tt.category)
			if len(got) != len(tt.want) {
				t.Errorf("GetBackupLabels() returned %d labels, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("GetBackupLabels()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestGetRestoreLabels(t *testing.T) {
	tests := []struct {
		name         string
		restoreOrder string
		category     string
		want         map[string]string
	}{
		{
			name:         "controlplane CR",
			restoreOrder: RestoreOrder30,
			category:     CategoryControlPlane,
			want: map[string]string{
				BackupRestoreLabel:      "true",
				BackupRestoreOrderLabel: "30",
				BackupCategoryLabel:     "controlplane",
			},
		},
		{
			name:         "without category",
			restoreOrder: RestoreOrder10,
			category:     "",
			want: map[string]string{
				BackupRestoreLabel:      "true",
				BackupRestoreOrderLabel: "10",
			},
		},
		{
			name:         "dataplane CR",
			restoreOrder: RestoreOrder60,
			category:     CategoryDataPlane,
			want: map[string]string{
				BackupRestoreLabel:      "true",
				BackupRestoreOrderLabel: "60",
				BackupCategoryLabel:     "dataplane",
			},
		},
		{
			name:         "PVC restore order",
			restoreOrder: RestoreOrder00,
			category:     CategoryControlPlane,
			want: map[string]string{
				BackupRestoreLabel:      "true",
				BackupRestoreOrderLabel: "00",
				BackupCategoryLabel:     "controlplane",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRestoreLabels(tt.restoreOrder, tt.category)
			if len(got) != len(tt.want) {
				t.Errorf("GetRestoreLabels() returned %d labels, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("GetRestoreLabels()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestShouldBackup(t *testing.T) {
	tests := []struct {
		name   string
		labels map[string]string
		want   bool
	}{
		{
			name:   "nil labels",
			labels: nil,
			want:   false,
		},
		{
			name:   "empty labels",
			labels: map[string]string{},
			want:   false,
		},
		{
			name: "backup label true",
			labels: map[string]string{
				BackupLabel: "true",
			},
			want: true,
		},
		{
			name: "backup label false",
			labels: map[string]string{
				BackupLabel: "false",
			},
			want: false,
		},
		{
			name: "no backup label",
			labels: map[string]string{
				"other": "label",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldBackup(tt.labels); got != tt.want {
				t.Errorf("ShouldBackup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLabelKeys(t *testing.T) {
	keys := LabelKeys()
	if len(keys) != 3 {
		t.Errorf("LabelKeys() returned %d keys, want 3", len(keys))
	}
	expected := map[string]bool{
		BackupLabel:             true,
		BackupRestoreLabel:      true,
		BackupRestoreOrderLabel: true,
	}
	for _, k := range keys {
		if !expected[k] {
			t.Errorf("LabelKeys() contains unexpected key %q", k)
		}
	}
}

func TestApplyAnnotationOverrides(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		labels      map[string]string
		want        map[string]string
	}{
		{
			name:        "nil annotations",
			annotations: nil,
			labels:      map[string]string{BackupRestoreLabel: "true", BackupRestoreOrderLabel: "00"},
			want:        map[string]string{BackupRestoreLabel: "true", BackupRestoreOrderLabel: "00"},
		},
		{
			name:        "no backup annotations",
			annotations: map[string]string{"other": "value"},
			labels:      map[string]string{BackupRestoreLabel: "true", BackupRestoreOrderLabel: "00"},
			want:        map[string]string{BackupRestoreLabel: "true", BackupRestoreOrderLabel: "00"},
		},
		{
			name:        "override restore to false",
			annotations: map[string]string{BackupRestoreLabel: "false"},
			labels:      map[string]string{BackupRestoreLabel: "true", BackupRestoreOrderLabel: "00"},
			want:        map[string]string{BackupRestoreLabel: "false", BackupRestoreOrderLabel: "00"},
		},
		{
			name:        "override restore order",
			annotations: map[string]string{BackupRestoreOrderLabel: "20"},
			labels:      map[string]string{BackupRestoreLabel: "true", BackupRestoreOrderLabel: "00"},
			want:        map[string]string{BackupRestoreLabel: "true", BackupRestoreOrderLabel: "20"},
		},
		{
			name:        "restore order implies restore true",
			annotations: map[string]string{BackupRestoreOrderLabel: "20"},
			labels:      map[string]string{BackupRestoreLabel: "false", BackupRestoreOrderLabel: "00"},
			want:        map[string]string{BackupRestoreLabel: "true", BackupRestoreOrderLabel: "20"},
		},
		{
			name:        "override backup to false",
			annotations: map[string]string{BackupLabel: "false"},
			labels:      map[string]string{BackupLabel: "true"},
			want:        map[string]string{BackupLabel: "false"},
		},
		{
			name:        "case insensitive",
			annotations: map[string]string{BackupRestoreLabel: "TRUE"},
			labels:      map[string]string{BackupRestoreLabel: "false"},
			want:        map[string]string{BackupRestoreLabel: "true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labels := make(map[string]string)
			for k, v := range tt.labels {
				labels[k] = v
			}
			ApplyAnnotationOverrides(tt.annotations, labels)
			for k, v := range tt.want {
				if labels[k] != v {
					t.Errorf("applyAnnotationOverrides() labels[%q] = %q, want %q", k, labels[k], v)
				}
			}
		})
	}
}

func TestGetRestoreLabelsWithOverrides(t *testing.T) {
	tests := []struct {
		name                string
		defaultRestoreOrder string
		overrides           map[string]string
		want                map[string]string
	}{
		{
			name:                "no overrides",
			defaultRestoreOrder: RestoreOrder30,
			overrides:           map[string]string{},
			want: map[string]string{
				BackupRestoreLabel:      "true",
				BackupRestoreOrderLabel: "30",
			},
		},
		{
			name:                "override restore order",
			defaultRestoreOrder: RestoreOrder30,
			overrides: map[string]string{
				BackupRestoreOrderLabel: "40",
			},
			want: map[string]string{
				BackupRestoreLabel:      "true",
				BackupRestoreOrderLabel: "40",
			},
		},
		{
			name:                "override category",
			defaultRestoreOrder: RestoreOrder30,
			overrides: map[string]string{
				BackupCategoryLabel: CategoryDataPlane,
			},
			want: map[string]string{
				BackupRestoreLabel:      "true",
				BackupRestoreOrderLabel: "30",
				BackupCategoryLabel:     "dataplane",
			},
		},
		{
			name:                "override both order and category",
			defaultRestoreOrder: RestoreOrder30,
			overrides: map[string]string{
				BackupRestoreOrderLabel: "50",
				BackupCategoryLabel:     CategoryDataPlane,
			},
			want: map[string]string{
				BackupRestoreLabel:      "true",
				BackupRestoreOrderLabel: "50",
				BackupCategoryLabel:     "dataplane",
			},
		},
		{
			name:                "nil overrides",
			defaultRestoreOrder: RestoreOrder10,
			overrides:           nil,
			want: map[string]string{
				BackupRestoreLabel:      "true",
				BackupRestoreOrderLabel: "10",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRestoreLabelsWithOverrides(tt.defaultRestoreOrder, tt.overrides)
			if len(got) != len(tt.want) {
				t.Errorf("GetRestoreLabelsWithOverrides() returned %d labels, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("GetRestoreLabelsWithOverrides()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestAnnotationChangedPredicate(t *testing.T) {
	labelSelector := "service-cert"
	p := AnnotationChangedPredicate(labelSelector)

	t.Run("create event returns false", func(t *testing.T) {
		if p.Create(event.CreateEvent{}) {
			t.Error("expected CreateEvent to return false")
		}
	})

	t.Run("delete event returns false", func(t *testing.T) {
		if p.Delete(event.DeleteEvent{}) {
			t.Error("expected DeleteEvent to return false")
		}
	})

	t.Run("generic event returns false", func(t *testing.T) {
		if p.Generic(event.GenericEvent{}) {
			t.Error("expected GenericEvent to return false")
		}
	})

	t.Run("update without label selector returns false", func(t *testing.T) {
		e := event.UpdateEvent{
			ObjectOld: &metav1.PartialObjectMetadata{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"other": "label"}},
			},
			ObjectNew: &metav1.PartialObjectMetadata{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"other": "label"},
					Annotations: map[string]string{BackupRestoreLabel: "true"},
				},
			},
		}
		if p.Update(e) {
			t.Error("expected update without label selector to return false")
		}
	})

	t.Run("update with label selector but no annotation change returns false", func(t *testing.T) {
		e := event.UpdateEvent{
			ObjectOld: &metav1.PartialObjectMetadata{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{labelSelector: ""},
					Annotations: map[string]string{BackupRestoreLabel: "false"},
				},
			},
			ObjectNew: &metav1.PartialObjectMetadata{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{labelSelector: ""},
					Annotations: map[string]string{BackupRestoreLabel: "false"},
				},
			},
		}
		if p.Update(e) {
			t.Error("expected update with no annotation change to return false")
		}
	})

	t.Run("update with restore annotation change returns true", func(t *testing.T) {
		e := event.UpdateEvent{
			ObjectOld: &metav1.PartialObjectMetadata{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{labelSelector: ""},
					Annotations: map[string]string{BackupRestoreLabel: "false"},
				},
			},
			ObjectNew: &metav1.PartialObjectMetadata{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{labelSelector: ""},
					Annotations: map[string]string{BackupRestoreLabel: "true"},
				},
			},
		}
		if !p.Update(e) {
			t.Error("expected update with restore annotation change to return true")
		}
	})

	t.Run("update with restore-order annotation added returns true", func(t *testing.T) {
		e := event.UpdateEvent{
			ObjectOld: &metav1.PartialObjectMetadata{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{labelSelector: ""},
				},
			},
			ObjectNew: &metav1.PartialObjectMetadata{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{labelSelector: ""},
					Annotations: map[string]string{BackupRestoreOrderLabel: "20"},
				},
			},
		}
		if !p.Update(e) {
			t.Error("expected update with restore-order annotation added to return true")
		}
	})

	t.Run("update with non-backup annotation change returns false", func(t *testing.T) {
		e := event.UpdateEvent{
			ObjectOld: &metav1.PartialObjectMetadata{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{labelSelector: ""},
					Annotations: map[string]string{"other": "old"},
				},
			},
			ObjectNew: &metav1.PartialObjectMetadata{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{labelSelector: ""},
					Annotations: map[string]string{"other": "new"},
				},
			},
		}
		if p.Update(e) {
			t.Error("expected update with non-backup annotation change to return false")
		}
	})
}
