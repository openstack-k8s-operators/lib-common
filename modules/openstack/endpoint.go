/*
Copyright 2022 Red Hat

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

package openstack

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	gophercloud "github.com/gophercloud/gophercloud/v2"
	endpoints "github.com/gophercloud/gophercloud/v2/openstack/identity/v3/endpoints"
)

// Endpoint -
type Endpoint struct {
	Name         string
	ServiceID    string
	Availability gophercloud.Availability
	URL          string
}

// CreateEndpoint - create endpoint
func (o *OpenStack) CreateEndpoint(
	ctx context.Context,
	log logr.Logger,
	e Endpoint,
) (string, error) {
	// validate if endpoint already exist
	allEndpoints, err := o.GetEndpoints(
		ctx,
		log,
		e.ServiceID,
		string(e.Availability))
	if err != nil {
		return "", err
	}

	if len(allEndpoints) > 0 {
		return allEndpoints[0].ID, nil
	}

	// Create the endpoint
	createOpts := endpoints.CreateOpts{
		Availability: e.Availability,
		Name:         e.Name,
		Region:       o.region,
		ServiceID:    e.ServiceID,
		URL:          e.URL,
	}
	createdEndpoint, err := endpoints.Create(ctx, o.osclient, createOpts).Extract()
	if err != nil {
		return "", err
	}
	return createdEndpoint.ID, nil
}

// GetEndpoints - get endpoints for the registered service. if endpointInterface
// is provided, just return the endpoint for that type.
func (o *OpenStack) GetEndpoints(
	ctx context.Context,
	log logr.Logger,
	serviceID string,
	endpointInterface string,
) ([]endpoints.Endpoint, error) {
	log.Info(fmt.Sprintf("Getting Endpoints for service %s %s ", serviceID, endpointInterface))

	listOpts := endpoints.ListOpts{
		ServiceID: serviceID,
		RegionID:  o.region,
	}
	if endpointInterface != "" {
		availability, err := GetAvailability(endpointInterface)
		if err != nil {
			return nil, err
		}

		listOpts.Availability = availability
	}

	allPages, err := endpoints.List(o.osclient, listOpts).AllPages(ctx)
	if err != nil {
		return nil, err
	}
	allEndpoints, err := endpoints.ExtractEndpoints(allPages)
	if err != nil {
		return nil, err
	}

	log.Info("Getting Endpoint successfully")

	return allEndpoints, nil
}

// DeleteEndpoint - delete endpoint
func (o *OpenStack) DeleteEndpoint(
	ctx context.Context,
	log logr.Logger,
	e Endpoint,
) error {
	log.Info(fmt.Sprintf("Deleting Endpoint %s %s ", e.Name, e.Availability))

	// get all registered endpoints for the service/endpointInterface
	allEndpoints, err := o.GetEndpoints(ctx, log, e.ServiceID, string(e.Availability))
	if err != nil {
		return err
	}

	for _, endpt := range allEndpoints {
		err = endpoints.Delete(ctx, o.osclient, endpt.ID).ExtractErr()
		if err != nil {
			return err
		}

		log.Info(fmt.Sprintf("Deleted endpoint %s %s - %s", endpt.Name, string(endpt.Availability), endpt.URL))
	}

	return nil
}

// UpdateEndpoint -
func (o *OpenStack) UpdateEndpoint(
	ctx context.Context,
	log logr.Logger,
	e Endpoint,
	endpointID string,
) (string, error) {
	log.Info(fmt.Sprintf("Updating Endpoint %s %s ", e.Name, e.Availability))

	// Update the endpoint
	updateOpts := endpoints.UpdateOpts{
		Availability: e.Availability,
		Name:         e.Name,
		Region:       o.region,
		ServiceID:    e.ServiceID,
		URL:          e.URL,
	}
	endpt, err := endpoints.Update(ctx, o.osclient, endpointID, updateOpts).Extract()
	if err != nil {
		return "", err
	}

	log.Info(fmt.Sprintf("Updated Endpoint %s %s ", e.Name, e.Availability))

	return endpt.ID, nil
}
