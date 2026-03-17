package errcode

import "errors"

var (
	ErrServerSidePodPending = errors.New("server side pod is pending")
	ErrPodPending           = errors.New("pod is pending")
	ErrHamiUnschedulable    = errors.New("hami-unschedulable")
)
