package util

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestStringInSlice(t *testing.T) {

	tests := []struct {
		name       string
		data       []string
		teststring string
		want       bool
	}{
		{
			name:       "String in slice",
			data:       []string{"foo", "bar"},
			teststring: "foo",
			want:       true,
		},
		{
			name:       "String not in slice",
			data:       []string{"foo", "bar"},
			teststring: "boo",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			b := StringInSlice(
				tt.teststring,
				tt.data,
			)

			g.Expect(b).To(BeIdenticalTo(tt.want))
		})
	}
}

func TestDumpListToString(t *testing.T) {
	tests := []struct {
		name string
		data []string
		want string
	}{
		{
			name: "Dump list to string",
			data: []string{"c", "a", "b"},
			want: "a,b,c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			b := DumpListToString(
				tt.data,
			)

			g.Expect(b).To(BeIdenticalTo(tt.want))
		})
	}
}

func TestLoadListFromString(t *testing.T) {
	tests := []struct {
		name string
		data string
		want []string
	}{
		{
			name: "Load list from string",
			data: "a,b,c",
			want: []string{"c", "a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			b := LoadListFromString(
				tt.data,
			)

			g.Expect(b).To(BeIdenticalTo(tt.want))
		})
	}
}
