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

// Package secret provides utilities for managing Kubernetes Secret resources and conditions
package secret

import condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"

// Conditions for status in web console
const (
	//
	// condition reasons
	//

	// ReasonSecretMissing - secret does not exist
	ReasonSecretMissing condition.Reason = "SecretMissing"
	// ReasonSecretError - secret error
	ReasonSecretError condition.Reason = "SecretError"
	// ReasonSecretDeleteError - secret deletion error
	ReasonSecretDeleteError condition.Reason = "SecretDeleteError"
)
