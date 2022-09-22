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
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	services "github.com/gophercloud/gophercloud/openstack/identity/v3/services"
)

// Service -
type Service struct {
	Name        string
	Type        string
	Description string
	Enabled     bool
}

//
// CreateService - create service
//
func (o *OpenStack) CreateService(
	log logr.Logger,
	s Service,
) (string, error) {
	var serviceID string

	service, err := o.GetService(
		log,
		s.Type,
		s.Name,
	)
	if err != nil {
		return serviceID, err
	}

	// if there is already a service, use it
	if service != nil {
		serviceID = service.ID
	} else {
		createOpts := services.CreateOpts{
			Type:    s.Type,
			Enabled: &s.Enabled,
			Extra: map[string]interface{}{
				"name":        s.Name,
				"description": s.Description,
			},
		}

		service, err := services.Create(o.GetOSClient(), createOpts).Extract()
		if err != nil {
			return serviceID, err
		}
		log.Info(fmt.Sprintf("Service Created - Servicename %s, ID %s", s.Name, service.ID))
		serviceID = service.ID
	}

	return serviceID, nil
}

//
// GetService - get service with type and name
//
func (o *OpenStack) GetService(
	log logr.Logger,
	serviceType string,
	serviceName string,
) (*services.Service, error) {
	listOpts := services.ListOpts{
		ServiceType: serviceType,
		Name:        serviceName,
	}

	allPages, err := services.List(o.osclient, listOpts).AllPages()
	if err != nil {
		return nil, err
	}
	allServices, err := services.ExtractServices(allPages)
	if err != nil {
		return nil, err
	}

	if len(allServices) == 0 {
		return nil, fmt.Errorf(fmt.Sprintf("%s service not found in keystone", serviceName))
	}

	return &allServices[0], nil
}

//
// UpdateService - update service with type and name
//
func (o *OpenStack) UpdateService(
	log logr.Logger,
	s Service,
	serviceID string,
) error {
	updateOpts := services.UpdateOpts{
		Type:    s.Type,
		Enabled: &s.Enabled,
		Extra: map[string]interface{}{
			"name":        s.Name,
			"description": s.Description,
		},
	}
	_, err := services.Update(o.GetOSClient(), serviceID, updateOpts).Extract()
	if err != nil {
		return err
	}
	return nil
}

//
// DeleteService - delete service with serviceID
//
func (o *OpenStack) DeleteService(
	log logr.Logger,
	serviceID string,
) error {
	log.Info(fmt.Sprintf("Delete service with id %s", serviceID))
	err := services.Delete(o.GetOSClient(), serviceID).ExtractErr()
	if err != nil && !strings.Contains(err.Error(), "Resource not found") {
		return err
	}

	return nil
}
