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

package endpoint

import "testing"

func TestLegacyPublicTLSPorts(t *testing.T) {
	expected := map[string]int32{
		LegacyServiceNameKeystone:        13000,
		LegacyServiceNamePlacement:       13778,
		LegacyServiceNameNova:            13774,
		LegacyServiceNameGlance:          13292,
		LegacyServiceNameCinder:          13776,
		LegacyServiceNameNeutron:         13696,
		LegacyServiceNameHeatAPI:         13004,
		LegacyServiceNameHeatCfnAPI:      13005,
		LegacyServiceNameSwift:           13808,
		LegacyServiceNameBarbican:        13311,
		LegacyServiceNameOctavia:         13876,
		LegacyServiceNameManila:          13786,
		LegacyServiceNameIronic:          13385,
		LegacyServiceNameIronicInspector: 13050,
		LegacyServiceNameDesignate:       13001,
		LegacyServiceNameAodh:            13042,
		LegacyServiceNameNovaNoVNCProxy:  13080,
		LegacyServiceNameHorizon:         443,
	}

	for name, port := range expected {
		if LegacyPublicTLSPorts[name] != port {
			t.Errorf("LegacyPublicTLSPorts[%q] = %d, want %d", name, LegacyPublicTLSPorts[name], port)
		}
	}

	if len(LegacyPublicTLSPorts) != len(expected) {
		t.Errorf("LegacyPublicTLSPorts has %d entries, expected %d", len(LegacyPublicTLSPorts), len(expected))
	}
}

func TestGetLegacyPublicTLSPort(t *testing.T) {
	port, ok := GetLegacyPublicTLSPort(LegacyServiceNamePlacement)
	if !ok || port != 13778 {
		t.Errorf("GetLegacyPublicTLSPort(%q) = (%d, %v), want (13778, true)", LegacyServiceNamePlacement, port, ok)
	}

	if _, ok := GetLegacyPublicTLSPort("does-not-exist"); ok {
		t.Errorf("GetLegacyPublicTLSPort(%q) ok = true, want false", "does-not-exist")
	}
}
