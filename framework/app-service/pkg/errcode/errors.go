package errcode

import "errors"

var (
	ErrPodPending = errors.New("pod is pending")
)
