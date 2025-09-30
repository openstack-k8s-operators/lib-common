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
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	services "github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v2/services"
)

// VolumeServiceCheck - Check particular cinder service is enabled and running or not
func (o *OpenStack) VolumeServiceCheck(
	ctx context.Context,
	log logr.Logger,
	serviceName string,
) (bool, error) {
	var serviceRunning bool
	serviceRunning = false
	log.Info(fmt.Sprintf("Checking %s service is running or not", serviceName))
	allPages, err := services.List(o.osclient, services.ListOpts{}).AllPages(ctx)
	if err != nil {
		return serviceRunning, err
	}
	allServices, err := services.ExtractServices(allPages)
	if err != nil {
		return serviceRunning, err
	}
	for _, service := range allServices {
		if strings.Contains(service.Binary, serviceName) && service.State == "up" && service.Status == "enabled" {
			serviceRunning = true
			break
		}
	}
	return serviceRunning, nil
}
