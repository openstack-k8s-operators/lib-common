package ceph

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestValidateMons(t *testing.T) {
	// Input as empty string should return false (malformed IP addresses)
	t.Run("Empty string", func(t *testing.T) {
		g := NewWithT(t)
		m := ValidateMons("")
		g.Expect(m).To(Equal(false))
	})
	// Valid use case
	t.Run("Validate Mons", func(t *testing.T) {
		g := NewWithT(t)
		m := ValidateMons("192.168.2.2,192.168.2.3, 192.168.2.4")
		g.Expect(m).To(Equal(true))
	})

	t.Run("Validate (Wrong) Mons", func(t *testing.T) {
		// Input with a wrong IP address
		g := NewWithT(t)
		m := ValidateMons("192.168.2.2,192.168.2.3,192.168.2")
		g.Expect(m).To(Equal(false))
	})
}
