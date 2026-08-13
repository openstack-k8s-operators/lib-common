/*
Copyright 2026 Red Hat

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

package users

import "testing"

func TestRegistryUIDs(t *testing.T) {
	expected := map[string]int64{
		"ansible":          AnsibleUID,
		"aodh":             AodhUID,
		"apache":           ApacheUID,
		"barbican":         BarbicanUID,
		"ceilometer":       CeilometerUID,
		"cinder":           CinderUID,
		"cloud-admin":      CloudAdminUID,
		"cloudkitty":       CloudkittyUID,
		"cyborg":           CyborgUID,
		"designate":        DesignateUID,
		"frr":              FrrUID,
		"frrvty":           FrrVtyUID,
		"glance":           GlanceUID,
		"haproxy":          HaproxyUID,
		"heat":             HeatUID,
		"horizon":          HorizonUID,
		"horizontest":      HorizontestUID,
		"hugetlbfs":        HugetlbfsUID,
		"ironic":           IronicUID,
		"ironic-inspector": IronicInspectorUID,
		"keystone":         KeystoneUID,
		"kolla":            KollaUID,
		"libvirt":          LibvirtUID,
		"manila":           ManilaUID,
		"memcached":        MemcachedUID,
		"mysql":            MysqlUID,
		"named":            NamedUID,
		"neutron":          NeutronUID,
		"nfast":            NfastUID,
		"nova":             NovaUID,
		"octavia":          OctaviaUID,
		"openvswitch":      OpenvswitchUID,
		"ovn-bgp":          OvnBgpUID,
		"placement":        PlacementUID,
		"qemu":             QemuUID,
		"rabbitmq":         RabbitmqUID,
		"rally":            RallyUID,
		"redis":            RedisUID,
		"swift":            SwiftUID,
		"tempest":          TempestUID,
		"tobiko":           TobikoUID,
		"tss":              TssUID,
		"valkey":           ValkeyUID,
		"watcher":          WatcherUID,
	}

	for name, uid := range expected {
		entry, ok := Registry[name]
		if !ok {
			t.Errorf("service %q missing from Registry", name)
			continue
		}
		if entry.UID != uid {
			t.Errorf("Registry[%q].UID = %d, want %d", name, entry.UID, uid)
		}
	}

	if len(Registry) != len(expected) {
		t.Errorf("Registry has %d entries, expected %d", len(Registry), len(expected))
	}
}

func TestRegistryNoDuplicateUIDs(t *testing.T) {
	// valkey (42460) is the redis replacement, intentionally shares its UID
	knownDupes := map[int64]bool{
		42460: true,
	}

	seen := map[int64]string{}
	for name, entry := range Registry {
		if prev, exists := seen[entry.UID]; exists && !knownDupes[entry.UID] {
			t.Errorf("duplicate UID %d: %q and %q", entry.UID, prev, name)
		}
		seen[entry.UID] = name
	}
}

func TestRegistryGIDMatchesUID(t *testing.T) {
	for name, entry := range Registry {
		if entry.GID != entry.UID {
			t.Errorf("Registry[%q]: GID %d != UID %d", name, entry.GID, entry.UID)
		}
	}
}
