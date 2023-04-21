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

// Common Condition Types used by API objects.
const (
	// ReadyCondition defines the Ready condition type that summarizes the operational state of an API object.
	ReadyCondition Type = "Ready"

	// InputReadyCondition Status=True condition which indicates if all required input sources are available, like e.g. secret holding passwords, other config maps providing input for the service.
	InputReadyCondition Type = "InputReady"

	// ServiceConfigReadyCondition Status=True Condition which indicates that all service config got rendered ok from the templates and stored in the ConfigMap
	ServiceConfigReadyCondition Type = "ServiceConfigReady"

	// DBReadyCondition Status=True condition is mirrored from the Ready condition in the mariadbdatabase ref object to the service API.
	DBReadyCondition Type = "DBReady"

	// DBSyncReadyCondition Status=True condition when dbsync job completed ok
	DBSyncReadyCondition Type = "DBSyncReady"

	// ExposeServiceReadyCondition Status=True condition when service/routes to expose the service created ok
	ExposeServiceReadyCondition Type = "ExposeServiceReady"

	// BootstrapReadyCondition Status=True condition when bootstrap job completed ok
	BootstrapReadyCondition Type = "BootstrapReady"

	// DeploymentReadyCondition Status=True condition when service deployment/statefulset created ok
	DeploymentReadyCondition Type = "DeploymentReady"

	// KeystoneServiceReadyCondition This condition is mirrored from the Ready condition in the keystoneservice ref object to the service API.
	KeystoneServiceReadyCondition Type = "KeystoneServiceReady"

	// KeystoneEndpointReadyCondition This condition is mirrored from the Ready condition in the keystoneendpoint ref object to the service API.
	KeystoneEndpointReadyCondition Type = "KeystoneEndpointReady"

	// NetworkAttachmentsReadyCondition Status=True condition when all pods k8s.v1.cni.cncf.io/network-status shows configured interfaces for all the NetworkAttachments with IP address
	NetworkAttachmentsReadyCondition Type = "NetworkAttachmentsReady"

	// CronJobReadyCondition Status=True condition when cron jobs created ok.
	CronJobReadyCondition Type = "CronJobReady"

	// AnsibleEECondition Status=True condition when the AnsibleEE run has been created ok
	AnsibleEECondition Type = "AnsibleEEReady"

	// ServiceAccountReadyCondition Status=True condition
	ServiceAccountReadyCondition Type = "ServiceAccountReady"

	// RoleReadyCondition Status=True condition
	RoleReadyCondition Type = "RoleReady"

	// RoleBindingReadyCondition Status=True condition
	RoleBindingReadyCondition Type = "RoleBindingReady"
)

// Common Reasons used by API objects.
const (
	// RequestedReason (Severity=Info) documents a condition not in Status=True because the underlying object is not ready.
	RequestedReason = "Requested"

	// CreationFailedReason (Severity=Error) documents a condition not in Status=True because the underlying object failed.
	CreationFailedReason = "CreationFailed"

	// ReadyReason documents a condition in `Status=True` when requested resource is ready.
	ReadyReason = "Ready"

	// InitReason documents a condition in `Status=Unknown` when reconcilation started.
	InitReason = "Init"

	// ErrorReason (Severity=Warning) documents a condition not in Status=True because the underlying object failed.
	// This is a warning because the reconciler will retry deletion.
	ErrorReason = "Error"

	// DeletingReason (Severity=Info) documents a condition not in Status=True because the underlying object it is currently being deleted.
	DeletingReason = "Deleting"

	// DeletionFailedReason (Severity=Warning) documents a condition not in Status=True because the underlying object
	// encountered problems during deletion. This is a warning because the reconciler will retry deletion.
	DeletionFailedReason = "DeletionFailed"

	// DeletedReason (Severity=Info) documents a condition not in Status=True because the underlying object was deleted.
	DeletedReason = "Deleted"
)

