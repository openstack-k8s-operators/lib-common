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
	"k8s.io/apimachinery/pkg/util/validation/field"
)

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
			validateHTTPProbe(t, "liveness", probeSet.Liveness, tt.expectedLiveness)
			// Validate readiness probe
			validateHTTPProbe(t, "readiness", probeSet.Readiness, tt.expectedReadiness)
			// Validate startup probe
			validateHTTPProbe(t, "startup", probeSet.Startup, tt.expectedStartup)
		})
	}
}

func validateHTTPProbe(t *testing.T, probeType string, actual *v1.Probe, expected ProbeConf) {
	t.Helper()
	if actual == nil {
		t.Fatalf("%s probe is nil", probeType)
	}
	if actual.HTTPGet == nil {
		t.Fatalf("%s probe HTTPGet is nil", probeType)
	}
	if actual.Exec != nil {
		t.Errorf("%s probe Exec should be nil for HTTP probe", probeType)
	}
	if actual.HTTPGet.Path != expected.Path {
		t.Errorf("%s probe path = %q, want %q", probeType, actual.HTTPGet.Path, expected.Path)
	}
	if expected.Port != 0 && actual.HTTPGet.Port.IntValue() != int(expected.Port) {
		t.Errorf("%s probe port = %d, want %d", probeType, actual.HTTPGet.Port.IntValue(), expected.Port)
	}
	if expected.Scheme != nil && actual.HTTPGet.Scheme != *expected.Scheme {
		t.Errorf("%s probe scheme = %v, want %v", probeType, actual.HTTPGet.Scheme, *expected.Scheme)
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

func validateExecProbe(t *testing.T, probeType string, actual *v1.Probe, expected ProbeConf) {
	t.Helper()
	if actual == nil {
		t.Fatalf("%s probe is nil", probeType)
	}
	if actual.Exec == nil {
		t.Fatalf("%s probe Exec is nil", probeType)
	}
	if actual.HTTPGet != nil {
		t.Errorf("%s probe HTTPGet should be nil for exec probe", probeType)
	}
	if len(actual.Exec.Command) != len(expected.Command) {
		t.Errorf("%s probe command length = %d, want %d", probeType, len(actual.Exec.Command), len(expected.Command))
	} else {
		for i, cmd := range actual.Exec.Command {
			if cmd != expected.Command[i] {
				t.Errorf("%s probe command[%d] = %q, want %q", probeType, i, cmd, expected.Command[i])
			}
		}
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

func TestSetProbeConfV2ExecProbe(t *testing.T) {
	config := ProbeConf{
		Type:                ProbeHandlerExec,
		Command:             []string{"/bin/sh", "-c", "mysqladmin ping"},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      5,
		PeriodSeconds:       15,
		FailureThreshold:    3,
	}

	probe, err := SetProbeConfV2(config)
	if err != nil {
		t.Fatalf("SetProbeConfV2() unexpected error: %v", err)
	}
	validateExecProbe(t, "exec", probe, config)
}

func TestSetProbeConfV2HTTPProbe(t *testing.T) {
	scheme := v1.URISchemeHTTPS
	config := ProbeConf{
		Type:                ProbeHandlerHTTP,
		Path:                "/healthz",
		Port:                8443,
		Scheme:              &scheme,
		InitialDelaySeconds: 30,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		FailureThreshold:    3,
	}

	probe, err := SetProbeConfV2(config)
	if err != nil {
		t.Fatalf("SetProbeConfV2() unexpected error: %v", err)
	}
	validateHTTPProbe(t, "http", probe, config)
}

func TestSetProbeConfV2DefaultType(t *testing.T) {
	config := ProbeConf{
		Path: "/health",
		Port: 8080,
	}

	probe, err := SetProbeConfV2(config)
	if err != nil {
		t.Fatalf("SetProbeConfV2() unexpected error: %v", err)
	}
	if probe.HTTPGet == nil {
		t.Fatal("SetProbeConfV2() should default to HTTP when type is empty")
	}
}

func TestSetProbeConfV2Errors(t *testing.T) {
	tests := []struct {
		name   string
		config ProbeConf
	}{
		{
			name: "Exec with empty command",
			config: ProbeConf{
				Type: ProbeHandlerExec,
			},
		},
		{
			name: "HTTP with invalid port",
			config: ProbeConf{
				Type: ProbeHandlerHTTP,
				Port: 0,
				Path: "/health",
			},
		},
		{
			name: "HTTP with port too large",
			config: ProbeConf{
				Type: ProbeHandlerHTTP,
				Port: 70000,
				Path: "/health",
			},
		},
		{
			name: "Default type with zero port",
			config: ProbeConf{
				Path: "/health",
				Port: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SetProbeConfV2(tt.config)
			if err == nil {
				t.Error("SetProbeConfV2() expected error but got none")
			}
		})
	}
}

func TestMergeHTTPOverrides(t *testing.T) {
	scheme := v1.URISchemeHTTPS
	base := ProbeConf{
		Type:             ProbeHandlerHTTP,
		Path:             "/healthcheck",
		Port:             8080,
		TimeoutSeconds:   5,
		PeriodSeconds:    10,
		FailureThreshold: 3,
	}

	overrides := ProbeConf{
		Path:           "/healthcheck/v2",
		Port:           8443,
		Scheme:         &scheme,
		TimeoutSeconds: 10,
	}

	base.Merge(overrides)

	if base.Type != ProbeHandlerHTTP {
		t.Errorf("Merge() type = %q, want %q", base.Type, ProbeHandlerHTTP)
	}
	if base.Path != "/healthcheck/v2" {
		t.Errorf("Merge() path = %q, want /healthcheck/v2", base.Path)
	}
	if base.Port != 8443 {
		t.Errorf("Merge() port = %d, want 8443", base.Port)
	}
	if base.Scheme == nil || *base.Scheme != v1.URISchemeHTTPS {
		t.Errorf("Merge() scheme = %v, want HTTPS", base.Scheme)
	}
	if base.TimeoutSeconds != 10 {
		t.Errorf("Merge() timeoutSeconds = %d, want 10", base.TimeoutSeconds)
	}
	if base.PeriodSeconds != 10 {
		t.Errorf("Merge() periodSeconds = %d, want 10 (should preserve non-overridden)", base.PeriodSeconds)
	}
	if base.FailureThreshold != 3 {
		t.Errorf("Merge() failureThreshold = %d, want 3 (should preserve non-overridden)", base.FailureThreshold)
	}
}

func TestMergeExecCommandOverride(t *testing.T) {
	base := ProbeConf{
		Type:             ProbeHandlerExec,
		Command:          []string{"/bin/sh", "-c", "mysqladmin ping"},
		TimeoutSeconds:   5,
		PeriodSeconds:    10,
		FailureThreshold: 3,
	}

	overrides := ProbeConf{
		Command:        []string{"/bin/sh", "-c", "mysqladmin ping --connect-timeout=10"},
		TimeoutSeconds: 15,
	}

	base.Merge(overrides)

	if base.Type != ProbeHandlerExec {
		t.Errorf("Merge() type = %q, want %q", base.Type, ProbeHandlerExec)
	}
	if len(base.Command) != 3 || base.Command[2] != "mysqladmin ping --connect-timeout=10" {
		t.Errorf("Merge() command = %v, want [/bin/sh -c mysqladmin ping --connect-timeout=10]", base.Command)
	}
	if base.TimeoutSeconds != 15 {
		t.Errorf("Merge() timeoutSeconds = %d, want 15", base.TimeoutSeconds)
	}
	if base.PeriodSeconds != 10 {
		t.Errorf("Merge() periodSeconds = %d, want 10 (should preserve non-overridden)", base.PeriodSeconds)
	}
}

func TestMergeExecTimingOnly(t *testing.T) {
	base := ProbeConf{
		Type:             ProbeHandlerExec,
		Command:          []string{"/bin/sh", "-c", "mysqladmin ping"},
		TimeoutSeconds:   5,
		PeriodSeconds:    10,
		FailureThreshold: 3,
	}

	overrides := ProbeConf{
		TimeoutSeconds:   30,
		FailureThreshold: 6,
	}

	base.Merge(overrides)

	if base.Type != ProbeHandlerExec {
		t.Errorf("Merge() type = %q, want %q", base.Type, ProbeHandlerExec)
	}
	if base.Command[2] != "mysqladmin ping" {
		t.Errorf("Merge() command should be preserved, got %v", base.Command)
	}
	if base.TimeoutSeconds != 30 {
		t.Errorf("Merge() timeoutSeconds = %d, want 30", base.TimeoutSeconds)
	}
	if base.FailureThreshold != 6 {
		t.Errorf("Merge() failureThreshold = %d, want 6", base.FailureThreshold)
	}
	if base.PeriodSeconds != 10 {
		t.Errorf("Merge() periodSeconds = %d, want 10 (should preserve non-overridden)", base.PeriodSeconds)
	}
}

func TestMergeEmptyOverrides(t *testing.T) {
	base := ProbeConf{
		Type:             ProbeHandlerExec,
		Command:          []string{"/bin/sh", "-c", "mysqladmin ping"},
		TimeoutSeconds:   5,
		PeriodSeconds:    10,
		FailureThreshold: 3,
	}

	base.Merge(ProbeConf{})

	if base.Type != ProbeHandlerExec {
		t.Errorf("Merge() type should not be overwritten by empty, got %q", base.Type)
	}
	if len(base.Command) != 3 {
		t.Errorf("Merge() command should not be overwritten by nil, got %v", base.Command)
	}
	if base.TimeoutSeconds != 5 {
		t.Errorf("Merge() timeoutSeconds should not be overwritten by zero, got %d", base.TimeoutSeconds)
	}
}

func TestCreateProbeSetV2(t *testing.T) {
	defaults := OverrideSpec{
		LivenessProbes: &ProbeConf{
			Type:             ProbeHandlerHTTP,
			Path:             "/healthz",
			Port:             8080,
			TimeoutSeconds:   5,
			PeriodSeconds:    10,
			FailureThreshold: 3,
		},
		ReadinessProbes: &ProbeConf{
			Type:             ProbeHandlerExec,
			Command:          []string{"/bin/sh", "-c", "mysql -e 'SELECT 1'"},
			TimeoutSeconds:   3,
			PeriodSeconds:    5,
			FailureThreshold: 1,
		},
		StartupProbes: &ProbeConf{
			Type:             ProbeHandlerExec,
			Command:          []string{"/bin/sh", "-c", "mysql -e 'SELECT 1'"},
			TimeoutSeconds:   10,
			PeriodSeconds:    5,
			FailureThreshold: 10,
		},
	}

	overrides := OverrideSpec{
		LivenessProbes: &ProbeConf{
			TimeoutSeconds: 15,
		},
		StartupProbes: &ProbeConf{
			FailureThreshold: 20,
		},
	}

	probeSet, err := CreateProbeSetV2(overrides, defaults)
	if err != nil {
		t.Fatalf("CreateProbeSetV2() unexpected error: %v", err)
	}

	validateHTTPProbe(t, "liveness", probeSet.Liveness, ProbeConf{
		Path:             "/healthz",
		Port:             8080,
		TimeoutSeconds:   15,
		PeriodSeconds:    10,
		FailureThreshold: 3,
	})
	validateExecProbe(t, "readiness", probeSet.Readiness, *defaults.ReadinessProbes)
	validateExecProbe(t, "startup", probeSet.Startup, ProbeConf{
		Command:          []string{"/bin/sh", "-c", "mysql -e 'SELECT 1'"},
		TimeoutSeconds:   10,
		PeriodSeconds:    5,
		FailureThreshold: 20,
	})
}

func TestValidateProbeConfExec(t *testing.T) {
	tests := []struct {
		name      string
		config    *ProbeConf
		wantCount int
	}{
		{
			name: "Valid exec probe",
			config: &ProbeConf{
				Type:           ProbeHandlerExec,
				Command:        []string{"/bin/check"},
				TimeoutSeconds: 5,
			},
			wantCount: 0,
		},
		{
			name: "Exec probe missing command",
			config: &ProbeConf{
				Type: ProbeHandlerExec,
			},
			wantCount: 1,
		},
		{
			name: "Invalid type",
			config: &ProbeConf{
				Type: "BadType",
			},
			wantCount: 1,
		},
		{
			name: "Exec probe with invalid timing",
			config: &ProbeConf{
				Type:                ProbeHandlerExec,
				Command:             []string{"/bin/check"},
				InitialDelaySeconds: -1,
			},
			wantCount: 1,
		},
		{
			name: "Exec probe ignores path validation",
			config: &ProbeConf{
				Type:    ProbeHandlerExec,
				Command: []string{"/bin/check"},
				Path:    "no-leading-slash",
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateProbeConf(field.NewPath("spec"), tt.config)
			if len(errs) != tt.wantCount {
				t.Errorf("ValidateProbeConf() error count = %d, want %d: %v", len(errs), tt.wantCount, errs)
			}
		})
	}
}
