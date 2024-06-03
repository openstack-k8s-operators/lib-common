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
			name: "Merge empty maps",
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

func TestMergeMaps(t *testing.T) {
	t.Run("Merge maps", func(t *testing.T) {
		g := NewWithT(t)

		m1 := map[string]string{
			"a": "a",
		}
		m2 := map[string]string{
			"b": "b",
			"c": "c",
		}

		mergedIntMap := MergeMaps(m1, m2)

		g.Expect(mergedIntMap).To(HaveKeyWithValue("a", "a"))
		g.Expect(mergedIntMap).To(HaveKeyWithValue("b", "b"))
		g.Expect(mergedIntMap).To(HaveKeyWithValue("c", "c"))
	})

	t.Run("Merge maps with existing key, the value in the first map is preserved", func(t *testing.T) {
		g := NewWithT(t)

		m1 := map[string]int{
			"a": 2,
			"b": 2,
		}

		m2 := map[string]int{
			"a": 2,
			"c": 3,
			"b": 4,
		}

		mergedIntMap := MergeMaps(m1, m2)

		g.Expect(mergedIntMap).To(HaveKeyWithValue("a", 2))
		g.Expect(mergedIntMap).To(HaveKeyWithValue("b", 2))
		g.Expect(mergedIntMap).To(HaveKeyWithValue("c", 3))
	})
}

func TestGetStringsFromMap(t *testing.T) {
	t.Run("Get List of strings from map", func(t *testing.T) {
		g := NewWithT(t)

		key := "additionalSubjectNamesKey"

		m1 := map[string]string{
			key: "*.foo.svc,*.bar.svc,example.svc.clusterlocal",
		}

		m2 := map[string]string{
			"otherkey": "*.foo.svc,*.bar.svc,example.svc.clusterlocal",
		}

		lstr := GetStringListFromMap(m1, key)
		g.Expect(lstr).To(HaveLen(3))

		lstr = GetStringListFromMap(m2, key)
		g.Expect(lstr).To(BeEmpty())
	})
}
