package api

import "errors"

var (
	// ErrResourceNotFound indicates that a resource is not found.
	ErrResourceNotFound = errors.New("resource not found")
)
