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

package openstack

import (
	"fmt"

	"github.com/go-logr/logr"
	limits "github.com/gophercloud/gophercloud/openstack/identity/v3/limits"
	registeredlimits "github.com/gophercloud/gophercloud/openstack/identity/v3/registeredlimits"
)

// Limit -
type Limit struct {
	// RegionID is the ID of the region where the limit is applied.
	RegionID string `json:"region_id,omitempty"`

	// DomainID is the ID of the domain where the limit is applied.
	DomainID string `json:"domain_id,omitempty"`

	// ProjectID is the ID of the project where the limit is applied.
	ProjectID string `json:"project_id,omitempty"`

	// ServiceID is the ID of the service where the limit is applied.
	ServiceID string `json:"service_id" required:"true"`

	// Description of the limit.
	Description string `json:"description,omitempty"`

	// ResourceName is the name of the resource that the limit is applied to.
	ResourceName string `json:"resource_name" required:"true"`

	// ResourceLimit is the override limit.
	ResourceLimit int `json:"resource_limit" required:"true"`
}

// CreateLimit - create a limit in keystone for particular project if it does not exist
func (o *OpenStack) CreateLimit(
	log logr.Logger,
	l Limit,
) (string, error) {
	var limitID string

	allPages, err := limits.List(o.osclient, limits.ListOpts{ResourceName: l.ResourceName}).AllPages()
	if err != nil {
		return limitID, err
	}

	allLimits, err := limits.ExtractLimits(allPages)
	if err != nil {
		return limitID, err
	}

	if len(allLimits) == 1 {
		limitID = allLimits[0].ID
	} else if len(allLimits) == 0 {
		createOpts := limits.BatchCreateOpts{
			limits.CreateOpts{
				ResourceName:  l.ResourceName,
				Description:   l.Description,
				ResourceLimit: l.ResourceLimit,
				ServiceID:     l.ServiceID,
				ProjectID:     l.ProjectID,
				DomainID:      l.DomainID,
				RegionID:      l.RegionID,
			},
		}
		log.Info(fmt.Sprintf("Creating limit %s", l.ResourceName))
		createdLimits, err := limits.BatchCreate(o.osclient, createOpts).Extract()
		if err != nil {
			return limitID, err
		}
		limitID = createdLimits[0].ID
	} else {
		return limitID, fmt.Errorf(fmt.Sprintf("multiple limits named \"%s\" found", l.ResourceName))
	}

	return limitID, nil
}

// RegisteredLimit -
type RegisteredLimit struct {
	// RegionID is the ID of the region where the limit is applied.
	RegionID string `json:"region_id,omitempty"`

	// ServiceID is the ID of the service where the limit is applied.
	ServiceID string `json:"service_id" required:"true"`

	// Description of the limit.
	Description string `json:"description,omitempty"`

	// ResourceName is the name of the resource that the limit is applied to.
	ResourceName string `json:"resource_name" required:"true"`

	// DefaultLimit is the default limit.
	DefaultLimit int `json:"default_limit"`
}

// CreateRegisteredLimit - create limit in keystone (global across projects) if it does not exist
func (o *OpenStack) CreateRegisteredLimit(
	log logr.Logger,
	l RegisteredLimit,
) (string, error) {
	var limitID string

	allPages, err := registeredlimits.List(o.osclient, registeredlimits.ListOpts{ResourceName: l.ResourceName}).AllPages()
	if err != nil {
		return limitID, err
	}

	allLimits, err := registeredlimits.ExtractRegisteredLimits(allPages)
	if err != nil {
		return limitID, err
	}

	if len(allLimits) == 1 {
		limitID = allLimits[0].ID
	} else if len(allLimits) == 0 {
		createOpts := registeredlimits.BatchCreateOpts{
			registeredlimits.CreateOpts{
				ResourceName: l.ResourceName,
				Description:  l.Description,
				DefaultLimit: l.DefaultLimit,
				ServiceID:    l.ServiceID,
				RegionID:     l.RegionID,
			},
		}
		log.Info(fmt.Sprintf("Creating registered limit %s", l.ResourceName))
		createdLimits, err := registeredlimits.BatchCreate(o.osclient, createOpts).Extract()
		if err != nil {
			return limitID, err
		}
		limitID = createdLimits[0].ID
	} else {
		return limitID, fmt.Errorf(fmt.Sprintf("multiple limits named \"%s\" found", l.ResourceName))
	}

	return limitID, nil
}
