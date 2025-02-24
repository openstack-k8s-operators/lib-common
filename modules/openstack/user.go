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
	users "github.com/gophercloud/gophercloud/openstack/identity/v3/users"
)

// UserNotFound - user not found error message"
const UserNotFound = "user not found in keystone"

// User -
type User struct {
	Name      string
	Password  string
	ProjectID string
	DomainID  string
}

// CreateUser - creates user with userName, password and default project projectID
func (o *OpenStack) CreateUser(
	log logr.Logger,
	u User,
) (string, error) {
	var userID string

	user, err := o.GetUser(
		log,
		u.Name,
		u.DomainID,
	)
	// If the user is not found, don't count that as an error here
	if err != nil && !strings.Contains(err.Error(), UserNotFound) {
		return userID, err
	}

	// if there is already a user registered use it
	if user != nil {
		// TODO support PWD change
		userID = user.ID
	} else {
		createOpts := users.CreateOpts{
			Name:     u.Name,
			Password: u.Password,
			DomainID: u.DomainID,
		}
		if u.ProjectID != "" {
			createOpts.DefaultProjectID = u.ProjectID
		}

		user, err := users.Create(o.GetOSClient(), createOpts).Extract()
		if err != nil {
			return userID, err
		}
		log.Info(fmt.Sprintf("User Created - Username %s, ID %s", user.Name, user.ID))
		userID = user.ID

	}

	return userID, nil
}

// GetUser - get user with userName
func (o *OpenStack) GetUser(
	log logr.Logger,
	userName string,
	domainID string,
) (*users.User, error) {
	allPages, err := users.List(o.GetOSClient(), users.ListOpts{Name: userName, DomainID: domainID}).AllPages()
	if err != nil {
		return nil, err
	}
	allUsers, err := users.ExtractUsers(allPages)
	if err != nil {
		return nil, err
	}

	if len(allUsers) == 0 {
		return nil, fmt.Errorf("%s %s", userName, UserNotFound) // nolint:err113
	} else if len(allUsers) > 1 {
		return nil, fmt.Errorf("multiple users named \"%s\" found", userName) // nolint:err113
	}

	return &allUsers[0], nil
}

// DeleteUser - deletes user with userName
func (o *OpenStack) DeleteUser(
	log logr.Logger,
	userName string,
	domainID string,
) error {
	user, err := o.GetUser(
		log,
		userName,
		domainID,
	)
	// If the user is not found, don't count that as an error here
	if err != nil && !strings.Contains(err.Error(), "user not found in keystone") {
		return err
	}

	if user != nil {
		log.Info(fmt.Sprintf("Deleting user %s in %s", user.Name, user.DomainID))
		err = users.Delete(o.GetOSClient(), user.ID).ExtractErr()
		if err != nil {
			return err
		}
	}

	log.Info("Deleting user successfully")
	return nil
}
