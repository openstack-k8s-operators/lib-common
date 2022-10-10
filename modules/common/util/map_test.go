package util

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestMergeStringMaps(t *testing.T) {

	tests := []struct {
		name string
		map1 map[string]string
		map2 map[string]string
		want map[string]string
	}{
		{
			name: "Merge maps",
			map1: map[string]string{
				"a": "a",
			},
			map2: map[string]string{
				"b": "b",
				"c": "c",
			},
			want: map[string]string{
				"a": "a",
				"b": "b",
				"c": "c",
			},
		},
		{
			name: "Merge maps with existing key, the value in the first map is preserved",
			map1: map[string]string{
				"a": "a",
				"b": "b",
			},
			map2: map[string]string{
				"a": "ax",
				"c": "c",
			},
			want: map[string]string{
				"a": "a",
				"b": "b",
				"c": "c",
			},
		},
		{
			name: "Merge maps with existing key, the value in the first map is preserved",
			map1: map[string]string{},
			map2: map[string]string{},
			want: map[string]string{},
		},
	}

	mergedMap := map[string]string{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			mergedMap = MergeStringMaps(
				tt.map1,
				tt.map2,
			)

			if mergedMap == nil {
				g.Expect(mergedMap).To(BeNil())
			}

			for k, v := range tt.want {
				g.Expect(mergedMap).To(HaveKeyWithValue(k, v))
			}
		})
	}
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
