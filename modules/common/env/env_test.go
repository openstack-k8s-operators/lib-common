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

	. "github.com/onsi/gomega" // nolint:revive
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

func containerWithEnv(env ...corev1.EnvVar) corev1.Container {
	return corev1.Container{Name: "c", Env: env}
}

func TestConfigHashMatches(t *testing.T) {
	tests := []struct {
		name       string
		podSpec    corev1.PodSpec
		configHash string
		want       bool
	}{
		{
			name: "matches a literal CONFIG_HASH on a regular container",
			podSpec: corev1.PodSpec{
				Containers: []corev1.Container{containerWithEnv(corev1.EnvVar{Name: ConfigHashEnvVar, Value: "abc"})},
			},
			configHash: "abc",
			want:       true,
		},
		{
			name: "no CONFIG_HASH env var present",
			podSpec: corev1.PodSpec{
				Containers: []corev1.Container{containerWithEnv(corev1.EnvVar{Name: "OTHER", Value: "abc"})},
			},
			configHash: "abc",
			want:       false,
		},
		{
			name: "CONFIG_HASH present but value differs",
			podSpec: corev1.PodSpec{
				Containers: []corev1.Container{containerWithEnv(corev1.EnvVar{Name: ConfigHashEnvVar, Value: "stale"})},
			},
			configHash: "abc",
			want:       false,
		},
		{
			name: "empty configHash never matches even against an empty CONFIG_HASH value",
			podSpec: corev1.PodSpec{
				Containers: []corev1.Container{containerWithEnv(corev1.EnvVar{Name: ConfigHashEnvVar, Value: ""})},
			},
			configHash: "",
			want:       false,
		},
		{
			name: "empty configHash never matches against a populated CONFIG_HASH value",
			podSpec: corev1.PodSpec{
				Containers: []corev1.Container{containerWithEnv(corev1.EnvVar{Name: ConfigHashEnvVar, Value: "abc"})},
			},
			configHash: "",
			want:       false,
		},
		{
			name: "matches a literal CONFIG_HASH carried only on an init container",
			podSpec: corev1.PodSpec{
				InitContainers: []corev1.Container{containerWithEnv(corev1.EnvVar{Name: ConfigHashEnvVar, Value: "abc"})},
				Containers:     []corev1.Container{containerWithEnv(corev1.EnvVar{Name: "OTHER", Value: "x"})},
			},
			configHash: "abc",
			want:       true,
		},
		{
			name: "CONFIG_HASH sourced via ValueFrom is not treated as a match",
			podSpec: corev1.PodSpec{
				Containers: []corev1.Container{containerWithEnv(corev1.EnvVar{
					Name:      ConfigHashEnvVar,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
				})},
			},
			configHash: "abc",
			want:       false,
		},
		{
			name:       "no containers at all",
			podSpec:    corev1.PodSpec{},
			configHash: "abc",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(ConfigHashMatches(tt.podSpec, tt.configHash)).To(Equal(tt.want))
		})
	}
}
