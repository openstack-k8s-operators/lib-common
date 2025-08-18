package util

import (
	"errors"
)

var (
	// ErrResourceIsNotReady indicates that the resource is not ready
	ErrResourceIsNotReady = errors.New("resource is not ready")
	// ErrInvalidPort indicates that the port is invalid
	ErrInvalidPort = errors.New("invalid port")
	// ErrNotFound indicates that the object was not found
	ErrNotFound = errors.New("object not found")
	// ErrInvalidStatus indicates that the status is invalid
	ErrInvalidStatus = errors.New("invalid status")
	// ErrInvalidEndpoint indicates that the endpoint type is invalid
	ErrInvalidEndpoint = errors.New("invalid endpoint type")
	// ErrCannotUpdateObject indicates that the object cannot be updated
	ErrCannotUpdateObject = errors.New("cannot update object")
	// ErrFieldNotFound indicates that the field was not found in the Secret
	ErrFieldNotFound = errors.New("field not found in Secret")
	// ErrMoreThanOne indicates that only one should exist
	ErrMoreThanOne = errors.New("only one should exist")
	// ErrNoPodSubdomain indicates that there is no Subdomain or Hostname
	ErrNoPodSubdomain = errors.New("no subdomain or hostname")
	// ErrPodsInterfaces indicates that pod interfaces aren't configured
	ErrPodsInterfaces = errors.New("not all pods have interfaces")
)
