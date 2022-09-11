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
	"net"
	"strings"
)

// ValidateMons is a function that validates the comma separated Mon list defined
// for the external ceph cluster; it also checks the provided IP addresses are not
// malformed
func ValidateMons(ipList string) bool {
	for _, ip := range strings.Split(ipList, ",") {
		if net.ParseIP(strings.Trim(ip, " ")) == nil {
			return false
		}
	}
	return true
}
