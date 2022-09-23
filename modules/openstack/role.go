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
	roles "github.com/gophercloud/gophercloud/openstack/identity/v3/roles"
)

// RoleNotFound - role not found error message"
const RoleNotFound = "role not found in keystone"

// Role -
type Role struct {
	Name string
}

//
// CreateRole - creates role with projectuserName, password and default project projectID
//
func (o *OpenStack) CreateRole(
	log logr.Logger,
	roleName string,
) (string, error) {
	var roleID string

	role, err := o.GetRole(
		log,
		roleName,
	)
	if err != nil && !strings.Contains(err.Error(), RoleNotFound) {
		return roleID, err
	}

	// if there is already a role, use it
	if role != nil {
		roleID = role.ID
	} else {
		createOpts := roles.CreateOpts{
			Name: roleName,
		}
		role, err := roles.Create(o.osclient, createOpts).Extract()
		if err != nil {
			return roleID, err
		}
		log.Info(fmt.Sprintf("Role Created - Rolename %s, ID %s", role.Name, role.ID))
		roleID = role.ID
	}

	return roleID, nil
}

//
// GetRole - gets role with roleName
//
func (o *OpenStack) GetRole(
	log logr.Logger,
	roleName string,
) (*roles.Role, error) {
	allPages, err := roles.List(o.osclient, roles.ListOpts{Name: roleName}).AllPages()
	if err != nil {
		return nil, err
	}
	allRoles, err := roles.ExtractRoles(allPages)
	if err != nil {
		return nil, err
	}

	if len(allRoles) == 0 {
		return nil, fmt.Errorf(fmt.Sprintf("%s %s", roleName, RoleNotFound))
	}

	return &allRoles[0], nil
}

//
// AssignUserRole - adds user with userID,projectID to role with roleName
//
func (o *OpenStack) AssignUserRole(
	log logr.Logger,
	roleName string,
	userID string,
	projectID string,
) error {
	role, err := o.GetRole(log, roleName)
	if err != nil {
		return err
	}

	// validate if user is already assigned to role
	listAssignmentsOpts := roles.ListAssignmentsOpts{
		ScopeProjectID: projectID,
		UserID:         userID,
		RoleID:         role.ID,
	}
	allPages, err := roles.ListAssignments(o.osclient, listAssignmentsOpts).AllPages()
	if err != nil {
		return err
	}

	assignUser, err := allPages.IsEmpty()
	if err != nil {
		return err
	}

	if assignUser {
		log.Info(fmt.Sprintf("Assigning userID %s to role %s - %s", userID, role.Name, role.ID))

		err = roles.Assign(o.osclient, role.ID, roles.AssignOpts{
			UserID:    userID,
			ProjectID: projectID}).ExtractErr()
		if err != nil {
			return err
		}
	}

	return nil
}
