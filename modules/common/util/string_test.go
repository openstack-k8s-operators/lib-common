package util // nolint:revive

import (
	"fmt"
	"regexp"
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

func TestRandomString(t *testing.T) {
	g := NewWithT(t)

	t.Run("Valid lengths", func(t *testing.T) {
		lengths := []int{1, 2, 5, 10, 16, 32, 64}

		for _, length := range lengths {
			t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
				result := RandomString(length)

				g.Expect(result).To(HaveLen(length))

				// Check that it contains only alphanumeric characters
				alphanumericPattern := regexp.MustCompile(`^[0-9a-zA-Z]+$`)
				g.Expect(alphanumericPattern.MatchString(result)).To(BeTrue())
			})
		}
	})

	t.Run("Zero and negative lengths", func(t *testing.T) {
		// Test zero length
		result := RandomString(0)
		g.Expect(result).To(Equal(""))

		// Test negative length
		result = RandomString(-1)
		g.Expect(result).To(Equal(""))
	})

	t.Run("Character set validation", func(t *testing.T) {
		result := RandomString(100)

		// Should contain only alphanumeric characters (0-9, a-z, A-Z)
		alphanumericPattern := regexp.MustCompile(`^[0-9a-zA-Z]+$`)
		g.Expect(alphanumericPattern.MatchString(result)).To(BeTrue())

		// Should not contain special characters
		specialPattern := regexp.MustCompile(`[^0-9a-zA-Z]`)
		g.Expect(specialPattern.MatchString(result)).To(BeFalse())
	})
}
