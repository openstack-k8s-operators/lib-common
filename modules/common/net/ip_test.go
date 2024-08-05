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

package net

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestSortIPs(t *testing.T) {

	tests := []struct {
		name string
		ips  []string
		want []string
	}{
		{
			name: "empty ip list",
			ips:  []string{},
			want: []string{},
		},
		{
			name: "IPv4 - single ip in list",
			ips:  []string{"1.1.1.1"},
			want: []string{"1.1.1.1"},
		},
		{
			name: "IPv4 - already sorted list",
			ips:  []string{"1.1.1.1", "2.2.2.2"},
			want: []string{"1.1.1.1", "2.2.2.2"},
		},
		{
			name: "IPv4 - unsorted sorted list",
			ips:  []string{"2.2.2.2", "1.1.1.1"},
			want: []string{"1.1.1.1", "2.2.2.2"},
		},
		{
			name: "IPv4 - another unsorted sorted list",
			ips:  []string{"2.2.2.2", "1.1.1.2", "1.1.1.1"},
			want: []string{"1.1.1.1", "1.1.1.2", "2.2.2.2"},
		},
		{
			name: "IPv6 - single ip in list",
			ips:  []string{"fd00:bbbb::1"},
			want: []string{"fd00:bbbb::1"},
		},
		{
			name: "IPv6 - already sorted list",
			ips:  []string{"fd00:bbbb::1", "fd00:bbbb::2"},
			want: []string{"fd00:bbbb::1", "fd00:bbbb::2"},
		},
		{
			name: "IPv6 - unsorted sorted list",
			ips:  []string{"fd00:bbbb::2", "fd00:bbbb::1"},
			want: []string{"fd00:bbbb::1", "fd00:bbbb::2"},
		},
		{
			name: "IPv6 - another unsorted sorted list",
			ips:  []string{"fd00:bbbb::2", "fd00:aaaa::1", "fd00:bbbb::1"},
			want: []string{"fd00:aaaa::1", "fd00:bbbb::1", "fd00:bbbb::2"},
		},
		{
			name: "IPV4 and IPv6 - unsorted sorted list",
			ips:  []string{"fd00:bbbb::2", "fd00:aaaa::1", "fd00:bbbb::1", "1.1.1.1"},
			want: []string{"1.1.1.1", "fd00:aaaa::1", "fd00:bbbb::1", "fd00:bbbb::2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			sortedIPs := SortIPs(tt.ips)
			g.Expect(sortedIPs).NotTo(BeNil())
			g.Expect(sortedIPs).To(HaveLen(len(tt.want)))
			g.Expect(sortedIPs).To(BeEquivalentTo(tt.want))
		})
	}
}
