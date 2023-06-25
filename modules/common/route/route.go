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
	"fmt"
	"time"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

// NewRoute returns an initialized Route.
func NewRoute(
	route *routev1.Route,
	labels map[string]string,
	timeout time.Duration,
) *Route {
	return &Route{
		route:   route,
		timeout: timeout,
	}
}

// GetHostname - returns the hostname of the created route
func (r *Route) GetHostname() string {
	return r.hostname
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
			Name:        routeInfo.Name,
			Namespace:   routeInfo.Namespace,
			Labels:      routeInfo.Labels,
			Annotations: routeInfo.Annotations,
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
			Name:        r.route.Name,
			Namespace:   r.route.Namespace,
			Annotations: r.route.Annotations,
		},
	}

	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), route, func() error {
		route.Labels = util.MergeStringMaps(route.Labels, r.route.Labels)
		route.Annotations = r.route.Annotations
		route.Spec = r.route.Spec
		if len(route.Spec.Host) == 0 && len(route.Status.Ingress) > 0 {
			route.Spec.Host = route.Status.Ingress[0].Host
		}

		err := controllerutil.SetControllerReference(h.GetBeforeObject(), route, h.GetScheme())
		if err != nil {
			return err
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
