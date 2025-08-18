/*
Copyright 2024 Red Hat

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

package webhook

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"

	. "github.com/onsi/gomega" // nolint:revive
)

func TestValidateDNS1123Label(t *testing.T) {
	tests := []struct {
		name string
		keys []string
		corr int
		want bool
	}{
		{
			name: "valid name",
			keys: []string{"foo123"},
			corr: 0,
			want: false,
		},
		{
			name: "valid max lenth",
			keys: []string{"foo-1234567890-1234567890-1234567890-1234567890-1234567890-1234"},
			corr: 0,
			want: false,
		},
		{
			name: "invalid max lenth",
			keys: []string{"foo-1234567890-1234567890-1234567890-1234567890-1234567890-1234567890"},
			corr: 0,
			want: true,
		},
		{
			name: "invalid max lenth with correction",
			keys: []string{"foo-1234567890-1234567890-1234567890-1234567890-1234567890-1234"},
			corr: 5,
			want: true,
		},
		{
			name: "invalid char",
			keys: []string{"foo_bar"},
			corr: 0,
			want: true,
		},
		{
			name: "invalid multiple reasons",
			keys: []string{"foo123", "foo-1234567890-1234567890-1234567890-1234567890-1234567890-1234567890", "foo_bar"},
			corr: 0,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			p := field.NewPath("foo")

			errs := ValidateDNS1123Label(p, tt.keys, tt.corr)
			if tt.want {
				g.Expect(errs).ToNot(BeEmpty())
			} else {
				g.Expect(errs).To(BeEmpty())
			}
		})
	}
}
