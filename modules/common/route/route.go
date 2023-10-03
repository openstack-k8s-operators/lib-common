/*
Copyright 2020 Red Hat

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

package route

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	ctrl "sigs.k8s.io/controller-runtime"
)

// NewRoute returns an initialized Route.
func NewRoute(
	route *routev1.Route,
	timeout time.Duration,
	overrides []OverrideSpec,
) (*Route, error) {
	r := &Route{
		route:   route,
		timeout: timeout,
	}

	// patch route with possible overrides of Labels, Annotations and Spec
	for _, override := range overrides {
		if override.EmbeddedLabelsAnnotations != nil {
			if override.Labels != nil {
				r.route.Labels = util.MergeStringMaps(override.Labels, r.route.Labels)
			}
			if override.Annotations != nil {
				r.route.Annotations = util.MergeStringMaps(override.Annotations, r.route.Annotations)
			}
		}
		if override.Spec != nil {
			originalSpec, err := json.Marshal(r.route.Spec)
			if err != nil {
				return r, fmt.Errorf("error marshalling Route Spec: %w", err)
			}

			patch, err := json.Marshal(override.Spec)
			if err != nil {
				return r, fmt.Errorf("error marshalling Route Spec override: %w", err)
			}

			patchedJSON, err := strategicpatch.StrategicMergePatch(originalSpec, patch, routev1.RouteSpec{})
			if err != nil {
				return r, fmt.Errorf("error patching Route Spec: %w", err)
			}

			patchedSpec := routev1.RouteSpec{}
			err = json.Unmarshal(patchedJSON, &patchedSpec)
			if err != nil {
				return r, fmt.Errorf("error unmarshalling patched Route Spec: %w", err)
			}
			r.route.Spec = patchedSpec
		}
	}

	return r, nil
}

// GetHostname - returns the hostname of the created route
func (r *Route) GetHostname() string {
	return r.hostname
}

// GetRoute - returns the route object
func (r *Route) GetRoute() *routev1.Route {
	return r.route
}

// GenericRoute func
func GenericRoute(routeInfo *GenericRouteDetails) *routev1.Route {
	serviceRef := routev1.RouteTargetReference{
		Kind: "Service",
		Name: routeInfo.ServiceName,
	}
	routePort := &routev1.RoutePort{
		TargetPort: intstr.FromString(routeInfo.TargetPortName),
	}

	result := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      routeInfo.Name,
			Namespace: routeInfo.Namespace,
			Labels:    routeInfo.Labels,
		},
		Spec: routev1.RouteSpec{
			To:   serviceRef,
			Port: routePort,
		},
	}
	if len(routeInfo.FQDN) > 0 {
		result.Spec.Host = routeInfo.FQDN
	}
	return result
}

// CreateOrPatch - creates or patches a route, reconciles after Xs if object won't exist.
func (r *Route) CreateOrPatch(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.route.Name,
			Namespace: r.route.Namespace,
		},
	}

	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), route, func() error {
		route.Labels = util.MergeStringMaps(route.Labels, r.route.Labels)
		route.Annotations = util.MergeStringMaps(route.Annotations, r.route.Annotations)
		route.Spec = r.route.Spec
		if len(route.Spec.Host) == 0 && len(route.Status.Ingress) > 0 {
			route.Spec.Host = route.Status.Ingress[0].Host
		}

		err := controllerutil.SetControllerReference(h.GetBeforeObject(), route, h.GetScheme())
		if err != nil {
			return err
		}

		// Add the service CR to the ownerRef list of the route to prevent the route being deleted
		// before the service is deleted. Otherwise this can result cleanup issues which require
		// the endpoint to be reachable.
		// If ALL objects in the list have been deleted, this object will be garbage collected.
		// https://github.com/kubernetes/apimachinery/blob/15d95c0b2af3f4fcf46dce24105e5fbb9379af5a/pkg/apis/meta/v1/types.go#L240-L247
		for _, owner := range r.OwnerReferences {
			err := controllerutil.SetOwnerReference(owner, route, h.GetScheme())
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("Route %s not found, reconcile in %s", route.Name, r.timeout))
			return ctrl.Result{RequeueAfter: r.timeout}, nil
		}
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("Route %s - %s", route.Name, op))
	}

	// update the route instance with the host
	r.hostname = route.Spec.Host

	return ctrl.Result{}, nil
}

// Delete - delete a service.
func (r *Route) Delete(
	ctx context.Context,
	h *helper.Helper,
) error {

	err := h.GetClient().Delete(ctx, r.route)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return fmt.Errorf("Error deleting route %s: %w", r.route.Name, err)
	}

	return nil
}
