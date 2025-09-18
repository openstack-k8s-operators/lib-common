package util // nolint:revive

import (
	"testing"

	. "github.com/onsi/gomega" // nolint:revive
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
