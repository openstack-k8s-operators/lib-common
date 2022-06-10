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

// +kubebuilder:object:generate:=true

package database

import condition "github.com/openstack-k8s-operators/lib-common/pkg/condition"

// Conditions for status in web console
const (

	//
	// condition reasones
	//

	// ReasonDBError - DB error
	ReasonDBError condition.Reason = "DatabaseError"
	// ReasonDBPatchError - new resource set to reason Init
	ReasonDBPatchError condition.Reason = "DatabasePatchError"
	// ReasonDBPathOK - DB object created or patched ok
	ReasonDBPatchOK condition.Reason = "DatabasePatchOK"
	// ReasonDBNotFound - DB object not found
	ReasonDBNotFound condition.Reason = "DatabaseNotFound"
	// ReasonDBWaitingInitialized - waiting for service DB to be initialized
	ReasonDBWaitingInitialized condition.Reason = "DatabaseWaitingInitialized"
	// ReasonDBServiceNameError - error getting the DB service hostname
	ReasonDBServiceNameError condition.Reason = "DatabaseServiceNameError"
)
