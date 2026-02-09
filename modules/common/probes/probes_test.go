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

package probes

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestSetProbeConf(t *testing.T) {
	tests := []struct {
		name     string
		port     int32
		scheme   *v1.URIScheme
		config   ProbeConf
		wantURI  v1.URIScheme
		wantErr  bool
		wantPath string
	}{
		{
			name:   "Valid port with HTTP scheme",
			port:   8080,
			scheme: &[]v1.URIScheme{v1.URISchemeHTTP}[0],
			config: ProbeConf{
				Path:                "/health",
				InitialDelaySeconds: 30,
				TimeoutSeconds:      5,
				PeriodSeconds:       10,
				FailureThreshold:    3,
			},
			wantURI:  v1.URISchemeHTTP,
			wantPath: "/health",
		},
		{
			name:   "Valid port with HTTPS scheme",
			port:   8443,
			scheme: &[]v1.URIScheme{v1.URISchemeHTTPS}[0],
			config: ProbeConf{
				Path:                "/ready",
				InitialDelaySeconds: 15,
				TimeoutSeconds:      10,
				PeriodSeconds:       5,
				FailureThreshold:    1,
			},
			wantURI:  v1.URISchemeHTTPS,
			wantPath: "/ready",
		},
		{
			name:   "Valid port without scheme",
			port:   9090,
			scheme: nil,
			config: ProbeConf{
				Path:                "/status",
				InitialDelaySeconds: 20,
				TimeoutSeconds:      3,
				PeriodSeconds:       15,
				FailureThreshold:    5,
			},
			wantPath: "/status",
		},
		{
			name:    "Negative Port",
			port:    -8080,
			scheme:  nil,
			config:  ProbeConf{Path: "/health"},
			wantErr: true,
		},
		{
			name:    "Port Larger than 65535",
			port:    70000,
			scheme:  nil,
			config:  ProbeConf{Path: "/health"},
			wantErr: true,
		},
		{
			name:    "Port zero",
			port:    0,
			scheme:  nil,
			config:  ProbeConf{Path: "/health"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			probe, err := SetProbeConf(tt.port, tt.scheme, tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SetProbeConf() expected error but got none")
				}
				return
			} else if err != nil {
				t.Errorf("SetProbeConf() unexpected error: %v", err)
				return
			}
			if probe.HTTPGet.Path != tt.wantPath {
				t.Errorf("SetProbeConf() path = %v, want %v", probe.HTTPGet.Path, tt.wantPath)
			}
			if tt.scheme != nil && probe.HTTPGet.Scheme != tt.wantURI {
				t.Errorf("SetProbeConf() scheme = %v, want %v", probe.HTTPGet.Scheme, tt.wantURI)
			}
			if probe.InitialDelaySeconds != tt.config.InitialDelaySeconds {
				t.Errorf("SetProbeConf() initialDelaySeconds = %v, want %v", probe.InitialDelaySeconds, tt.config.InitialDelaySeconds)
			}
			if probe.TimeoutSeconds != tt.config.TimeoutSeconds {
				t.Errorf("SetProbeConf() timeoutSeconds = %v, want %v", probe.TimeoutSeconds, tt.config.TimeoutSeconds)
			}
			if probe.PeriodSeconds != tt.config.PeriodSeconds {
				t.Errorf("SetProbeConf() periodSeconds = %v, want %v", probe.PeriodSeconds, tt.config.PeriodSeconds)
			}
			if probe.FailureThreshold != tt.config.FailureThreshold {
				t.Errorf("SetProbeConf() failureThreshold = %v, want %v", probe.FailureThreshold, tt.config.FailureThreshold)
			}
		})
	}
}

