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
	users "github.com/gophercloud/gophercloud/openstack/identity/v3/users"
)

// User -
type User struct {
	Name      string
	Password  string
	ProjectID string
}

//
// CreateUser - creates user with userName, password and default project projectID
//
func (o *OpenStack) CreateUser(
	log logr.Logger,
	u User,
) (string, error) {
	var userID string

	user, err := o.GetUser(
		log,
		u.Name,
	)
	if err != nil {
		return userID, err
	}

	// if there is already a user registered use it
	if user != nil {
		// TODO support PWD change
		userID = user.ID
	} else {
		createOpts := users.CreateOpts{
			Name:             u.Name,
			DefaultProjectID: u.ProjectID,
			Password:         u.Password,
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

//
// GetUser - get user with userName
//
func (o *OpenStack) GetUser(
	log logr.Logger,
	userName string,
) (*users.User, error) {
	allPages, err := users.List(o.GetOSClient(), users.ListOpts{Name: userName}).AllPages()
	if err != nil {
		return nil, err
	}
	allUsers, err := users.ExtractUsers(allPages)
	if err != nil {
		return nil, err
	}

	if len(allUsers) == 0 {
		log.Info(fmt.Sprintf("%s user not found in keystone", userName))

		return nil, nil
	}

	return &allUsers[0], nil
}

//
// DeleteUser - deletes user with userName
//
func (o *OpenStack) DeleteUser(
	log logr.Logger,
	userName string,
) error {
	user, err := o.GetUser(
		log,
		userName,
	)
	if err != nil {
		return err
	}

	if user != nil {
		log.Info(fmt.Sprintf("Deleting user %s", user.Name))
		err = users.Delete(o.GetOSClient(), user.ID).ExtractErr()
		if err != nil {
			return err
		}
	}

	log.Info("Deleting user successfully")
	return nil
}
