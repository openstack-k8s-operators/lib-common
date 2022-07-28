package common

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestMergeStringMaps(t *testing.T) {
	t.Run("Merge maps", func(t *testing.T) {
		g := NewWithT(t)

		m := MergeStringMaps(
			map[string]string{
				"a": "a",
				"b": "b",
			}, map[string]string{
				"a": "ax",
				"c": "c",
			},
		)
		g.Expect(m).To(HaveKeyWithValue("a", "a"))
		g.Expect(m).To(HaveKeyWithValue("b", "b"))
		g.Expect(m).To(HaveKeyWithValue("c", "c"))
	})
	t.Run("Nils empty maps", func(t *testing.T) {
		g := NewWithT(t)

		m := MergeStringMaps(map[string]string{}, map[string]string{})
		g.Expect(m).To(BeNil())
	})
}

func TestSortStringMapByValue(t *testing.T) {
	t.Run("Sort map", func(t *testing.T) {
		g := NewWithT(t)

		l := SortStringMapByValue(
			map[string]string{
				"b": "b",
				"a": "a",
			},
		)

		g.Expect(l[0]).To(HaveField("Key", "a"))
		g.Expect(l[0]).To(HaveField("Value", "a"))
		g.Expect(l[1]).To(HaveField("Key", "b"))
		g.Expect(l[1]).To(HaveField("Value", "b"))
	})
}