func TestCreateProbeSet(t *testing.T) {
	defaultLiveness := &ProbeConf{
		Path:                "/health",
		InitialDelaySeconds: 30,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		FailureThreshold:    3,
	}
	defaultReadiness := &ProbeConf{
		Path:                "/ready",
		InitialDelaySeconds: 15,
		TimeoutSeconds:      3,
		PeriodSeconds:       5,
		FailureThreshold:    1,
	}
	defaultStartup := &ProbeConf{
		Path:                "/startup",
		InitialDelaySeconds: 10,
		TimeoutSeconds:      10,
		PeriodSeconds:       5,
		FailureThreshold:    10,
	}

	defaults := OverrideSpec{
		LivenessProbes:  defaultLiveness,
		ReadinessProbes: defaultReadiness,
		StartupProbes:   defaultStartup,
	}

	tests := []struct {
		name      string
		port      int32
		scheme    *v1.URIScheme
		overrides ProbeOverrides
		defaults  OverrideSpec
		wantErr   bool
	}{
		{
			name:      "Valid configuration with no overrides",
			port:      8080,
			scheme:    &[]v1.URIScheme{v1.URISchemeHTTP}[0],
			overrides: OverrideSpec{},
			defaults:  defaults,
			wantErr:   false,
		},
		{
			name:   "Valid configuration with overrides",
			port:   8443,
			scheme: &[]v1.URIScheme{v1.URISchemeHTTPS}[0],
			overrides: OverrideSpec{
				LivenessProbes: &ProbeConf{
					Path:                "/custom-health",
					InitialDelaySeconds: 60,
					TimeoutSeconds:      10,
					PeriodSeconds:       20,
					FailureThreshold:    5,
				},
			},
			defaults: defaults,
			wantErr:  false,
		},
		{
			name:      "Invalid port",
			port:      -1,
			scheme:    nil,
			overrides: OverrideSpec{},
			defaults:  defaults,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			probeSet, err := CreateProbeSet(tt.port, tt.scheme, tt.overrides, tt.defaults)
			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateProbeSet() expected error but got none")
				}
				return
			} else if err != nil {
				t.Errorf("CreateProbeSet() unexpected error: %v", err)
				return
			}
			if probeSet == nil {
				t.Errorf("CreateProbeSet() returned nil probe set")
				return
			}
			if probeSet.Liveness == nil {
				t.Errorf("CreateProbeSet() liveness probe is nil")
			}
			if probeSet.Readiness == nil {
				t.Errorf("CreateProbeSet() readiness probe is nil")
			}

			if probeSet.Startup == nil {
				t.Errorf("CreateProbeSet() startup probe is nil")
			}
		})
	}
}

func TestOverrideSpecMethods(t *testing.T) {
	livenessProbe := &ProbeConf{Path: "/health"}
	readinessProbe := &ProbeConf{Path: "/ready"}
	startupProbe := &ProbeConf{Path: "/startup"}

	override := OverrideSpec{
		LivenessProbes:  livenessProbe,
		ReadinessProbes: readinessProbe,
		StartupProbes:   startupProbe,
	}
	if override.GetLivenessProbes() != livenessProbe {
		t.Errorf("GetLivenessProbes() = %v, want %v", override.GetLivenessProbes(), livenessProbe)
	}
	if override.GetReadinessProbes() != readinessProbe {
		t.Errorf("GetReadinessProbes() = %v, want %v", override.GetReadinessProbes(), readinessProbe)
	}
	if override.GetStartupProbes() != startupProbe {
		t.Errorf("GetStartupProbes() = %v, want %v", override.GetStartupProbes(), startupProbe)
	}
	emptyOverride := OverrideSpec{}
	if emptyOverride.GetLivenessProbes() != nil {
		t.Errorf("GetLivenessProbes() on empty override = %v, want nil", emptyOverride.GetLivenessProbes())
	}
	if emptyOverride.GetReadinessProbes() != nil {
		t.Errorf("GetReadinessProbes() on empty override = %v, want nil", emptyOverride.GetReadinessProbes())
	}
	if emptyOverride.GetStartupProbes() != nil {
		t.Errorf("GetStartupProbes() on empty override = %v, want nil", emptyOverride.GetStartupProbes())
	}
}

