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

package ceph

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// GetPool is a function that validates the pool passed as input, and return
// the default if no pool is given
func GetPool(pools map[string]PoolSpec, service string) (string, error) {
	if pool, found := pools[service]; found {
		return pool.PoolName, nil
	}
	switch service {
	case "cinder":
		return string(DefaultCinderPool), nil
	case "backup":
		return string(DefaultCinderBackupPool), nil
	case "nova":
		return string(DefaultNovaPool), nil
	case "glance":
		return string(DefaultGlancePool), nil
	default:
		return string(CError), errors.New(string("no default pool found")) // nolint:err113
	}
}

// GetRbdUser is a function that validate the user passed as input, and return
// openstack if no user is given
func GetRbdUser(user string) string {

	if user == "" {
		return string(DefaultUser)
	}
	// User is valid and can be used in config
	return user
}

// GetOsdCaps is a function that returns the Caps for each defined pool
func GetOsdCaps(pools map[string]PoolSpec) string {

	var osdCaps string // the resulting string containing caps

	/**
	A map of strings (pool service/name in this case) is, by definition, an
	unordered structure, and let the function return a different pattern
	each time. This causes the ConfigMap hash to change, and the pod being
	redeployed because the operator detects the different hash. Sorting the
	resulting array of pools makes everything predictable
	**/
	var plist []string
	for _, pool := range pools {
		plist = append(plist, pool.PoolName)
	}
	// sort the pool names
	sort.Strings(plist)

	// Build the resulting caps from the _ordered_ array applying the template
	for _, pool := range plist {
		if pool != "" {
			osdCaps += fmt.Sprintf("profile rbd pool=%s,", pool)
		}
	}
	// Default case, no pools are specified, adding just "volumes" (the default)
	if osdCaps == "" {
		osdCaps = "profile rbd pool=" + string(DefaultCinderPool) + ","
	}

	return strings.TrimSuffix(osdCaps, ",")
}
