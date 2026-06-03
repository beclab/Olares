package validation

import (
	"context"
)

// Run executes the supplied validators in order, returning the first
// non-OK Decision (or an OK Decision after all of them pass). A
// non-nil error means the chain couldn't be evaluated (e.g. an API
// call inside a validator failed); callers should treat that the same
// as an "unknown" outcome and surface it to the user / handler.
//
// Validators whose AppliesTo returns false for in.Op are skipped, so
// callers can pass the same canonical chain (DefaultValidators) for
// every op and let the chain self-select.
func Run(ctx context.Context, in Input, validators ...Validator) (Decision, error) {
	if len(validators) == 0 {
		validators = DefaultValidators()
	}
	for _, v := range validators {
		if !v.AppliesTo(in.Op) {
			continue
		}
		d, err := v.Validate(ctx, in)
		if err != nil {
			d.Validator = v.Name()
			return d, err
		}
		if !d.OK {
			d.Validator = v.Name()
			return d, nil
		}
	}
	return Decision{OK: true}, nil
}
