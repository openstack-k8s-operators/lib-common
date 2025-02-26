package util

import (
	"errors"
)

var (
	// ErrResourceIsNotReady indicates that the resource is not ready
	ErrResourceIsNotReady = errors.New("Resource is not ready")
	// ErrInvalidPort indicates that the port is invalid
	ErrInvalidPort = errors.New("invalid port")
	// ErrNotFound indicates that the object was not found
	ErrNotFound = errors.New("Object not found")
	// ErrInvalidStatus indicates that the status is invalid
	ErrInvalidStatus = errors.New("invalid status")
	// ErrInvalidEndpoint indicates that the endpoint type is invalid
	ErrInvalidEndpoint = errors.New("invalid endpoint type")
	// ErrCannotUpdateObject indicates that the object cannot be updated
	ErrCannotUpdateObject = errors.New("cannot update object")
	// ErrFieldNotFound indicates that the field was not found in the Secret
	ErrFieldNotFound = errors.New("field not found in Secret")
	// ErrMoreThanOne indicates that only one should exist
	ErrMoreThanOne = errors.New("Only one should exist")
	// ErrNoPodSubdomain indicates that there is no Subdomain or Hostname
	ErrNoPodSubdomain = errors.New("No Subdomain or Hostname")
	// ErrCreateFailed indicates that the resource failed to create
	ErrCreateFailed = errors.New("failed to create")
	// ErrFetchFailed indicates that the resource failed to fetch
	ErrFetchFailed = errors.New("failed to fetch")
	// ErrDeleteFailed indicates that the resource failed to delete
	ErrDeleteFailed = errors.New("failed to delete")
	// ErrNoPodSubdomain indicates that ensure failed
	ErrEnsureFailed = errors.New("failed to ensure")
	// ErrNoPodSubdomain indicates that the resource failed to reconcile
	ErrReconcileFailed = errors.New("failed to reconcile")
	// ErrNoPodSubdomain indicates that the key can't be found
	ErrKeyNotFound = errors.New("key not found in")
	// ErrNoPodSubdomain indicates that pod interfaces aren't configured
	ErrPodsInterfaces = errors.New("not all pods have interfaces")
)
