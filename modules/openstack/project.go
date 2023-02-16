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

	"github.com/go-logr/logr"
	projects "github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
)

// Project -
type Project struct {
	Name        string
	Description string
}

// CreateProject - creates project with projectName and projectDescription if it does not exist
func (o *OpenStack) CreateProject(
	log logr.Logger,
	p Project,
) (string, error) {
	var projectID string
	allPages, err := projects.List(o.osclient, projects.ListOpts{Name: p.Name}).AllPages()
	if err != nil {
		return projectID, err
	}
	allProjects, err := projects.ExtractProjects(allPages)
	if err != nil {
		return projectID, err
	}
	if len(allProjects) == 1 {
		projectID = allProjects[0].ID
	} else if len(allProjects) == 0 {
		createOpts := projects.CreateOpts{
			Name:        p.Name,
			Description: p.Description,
		}
		log.Info(fmt.Sprintf("Creating project %s", p.Name))
		project, err := projects.Create(o.osclient, createOpts).Extract()
		if err != nil {
			return projectID, err
		}
		projectID = project.ID
	} else {
		return projectID, fmt.Errorf(fmt.Sprintf("multiple projects named \"%s\" found", p.Name))
	}

	return projectID, nil
}
