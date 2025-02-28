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

	. "github.com/onsi/gomega"
)

func TestValidateStorageRequest(t *testing.T) {
	tests := []struct {
		name     string
		req      string
		min      string
		err      bool
		wantErr  bool
		wantWarn bool
	}{
		{
			name:     "req is higher then min",
			req:      "500M",
			min:      "400M",
			err:      true,
			wantErr:  false,
			wantWarn: false,
		},
		{
			name:     "req is lower then min, want error",
			req:      "500M",
			min:      "1G",
			err:      true,
			wantErr:  true,
			wantWarn: false,
		},
		{
			name:     "req is lower then min, want warn",
			req:      "500M",
			min:      "1G",
			err:      false,
			wantErr:  false,
			wantWarn: true,
		},
		{
			name:     "req is equal min",
			req:      "500M",
			min:      "500M",
			err:      true,
			wantErr:  false,
			wantWarn: false,
		},
		{
			name:     "req is a wrong string, want err",
			req:      "foo",
			min:      "500M",
			err:      true,
			wantErr:  true,
			wantWarn: false,
		},
		{
			name:     "min is a wrong string, want warn",
			req:      "500M",
			min:      "foo",
			err:      false,
			wantErr:  false,
			wantWarn: true,
		},
		{
			name:     "both are wrong strings, want err",
			req:      "foo",
			min:      "bar",
			err:      true,
			wantErr:  true,
			wantWarn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			p := field.NewPath("foo")

			warns, errs := ValidateStorageRequest(p, tt.req, tt.min, tt.err)
			if tt.wantWarn {
				g.Expect(warns).To(HaveLen(1))
			} else {
				g.Expect(warns).To(BeEmpty())
			}
			if tt.wantErr {
				g.Expect(errs).To(HaveLen(1))
			} else {
				g.Expect(errs).To(BeEmpty())
			}
		})
	}
}