// Common Messages used by API objects.
const (
	//
	// Overall Ready Condition messages
	//
	// ReadyInitMessage
	ReadyInitMessage = "Setup started"

	// ReadyMessage
	ReadyMessage = "Setup complete"

	//
	// InputReady condition messages
	//
	// InputReadyInitMessage
	InputReadyInitMessage = "Input data not checked"

	// InputReadyMessage
	InputReadyMessage = "Input data complete"

	// InputReadyWaiting
	InputReadyWaitingMessage = "Input data resources missing"

	// InputReadyErrorMessage
	InputReadyErrorMessage = "Input data error occurred %s"

	//
	// ServiceConfig condition messages
	//
	// ServiceConfigReadyInitMessage
	ServiceConfigReadyInitMessage = "Service config create not started"

	// ServiceConfigReadyMessage
	ServiceConfigReadyMessage = "Service config create completed"

	// ServiceConfigReadyErrorMessage
	ServiceConfigReadyErrorMessage = "Service config create error occurred %s"

	//
	// DBReady condition messages
	//
	// DBReadyInitMessage
	DBReadyInitMessage = "DB create not started"

	// DBReadyMessage
	DBReadyMessage = "DB create completed"

	// DBReadyRunningMessage
	DBReadyRunningMessage = "DB create job still running"

	// DBReadyErrorMessage
	DBReadyErrorMessage = "DB create job error occurred %s"

	//
	// DBSync condition messages
	//
	// DBSyncReadyInitMessage
	DBSyncReadyInitMessage = "DBsync not started"

	// DBSyncReadyMessage
	DBSyncReadyMessage = "DBsync completed"

	// DBSyncReadyRunning
	DBSyncReadyRunningMessage = "DBsync job still running"

	// DBSyncReadyErrorMessage
	DBSyncReadyErrorMessage = "DBsync job error occurred %s"

	//
	// ExposeService condition messages
	//
	// ExposeServiceReadyInitMessage
	ExposeServiceReadyInitMessage = "Exposing service not started"

	// ExposeServiceReadyMessage
	ExposeServiceReadyMessage = "Exposing service completed"

	// ExposeServiceReadyRunningMessage
	ExposeServiceReadyRunningMessage = "Exposing service in progress"

	// ExposeServiceReadyErrorMessage
	ExposeServiceReadyErrorMessage = "Exposing service error occurred %s"

	//
	// BootstrapReady condition messages
	//
	// BootstrapReadyInitMessage
	BootstrapReadyInitMessage = "Bootstrap not started"

	// BootstrapReadyMessage
	BootstrapReadyMessage = "Bootstrap completed"

	// BootstrapReadyRunningMessage
	BootstrapReadyRunningMessage = "Bootstrap in progress"

	// BootstrapReadyErrorMessage
	BootstrapReadyErrorMessage = "Bootstrap error occurred %s"

	//
	// DeploymentReady condition messages
	//
	// DeploymentReadyInitMessage
	DeploymentReadyInitMessage = "Deployment not started"

	// DeploymentReadyMessage
	DeploymentReadyMessage = "Deployment completed"

	// DeploymentReadyRunningMessage
	DeploymentReadyRunningMessage = "Deployment in progress"

	// DeploymentReadyErrorMessage
	DeploymentReadyErrorMessage = "Deployment error occurred %s"

	//
	// NetworkAttachmentsReady condition messages
	//
	// NetworkAttachmentsReadyInitMessage
	NetworkAttachmentsReadyInitMessage = "NetworkAttachments not started"

	// NetworkAttachmentsReadyMessage
	NetworkAttachmentsReadyMessage = "NetworkAttachments completed"

	// NetworkAttachmentsReadyWaitingMessage
	NetworkAttachmentsReadyWaitingMessage = "NetworkAttachment resources missing: %s"

	// NetworkAttachmentsReadyErrorMessage
	NetworkAttachmentsReadyErrorMessage = "NetworkAttachments error occurred %s"

	//
	// CronJobReady condition messages
	//
	// CronJobReadyInitMessage
	CronJobReadyInitMessage = "CronJob not started"

	// CronJobReadyMessage
	CronJobReadyMessage = "CronJob completed"

	// CronJobReadyRunningMessage
	CronJobReadyRunningMessage = "CronJob in progress"

	// CronJobReadyErrorMessage
	CronJobReadyErrorMessage = "CronJob error occurred %s"

	//
	// AnsibleEEReady condition messages
	//
	// AnsibleEEReadyInitMessage
	AnsibleEEReadyInitMessage = "AnsibleEE not started"

	// AnsibleEEReadyMessage
	AnsibleEEReadyMessage = "AnsibleEE completed"

	// AnsibleEEReadyRunningMessage
	AnsibleEEReadyRunningMessage = "AnsibleEE in progress"

	// AnsibleEEReadyErrorMessage
	AnsibleEEReadyErrorMessage = "AnsibleEE error occurred %s"
)

// Common Messages used for service accounts, roles, role bindings
const (
	// ServiceAccountReadyErrorMessage
	ServiceAccountReadyErrorMessage = "ServiceAccount error occurred %s"

	// ServiceAccountCreatingMessage
	ServiceAccountCreatingMessage = "ServiceAccount creation in progress"

	// ServiceAccountReadyInitMessage
	ServiceAccountReadyInitMessage = "ServiceAccount not created"

	// ServiceAccountReadyMessage
	ServiceAccountReadyMessage = "ServiceAccount created"

	// RoleReadyErrorMessage
	RoleReadyErrorMessage = "Role error occurred %s"

	// RoleCreatingMessage
	RoleCreatingMessage = "Role creation in progress"

	// RoleReadyInitMessage
	RoleReadyInitMessage = "Role not created"

	// RoleReadyMessage
	RoleReadyMessage = "Role created"

	// RoleBindingReadyErrorMessage
	RoleBindingReadyErrorMessage = "RoleBinding error occurred %s"

	// RoleBindingCreatingMessage
	RoleBindingCreatingMessage = "RoleBinding creation in progress"

	// RoleBindingReadyInitMessage
	RoleBindingReadyInitMessage = "RoleBinding not created"

	// RoleBindingReadyMessage
	RoleBindingReadyMessage = "RoleBinding created"
)
