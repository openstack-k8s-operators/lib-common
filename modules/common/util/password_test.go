package util // nolint:revive

import (
	"testing"

	. "github.com/onsi/gomega" // nolint:revive
)

func TestGeneratePassword(t *testing.T) {

	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "Generate 16 character password",
			length: 16,
		},
		{
			name:   "Generate 32 character password",
			length: 32,
		},
		{
			name:   "Generate 64 character password",
			length: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			password, err := GeneratePassword(tt.length)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(password).To(HaveLen(tt.length))
		})
	}
}

func TestGeneratePasswordUniqueness(t *testing.T) {
	g := NewWithT(t)

	// Generate multiple passwords and ensure they're different
	passwords := make(map[string]bool)
	for i := 0; i < 100; i++ {
		password, err := GeneratePassword(32)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(passwords[password]).To(BeFalse(), "Generated duplicate password")
		passwords[password] = true
	}
}
