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
	"go/build"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

// getDependencyVersion returns the name and version of the given module
// specified in the go.mod file. The returned name follows the "replace"
// directives from go.mod
func getDependencyVersion(moduleName string, goModPath string) (string, string, error) {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", "", err
	}

	f, err := modfile.Parse("go.mod", content, nil)
	if err != nil {
		return "", "", err
	}

	version := ""
	name := moduleName

	for _, r := range f.Require {
		if r.Mod.Path == moduleName {
			version = r.Mod.Version
		}
	}

	// check for replacement config in go.mod for the named module
	for _, r := range f.Replace {
		if r.Old.Path == moduleName {
			version = r.New.Version
			name = r.New.Path
		}
	}

	if version != "" {
		return name, version, nil
	}

	return name, "", fmt.Errorf("cannot find %s in %s file", moduleName, goModPath)
}

// GetCRDDirFromModule returns the absolute path of the directory that holds the
// custom resource definitions for the given module name. It will use the
// version of the module specified in the go.mod file.
func GetCRDDirFromModule(moduleName string, goModPath string, relativeCRDPath string) (string, error) {
	moduleName, version, err := getDependencyVersion(moduleName, goModPath)
	if err != nil {
		return "", err
	}
	versionedModule := fmt.Sprintf("%s@%s", moduleName, version)
	path := filepath.Join(build.Default.GOPATH, "pkg", "mod", versionedModule, relativeCRDPath)
	return path, nil
}

// GetOpenShiftCRDDir returns the absolute path of the directory holding the
// OpenShift custom resource definition. It will look the CRD path up from
// lib-common module.
func GetOpenShiftCRDDir(crdName string, goModPath string) (string, error) {
	// OpenShift CRDs are stored within lib-common so get them from there. To
	// call GetOpenShiftCRDDir the caller needs to have the test module
	// imported so we can use that as anchor for the openshift_crds directory
	libCommon := "github.com/openstack-k8s-operators/lib-common/modules/test"
	libCommon, version, err := getDependencyVersion(libCommon, goModPath)
	if err != nil {
		return "", err
	}
	versionedModule := fmt.Sprintf("%s@%s", libCommon, version)
	path := filepath.Join(build.Default.GOPATH, "pkg", "mod", versionedModule, "openshift_crds", crdName)
	return path, nil
}