func TestCreateProbeSetWithActualOverrides(t *testing.T) {
	defaultLiveness := &ProbeConf{
		Path:                "/health",
		InitialDelaySeconds: 30,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		FailureThreshold:    3,
	}
	defaultReadiness := &ProbeConf{
		Path:                "/ready",
		InitialDelaySeconds: 15,
		TimeoutSeconds:      3,
		PeriodSeconds:       5,
		FailureThreshold:    1,
	}
	defaultStartup := &ProbeConf{
		Path:                "/startup",
		InitialDelaySeconds: 10,
		TimeoutSeconds:      10,
		PeriodSeconds:       5,
		FailureThreshold:    10,
	}

	defaults := OverrideSpec{
		LivenessProbes:  defaultLiveness,
		ReadinessProbes: defaultReadiness,
		StartupProbes:   defaultStartup,
	}

	tests := []struct {
		name              string
		overrides         OverrideSpec
		expectedLiveness  ProbeConf
		expectedReadiness ProbeConf
		expectedStartup   ProbeConf
	}{
		{
			name: "Override only liveness probe",
			overrides: OverrideSpec{
				LivenessProbes: &ProbeConf{
					Path:                "/custom-health",
					InitialDelaySeconds: 60,
					TimeoutSeconds:      10,
					PeriodSeconds:       20,
					FailureThreshold:    5,
				},
			},
			expectedLiveness: ProbeConf{
				Path:                "/custom-health",
				InitialDelaySeconds: 60,
				TimeoutSeconds:      10,
				PeriodSeconds:       20,
				FailureThreshold:    5,
			},
			expectedReadiness: *defaultReadiness,
			expectedStartup:   *defaultStartup,
		},
		{
			name: "Override all probes",
			overrides: OverrideSpec{
				LivenessProbes: &ProbeConf{
					Path:                "/override-health",
					InitialDelaySeconds: 45,
					TimeoutSeconds:      8,
					PeriodSeconds:       15,
					FailureThreshold:    4,
				},
				ReadinessProbes: &ProbeConf{
					Path:                "/override-ready",
					InitialDelaySeconds: 20,
					TimeoutSeconds:      6,
					PeriodSeconds:       8,
					FailureThreshold:    2,
				},
				StartupProbes: &ProbeConf{
					Path:                "/override-startup",
					InitialDelaySeconds: 25,
					TimeoutSeconds:      12,
					PeriodSeconds:       7,
					FailureThreshold:    15,
				},
			},
			expectedLiveness: ProbeConf{
				Path:                "/override-health",
				InitialDelaySeconds: 45,
				TimeoutSeconds:      8,
				PeriodSeconds:       15,
				FailureThreshold:    4,
			},
			expectedReadiness: ProbeConf{
				Path:                "/override-ready",
				InitialDelaySeconds: 20,
				TimeoutSeconds:      6,
				PeriodSeconds:       8,
				FailureThreshold:    2,
			},
			expectedStartup: ProbeConf{
				Path:                "/override-startup",
				InitialDelaySeconds: 25,
				TimeoutSeconds:      12,
				PeriodSeconds:       7,
				FailureThreshold:    15,
			},
		},
		{
			name:              "No overrides - use defaults",
			overrides:         OverrideSpec{},
			expectedLiveness:  *defaultLiveness,
			expectedReadiness: *defaultReadiness,
			expectedStartup:   *defaultStartup,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			probeSet, err := CreateProbeSet(8080, nil, tt.overrides, defaults)
			if err != nil {
				t.Fatalf("CreateProbeSet() unexpected error: %v", err)
			}
			// Validate liveness probe
			validateProbe(t, "liveness", probeSet.Liveness, tt.expectedLiveness)
			// Validate readiness probe
			validateProbe(t, "readiness", probeSet.Readiness, tt.expectedReadiness)
			// Validate startup probe
			validateProbe(t, "startup", probeSet.Startup, tt.expectedStartup)
		})
	}
}

func validateProbe(t *testing.T, probeType string, actual *v1.Probe, expected ProbeConf) {
	t.Helper()
	if actual == nil {
		t.Errorf("%s probe is nil", probeType)
		return
	}
	if actual.HTTPGet.Path != expected.Path {
		t.Errorf("%s probe path = %q, want %q", probeType, actual.HTTPGet.Path, expected.Path)
	}
	if actual.InitialDelaySeconds != expected.InitialDelaySeconds {
		t.Errorf("%s probe initialDelaySeconds = %d, want %d", probeType, actual.InitialDelaySeconds, expected.InitialDelaySeconds)
	}
	if actual.TimeoutSeconds != expected.TimeoutSeconds {
		t.Errorf("%s probe timeoutSeconds = %d, want %d", probeType, actual.TimeoutSeconds, expected.TimeoutSeconds)
	}
	if actual.PeriodSeconds != expected.PeriodSeconds {
		t.Errorf("%s probe periodSeconds = %d, want %d", probeType, actual.PeriodSeconds, expected.PeriodSeconds)
	}
	if actual.FailureThreshold != expected.FailureThreshold {
		t.Errorf("%s probe failureThreshold = %d, want %d", probeType, actual.FailureThreshold, expected.FailureThreshold)
	}
}
