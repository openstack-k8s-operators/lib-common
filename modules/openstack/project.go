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
	DomainID    string
}

// ProjectNotFound - project not found error message"
const ProjectNotFound = "project not found"

// CreateProject - creates project with projectName and projectDescription if it does not exist
func (o *OpenStack) CreateProject(
	log logr.Logger,
	p Project,
) (string, error) {
	var projectID string
	allPages, err := projects.List(o.osclient, projects.ListOpts{Name: p.Name, DomainID: p.DomainID}).AllPages()
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
			DomainID:    p.DomainID,
		}
		log.Info(fmt.Sprintf("Creating project %s in %s", p.Name, p.DomainID))
		project, err := projects.Create(o.osclient, createOpts).Extract()
		if err != nil {
			return projectID, err
		}
		projectID = project.ID
	} else {
		return projectID, fmt.Errorf("multiple projects named \"%s\" found", p.Name)
	}

	return projectID, nil
}

// GetProject - gets project with projectName
func (o *OpenStack) GetProject(
	log logr.Logger,
	projectName string,
	domainID string,
) (*projects.Project, error) {
	allPages, err := projects.List(o.GetOSClient(), projects.ListOpts{Name: projectName, DomainID: domainID}).AllPages()
	if err != nil {
		return nil, err
	}
	allProjects, err := projects.ExtractProjects(allPages)
	if err != nil {
		return nil, err
	}

	if len(allProjects) == 0 {
		return nil, fmt.Errorf("%s %s", projectName, ProjectNotFound)
	} else if len(allProjects) > 1 {
		return nil, fmt.Errorf("multiple project named \"%s\" found", projectName)
	}

	return &allProjects[0], nil
}
