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
	"net"
	"net/url"
	"slices"
	"strings"

	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"k8s.io/apimachinery/pkg/util/validation"
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

// ValidateEndpointURLPort rejects an override whose endpointURL implies a
// port (explicit :port, or the scheme's standard port when omitted - see
// service.GetEndpointURLPort) that disagrees with a spec.ports[].port it also
// sets. No error when only one side is set, the URL's scheme has no implied
// port, or the URL is unparseable (URL shape validation is a separate
// concern, not this helper's). Service-agnostic: a pure int compare after
// service.GetEndpointURLPort. Nil-safe on override.Spec.
//
// basePath should point at the endpoint override, e.g.
// field.NewPath("spec", "override", "service", "public"); this appends
// .spec.ports itself.
func ValidateEndpointURLPort(basePath *field.Path, override service.RoutedOverrideSpec) field.ErrorList {
	allErrs := field.ErrorList{}

	if override.Spec == nil {
		return allErrs
	}

	urlPort, ok, err := service.GetEndpointURLPort(override.EndpointURL)
	if err != nil || !ok {
		return allErrs
	}

	for i, p := range override.Spec.Ports {
		if p.Port != urlPort {
			allErrs = append(allErrs, field.Invalid(
				basePath.Child("spec").Child("ports").Index(i).Child("port"),
				p.Port,
				fmt.Sprintf("must match the port in endpointURL %q", *override.EndpointURL),
			))
		}
	}

	return allErrs
}

// ValidateEndpointURL rejects a non-empty override.EndpointURL whose host
// cannot serve as an SNI ServerName / DNS SAN: the URL must parse with an
// http or https scheme and a host, the host must be RFC 1123 subdomain valid with at least
// two dot-separated segments (rejects a bare short name like
// "https://placement"), and the host must not be an IP literal (an IP cannot
// be an SNI ServerName / DNS SAN), and an explicit port must be 1-65535 (this
// is where an out-of-range port is reported; ValidateEndpointURLPort skips
// URLs whose port does not parse). No error when endpointURL is nil or empty.
// Nil-safe on override.EndpointURL.
//
// basePath should point at the endpoint override, e.g.
// field.NewPath("spec", "override", "service", "public"); this appends
// .endpointURL itself.
func ValidateEndpointURL(basePath *field.Path, override service.RoutedOverrideSpec) field.ErrorList {
	if override.EndpointURL == nil || *override.EndpointURL == "" {
		return nil
	}

	raw := *override.EndpointURL
	fldPath := basePath.Child("endpointURL")

	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Hostname() == "" {
		return field.ErrorList{field.Invalid(fldPath, raw, "must be a valid URL with a scheme and host")}
	}

	if scheme := service.Protocol(u.Scheme); scheme != service.ProtocolHTTP && scheme != service.ProtocolHTTPS {
		return field.ErrorList{field.Invalid(fldPath, raw, "scheme must be http or https")}
	}

	if _, _, err := service.GetEndpointURLPort(&raw); err != nil {
		return field.ErrorList{field.Invalid(fldPath, raw, "port must be a number between 1 and 65535")}
	}

	host := u.Hostname()
	if net.ParseIP(strings.TrimSuffix(host, ".")) != nil {
		return field.ErrorList{field.Invalid(fldPath, raw, "host must not be an IP literal")}
	}

	errs := validation.IsFullyQualifiedDomainName(fldPath, host)
	if len(errs) == 0 {
		return nil
	}
	// Report the raw endpointURL as the offending value, matching the other
	// branches above, rather than IsFullyQualifiedDomainName's bare host.
	out := make(field.ErrorList, len(errs))
	for i, e := range errs {
		out[i] = field.Invalid(fldPath, raw, e.Detail)
	}
	return out
}

// ValidatePortNameOverrides - validates, for every endpoint in overrides, that
// EndpointURL (if set) is valid (see ValidateEndpointURL), and, for endpoints
// also present in validNames, that Spec.Ports (if set) only uses names from
// validNames[endpoint] and agrees with any port implied by endpointURL.
// Endpoints absent from validNames get only the endpointURL check (not every
// service override needs a port-name guard; e.g. a service with no director
// compat-port use case on a given endpoint).
//
// basePath should point at the service override map, e.g.
// field.NewPath("spec", "override", "service"); this appends
// .<endpoint>.spec.ports itself.
//
// Endpoints are visited in sorted order so the returned field.ErrorList has a
// deterministic order regardless of overrides' map iteration order.
func ValidatePortNameOverrides(
	basePath *field.Path,
	overrides map[service.Endpoint]service.RoutedOverrideSpec,
	validNames map[service.Endpoint][]string,
) field.ErrorList {
	allErrs := field.ErrorList{}

	endpoints := make([]service.Endpoint, 0, len(overrides))
	for e := range overrides {
		endpoints = append(endpoints, e)
	}
	slices.Sort(endpoints)

	for _, e := range endpoints {
		ov := overrides[e]
		allErrs = append(allErrs, ValidateEndpointURL(basePath.Key(string(e)), ov)...)
		if _, ok := validNames[e]; !ok || ov.Spec == nil {
			continue
		}
		allErrs = append(allErrs, ValidateServicePortNames(
			basePath.Key(string(e)).Child("spec").Child("ports"),
			ov.Spec.Ports,
			validNames[e]...,
		)...)
		allErrs = append(allErrs, ValidateEndpointURLPort(basePath.Key(string(e)), ov)...)
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
