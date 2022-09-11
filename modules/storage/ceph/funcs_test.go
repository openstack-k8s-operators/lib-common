package ceph

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestGetPool(t *testing.T) {
	// Wrong input: the output is an empty string
	// as no processing is performed
	t.Run("Empty string", func(t *testing.T) {
		g := NewWithT(t)
		m, _ := GetPool(map[string]PoolSpec{}, "foo")
		g.Expect(m).To(BeEmpty())
	})

	// Valid use case: get the Pool defined as input
	t.Run("Get cinder pool", func(t *testing.T) {
		g := NewWithT(t)
		m, _ := GetPool(
			map[string]PoolSpec{
				"cinder": PoolSpec{"volumes"},
			}, "cinder")

		g.Expect(m).To(Equal("volumes"))
	})
}

func TestGetRbdUser(t *testing.T) {
	// If no User is defined, "openstack" should be returned
	t.Run("Get Default User (empty input)", func(t *testing.T) {
		g := NewWithT(t)
		m := GetRbdUser("")
		g.Expect(m).To(Equal("openstack"))
	})
}

func TestGetOsdCaps(t *testing.T) {
	// No pools specified for a given service: the DefaultPool defined
	// for that service should be returned
	t.Run("Get Default Caps", func(t *testing.T) {
		g := NewWithT(t)
		m := GetOsdCaps(map[string]PoolSpec{})
		expectedCaps := "profile rbd pool=" + string(DefaultCinderPool)
		g.Expect(m).To(Equal(expectedCaps))
	})
	// Valid use case: given a list of pools, OSDCaps are built accordingly
	t.Run("Build OSDCaps", func(t *testing.T) {
		g := NewWithT(t)
		m := GetOsdCaps(
			map[string]PoolSpec{
				"cinder": PoolSpec{"volumes"},
				"nova":   PoolSpec{"vms"},
			})
		// We expect caps produced in an ordered list
		expectedCaps := "profile rbd pool=vms,profile rbd pool=volumes"
		g.Expect(m).To(Equal(expectedCaps))
	})
}
