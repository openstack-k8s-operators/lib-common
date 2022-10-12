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

package env

import (
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func TestMergeEnvs(t *testing.T) {

	tests := []struct {
		name string
		envs SetterMap
		want []corev1.EnvVar
	}{
		{
			name: "Add first env",
			envs: map[string]Setter{"01": SetValue("FIRST_VALUE")},
			want: []corev1.EnvVar{
				{Name: "01", Value: "FIRST_VALUE"},
			},
		},
		{
			name: "Add another env",
			envs: map[string]Setter{"02": SetValue("SECOND_VALUE")},
			want: []corev1.EnvVar{
				{Name: "01", Value: "FIRST_VALUE"},
				{Name: "02", Value: "SECOND_VALUE"},
			},
		},
		{
			name: "Add multiple not sorted envs",
			envs: map[string]Setter{
				"04": SetValue("FOURTH_VALUE"),
				"03": SetValue("THIRD_VALUE"),
			},
			want: []corev1.EnvVar{
				{Name: "01", Value: "FIRST_VALUE"},
				{Name: "02", Value: "SECOND_VALUE"},
				{Name: "03", Value: "THIRD_VALUE"},
				{Name: "04", Value: "FOURTH_VALUE"},
			},
		},
		{
			name: "Update an existing value",
			envs: map[string]Setter{"02": SetValue("SECOND_UPDATED_VALUE")},
			want: []corev1.EnvVar{
				{Name: "01", Value: "FIRST_VALUE"},
				{Name: "02", Value: "SECOND_UPDATED_VALUE"},
				{Name: "03", Value: "THIRD_VALUE"},
				{Name: "04", Value: "FOURTH_VALUE"},
			},
		},
	}

	mergedEnvs := []corev1.EnvVar{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			mergedEnvs = MergeEnvs(mergedEnvs, tt.envs)

			g.Expect(mergedEnvs).To(HaveLen(len(tt.want)))
			g.Expect(mergedEnvs).To(BeEquivalentTo(tt.want))
		})
	}
}
