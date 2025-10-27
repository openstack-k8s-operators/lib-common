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

package util // nolint:revive

import (
	"testing"

	. "github.com/onsi/gomega" // nolint:revive
)

func TestGetOr(t *testing.T) {

	tests := []struct {
		name string
		data map[string]any
		key  string
		want any
	}{
		{
			name: "Key exists with value 111, returns 111",
			data: map[string]any{"one": "111"},
			key:  "one",
			want: "111",
		},
		{
			name: "Key exists and empty string value, returns fallback",
			data: map[string]any{"one": ""},
			key:  "one",
			want: "fallback",
		},
		{
			name: "Key does not exist, returns the fallback",
			data: map[string]any{"one": "111"},
			key:  "four",
			want: "fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			newData := GetOr(tt.data, tt.key, "fallback")
			g.Expect(newData).To(BeIdenticalTo(tt.want))
		})
	}
}

func TestIsSet(t *testing.T) {

	tests := []struct {
		name string
		data map[string]any
		key  string
		want any
	}{
		{
			name: "Key exists, returns 111",
			data: map[string]any{"one": "111"},
			key:  "one",
			want: "111",
		},
		{
			name: "Key does not exist, returns false",
			data: map[string]any{"one": "111"},
			key:  "four",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			newData := IsSet(tt.data, tt.key)
			g.Expect(newData).To(BeIdenticalTo(tt.want))
		})
	}
}

func TestIsJSON(t *testing.T) {

	tests := []struct {
		name  string
		data  string
		error bool
	}{
		{
			name:  "Valid json string",
			data:  `{"some":"json"}`,
			error: false,
		},
		{
			name:  "Empty string",
			data:  "",
			error: true,
		},
		{
			name:  "Not valid json string",
			data:  "not valid json",
			error: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			err := IsJSON(tt.data)
			if tt.error {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
			}
		})
	}
}

func TestRemoveIndex(t *testing.T) {

	tests := []struct {
		name  string
		data  []string
		index int
		want  []string
	}{
		{
			name:  "Remove inx 0",
			data:  []string{"111", "222", "333"},
			index: 0,
			want:  []string{"222", "333"},
		},
		{
			name:  "Remove inx 1",
			data:  []string{"111", "222", "333"},
			index: 1,
			want:  []string{"111", "333"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			newData := RemoveIndex(tt.data, tt.index)
			for idx, d := range newData {
				g.Expect(d).To(BeIdenticalTo(tt.want[idx]))
			}
		})
	}
}
