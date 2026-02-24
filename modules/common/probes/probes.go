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

func (p *ProbeConf) merge(overrides ProbeConf) {
	// Override path if provided
	if overrides.Path != "" {
		p.Path = overrides.Path
	}
	// Override timing values if they are non-zero
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

// CreateProbeSet - creates all probes at once using the interface
func CreateProbeSet(
	port int32,
	scheme *v1.URIScheme,
	overrides ProbeOverrides,
	defaults OverrideSpec,
) (*ProbeSet, error) {

	livenessProbe, err := SetProbeConf(
		port,
		scheme,
		func() ProbeConf {
			if defaults.LivenessProbes == nil {
				defaults.LivenessProbes = &ProbeConf{}
			}
			baseConf := *defaults.LivenessProbes
			if p := overrides.GetLivenessProbes(); p != nil {
				baseConf.merge(*p)
			}
			return baseConf
		}(),
	)

	// Could not process probes config
	if err != nil {
		return nil, err
	}

	readinessProbe, err := SetProbeConf(
		port,
		scheme,
		func() ProbeConf {
			if defaults.ReadinessProbes == nil {
				defaults.ReadinessProbes = &ProbeConf{}
			}
			baseConf := *defaults.ReadinessProbes
			if p := overrides.GetReadinessProbes(); p != nil {
				baseConf.merge(*p)
			}
			return baseConf
		}(),
	)

	// Could not process probes config
	if err != nil {
		return nil, err
	}

	startupProbe, err := SetProbeConf(
		port,
		scheme,
		func() ProbeConf {
			if defaults.StartupProbes == nil {
				defaults.StartupProbes = &ProbeConf{}
			}
			baseConf := *defaults.StartupProbes
			if p := overrides.GetStartupProbes(); p != nil {
				baseConf.merge(*p)
			}
			return baseConf
		}(),
	)

	// Could not process probes config
	if err != nil {
		return nil, err
	}

	return &ProbeSet{
		Liveness:  livenessProbe,
		Readiness: readinessProbe,
		Startup:   startupProbe,
	}, nil
}

// SetProbeConf configures and returns liveness and readiness probes based on
// the provided settings
func SetProbeConf(port int32, scheme *v1.URIScheme, config ProbeConf) (*v1.Probe, error) {
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("%w: %d", util.ErrInvalidPort, port)
	}
	probe := &v1.Probe{
		ProbeHandler: v1.ProbeHandler{
			HTTPGet: &v1.HTTPGetAction{
				Path: config.Path,
				Port: intstr.FromInt32(port),
			},
		},
		InitialDelaySeconds: config.InitialDelaySeconds,
		TimeoutSeconds:      config.TimeoutSeconds,
		PeriodSeconds:       config.PeriodSeconds,
		FailureThreshold:    config.FailureThreshold,
	}
	if scheme != nil {
		probe.HTTPGet.Scheme = *scheme
	}
	return probe, nil
}

// ValidateProbeConf - This function can be used at webhooks level to explicitly
// validate the overrides
func ValidateProbeConf(basePath *field.Path, config *ProbeConf) field.ErrorList {
	errorList := field.ErrorList{}
	// nothing to validate, return an empty errorList
	if config == nil {
		return errorList
	}
	// Path validation: fail is explicitly set as an empty string
	// or the endpoint does't start with "/"
	if config.Path != "" && !strings.HasPrefix(config.Path, "/") {
		err := field.Invalid(basePath.Child("path"), config.Path,
			"path must start with '/' if specified")
		errorList = append(errorList, err)
	}

	// InitialDelaySeconds validation: must be > 0
	if config.InitialDelaySeconds < 0 {
		err := field.Invalid(basePath.Child("initialDelaySeconds"), config.InitialDelaySeconds,
			"initialDelaySeconds must be non-negative")
		errorList = append(errorList, err)
	}

	// TimeoutSeconds validation: fail if it's a negative number
	if config.TimeoutSeconds != 0 && config.TimeoutSeconds < 1 {
		err := field.Invalid(basePath.Child("timeoutSeconds"), config.TimeoutSeconds,
			"timeoutSeconds must be at least 1 second when set")
		errorList = append(errorList, err)
	}

	// PeriodSeconds validation: fail if it's set as a negative number
	if config.PeriodSeconds != 0 && config.PeriodSeconds < 1 {
		err := field.Invalid(basePath.Child("periodSeconds"), config.PeriodSeconds,
			"periodSeconds must be at least 1 second when set")
		errorList = append(errorList, err)
	}

	// FailureThreshold validation: fail if it's set as a negative number
	if config.FailureThreshold != 0 && config.FailureThreshold < 1 {
		err := field.Invalid(basePath.Child("failureThreshold"), config.FailureThreshold,
			"failureThreshold must be at least 1 when set")
		errorList = append(errorList, err)
	}

	return errorList
}
