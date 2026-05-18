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
// +kubebuilder:object:generate:=true

// The probes package provides utilities for configuring Kubernetes liveness
// and readiness probes

package probes

import (
	"fmt"

	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"strings"
)

// Merge applies non-zero override values onto the receiver
func (p *ProbeConf) Merge(overrides ProbeConf) {
	if overrides.Type != "" {
		p.Type = overrides.Type
	}
	if overrides.Path != "" {
		p.Path = overrides.Path
	}
	if len(overrides.Command) > 0 {
		p.Command = overrides.Command
	}
	if overrides.Port > 0 {
		p.Port = overrides.Port
	}
	if overrides.Scheme != nil {
		p.Scheme = overrides.Scheme
	}
	if overrides.InitialDelaySeconds > 0 {
		p.InitialDelaySeconds = overrides.InitialDelaySeconds
	}
	if overrides.TimeoutSeconds > 0 {
		p.TimeoutSeconds = overrides.TimeoutSeconds
	}
	if overrides.PeriodSeconds > 0 {
		p.PeriodSeconds = overrides.PeriodSeconds
	}
	if overrides.FailureThreshold > 0 {
		p.FailureThreshold = overrides.FailureThreshold
	}
}

// CreateProbeSet creates all probes at once using the interface.
// Port and scheme are applied to all probes as HTTP GET handler parameters.
// For mixed probe types (e.g. HTTP and exec), use CreateProbeSetV2 instead
// with Port/Scheme set in the individual ProbeConf defaults.
func CreateProbeSet(
	port int32,
	scheme *v1.URIScheme,
	overrides ProbeOverrides,
	defaults OverrideSpec,
) (*ProbeSet, error) {
	for _, p := range []*ProbeConf{defaults.LivenessProbes, defaults.ReadinessProbes, defaults.StartupProbes} {
		if p != nil {
			p.Port = port
			p.Scheme = scheme
		}
	}
	return CreateProbeSetV2(overrides, defaults)
}

// CreateProbeSetV2 creates all probes at once using the interface.
// Each probe's handler type, port, scheme, path, and command are determined
// by the ProbeConf fields in the defaults and overrides.
func CreateProbeSetV2(
	overrides ProbeOverrides,
	defaults OverrideSpec,
) (*ProbeSet, error) {

	mergeConf := func(base *ProbeConf, override *ProbeConf) ProbeConf {
		if base == nil {
			base = &ProbeConf{}
		}
		conf := *base
		if override != nil {
			conf.Merge(*override)
		}
		return conf
	}

	livenessProbe, err := SetProbeConfV2(
		mergeConf(defaults.LivenessProbes, overrides.GetLivenessProbes()),
	)
	if err != nil {
		return nil, err
	}

	readinessProbe, err := SetProbeConfV2(
		mergeConf(defaults.ReadinessProbes, overrides.GetReadinessProbes()),
	)
	if err != nil {
		return nil, err
	}

	startupProbe, err := SetProbeConfV2(
		mergeConf(defaults.StartupProbes, overrides.GetStartupProbes()),
	)
	if err != nil {
		return nil, err
	}

	return &ProbeSet{
		Liveness:  livenessProbe,
		Readiness: readinessProbe,
		Startup:   startupProbe,
	}, nil
}

// SetProbeConfV2 configures and returns a probe based on the ProbeConf
// settings. The probe handler type is determined by ProbeConf.Type:
// ProbeHandlerExec creates an exec probe using ProbeConf.Command, while
// ProbeHandlerHTTP (or unset) creates an HTTP GET probe using ProbeConf.Port,
// ProbeConf.Path, and ProbeConf.Scheme.
func SetProbeConfV2(config ProbeConf) (*v1.Probe, error) {
	probe := &v1.Probe{
		InitialDelaySeconds: config.InitialDelaySeconds,
		TimeoutSeconds:      config.TimeoutSeconds,
		PeriodSeconds:       config.PeriodSeconds,
		FailureThreshold:    config.FailureThreshold,
	}

	switch config.Type {
	case ProbeHandlerExec:
		if len(config.Command) == 0 {
			return nil, util.ErrExecProbeCommandRequired
		}
		probe.ProbeHandler = v1.ProbeHandler{
			Exec: &v1.ExecAction{
				Command: config.Command,
			},
		}
	default:
		if config.Port < 1 || config.Port > 65535 {
			return nil, fmt.Errorf("%w: %d", util.ErrInvalidPort, config.Port)
		}
		probe.ProbeHandler = v1.ProbeHandler{
			HTTPGet: &v1.HTTPGetAction{
				Path: config.Path,
				Port: intstr.FromInt32(config.Port),
			},
		}
		if config.Scheme != nil {
			probe.HTTPGet.Scheme = *config.Scheme
		}
	}
	return probe, nil
}

// ValidateProbeConf validates probe configuration overrides for use at the
// webhook level
func ValidateProbeConf(basePath *field.Path, config *ProbeConf) field.ErrorList {
	errorList := field.ErrorList{}
	if config == nil {
		return errorList
	}

	switch config.Type {
	case ProbeHandlerExec:
		if len(config.Command) == 0 {
			errorList = append(errorList, field.Required(basePath.Child("command"),
				"command is required for exec probe type"))
		}
	case "", ProbeHandlerHTTP:
		if config.Path != "" && !strings.HasPrefix(config.Path, "/") {
			errorList = append(errorList, field.Invalid(basePath.Child("path"), config.Path,
				"path must start with '/' if specified"))
		}
	default:
		errorList = append(errorList, field.Invalid(basePath.Child("type"), config.Type,
			"type must be one of: HTTP, Exec, or unset"))
	}

	if config.InitialDelaySeconds < 0 {
		errorList = append(errorList, field.Invalid(basePath.Child("initialDelaySeconds"), config.InitialDelaySeconds,
			"initialDelaySeconds must be non-negative"))
	}

	if config.TimeoutSeconds != 0 && config.TimeoutSeconds < 1 {
		errorList = append(errorList, field.Invalid(basePath.Child("timeoutSeconds"), config.TimeoutSeconds,
			"timeoutSeconds must be at least 1 second when set"))
	}

	if config.PeriodSeconds != 0 && config.PeriodSeconds < 1 {
		errorList = append(errorList, field.Invalid(basePath.Child("periodSeconds"), config.PeriodSeconds,
			"periodSeconds must be at least 1 second when set"))
	}

	if config.FailureThreshold != 0 && config.FailureThreshold < 1 {
		errorList = append(errorList, field.Invalid(basePath.Child("failureThreshold"), config.FailureThreshold,
			"failureThreshold must be at least 1 when set"))
	}

	return errorList
}
