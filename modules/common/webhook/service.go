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

package webhook

import (
	"fmt"
	"slices"

	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateServicePortNames - validates that every port.Name in ports is one of
// validNames. Used by service operators' admission webhooks to guard a raw
// override.service.<endpoint>.spec.ports[].name escape hatch: lib-common's
// Service by-name port merge (OverrideServiceSpec.Ports) only replaces a base
// port when the name matches exactly; a mismatched name instead fails the
// Service reconcile with ErrUnknownServicePort. This check surfaces that as a
// clear admission-time rejection instead.
//
// basePath should already point at the ports field, e.g.
// field.NewPath("spec", "override", "service", "public", "spec", "ports").
//
// Callers compute validNames themselves since it is not derivable generically:
// most services use a single "<ServiceName>-<endpoint>" name, but some (e.g.
// glance, keyed per API instance, or horizon, which uses the bare "horizon")
// do not follow that pattern.
func ValidateServicePortNames(basePath *field.Path, ports []service.OverrideServicePort, validNames ...string) field.ErrorList {
	allErrs := field.ErrorList{}

	var detail string
	switch len(validNames) {
	case 0:
		detail = "no override port names are valid here"
	case 1:
		detail = fmt.Sprintf("must be %q", validNames[0])
	default:
		detail = fmt.Sprintf("must be one of %q", validNames)
	}

	for i, p := range ports {
		if !slices.Contains(validNames, p.Name) {
			allErrs = append(allErrs, field.Invalid(basePath.Index(i).Child("name"), p.Name, detail))
		}
	}

	return allErrs
}

// ValidatePortNameOverrides - validates, for every endpoint present in
// validNames, that overrides[endpoint].Spec.Ports (if set) only uses names
// from validNames[endpoint]. Endpoints absent from validNames are not
// checked (not every service override needs a port-name guard; e.g. a
// service with no director compat-port use case on a given endpoint).
//
// basePath should point at the service override map, e.g.
// field.NewPath("spec", "override", "service"); this appends
// .<endpoint>.spec.ports itself.
//
// Endpoints are visited in sorted order so the returned field.ErrorList has a
// deterministic order regardless of validNames' map iteration order.
func ValidatePortNameOverrides(
	basePath *field.Path,
	overrides map[service.Endpoint]service.RoutedOverrideSpec,
	validNames map[service.Endpoint][]string,
) field.ErrorList {
	allErrs := field.ErrorList{}

	endpoints := make([]service.Endpoint, 0, len(validNames))
	for e := range validNames {
		endpoints = append(endpoints, e)
	}
	slices.Sort(endpoints)

	for _, e := range endpoints {
		ov, ok := overrides[e]
		if !ok || ov.Spec == nil {
			continue
		}
		allErrs = append(allErrs, ValidateServicePortNames(
			basePath.Key(string(e)).Child("spec").Child("ports"),
			ov.Spec.Ports,
			validNames[e]...,
		)...)
	}

	return allErrs
}

// DefaultServicePortNames returns the conventional Service port name for each
// endpoint ("public"/"internal"). This is an opt-in convenience for callers —
// ValidateServicePortNames and ValidatePortNameOverrides always take validNames
// explicitly and never assume this (or any) naming convention internally.
func DefaultServicePortNames() map[service.Endpoint][]string {
	return map[service.Endpoint][]string{
		service.EndpointPublic:   {string(service.EndpointPublic)},
		service.EndpointInternal: {string(service.EndpointInternal)},
	}
}
