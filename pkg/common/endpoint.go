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

package common

import (
	"context"
	"fmt"
	"strings"

	"github.com/openstack-k8s-operators/lib-common/pkg/helper"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Endpoint - typedef to enumerate Endpoint verbs
type Endpoint string

const (
	// EndpointAdmin - admin endpoint
	EndpointAdmin Endpoint = "admin"
	// EndpointInternal - internal endpoint
	EndpointInternal Endpoint = "internal"
	// EndpointPublic - public endpoint
	EndpointPublic Endpoint = "public"
)

// EndpointData - information for generation of K8S services and Keystone endpoint URLs
type EndpointData struct {
	// Used in k8s service definition
	Port int32
	// An optional path suffix to append to route hostname when forming Keystone endpoint URLs
	Path string
}

//
// ExposeEndpoints - creates services, routes and returns a map of created openstack endpoint
//
func ExposeEndpoints(
	ctx context.Context,
	h *helper.Helper,
	serviceName string,
	endpointSelector map[string]string,
	endpoints map[Endpoint]EndpointData,
) (map[string]string, ctrl.Result, error) {
	endpointMap := make(map[string]string)

	for endpointType, data := range endpoints {

		endpointName := serviceName + "-" + string(endpointType)
		exportLabels := MergeStringMaps(
			endpointSelector,
			map[string]string{
				string(endpointType): "true",
			},
		)
		//
		// Create the service if none exists
		//
		svc := NewService(
			GenericService(&GenericServiceDetails{
				Name:      endpointName,
				Namespace: h.GetBeforeObject().GetNamespace(),
				Labels:    exportLabels,
				Selector:  endpointSelector,
				Port: GenericServicePort{
					Name:     endpointName,
					Port:     data.Port,
					Protocol: corev1.ProtocolTCP,
				}}),
			exportLabels,
			5,
		)
		ctrlResult, err := svc.CreateOrPatch(ctx, h)
		if err != nil {
			return endpointMap, ctrlResult, err
		} else if (ctrlResult != ctrl.Result{}) {
			return endpointMap, ctrlResult, nil
		}
		// create service - end

		// Create the route if none exists
		// TODO TLS
		route := NewRoute(
			GenericRoute(&GenericRouteDetails{
				Name:           endpointName,
				Namespace:      h.GetBeforeObject().GetNamespace(),
				Labels:         exportLabels,
				ServiceName:    endpointName,
				TargetPortName: endpointName,
			}),
			exportLabels,
			5,
		)

		ctrlResult, err = route.CreateOrPatch(ctx, h)
		if err != nil {
			return endpointMap, ctrlResult, err
		} else if (ctrlResult != ctrl.Result{}) {
			return endpointMap, ctrlResult, nil
		}
		// create route - end

		//
		// Update instance status with service endpoint url from route host information
		//
		// TODO: need to support https default here
		var apiEndpoint string
		if !strings.HasPrefix(route.GetHostname(), "http") {
			apiEndpoint = fmt.Sprintf("http://%s%s", route.GetHostname(), data.Path)
		} else {
			apiEndpoint = fmt.Sprintf("%s%s", route.GetHostname(), data.Path)
		}
		// TODO: This check does not allow for certain "special" OpenStack endpoint
		//       formats.  For instance, "http://<ip>:<port>/v2/%(project_id)s" fails
		//       to parse because "%(" is an invalid escape character
		// u, err := url.Parse(apiEndpoint)
		// if err != nil {
		// 	return endpointMap, ctrlResult, err
		// }

		endpointMap[string(endpointType)] = apiEndpoint
	}

	return endpointMap, ctrl.Result{}, nil
}
