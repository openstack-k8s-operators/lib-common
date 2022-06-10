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

package condition

// Conditions for status in web console
const (
	//
	// condition types
	//

	// TypeEmpty - special state for 0 requested resources and 0 already provisioned
	TypeEmpty Type = "Empty"
	// TypeWaiting - something is causing the CR to wait
	TypeWaiting Type = "Waiting"
	// TypeProvisioning - one or more resoources are provisioning
	TypeProvisioning Type = "Provisioning"
	// TypeProvisioned - the requested resource count has been satisfied
	TypeProvisioned Type = "Provisioned"
	// TypeDeprovisioning - one or more resources are deprovisioning
	TypeDeprovisioning Type = "Deprovisioning"
	// TypeError - general catch-all for actual errors
	TypeError Type = "Error"
	// TypeCreated - general resource created
	TypeCreated Type = "Created"

	//
	// condition reasones
	//

	// ReasonInit - new resource set to reason Init
	ReasonInit Reason = "CommonInit"
	// ReasonSecretMissing - secret does not exist
	ReasonSecretMissing Reason = "SecretMissing"
	// ReasonSecretError - secret error
	ReasonSecretError Reason = "SecretError"
	// ReasonSecretDeleteError - secret deletion error
	ReasonSecretDeleteError Reason = "SecretDeleteError"
	// ReasonConfigMapMissing - config map does not exist
	ReasonConfigMapMissing Reason = "ConfigMapMissing"
	// ReasonConfigMapError - config map error
	ReasonConfigMapError Reason = "ConfigMapError"
	// ReasonCRStatusUpdateError - error updating CR status
	ReasonCRStatusUpdateError Reason = "CRStatusUpdateError"
	// ReasonControllerReferenceError - error set controller reference on object
	ReasonControllerReferenceError Reason = "ControllerReferenceError"
	// ReasonOwnerRefLabeledObjectsDeleteError - error deleting object using OwnerRef label
	ReasonOwnerRefLabeledObjectsDeleteError Reason = "OwnerRefLabeledObjectsDeleteError"
	// ReasonRemoveFinalizerError - error removing finalizer from object
	ReasonRemoveFinalizerError Reason = "RemoveFinalizerError"
	// ReasonAddRefLabelError - error adding reference label
	ReasonAddRefLabelError Reason = "AddRefLabelError"
	// ReasonServiceNotFound - service not found
	ReasonServiceNotFound Reason = "ServiceNotFound"
)
