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
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ProbeConfig - the configuration for liveness and readiness probes
// LivenessPath - Endpoint path for the liveness probe
// ReadinessPath - Endpoint path for the readiness probe
// InitialDelaySeconds - Number of seconds after the container starts before liveness/readiness probes are initiated
// TimeoutSeconds - Number of seconds after which the probe times out
// PeriodSeconds - How often (in seconds) to perform the probe
type ProbeConfig struct {
	LivenessPath        string
	ReadinessPath       string
	InitialDelaySeconds int32
	TimeoutSeconds      int32
	PeriodSeconds       int32
}

// SetProbes - configures and returns liveness and readiness probes based on the provided settings
func SetProbes(port int, disableNonTLSListeners bool, config ProbeConfig) (*v1.Probe, *v1.Probe, error) {

	if port < 1 || port > 65535 {
		return nil, nil, fmt.Errorf("invalid port: %d", port)
	}

	var scheme v1.URIScheme
	if disableNonTLSListeners {
		scheme = v1.URISchemeHTTPS
	} else {
		scheme = v1.URISchemeHTTP
	}

	livenessProbe := &v1.Probe{
		ProbeHandler: v1.ProbeHandler{
			HTTPGet: &v1.HTTPGetAction{
				Path:   config.LivenessPath,
				Port:   intstr.FromInt(port),
				Scheme: scheme,
			},
		},
		InitialDelaySeconds: config.InitialDelaySeconds,
		TimeoutSeconds:      config.TimeoutSeconds,
		PeriodSeconds:       config.PeriodSeconds,
	}

	readinessProbe := &v1.Probe{
		ProbeHandler: v1.ProbeHandler{
			HTTPGet: &v1.HTTPGetAction{
				Path:   config.ReadinessPath,
				Port:   intstr.FromInt(port),
				Scheme: scheme,
			},
		},
		InitialDelaySeconds: config.InitialDelaySeconds,
		TimeoutSeconds:      config.TimeoutSeconds,
		PeriodSeconds:       config.PeriodSeconds,
	}

	return livenessProbe, readinessProbe, nil
}
