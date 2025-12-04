package api

import "errors"

var (
	// ErrResourceNotFound indicates that a resource is not found.
	ErrResourceNotFound           = errors.New("resource not found")
	ErrGPUNodeNotFound            = errors.New("no available gpu node found")
	ErrStartUpFailed              = errors.New("app started up failed")
	ErrLaunchFailed               = errors.New("app launched failed")
	ErrNotSupportOperation        = errors.New("not support operation")
	ErrApplicationManagerNotFound = errors.New("application-manager not found")
)
