package util

import (
	corev1 "k8s.io/api/core/v1"
)

/*
Common functionality to add readiness/liveness Probe to a container using:

* default values:
readinessProbe 	:= common.Probe{ProbeType: "readiness"}
livenessProbe 	:= common.Probe{ProbeType: "liveness"}

* custom values:
readinessProbe := common.Probe{
	ProbeType: 		"readiness",
	Command:		"/to/some/command",
	InitialDelaySeconds: 	20,
 }
livenessProbe := common.Probe{ProbeType: "liveness"}

In container definition
Containers: []corev1.Container{
	{
		Name:  "nova-conductor",
		ReadinessProbe: readinessProbe.GetProbe(),
		LivenessProbe:  livenessProbe.GetProbe(),
		...
*/

const (
	// default Readiness values
	defaultReadinessInitialDelaySeconds int32 = 5
	defaultReadinessPeriodSeconds       int32 = 15
	defaultReadinessTimeoutSeconds      int32 = 3
	defaultReadinessFailureThreshold    int32 = 3
	// default Liveness values
	defaultLivenessInitialDelaySeconds int32  = 30
	defaultLivenessPeriodSeconds       int32  = 60
	defaultLivenessTimeoutSeconds      int32  = 3
	defaultLivenessFailureThreshold    int32  = 5
	defaultCommand                     string = "/openstack/healthcheck"
)

const (
	// ProbeTypeReadiness - readiness
	ProbeTypeReadiness ProbeType = "readiness"
	// ProbeTypeLiveness - liveness
	ProbeTypeLiveness ProbeType = "liveness"
)

// ProbeType -
type ProbeType string

// Probe details
type Probe struct {
	// ProbeType, either readiness, or liveness
	ProbeType           ProbeType
	Command             string
	InitialDelaySeconds int32 // min value 1
	PeriodSeconds       int32 // min value 1
	TimeoutSeconds      int32 // min value 1
	FailureThreshold    int32 // min value 1
}

// GetProbe -
func (p *Probe) GetProbe() *corev1.Probe {

	switch p.ProbeType {
	case ProbeTypeReadiness:
		if p.InitialDelaySeconds == 0 {
			p.InitialDelaySeconds = defaultReadinessInitialDelaySeconds
		}
		if p.PeriodSeconds == 0 {
			p.PeriodSeconds = defaultReadinessPeriodSeconds
		}
		if p.TimeoutSeconds == 0 {
			p.TimeoutSeconds = defaultReadinessTimeoutSeconds
		}
		if p.FailureThreshold == 0 {
			p.FailureThreshold = defaultReadinessFailureThreshold
		}
	case ProbeTypeLiveness:
		if p.InitialDelaySeconds == 0 {
			p.InitialDelaySeconds = defaultLivenessInitialDelaySeconds
		}
		if p.PeriodSeconds == 0 {
			p.PeriodSeconds = defaultLivenessPeriodSeconds
		}
		if p.TimeoutSeconds == 0 {
			p.TimeoutSeconds = defaultLivenessTimeoutSeconds
		}
		if p.FailureThreshold == 0 {
			p.FailureThreshold = defaultLivenessFailureThreshold
		}
	}
	if p.Command == "" {
		p.Command = defaultCommand
	}

	return &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{
					p.Command,
				},
			},
		},
		InitialDelaySeconds: p.InitialDelaySeconds,
		PeriodSeconds:       p.PeriodSeconds,
		TimeoutSeconds:      p.TimeoutSeconds,
		FailureThreshold:    p.FailureThreshold,
	}
}
