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

package util

import (
	"sort"
	"strings"
)

// InitMap - Inititialise a map to an empty map if it is nil.
func InitMap(m *map[string]string) {
	if *m == nil {
		*m = make(map[string]string)
	}
}

// MergeStringMaps - merge two or more string->map maps
// NOTE: In case a key exists, the value in the first map is preserved.
func MergeStringMaps(baseMap map[string]string, extraMaps ...map[string]string) map[string]string {
	mergedMap := make(map[string]string)

	// Copy from the original map to the target map
	for key, value := range baseMap {
		mergedMap[key] = value
	}

	for _, extraMap := range extraMaps {
		for key, value := range extraMap {
			if _, ok := mergedMap[key]; !ok {
				mergedMap[key] = value
			}
		}
	}

	// Nil the result if the map is empty, thus avoiding triggering infinite reconcile
	// given that at json level label: {} or annotation: {} is different from no field, which is the
	// corresponding value stored in etcd given that those fields are defined as omitempty.
	if len(mergedMap) == 0 {
		return nil
	}
	return mergedMap
}

// Pair -
type Pair struct {
	Key   string
	Value string
}

// List -
type List []Pair

func (p List) Len() int           { return len(p) }
func (p List) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p List) Less(i, j int) bool { return p[i].Key < p[j].Key }

// SortStringMapByValue - Creates a sorted List contain key/value of a map[string]string sorted by key
func SortStringMapByValue(in map[string]string) List {

	sorted := make(List, len(in))

	i := 0
	for k, v := range in {
		sorted[i] = Pair{k, v}
		i++
	}

	sort.Sort(sorted)

	return sorted
}

// MergeMaps - merge two or more maps
// NOTE: In case a key exists, the value in the first map is preserved.
func MergeMaps[K comparable, V any](baseMap map[K]V, extraMaps ...map[K]V) map[K]V {
	mergedMap := make(map[K]V)
	for key, value := range baseMap {
		mergedMap[key] = value
	}

	for _, extraMap := range extraMaps {
		for key, value := range extraMap {
			if _, ok := mergedMap[key]; !ok {
				mergedMap[key] = value
			}
		}
	}

	return mergedMap
}

// GetStringListFromMap - It returns a list of strings based on a comma
// separated list assigned to the map key. This is usually invoked to normalize
// annotation fields where a list of items is expressed with a comma separated
// list of strings.
// Example:
// input: in["additionalSubjectNamesKey"] = "foo.bar,bar.svc,*.foo.bar"
// output: [foo.bar bar.svc *.foo.bar]
func GetStringListFromMap(in map[string]string, key string) []string {
	strList := []string{}
	if h, ok := in[key]; ok {
		if h != "" {
			strList = strings.Split(h, ",")
		}
	}
	return strList
}
