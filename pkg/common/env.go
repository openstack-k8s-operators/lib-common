/*
Copyright 2020 Red Hat

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

package common

import (
	"sort"

	corev1 "k8s.io/api/core/v1"
)

// Update a list of corev1.EnvVar in place

// EnvSetter -
type EnvSetter func(*corev1.EnvVar)

// EnvSetterMap -
type EnvSetterMap map[string]EnvSetter

// MergeEnvs - merge envs
func MergeEnvs(envs []corev1.EnvVar, newEnvs EnvSetterMap) []corev1.EnvVar {

	// as there is no sorted order when look over hashmap,
	// the resulting env list can have different order and therefore
	// different hash sum, to provent this we create a sorted setter map
	sortedNewEnvSetterMap := SortSetterMapByKey(newEnvs)

	for _, newEnv := range sortedNewEnvSetterMap {
		updated := false
		for i := 0; i < len(envs); i++ {
			if envs[i].Name == newEnv.Key {
				newEnv.Value(&envs[i])
				updated = true
				break
			}
		}

		if !updated {
			envs = append(envs, corev1.EnvVar{Name: newEnv.Key})
			newEnv.Value(&envs[len(envs)-1])
		}
	}

	return envs
}

// EnvDownwardAPI - set env from FieldRef->FieldPath, e.g. status.podIP
func EnvDownwardAPI(field string) EnvSetter {
	return func(env *corev1.EnvVar) {
		if env.ValueFrom == nil {
			env.ValueFrom = &corev1.EnvVarSource{}
		}
		env.Value = ""

		if env.ValueFrom.FieldRef == nil {
			env.ValueFrom.FieldRef = &corev1.ObjectFieldSelector{}
		}

		env.ValueFrom.FieldRef.FieldPath = field
	}
}

// EnvValue -
func EnvValue(value string) EnvSetter {
	return func(env *corev1.EnvVar) {
		env.Value = value
		env.ValueFrom = nil
	}
}

// SetterPair -
type SetterPair struct {
	Key   string
	Value EnvSetter
}

// SetterList -
type SetterList []SetterPair

func (p SetterList) Len() int           { return len(p) }
func (p SetterList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p SetterList) Less(i, j int) bool { return p[i].Key < p[j].Key }

// SortSetterMapByKey - Creates a sorted List contain key/value of a map[string]string sorted by key
func SortSetterMapByKey(in map[string]EnvSetter) SetterList {

	sorted := make(SetterList, len(in))

	i := 0
	for k, v := range in {
		sorted[i] = SetterPair{k, v}
		i++
	}

	sort.Sort(sorted)

	return sorted
}
