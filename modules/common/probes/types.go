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
// +kubebuilder:object:generate:=true

// Package probes provides utilities for configuring Kubernetes liveness and
// readiness probes
package probes

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ProbeHandlerType defines the type of probe handler
// +kubebuilder:validation:Enum=HTTP;Exec;""
type ProbeHandlerType string

const (
	// ProbeHandlerHTTP configures an HTTP GET probe handler
	ProbeHandlerHTTP ProbeHandlerType = "HTTP"
	// ProbeHandlerExec configures an exec probe handler
	ProbeHandlerExec ProbeHandlerType = "Exec"
)

// ProbeConf - the configuration for liveness and readiness probes
type ProbeConf struct {
	// +kubebuilder:validation:Optional
	Type ProbeHandlerType `json:"type,omitempty"`
	// +kubebuilder:validation:Pattern=`^(/.*)?$`
	Path string `json:"path,omitempty"`
	// +listType=atomic
	Command []string `json:"command,omitempty"`
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port,omitempty"`
	// +kubebuilder:validation:Optional
	Scheme *v1.URIScheme `json:"scheme,omitempty"`
	// +kubebuilder:validation:Minimum=0
	InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty"`
	// +kubebuilder:validation:Minimum=1
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty"`
	// +kubebuilder:validation:Minimum=1
	PeriodSeconds int32 `json:"periodSeconds,omitempty"`
	// +kubebuilder:validation:Minimum=1
	FailureThreshold int32 `json:"failureThreshold,omitempty"`
}

// OverrideSpec to override StatefulSet fields
type OverrideSpec struct {
	// Override configuration for the StatefulSet like Probes and other tunable
	// fields
	LivenessProbes  *ProbeConf `json:"livenessProbes,omitempty"`
	ReadinessProbes *ProbeConf `json:"readinessProbes,omitempty"`
	StartupProbes   *ProbeConf `json:"startupProbes,omitempty"`
}

// ProbeSet holds all the probes for a service
type ProbeSet struct {
	Liveness  *v1.Probe
	Readiness *v1.Probe
	Startup   *v1.Probe
}

// ProbeOverrides interface that all override specs can implement
// +kubebuilder:object:generate:=false
type ProbeOverrides interface {
	GetLivenessProbes() *ProbeConf
	GetReadinessProbes() *ProbeConf
	GetStartupProbes() *ProbeConf
	ValidateProbes(basePath *field.Path) field.ErrorList
}

// GetLivenessProbes -
func (o OverrideSpec) GetLivenessProbes() *ProbeConf {
	return o.LivenessProbes
}

// GetReadinessProbes -
func (o OverrideSpec) GetReadinessProbes() *ProbeConf {
	return o.ReadinessProbes
}

// GetStartupProbes -
func (o OverrideSpec) GetStartupProbes() *ProbeConf {
	return o.StartupProbes
}

// ValidateProbes represents the entrypoint for webhook validation. It processes
// the ProbeSet (liveness, readiness and startup probes) and performs a validation
// over the overrides passed as input
func (o OverrideSpec) ValidateProbes(basePath *field.Path) field.ErrorList {
	errorList := field.ErrorList{}

	if o.GetLivenessProbes() != nil {
		errorList = append(errorList,
			ValidateProbeConf(basePath.Child("livenessProbe"), o.GetLivenessProbes())...)
	}

	if o.GetReadinessProbes() != nil {
		errorList = append(errorList,
			ValidateProbeConf(basePath.Child("readinessProbe"), o.GetReadinessProbes())...)
	}

	if o.GetStartupProbes() != nil {
		errorList = append(errorList,
			ValidateProbeConf(basePath.Child("startupProbe"), o.GetStartupProbes())...)
	}
	return errorList
}
