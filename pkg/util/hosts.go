package util

import (
	"fmt"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

var ansibleBlockRE = regexp.MustCompile(`(?s).*BEGIN ANSIBLE MANAGED BLOCK\n(.*)# END ANSIBLE MANAGED BLOCK.*`)

// CreateOspHostsEntries creates hostAliases from added /etc/hosts file of the OSP environment.
// All entries inside the ANSIBLE MANAGED BLOCK gets added. Important is that
// the /etc/hosts file in the config map has opening and closing tag.
func CreateOspHostsEntries(commonConfigMap *corev1.ConfigMap) ([]corev1.HostAlias, error) {
	hostAliases := []corev1.HostAlias{}

	hostsFile, isset := commonConfigMap.Data["hosts"]
	if !isset {
		return nil, fmt.Errorf("No hosts file in common-config map")
	}

	hostsEntries := ansibleBlockRE.FindStringSubmatch(hostsFile)
	if len(hostsEntries) == 0 {
		return nil, fmt.Errorf("Ansible tags not found in hosts file of common-config map")
	}

	for _, hostRecord := range strings.Split(hostsEntries[1], "\n") {
		fields := strings.Fields(hostRecord)
		if len(fields) > 0 {
			ip := fields[0]
			names := fields[1:]

			hostAliases = append(hostAliases, corev1.HostAlias{IP: ip, Hostnames: names})
		}
	}

	return hostAliases, nil
}
