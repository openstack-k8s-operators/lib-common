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

package endpoint

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestEndptProtocol(t *testing.T) {

	tests := []struct {
		name  string
		proto *Protocol
		want  string
	}{
		{
			name:  "None protocol",
			proto: PtrProtocol(ProtocolNone),
			want:  "",
		},
		{
			name:  "http protocol",
			proto: PtrProtocol(ProtocolHTTP),
			want:  "http://",
		},
		{
			name:  "https protocol",
			proto: PtrProtocol(ProtocolHTTPS),
			want:  "https://",
		},
		{
			name:  "Nil protocol",
			proto: nil,
			want:  "http://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(endptProtocol(tt.proto)).To(Equal(tt.want))
		})
	}
}
