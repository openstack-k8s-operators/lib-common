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

// Legacy service names for LegacyPublicTLSPorts, matching each operator's real
// ServiceName const (confirmed against each operator's const.go, not the THT
// endpoint name - e.g. LegacyServiceNameHeatCfnAPI is "heat-cfnapi", not "heat-cfn";
// LegacyServiceNameNovaNoVNCProxy is "nova-novncproxy", not "novncproxy"). Callers
// should use these constants rather than raw string literals so a typo fails to
// compile instead of silently missing the LegacyPublicTLSPorts lookup.
const (
	LegacyServiceNameKeystone        = "keystone"
	LegacyServiceNamePlacement       = "placement"
	LegacyServiceNameNova            = "nova"
	LegacyServiceNameGlance          = "glance"
	LegacyServiceNameCinder          = "cinder"
	LegacyServiceNameNeutron         = "neutron"
	LegacyServiceNameHeatAPI         = "heat-api"
	LegacyServiceNameHeatCfnAPI      = "heat-cfnapi"
	LegacyServiceNameSwift           = "swift"
	LegacyServiceNameBarbican        = "barbican"
	LegacyServiceNameOctavia         = "octavia"
	LegacyServiceNameManila          = "manila"
	LegacyServiceNameIronic          = "ironic"
	LegacyServiceNameIronicInspector = "ironic-inspector"
	LegacyServiceNameDesignate       = "designate"
	LegacyServiceNameAodh            = "aodh"
	LegacyServiceNameNovaNoVNCProxy  = "nova-novncproxy"
	LegacyServiceNameHorizon         = "horizon"
)

// LegacyPublicTLSPorts holds the RHOSP/director public TLS front ports
// (tripleo-heat-templates wallaby environments/ssl/tls-endpoints-public-ip.yaml),
// used for RHOSP 17.1 -> RHOSO 18 endpoint-URL preservation. Keyed by the
// LegacyServiceName* constants above.
var LegacyPublicTLSPorts = map[string]int32{
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

// GetLegacyPublicTLSPort returns the director-era public TLS port for
// serviceName and whether an entry exists.
func GetLegacyPublicTLSPort(serviceName string) (int32, bool) {
	port, ok := LegacyPublicTLSPorts[serviceName]
	return port, ok
}
