// Package wait is a stub of k8s.io/apimachinery/pkg/util/wait used by
// the sibling "poll" testdata package. Only the surface the analyzer
// looks at (function names) needs to be present.
package wait

import "time"

type ConditionFunc func() (done bool, err error)

func Poll(interval, timeout time.Duration, condition ConditionFunc) error {
	return nil
}

func PollImmediate(interval, timeout time.Duration, condition ConditionFunc) error {
	return nil
}

func PollUntil(interval time.Duration, condition ConditionFunc, stop <-chan struct{}) error {
	return nil
}

func PollImmediateUntil(interval time.Duration, condition ConditionFunc, stop <-chan struct{}) error {
	return nil
}
