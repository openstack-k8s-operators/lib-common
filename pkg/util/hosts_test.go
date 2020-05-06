package util

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func configMapWithHosts(hosts string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		Data: map[string]string{"hosts": hosts},
	}
}

func TestCreateOSPHostsEntries(t *testing.T) {
	hosts := []string{
		// Pre-amble
		// Post-amble
		// Space-separated fields
		// Tab-separated fields
		// Blank line
		// Line containing only whitespace
		// IP with no hostnames
		`127.0.0.1 localhost
# BEGIN ANSIBLE MANAGED BLOCK
192.168.0.1 one1 one2
192.168.0.2  two1	two2
192.168.0.3
  

# END ANSIBLE MANAGED BLOCK
127.0.0.2 localhost2
`,
		// No pre-amble
		// No post-amble
		`# BEGIN ANSIBLE MANAGED BLOCK
192.168.0.1 one
# END ANSIBLE MANAGED BLOCK`,
		// No ansible managed block
		`127.0.0.1 localhost`,
	}
	tests := []struct {
		hosts       *string
		err         bool
		hostAliases []corev1.HostAlias
	}{
		{&hosts[0], false, []corev1.HostAlias{{IP: "192.168.0.1", Hostnames: []string{"one1", "one2"}},
			{IP: "192.168.0.2", Hostnames: []string{"two1", "two2"}},
			{IP: "192.168.0.3", Hostnames: []string{}}}},
		{&hosts[1], false, []corev1.HostAlias{{IP: "192.168.0.1", Hostnames: []string{"one"}}}},
		{&hosts[2], true, nil},
	}

	for _, test := range tests {
		hostAliases, err := CreateOspHostsEntries(configMapWithHosts(*test.hosts))
		switch {
		case !test.err && err != nil:
			t.Errorf("Unexpected error parsing hosts `%s`: %v", *test.hosts, err)
		case test.err && err == nil:
			t.Errorf("Didn't get expected error parsing hosts `%s`", *test.hosts)
		case !test.err && !reflect.DeepEqual(hostAliases, test.hostAliases):
			t.Errorf("Parsing `%s`; Expected: %v; Got: %v", *test.hosts, test.hostAliases, hostAliases)
		}
	}
}
