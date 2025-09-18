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

package util

import (
	"testing"

	. "github.com/onsi/gomega" // nolint:revive

	env "github.com/openstack-k8s-operators/lib-common/modules/common/env"
)

func TestObjectHash(t *testing.T) {

	tests := []struct {
		name string
		data map[string]string
		want string
	}{
		{
			name: "Create hash",
			data: map[string]string{"a": "a"},
			want: "n548h65h79hffh74h59hf7h9ch8h65bh56fh665h66h98h575hdh74h58hbfh5c9h65dh655hbch55dh699hf5h689h695h5c7h5c7h5bbh5ffq",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			hash, err := ObjectHash(tt.data)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(hash).To(BeIdenticalTo(tt.want))
		})
	}
}

func TestSetHash(t *testing.T) {

	tests := []struct {
		name     string
		hashType string
		hashStr  string
		changed  bool
		want     map[string]string
	}{
		{
			name:     "Add new hashtype and hash - a:a",
			hashType: "a",
			hashStr:  "a",
			changed:  true,
			want:     map[string]string{"a": "a"},
		},
		{
			name:     "Add new hashtype and hash - b:b",
			hashType: "b",
			hashStr:  "b",
			changed:  true,
			want:     map[string]string{"a": "a", "b": "b"},
		},
		{
			name:     "Change existing hashtype with hash - a:aa",
			hashType: "a",
			hashStr:  "aa",
			changed:  true,
			want:     map[string]string{"a": "aa", "b": "b"},
		},
		{
			name:     "No change to existing hashtype with hash - b:b",
			hashType: "b",
			hashStr:  "b",
			changed:  false,
			want:     map[string]string{"a": "aa", "b": "b"},
		},
	}

	hashMap := map[string]string{}
	var changed bool

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			hashMap, changed = SetHash(
				hashMap,
				tt.hashType,
				tt.hashStr,
			)

			g.Expect(changed).To(BeIdenticalTo(tt.changed))
			for k, v := range tt.want {
				g.Expect(hashMap).To(HaveKeyWithValue(k, v))
			}
		})
	}
}

func TestHashOfInputHashes(t *testing.T) {

	tests := []struct {
		name string
		envs env.SetterMap
		want string
	}{
		{
			name: "Add first env",
			envs: map[string]env.Setter{"a": env.SetValue("a")},
			want: "n8dh59h59chfch665h5fch555h5f8h89h65dh568h5bbh8bh5f9h59dh56bh65ch685h589h655hcch599h698h564h8bh59h5fh5f8h97h55fh5d9h699q",
		},
		{
			name: "Add another env",
			envs: map[string]env.Setter{"b": env.SetValue("b")},
			want: "n57h5d4h56h88h656h689hfdhf6h677h5c8hb6h659hc8h5c6h8h5dh8h97hb7h5c4h589hbdh56bh564h8h5cbhd8h5chch674h5d8h588q",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			hash, err := HashOfInputHashes(tt.envs)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(hash).To(BeEquivalentTo(tt.want))
		})
	}
}
