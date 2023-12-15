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

package test

import (
	"fmt"

	. "github.com/onsi/gomega"

	"os"
	"path/filepath"
	"testing"
)

func TestGetCRDDirFromModule(t *testing.T) {
	t.Run("with an existing module in go.mod", func(t *testing.T) {
		g := NewWithT(t)

		// NOTE(gibi): this is not a good use case as k8s.io does not have
		// CRDs stored. But in lib-common we have no go.mod deps on a repo
		// that has such CRDs to test on.
		path, err := GetCRDDirFromModule("k8s.io/api", "../common/go.mod", "bases")
		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(path).Should(MatchRegexp("/.*/k8s.io/api@v.*/bases"))
	})
	t.Run("with a wrong go.mod path", func(t *testing.T) {
		g := NewWithT(t)

		_, err := GetCRDDirFromModule("github.com/openstack-k8s-operators/mariadb-operator/api", "../nonexistent/go.mod", "bases")
		g.Expect(err).Should(HaveOccurred())
		g.Expect(err).Should(MatchError("open ../nonexistent/go.mod: no such file or directory"))
	})
	t.Run("with a module not in go.mod", func(t *testing.T) {
		g := NewWithT(t)

		_, err := GetCRDDirFromModule("foobar", "go.mod", "bases")
		g.Expect(err).Should(HaveOccurred())
		g.Expect(err).Should(MatchError("cannot find foobar in go.mod file"))
	})
	t.Run("with a module in go.mod having a local replacement", func(t *testing.T) {
		g := NewWithT(t)

		// Generate a go.mod with a relative path replace statement
		dir := t.TempDir()
		mod := []byte(`module foo
go 1.19
require (
	github.com/openstack-k8s-operators/infra-operator/apis v0.1.1-0.20231001103054-f74a88ed4971
)
replace github.com/openstack-k8s-operators/infra-operator/apis => ../../infra-operator/apis
		`)
		modPath := filepath.Join(dir, "go.mod")
		err := os.WriteFile(modPath, mod, 0644)
		g.Expect(err).ShouldNot(HaveOccurred())

		path, err := GetCRDDirFromModule("github.com/openstack-k8s-operators/infra-operator/apis", modPath, "bases")
		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(path).Should(MatchRegexp(".*/infra-operator/config/crd/bases$"))
	})

	t.Run("with a module in go.mod having a local replacement with mixed case", func(t *testing.T) {
		g := NewWithT(t)

		// Generate a go.mod with replacement pointing to a mixed case github ID
		dir := t.TempDir()
		mod := []byte(`module foo
go 1.19
require (
	github.com/openstack-k8s-operators/infra-operator/apis v0.1.1-0.20231001103054-f74a88ed4971
)
replace github.com/openstack-k8s-operators/infra-operator/apis => ../../Infra-Operator/apis
		`)
		modPath := filepath.Join(dir, "go.mod")
		err := os.WriteFile(modPath, mod, 0644)
		g.Expect(err).ShouldNot(HaveOccurred())

		path, err := GetCRDDirFromModule("github.com/openstack-k8s-operators/infra-operator/apis", modPath, "bases")
		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(path).Should(MatchRegexp(".*/!infra-!operator/config/crd/bases$"))
	})

	t.Run("with a module in go.mod having a remote replacement with mixed case", func(t *testing.T) {
		g := NewWithT(t)

		// Generate a go.mod with replacement pointing to a mixed case github ID
		dir := t.TempDir()
		mod := []byte(`module foo
go 1.19
require (
	github.com/openstack-k8s-operators/infra-operator/apis v0.1.1-0.20231001103054-f74a88ed4971
)
replace github.com/openstack-k8s-operators/infra-operator/apis => github.com/MixedUser/infra-operator/apis v0.1.1-0.20231001103054-fffa88ed4971
		`)
		modPath := filepath.Join(dir, "go.mod")
		err := os.WriteFile(modPath, mod, 0644)
		g.Expect(err).ShouldNot(HaveOccurred())

		path, err := GetCRDDirFromModule("github.com/openstack-k8s-operators/infra-operator/apis", modPath, "bases")
		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(path).Should(MatchRegexp("github.com/!mixed!user/infra-operator/apis@v0.1.1-0.20231001103054-fffa88ed4971/bases$"))
	})
}

func TestGetOpenShiftCRDDir(t *testing.T) {
	t.Run("with a CRD and valid go.mod", func(t *testing.T) {
		g := NewWithT(t)

		// We need to generate a go.mod that has lib-common dependency in it
		dir := t.TempDir()
		mod := []byte(`module foo
go 1.19
require (
	github.com/openstack-k8s-operators/lib-common/modules/test v0.0.0-20220630111354-9f8383d4a2ea
)
		`)
		modPath := filepath.Join(dir, "go.mod")
		err := os.WriteFile(modPath, mod, 0644)
		g.Expect(err).ShouldNot(HaveOccurred())

		path, err := GetOpenShiftCRDDir("route/v1", modPath)
		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(path).Should(MatchRegexp("/.*/github.com/openstack-k8s-operators/lib-common/modules/test@v.*/openshift_crds/route/v1"))
	})
	t.Run("with a CRD without having lib-common in go.mod", func(t *testing.T) {
		g := NewWithT(t)

		// Our own lib-common go.mod will never have lib-common in it so we can use that as test input
		_, err := GetOpenShiftCRDDir("route/v1", "go.mod")
		g.Expect(err).Should(HaveOccurred())
		fmt.Printf("%s", err)
		g.Expect(err).Should(MatchError("cannot find github.com/openstack-k8s-operators/lib-common/modules/test in go.mod file"))
	})
}
