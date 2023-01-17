/*
Copyright 2023 Red Hat

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

	. "github.com/onsi/gomega"
)

func TestGetNADAnnotation(t *testing.T) {

	tests := []struct {
		name      string
		networks  []string
		namespace string
		want      map[string]string
	}{
		{
			name:      "Single network",
			networks:  []string{},
			namespace: "foo",
			want:      map[string]string{NetworkAttachmentAnnot: "[]"},
		},
		{
			name:      "Single network",
			networks:  []string{"one"},
			namespace: "foo",
			want:      map[string]string{NetworkAttachmentAnnot: "[{\"Name\":\"one\",\"Namespace\":\"foo\"}]"},
		},
		{
			name:      "Multiple networks",
			networks:  []string{"one", "two"},
			namespace: "foo",
			want:      map[string]string{NetworkAttachmentAnnot: "[{\"Name\":\"one\",\"Namespace\":\"foo\"},{\"Name\":\"two\",\"Namespace\":\"foo\"}]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			networkAnnotation, err := GetNADAnnotation(tt.namespace, tt.networks)
			g.Expect(err).To(BeNil())
			g.Expect(networkAnnotation).To(HaveLen(len(tt.want)))
			g.Expect(networkAnnotation).To(BeEquivalentTo(tt.want))
		})
	}
}
