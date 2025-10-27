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

package util // nolint:revive

import "encoding/json"

// GetOr returns the value of m[key] if it exists, fallback otherwise.
// As a special case, it also returns fallback if the value of m[key] is
// the empty string
func GetOr(m map[string]any, key, fallback string) any {
	val, ok := m[key]
	if !ok {
		return fallback
	}

	s, ok := val.(string)
	if ok && s == "" {
		return fallback
	}

	return val
}

// IsSet returns the value of m[key] if key exists, otherwise false
// Different from getOr because it will return zero values.
func IsSet(m map[string]any, key string) any {
	val, ok := m[key]
	if !ok {
		return false
	}
	return val
}

// IsJSON check if string is json format
func IsJSON(s string) error {
	var js map[string]any
	return json.Unmarshal([]byte(s), &js)
}

// RemoveIndex - remove int from slice
func RemoveIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}
