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

	. "github.com/onsi/gomega" // nolint:revive
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
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(networkAnnotation).To(HaveLen(len(tt.want)))
			g.Expect(networkAnnotation).To(BeEquivalentTo(tt.want))
		})
	}
}

func TestGetBoolFromAnnotation(t *testing.T) {
	ann := map[string]string{}
	testKey := "service.example.org/key"
	var value bool
	var exists bool
	var err error

	t.Run("", func(t *testing.T) {
		g := NewWithT(t)

		// Case 1: empty annotation map (the key does not exist)
		value, exists, err = GetBoolFromAnnotation(ann, testKey)
		g.Expect(exists).To(BeFalse())
		g.Expect(value).To(BeFalse())
		g.Expect(err).NotTo(HaveOccurred())

		// Case 2: testKey exists but is not a valid bool
		ann[testKey] = "foo"
		value, exists, err = GetBoolFromAnnotation(ann, testKey)
		g.Expect(value).To(BeFalse())
		g.Expect(exists).To(BeTrue())
		g.Expect(err).To(HaveOccurred())

		// Case 3: testKey exists and is False
		ann[testKey] = "false"
		value, exists, err = GetBoolFromAnnotation(ann, testKey)
		g.Expect(value).To(BeFalse())
		g.Expect(exists).To(BeTrue())
		g.Expect(err).ToNot(HaveOccurred())

		// Case 4: testKey exists and is True
		ann[testKey] = "true"
		value, exists, err = GetBoolFromAnnotation(ann, testKey)
		g.Expect(value).To(BeTrue())
		g.Expect(exists).To(BeTrue())
		g.Expect(err).ToNot(HaveOccurred())
	})
}
