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
	"testing"

	. "github.com/onsi/gomega" // nolint:revive
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
)

func TestIsPaused(t *testing.T) {
	t.Run("returns false when annotations are nil", func(t *testing.T) {
		g := NewWithT(t)
		obj := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}
		g.Expect(IsPaused(obj)).To(BeFalse())
	})

	t.Run("returns false when annotation is not present", func(t *testing.T) {
		g := NewWithT(t)
		obj := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "test",
				Annotations: map[string]string{"other": "value"},
			},
		}
		g.Expect(IsPaused(obj)).To(BeFalse())
	})

	t.Run("returns true when annotation is present with empty value", func(t *testing.T) {
		g := NewWithT(t)
		obj := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "test",
				Annotations: map[string]string{PausedAnnotation: ""},
			},
		}
		g.Expect(IsPaused(obj)).To(BeTrue())
	})

	t.Run("returns true when annotation is present with any value", func(t *testing.T) {
		g := NewWithT(t)
		obj := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "test",
				Annotations: map[string]string{PausedAnnotation: "true"},
			},
		}
		g.Expect(IsPaused(obj)).To(BeTrue())
	})
}
