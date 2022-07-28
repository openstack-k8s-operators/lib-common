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

	. "github.com/onsi/gomega"
)

//
// TestSetHash - create or patch the service DB instance
//
func TestSetHash(t *testing.T) {
	hashMap := map[string]string{
		"a": "a",
		"b": "b",
	}
	var changed bool

	t.Run("Add new hashtype and hash", func(t *testing.T) {
		g := NewWithT(t)

		hashMap, changed = SetHash(
			hashMap,
			"c",
			"c",
		)
		g.Expect(changed).To(BeTrue())
		g.Expect(hashMap).To(HaveKeyWithValue("a", "a"))
		g.Expect(hashMap).To(HaveKeyWithValue("b", "b"))
		g.Expect(hashMap).To(HaveKeyWithValue("c", "c"))

	})
	t.Run("Change existing hashtype with hash", func(t *testing.T) {
		g := NewWithT(t)

		hashMap, changed = SetHash(
			hashMap,
			"a",
			"aa",
		)
		g.Expect(changed).To(BeTrue())
		g.Expect(hashMap).To(HaveKeyWithValue("a", "aa"))
		g.Expect(hashMap).To(HaveKeyWithValue("b", "b"))
		g.Expect(hashMap).To(HaveKeyWithValue("c", "c"))
	})
	t.Run("No change to existing hashtype with hash", func(t *testing.T) {
		g := NewWithT(t)

		hashMap, changed = SetHash(
			hashMap,
			"b",
			"b",
		)
		g.Expect(changed).To(BeFalse())
		g.Expect(hashMap).To(HaveKeyWithValue("a", "aa"))
		g.Expect(hashMap).To(HaveKeyWithValue("b", "b"))
		g.Expect(hashMap).To(HaveKeyWithValue("c", "c"))
	})

}
