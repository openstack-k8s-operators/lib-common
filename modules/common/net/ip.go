/*
Copyright 2023 Red Hat

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

package net

import (
	"bytes"
	"net"
	"sort"
)

// SortIPs - Get network-attachment-definition with name in namespace
func SortIPs(
	ips []string,
) []string {
	netIPs := make([]net.IP, 0, len(ips))

	for _, ip := range ips {
		netIPs = append(netIPs, net.ParseIP(ip))
	}

	sort.Slice(netIPs, func(i, j int) bool {
		return bytes.Compare(netIPs[i], netIPs[j]) < 0
	})

	sortedIPs := make([]string, 0, len(netIPs))

	for _, ip := range netIPs {
		sortedIPs = append(sortedIPs, ip.String())
	}

	return sortedIPs
}
