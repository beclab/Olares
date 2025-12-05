package utils

import "errors"

// AggregateErrs aggregates a slice of errors into a single error.
func AggregateErrs(errs []error) error {
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		var errStr string
		for _, e := range errs {
			errStr += e.Error() + "\t"
		}
		return errors.New(errStr[:len(errStr)-1])
	}
}
